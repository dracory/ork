package security

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/skills/fs"
	"github.com/dracory/ork/skills/systemctl"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// phpFpmHardenMarkerBegin and phpFpmHardenMarkerEnd delimit the managed block
// appended to the FPM pool config (www.conf). They are used by the blockinfile
// replacement so the block is re-rendered on every run instead of appended
// once (which would silently ignore later changes to the desired directives).
const (
	phpFpmHardenMarkerBegin = "; BEGIN ork-php-fpm-harden"
	phpFpmHardenMarkerEnd   = "; END ork-php-fpm-harden"
)

// PhpFpmHarden applies security hardening to PHP-FPM. It is FPM-only: the
// hardened settings are written to the FPM conf.d drop-in and the FPM pool,
// so the CLI php.ini is untouched and artisan/composer remain unrestricted.
//
// This is the ork equivalent of an Ansible role built from template +
// blockinfile + a validate-gated service handler:
//
//   - conf.d/99-hardening.ini: written via fs.FileCreate with overwrite=true
//     (the template equivalent — replace-on-change, idempotent).
//   - pool.d/www.conf: the pool-level php_admin_value/php_admin_flag directives
//     are managed as a block between markers using a delete-then-append
//     (blockinfile) replacement, so editing the desired directives updates
//     the file instead of leaving stale content behind. A separate pool.d
//     drop-in cannot be used here because FPM treats every pool.d/*.conf as a
//     complete pool definition; admin values must live inside the [www] pool.
//   - validation: `php-fpm<version> -t` is run before the restart. On failure
//     the conf.d drop-in is removed and the pool config is restored from a
//     timestamped backup, mirroring the ssh-harden visudo -cf rollback.
//   - restart: systemctl restart php<version>-fpm, only after validation passes.
//
// Usage:
//
//	node.Run(security.NewPhpFpmHarden().
//	    SetVersion("8.5").
//	    SetOpenBasedirPaths("/var/www/app:/var/www/media:/tmp"))
//	// non-Laravel app: also disable proc_open
//	node.Run(security.NewPhpFpmHarden().
//	    SetVersion("8.3").
//	    SetOpenBasedirPaths("/var/www/app:/tmp").
//	    SetDisableFunctions("exec, shell_exec, system, passthru, popen, proc_open"))
//
// Args:
//   - php-version (required): PHP version, e.g. "8.5"
//   - open-basedir-paths (required): colon-separated paths for open_basedir
//   - disable-functions: comma-separated disable_functions (default Laravel-safe)
//   - memory-limit, upload-max-filesize, post-max-size, max-execution-time,
//     max-input-time, opcache-memory-consumption, opcache-max-accelerated-files
//   - conf-d-path, pool-path, error-log: path overrides (default derived from version)
//
// Execution Flow:
//  1. Validates php-version and open-basedir-paths
//  2. Backs up the FPM pool config (www.conf) to a timestamped copy
//  3. Writes the conf.d drop-in (99-hardening.ini) with overwrite=true
//  4. Replaces the managed block in www.conf (delete markers, re-append)
//  5. Validates the full FPM config with `php-fpm<version> -t`
//  6. On validation failure: removes the drop-in, restores www.conf, returns error
//  7. On success: restarts php<version>-fpm so the settings take effect
//
// Prerequisites:
//   - PHP-FPM must be installed (see the php.Install skill)
//   - Root SSH access required to write under /etc/php and /var/log
//
// Related Skills:
//   - php.Install: Install PHP with extensions and configure the FPM pool
//   - security.SshHarden: SSH server hardening (same backup/validate/restore shape)
type PhpFpmHarden struct {
	*types.BaseSkill
}

// Compile-time assertion that PhpFpmHarden implements types.RunnableInterface.
var _ types.RunnableInterface = (*PhpFpmHarden)(nil)

// Check returns true when the desired state is not yet present. It is
// conservative: the conf.d drop-in content diff is handled by fs.FileCreate,
// and the pool block is always re-rendered (the blockinfile replacement is
// itself idempotent). Returns true in dry-run mode.
func (p *PhpFpmHarden) Check() (bool, error) {
	if err := p.validateArgs(); err != nil {
		return false, err
	}
	cfg := p.GetNodeConfig()
	if cfg.IsDryRunMode {
		return true, nil
	}
	// Re-apply to guarantee the conf.d drop-in and the pool block match the
	// configured args. fs.FileCreate skips the write when content/mode/owner
	// already match; the blockinfile replacement is a no-op when the block is
	// already current.
	return true, nil
}

// Run applies PHP-FPM hardening.
func (p *PhpFpmHarden) Run() types.Result {
	if err := p.validateArgs(); err != nil {
		return types.Result{Changed: false, Message: "Invalid PHP-FPM harden args", Error: err}
	}

	cfg := p.GetNodeConfig()
	version := p.GetArg(ArgPhpVersion)
	openBasedir := p.GetArg(ArgOpenBasedirPaths)
	disableFunctions := p.GetArg(ArgDisableFunctions)
	if disableFunctions == "" {
		disableFunctions = DefaultDisableFunctions
	}
	memoryLimit := p.GetArg(ArgMemoryLimit)
	if memoryLimit == "" {
		memoryLimit = DefaultMemoryLimit
	}
	uploadMaxFilesize := p.GetArg(ArgUploadMaxFilesize)
	if uploadMaxFilesize == "" {
		uploadMaxFilesize = DefaultUploadMaxFilesize
	}
	postMaxSize := p.GetArg(ArgPostMaxSize)
	if postMaxSize == "" {
		postMaxSize = DefaultPostMaxSize
	}
	maxExecutionTime := p.GetArg(ArgMaxExecutionTime)
	if maxExecutionTime == "" {
		maxExecutionTime = DefaultMaxExecutionTime
	}
	maxInputTime := p.GetArg(ArgMaxInputTime)
	if maxInputTime == "" {
		maxInputTime = DefaultMaxInputTime
	}
	opcacheMemory := p.GetArg(ArgOpcacheMemory)
	if opcacheMemory == "" {
		opcacheMemory = DefaultOpcacheMemory
	}
	opcacheMaxFiles := p.GetArg(ArgOpcacheMaxFiles)
	if opcacheMaxFiles == "" {
		opcacheMaxFiles = DefaultOpcacheMaxFiles
	}
	confDPath := p.GetArg(ArgConfDPath)
	if confDPath == "" {
		confDPath = fmt.Sprintf(DefaultConfDPathPattern, version)
	}
	poolPath := p.GetArg(ArgPoolPath)
	if poolPath == "" {
		poolPath = fmt.Sprintf(DefaultPoolPathPattern, version)
	}
	errorLog := p.GetArg(ArgErrorLog)
	if errorLog == "" {
		errorLog = fmt.Sprintf(DefaultErrorLogPattern, version)
	}

	phpFpmService := "php" + version + "-fpm"
	validateCmd := fmt.Sprintf("php-fpm%s -t", version)

	cfg.GetLoggerOrDefault().Info("PHP-FPM hardening started", "version", version)

	// Dry-run: report the commands that would run, change nothing.
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would back up pool config", "path", poolPath)
		cfg.GetLoggerOrDefault().Info("dry-run: would write conf.d drop-in", "path", confDPath)
		cfg.GetLoggerOrDefault().Info("dry-run: would replace managed block in pool config", "path", poolPath)
		cfg.GetLoggerOrDefault().Info("dry-run: would validate FPM config", "cmd", validateCmd)
		cfg.GetLoggerOrDefault().Info("dry-run: would restart service", "service", phpFpmService)
		return types.Result{
			Changed: true,
			Message: "Would harden PHP-FPM configuration",
		}
	}

	// Step 1: Back up the pool config so it can be restored if validation fails.
	cfg.GetLoggerOrDefault().Info("backing up FPM pool config", "path", poolPath)
	cmdBackup := types.Command{
		Command:     fmt.Sprintf("cp %s %s.backup.$(date +%s)", skills.ShellEscapeArg(poolPath), skills.ShellEscapeArg(poolPath), "%Y%m%d%H%M%S"),
		Description: "Backup FPM pool config",
	}
	if _, err := ssh.Run(cfg, cmdBackup); err != nil {
		return types.Result{Changed: false, Message: "Failed to back up FPM pool config", Error: err}
	}

	// Step 2: Write the conf.d drop-in (template equivalent: replace-on-change).
	confDContent := buildConfDContent(errorLog, openBasedir, disableFunctions, memoryLimit,
		uploadMaxFilesize, postMaxSize, maxExecutionTime, maxInputTime, opcacheMemory, opcacheMaxFiles)

	cfg.GetLoggerOrDefault().Info("writing conf.d drop-in", "path", confDPath)
	confDResult := runSub(fs.NewFileCreate().
		SetPath(confDPath).
		SetContent(confDContent).
		SetOwner("root:root").
		SetMode("644").
		SetOverwrite(true), cfg)
	if confDResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to write PHP-FPM hardening drop-in",
			Error:   confDResult.Error,
		}
	}

	// Step 3: Replace the managed block in the pool config (blockinfile).
	// sed deletes any existing block between the markers (no-op if absent),
	// then the new block is appended. This re-renders the block on every run
	// so changes to the desired directives are applied, unlike a one-shot
	// `grep -q || cat >>` which silently skips once the marker exists.
	poolBlock := buildPoolBlock(disableFunctions, openBasedir)
	escPoolPath := skills.ShellEscapeArg(poolPath)
	escPoolBlock := skills.ShellEscapeContent(poolBlock)
	cmdPoolHarden := types.Command{
		Command:     fmt.Sprintf("sed -i '/%s/,/%s/d' %s && printf '%%s\\n' %s >> %s", phpFpmHardenMarkerBegin, phpFpmHardenMarkerEnd, escPoolPath, escPoolBlock, escPoolPath),
		Description: "Replace managed hardening block in FPM pool config (blockinfile)",
		Required:    true,
	}
	cfg.GetLoggerOrDefault().Info("replacing managed block in pool config", "path", poolPath)
	if output, err := ssh.Run(cfg, cmdPoolHarden); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to apply pool-level PHP-FPM hardening",
			Error:   fmt.Errorf("%w: %s", err, output),
		}
	}

	// Step 4: Validate the full FPM config before restarting. A syntax error in
	// either the drop-in or the pool block would otherwise take the pool down.
	cfg.GetLoggerOrDefault().Info("validating FPM config", "cmd", validateCmd)
	cmdValidate := types.Command{
		Command:     validateCmd,
		Description: "Validate PHP-FPM configuration",
		Required:    true,
	}
	if output, err := ssh.Run(cfg, cmdValidate); err != nil {
		// Rollback: remove the drop-in and restore the pool config from the
		// latest backup, so FPM can still start with the previous configuration.
		cfg.GetLoggerOrDefault().Error("FPM config validation failed; rolling back", "error", err, "output", output)
		_, _ = ssh.Run(cfg, types.Command{
			Command:     fmt.Sprintf("rm -f %s", skills.ShellEscapeArg(confDPath)),
			Description: "Remove invalid conf.d drop-in (rollback)",
		})
		_, _ = ssh.Run(cfg, types.Command{
			Command:     fmt.Sprintf("latest=$(ls -t %s.backup.* 2>/dev/null | head -1); [ -n \"$latest\" ] && cp \"$latest\" %s", escPoolPath+".backup.*", escPoolPath),
			Description: "Restore FPM pool config from latest backup (rollback)",
		})
		return types.Result{
			Changed: false,
			Message: "PHP-FPM config validation failed; drop-in removed and pool config restored",
			Error:   fmt.Errorf("%w: %s", err, output),
		}
	}

	// Step 5: Restart PHP-FPM so the hardened settings take effect.
	cfg.GetLoggerOrDefault().Info("restarting PHP-FPM", "service", phpFpmService)
	restartResult := runSub(systemctl.NewRestart().SetService(phpFpmService), cfg)
	if restartResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to restart PHP-FPM after hardening",
			Error:   restartResult.Error,
		}
	}

	cfg.GetLoggerOrDefault().Info("PHP-FPM hardening complete", "version", version)
	return types.Result{
		Changed: true,
		Message: "PHP-FPM hardened successfully (FPM-only, CLI unrestricted)",
		Details: map[string]string{
			"version":            version,
			"conf-d-path":        confDPath,
			"pool-path":          poolPath,
			"open-basedir-paths": openBasedir,
			"disable-functions":  disableFunctions,
		},
	}
}

// phpVersionRe validates that the php-version arg is a simple major.minor
// version string (e.g. "8.5", "8.3"). This prevents shell injection via the
// version string, which is interpolated into paths, service names, and
// commands.
var phpVersionRe = regexp.MustCompile(`^\d+\.\d+$`)

// validateArgs checks the required args.
func (p *PhpFpmHarden) validateArgs() error {
	version := p.GetArg(ArgPhpVersion)
	if version == "" {
		return fmt.Errorf("no php-version specified: set the %q argument", ArgPhpVersion)
	}
	if !phpVersionRe.MatchString(version) {
		return fmt.Errorf("invalid php-version %q: expected format like '8.5' or '8.3'", version)
	}
	if p.GetArg(ArgOpenBasedirPaths) == "" {
		return fmt.Errorf("no open-basedir-paths specified: set the %q argument (restricting file access is the core of the hardening)", ArgOpenBasedirPaths)
	}
	return nil
}

// buildConfDContent renders the conf.d drop-in (regular php.ini directives).
// These apply FPM-wide; the CLI php.ini is not affected.
func buildConfDContent(errorLog, openBasedir, disableFunctions, memoryLimit,
	uploadMaxFilesize, postMaxSize, maxExecutionTime, maxInputTime, opcacheMemory, opcacheMaxFiles string) string {
	var b strings.Builder
	b.WriteString("; PHP-FPM hardening for production\n")
	b.WriteString("; Managed by ork security.PhpFpmHarden — do not edit by hand\n\n")

	b.WriteString("; Hide PHP version from responses\n")
	b.WriteString("expose_php = Off\n\n")

	b.WriteString("; Don't display errors to users (log them instead)\n")
	b.WriteString("display_errors = Off\n")
	b.WriteString("display_startup_errors = Off\n")
	b.WriteString("log_errors = On\n")
	b.WriteString(fmt.Sprintf("error_log = %s\n\n", errorLog))

	b.WriteString("; Restrict file access to the configured paths and /tmp\n")
	b.WriteString(fmt.Sprintf("open_basedir = %s\n\n", openBasedir))

	b.WriteString("; Disable dangerous functions in FPM only (NOT proc_open — Laravel needs it)\n")
	b.WriteString(fmt.Sprintf("disable_functions = %s\n\n", disableFunctions))

	b.WriteString("; Session hardening\n")
	b.WriteString("session.cookie_secure = On\n")
	b.WriteString("session.cookie_httponly = On\n")
	b.WriteString("session.cookie_samesite = Lax\n")
	b.WriteString("session.use_strict_mode = On\n\n")

	b.WriteString("; OPcache for production performance\n")
	b.WriteString("opcache.enable = On\n")
	b.WriteString("opcache.validate_timestamps = Off\n")
	b.WriteString(fmt.Sprintf("opcache.memory_consumption = %s\n", opcacheMemory))
	b.WriteString(fmt.Sprintf("opcache.max_accelerated_files = %s\n", opcacheMaxFiles))
	b.WriteString("opcache.revalidate_freq = 2\n")
	b.WriteString("opcache.save_comments = On\n\n")

	b.WriteString("; Upload limits\n")
	b.WriteString("file_uploads = On\n")
	b.WriteString(fmt.Sprintf("upload_max_filesize = %s\n", uploadMaxFilesize))
	b.WriteString(fmt.Sprintf("post_max_size = %s\n\n", postMaxSize))

	b.WriteString("; Execution limits\n")
	b.WriteString(fmt.Sprintf("max_execution_time = %s\n", maxExecutionTime))
	b.WriteString(fmt.Sprintf("max_input_time = %s\n", maxInputTime))
	b.WriteString(fmt.Sprintf("memory_limit = %s\n", memoryLimit))
	return b.String()
}

// buildPoolBlock renders the managed block appended to the FPM pool config.
// These php_admin_value/php_admin_flag directives lock the settings at the
// pool level so they cannot be overridden by per-pool or per-script config.
func buildPoolBlock(disableFunctions, openBasedir string) string {
	var b strings.Builder
	b.WriteString(phpFpmHardenMarkerBegin + "\n")
	b.WriteString(fmt.Sprintf("php_admin_value[disable_functions] = %s\n", disableFunctions))
	b.WriteString("php_admin_flag[expose_php] = Off\n")
	b.WriteString("php_admin_flag[display_errors] = Off\n")
	b.WriteString(fmt.Sprintf("php_admin_value[open_basedir] = %s\n", openBasedir))
	b.WriteString(phpFpmHardenMarkerEnd)
	return b.String()
}

// SetArgs sets the arguments for PHP-FPM hardening.
// Returns PhpFpmHarden for fluent method chaining.
func (p *PhpFpmHarden) SetArgs(args map[string]string) types.RunnableInterface {
	p.BaseSkill.SetArgs(args)
	return p
}

// SetArg sets a single argument for PHP-FPM hardening.
// Returns PhpFpmHarden for fluent method chaining.
func (p *PhpFpmHarden) SetArg(key, value string) types.RunnableInterface {
	p.BaseSkill.SetArg(key, value)
	return p
}

// SetVersion sets the PHP version (e.g. "8.5") and returns PhpFpmHarden for
// chaining. Required.
func (p *PhpFpmHarden) SetVersion(version string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgPhpVersion, version)
	return p
}

// SetOpenBasedirPaths sets the colon-separated open_basedir paths (e.g.
// "/var/www/app:/var/www/media:/tmp") and returns PhpFpmHarden for chaining.
// Required — restricting file access is the core of the hardening.
func (p *PhpFpmHarden) SetOpenBasedirPaths(paths string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgOpenBasedirPaths, paths)
	return p
}

// SetDisableFunctions sets the comma-separated disable_functions list and
// returns PhpFpmHarden for chaining. Defaults to DefaultDisableFunctions
// (Laravel-safe). Set to a stricter list (including proc_open) for non-Laravel
// apps.
func (p *PhpFpmHarden) SetDisableFunctions(functions string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgDisableFunctions, functions)
	return p
}

// SetMemoryLimit sets the FPM memory_limit (e.g. "256M") and returns
// PhpFpmHarden for chaining.
func (p *PhpFpmHarden) SetMemoryLimit(limit string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgMemoryLimit, limit)
	return p
}

// SetUploadMaxFilesize sets the FPM upload_max_filesize (e.g. "10M") and
// returns PhpFpmHarden for chaining.
func (p *PhpFpmHarden) SetUploadMaxFilesize(size string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgUploadMaxFilesize, size)
	return p
}

// SetPostMaxSize sets the FPM post_max_size (e.g. "12M") and returns
// PhpFpmHarden for chaining.
func (p *PhpFpmHarden) SetPostMaxSize(size string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgPostMaxSize, size)
	return p
}

// SetMaxExecutionTime sets the FPM max_execution_time (seconds) and returns
// PhpFpmHarden for chaining.
func (p *PhpFpmHarden) SetMaxExecutionTime(seconds string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgMaxExecutionTime, seconds)
	return p
}

// SetMaxInputTime sets the FPM max_input_time (seconds) and returns
// PhpFpmHarden for chaining.
func (p *PhpFpmHarden) SetMaxInputTime(seconds string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgMaxInputTime, seconds)
	return p
}

// SetOpcacheMemory sets opcache.memory_consumption and returns PhpFpmHarden
// for chaining.
func (p *PhpFpmHarden) SetOpcacheMemory(mb string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgOpcacheMemory, mb)
	return p
}

// SetOpcacheMaxFiles sets opcache.max_accelerated_files and returns
// PhpFpmHarden for chaining.
func (p *PhpFpmHarden) SetOpcacheMaxFiles(count string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgOpcacheMaxFiles, count)
	return p
}

// SetConfDPath overrides the conf.d drop-in path and returns PhpFpmHarden for
// chaining. Defaults to /etc/php/<version>/fpm/conf.d/99-hardening.ini.
func (p *PhpFpmHarden) SetConfDPath(path string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgConfDPath, path)
	return p
}

// SetPoolPath overrides the FPM pool config path and returns PhpFpmHarden for
// chaining. Defaults to /etc/php/<version>/fpm/pool.d/www.conf.
func (p *PhpFpmHarden) SetPoolPath(path string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgPoolPath, path)
	return p
}

// SetErrorLog overrides the FPM error_log path and returns PhpFpmHarden for
// chaining. Defaults to /var/log/php<version>-fpm.log.
func (p *PhpFpmHarden) SetErrorLog(path string) *PhpFpmHarden {
	p.BaseSkill.SetArg(ArgErrorLog, path)
	return p
}

// SetID sets the ID for PHP-FPM hardening.
// Returns PhpFpmHarden for fluent method chaining.
func (p *PhpFpmHarden) SetID(id string) types.RunnableInterface {
	p.BaseSkill.SetID(id)
	return p
}

// SetDescription sets the description for PHP-FPM hardening.
// Returns PhpFpmHarden for fluent method chaining.
func (p *PhpFpmHarden) SetDescription(description string) types.RunnableInterface {
	p.BaseSkill.SetDescription(description)
	return p
}

// SetTimeout sets the timeout for PHP-FPM hardening.
// Returns PhpFpmHarden for fluent method chaining.
func (p *PhpFpmHarden) SetTimeout(timeout time.Duration) types.RunnableInterface {
	p.BaseSkill.SetTimeout(timeout)
	return p
}

// NewPhpFpmHarden creates a new php-fpm-harden skill.
//
// Returns a PhpFpmHarden skill configured with skills.IDPhpFpmHarden identifier
// and description "Harden PHP-FPM configuration for production security".
//
// Example:
//
//	node.Run(security.NewPhpFpmHarden().
//	    SetVersion("8.5").
//	    SetOpenBasedirPaths("/var/www/app:/var/www/media:/tmp"))
func NewPhpFpmHarden() *PhpFpmHarden {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDPhpFpmHarden)
	pb.SetDescription("Harden PHP-FPM configuration for production security")
	return &PhpFpmHarden{BaseSkill: pb}
}
