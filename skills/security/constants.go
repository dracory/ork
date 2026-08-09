// Package security provides playbooks for system security hardening and configuration.
// These playbooks help secure servers by applying industry-standard security settings
// to SSH, kernel parameters, and installing security monitoring tools.
package security

// Argument key constants for SSH hardening skill.
const (
	// ArgNonRootUser specifies the non-root user to verify before disabling root login
	ArgNonRootUser = "non-root-user"

	// ArgSSHConfigPath specifies the SSH configuration file path
	ArgSSHConfigPath = "ssh-config-path"

	// ArgMaxAuthTries specifies the maximum authentication attempts
	ArgMaxAuthTries = "max-auth-tries"

	// ArgClientAliveInterval specifies the client alive interval in seconds
	ArgClientAliveInterval = "client-alive-interval"

	// ArgClientAliveCountMax specifies the client alive count max
	ArgClientAliveCountMax = "client-alive-count-max"

	// ArgSysctlConfigPath specifies the sysctl configuration file path
	ArgSysctlConfigPath = "sysctl-config-path"

	// ArgSysctlDropInPath specifies the sysctl drop-in file path
	ArgSysctlDropInPath = "sysctl-dropin-path"

	// ArgUsername specifies the user account to generate the SSH keypair for
	ArgUsername = "username"

	// ArgKeyType specifies the SSH key type (ed25519, rsa, ecdsa)
	ArgKeyType = "key-type"

	// ArgComment specifies the comment embedded in the public key (-C)
	ArgComment = "comment"

	// ArgKeyPath specifies the private key file path
	ArgKeyPath = "key-path"

	// PHP-FPM hardening argument keys.
	//
	// ArgPhpVersion specifies the PHP version (e.g. "8.3", "8.5"). Required.
	ArgPhpVersion = "php-version"

	// ArgOpenBasedirPaths specifies the colon-separated open_basedir paths
	// (e.g. "/var/www/app:/var/www/media:/tmp"). Required — restricting file
	// access is the core of the hardening; an empty value would disable it.
	ArgOpenBasedirPaths = "open-basedir-paths"

	// ArgDisableFunctions specifies the comma-separated disable_functions list
	// applied to FPM. Defaults to DefaultDisableFunctions (Laravel-safe: it
	// omits proc_open, which Laravel needs). Override for non-Laravel apps.
	ArgDisableFunctions = "disable-functions"

	// ArgMemoryLimit specifies the FPM memory_limit (e.g. "256M").
	ArgMemoryLimit = "memory-limit"

	// ArgUploadMaxFilesize specifies the FPM upload_max_filesize (e.g. "10M").
	ArgUploadMaxFilesize = "upload-max-filesize"

	// ArgPostMaxSize specifies the FPM post_max_size (e.g. "12M").
	ArgPostMaxSize = "post-max-size"

	// ArgMaxExecutionTime specifies the FPM max_execution_time in seconds.
	ArgMaxExecutionTime = "max-execution-time"

	// ArgMaxInputTime specifies the FPM max_input_time in seconds.
	ArgMaxInputTime = "max-input-time"

	// ArgOpcacheMemory specifies the opcache.memory_consumption value.
	ArgOpcacheMemory = "opcache-memory-consumption"

	// ArgOpcacheMaxFiles specifies the opcache.max_accelerated_files value.
	ArgOpcacheMaxFiles = "opcache-max-accelerated-files"

	// ArgOpcacheValidateTimestamps specifies whether OPcache checks file
	// mtimes for changes ("On" = dev-friendly auto-reload, "Off" = max
	// production performance, requires FPM restart after deploys).
	ArgOpcacheValidateTimestamps = "opcache-validate-timestamps"

	// ArgConfDPath overrides the conf.d drop-in path (default derived from
	// php-version via DefaultConfDPathPattern).
	ArgConfDPath = "conf-d-path"

	// ArgPoolPath overrides the FPM pool config path (default derived from
	// php-version via DefaultPoolPathPattern).
	ArgPoolPath = "pool-path"

	// ArgErrorLog overrides the FPM error_log path (default derived from
	// php-version via DefaultErrorLogPattern).
	ArgErrorLog = "error-log"
)

// Default configuration constants for security playbooks.
const (
	// DefaultNonRootUser is the default non-root username to verify
	DefaultNonRootUser = "deploy"

	// DefaultSSHConfigPath is the default SSH configuration file path
	DefaultSSHConfigPath = "/etc/ssh/sshd_config"

	// DefaultMaxAuthTries is the default maximum authentication attempts
	DefaultMaxAuthTries = "3"

	// DefaultClientAliveInterval is the default client alive interval (seconds)
	DefaultClientAliveInterval = "300"

	// DefaultClientAliveCountMax is the default client alive count max
	DefaultClientAliveCountMax = "2"

	// DefaultSysctlConfigPath is the default sysctl configuration file path
	DefaultSysctlConfigPath = "/etc/sysctl.conf"

	// DefaultSysctlDropInPath is the default sysctl drop-in directory path
	DefaultSysctlDropInPath = "/etc/sysctl.d/99-security-hardening.conf"

	// DefaultKeyType is the default SSH key type
	DefaultKeyType = "ed25519"

	// DefaultKeyPath is empty — derived from username + key type when not set
	DefaultKeyPath = ""

	// PHP-FPM hardening defaults.
	//
	// DefaultDisableFunctions is the Laravel-safe disable_functions set. It
	// omits proc_open because Laravel's process-based components (e.g. Symfony
	// Process, queue workers) need it. For a non-Laravel app, override with a
	// stricter list that includes proc_open.
	DefaultDisableFunctions = "exec, shell_exec, system, passthru, popen"

	// DefaultMemoryLimit is the default FPM memory_limit.
	DefaultMemoryLimit = "256M"

	// DefaultUploadMaxFilesize is the default FPM upload_max_filesize.
	DefaultUploadMaxFilesize = "10M"

	// DefaultPostMaxSize is the default FPM post_max_size.
	DefaultPostMaxSize = "12M"

	// DefaultMaxExecutionTime is the default FPM max_execution_time (seconds).
	DefaultMaxExecutionTime = "60"

	// DefaultMaxInputTime is the default FPM max_input_time (seconds).
	DefaultMaxInputTime = "60"

	// DefaultOpcacheMemory is the default opcache.memory_consumption (MB).
	DefaultOpcacheMemory = "128"

	// DefaultOpcacheMaxFiles is the default opcache.max_accelerated_files.
	DefaultOpcacheMaxFiles = "10000"

	// DefaultOpcacheValidateTimestamps is the default for
	// opcache.validate_timestamps. "Off" is the production best practice
	// (max performance, requires FPM restart after deploys). Set to "On"
	// for low-traffic or dev sites where git pull should be visible without
	// a restart.
	DefaultOpcacheValidateTimestamps = "Off"

	// DefaultConfDPathPattern is the conf.d drop-in path pattern (%s = version).
	DefaultConfDPathPattern = "/etc/php/%s/fpm/conf.d/99-hardening.ini"

	// DefaultPoolPathPattern is the FPM pool config path pattern (%s = version).
	DefaultPoolPathPattern = "/etc/php/%s/fpm/pool.d/www.conf"

	// DefaultErrorLogPattern is the FPM error_log path pattern (%s = version).
	DefaultErrorLogPattern = "/var/log/php%s-fpm.log"
)
