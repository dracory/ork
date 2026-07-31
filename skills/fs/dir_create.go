package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// DirCreate creates a directory with optional ownership and permissions.
// This is a readable, idempotent abstraction over mkdir, chown, and chmod.
//
// Usage:
//
//	node.Run(fs.NewDirCreate().SetArgs(map[string]string{
//	    fs.ArgPath:  "/var/www/myapp",
//	    fs.ArgOwner: "www-data:www-data",
//	    fs.ArgMode:  "755",
//	}))
//
// Args:
//   - path: Directory path to create (required, must be absolute)
//   - owner: Owner in user:group format (optional, e.g. "www-data:www-data")
//   - mode: Permissions in octal (optional, default "755")
//   - parents: Create parent directories if needed (optional, default "true")
//
// Idempotency:
//   - Check() returns false if directory already exists with correct owner and mode
//   - Check() returns true if directory doesn't exist or owner/mode mismatch
type DirCreate struct {
	*types.BaseSkill
}

// Compile-time assertion that DirCreate implements types.RunnableInterface.
var _ types.RunnableInterface = (*DirCreate)(nil)

// Check determines if the directory needs to be created or fixed.
// Returns true if the directory doesn't exist or ownership/permissions mismatch.
func (d *DirCreate) Check() (bool, error) {
	cfg := d.GetNodeConfig()
	path := d.GetArg(ArgPath)

	if err := validatePath(path); err != nil {
		return false, err
	}

	// In dry-run mode, assume changes are needed without running SSH commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if directory creation is needed")
		return true, nil
	}

	// Check if directory exists
	if !dirExists(cfg, path) {
		// Directory doesn't exist â€” needs creation
		return true, nil
	}

	// Directory exists â€” check if owner/mode need fixing
	owner := d.GetArg(ArgOwner)
	mode := d.GetArg(ArgMode)
	if mode == "" {
		mode = DefaultDirMode
	}

	if owner != "" {
		if err := validateOwner(owner); err != nil {
			return false, err
		}
		currentOwner := getOwner(cfg, path)
		if currentOwner != owner {
			return true, nil
		}
	}

	if err := validateMode(mode); err != nil {
		return false, err
	}
	currentMode := getMode(cfg, path)
	if currentMode != mode {
		return true, nil
	}

	return false, nil
}

// Run creates the directory and sets ownership/permissions.
// Changed is true when the directory was created or its owner/mode was changed.
func (d *DirCreate) Run() types.Result {
	path := d.GetArg(ArgPath)

	if err := validatePath(path); err != nil {
		return types.Result{
			Changed: false,
			Message: "Invalid path",
			Error:   err,
		}
	}

	owner := d.GetArg(ArgOwner)
	mode := d.GetArg(ArgMode)
	if mode == "" {
		mode = DefaultDirMode
	}
	parents := d.GetArg(ArgParents)
	if parents == "" {
		parents = DefaultParents
	}

	if owner != "" {
		if err := validateOwner(owner); err != nil {
			return types.Result{
				Changed: false,
				Message: "Invalid owner",
				Error:   err,
			}
		}
	}
	if err := validateMode(mode); err != nil {
		return types.Result{
			Changed: false,
			Message: "Invalid mode",
			Error:   err,
		}
	}

	cfg := d.GetNodeConfig()
	escPath := skills.ShellEscapeArg(path)

	// Build mkdir command
	mkdirCmd := "mkdir"
	if isTrue(parents) {
		mkdirCmd += " -p"
	}
	mkdirCmd += " " + escPath

	// Check if already exists and correct
	needsChange, err := d.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check directory state",
			Error:   err,
		}
	}

	if !needsChange {
		return types.Result{
			Changed: false,
			Message: "Directory already exists with correct owner and mode: " + path,
		}
	}

	// Build all commands
	cmdMkdir := types.Command{Command: mkdirCmd, Description: "Create directory: " + path}

	var cmdChown, cmdChmod types.Command
	if owner != "" {
		cmdChown = types.Command{
			Command:     fmt.Sprintf("chown %s %s", skills.ShellEscapeArg(owner), escPath),
			Description: "Set owner: " + owner + " on " + path,
		}
	}
	cmdChmod = types.Command{
		Command:     fmt.Sprintf("chmod %s %s", skills.ShellEscapeArg(mode), escPath),
		Description: "Set mode: " + mode + " on " + path,
	}

	// Dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdMkdir.Command)
		if cmdChown.Command != "" {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdChown.Command)
		}
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdChmod.Command)
		return types.Result{
			Changed: true,
			Message: "Would create directory: " + path,
		}
	}

	cfg.GetLoggerOrDefault().Info("creating directory", "path", path)

	// Create directory (use mkdir -p which is idempotent â€” won't fail if exists)
	output, err := ssh.Run(cfg, cmdMkdir)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to create directory",
			Error:   fmt.Errorf("failed to create directory: %w\nOutput: %s", err, output),
		}
	}

	// Set owner
	if cmdChown.Command != "" {
		output, err = ssh.Run(cfg, cmdChown)
		if err != nil {
			return types.Result{
				Changed: true,
				Message: "Directory created but failed to set owner: " + path,
				Error:   fmt.Errorf("failed to set owner: %w\nOutput: %s", err, output),
				Details: map[string]string{"path": path, "owner": owner},
			}
		}
	}

	// Set mode
	output, err = ssh.Run(cfg, cmdChmod)
	if err != nil {
		return types.Result{
			Changed: true,
			Message: "Directory created but failed to set mode: " + path,
			Error:   fmt.Errorf("failed to set mode: %w\nOutput: %s", err, output),
			Details: map[string]string{"path": path, "mode": mode},
		}
	}

	cfg.GetLoggerOrDefault().Info("directory created", "path", path)
	return types.Result{
		Changed: true,
		Message: "Directory created: " + path,
		Details: map[string]string{
			"path":  path,
			"mode":  mode,
			"owner": owner,
		},
	}
}

// SetPath sets the directory path and returns DirCreate for chaining.
func (d *DirCreate) SetPath(path string) *DirCreate {
	d.BaseSkill.SetArg(ArgPath, path)
	return d
}

// SetOwner sets the owner (user:group) and returns DirCreate for chaining.
func (d *DirCreate) SetOwner(owner string) *DirCreate {
	d.BaseSkill.SetArg(ArgOwner, owner)
	return d
}

// SetMode sets the permissions (octal, e.g. "755") and returns DirCreate for chaining.
func (d *DirCreate) SetMode(mode string) *DirCreate {
	d.BaseSkill.SetArg(ArgMode, mode)
	return d
}

// SetParents sets whether to create parent directories and returns DirCreate for chaining.
func (d *DirCreate) SetParents(parents bool) *DirCreate {
	d.BaseSkill.SetArg(ArgParents, fmt.Sprintf("%v", parents))
	return d
}

// WithNodeConfig sets the node config and returns DirCreate for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (d *DirCreate) WithNodeConfig(cfg types.NodeConfig) *DirCreate {
	d.BaseSkill.SetNodeConfig(cfg)
	return d
}

// SetArgs sets the arguments and returns DirCreate for fluent chaining.
func (d *DirCreate) SetArgs(args map[string]string) types.RunnableInterface {
	d.BaseSkill.SetArgs(args)
	return d
}

// SetArg sets a single argument and returns DirCreate for fluent chaining.
func (d *DirCreate) SetArg(key, value string) types.RunnableInterface {
	d.BaseSkill.SetArg(key, value)
	return d
}

// SetID sets the ID and returns DirCreate for fluent chaining.
func (d *DirCreate) SetID(id string) types.RunnableInterface {
	d.BaseSkill.SetID(id)
	return d
}

// SetDescription sets the description and returns DirCreate for fluent chaining.
func (d *DirCreate) SetDescription(description string) types.RunnableInterface {
	d.BaseSkill.SetDescription(description)
	return d
}

// SetTimeout sets the timeout and returns DirCreate for fluent chaining.
func (d *DirCreate) SetTimeout(timeout time.Duration) types.RunnableInterface {
	d.BaseSkill.SetTimeout(timeout)
	return d
}

// NewDirCreate creates a new fs-dir-create skill.
func NewDirCreate() *DirCreate {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSDirCreate)
	pb.SetDescription("Create directory with ownership and permissions")
	return &DirCreate{BaseSkill: pb}
}
