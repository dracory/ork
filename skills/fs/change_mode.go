package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// ChangeMode changes permissions of a file or directory (chmod).
// Optionally recursive with chmod -R.
//
// Usage:
//
//	node.Run(fs.NewChangeMode().SetArgs(map[string]string{
//	    fs.ArgPath:  "/var/www/myapp/.ssh",
//	    fs.ArgMode:  "700",
//	}))
//
// Args:
//   - path: File or directory path (required, must be absolute)
//   - mode: Permissions in octal (required, e.g. "755", "600")
//   - recursive: Apply recursively (optional, default "false")
//
// Idempotency:
//   - Check() returns false if current mode already matches desired mode
//   - Check() returns true if mode mismatch
type ChangeMode struct {
	*types.BaseSkill
}

// Compile-time assertion that ChangeMode implements types.RunnableInterface.
var _ types.RunnableInterface = (*ChangeMode)(nil)

// Check determines if mode needs to be changed.
// Returns true if current mode doesn't match desired mode.
func (c *ChangeMode) Check() (bool, error) {
	cfg := c.GetNodeConfig()
	path := c.GetArg(ArgPath)
	mode := c.GetArg(ArgMode)

	if err := validatePath(path); err != nil {
		return false, err
	}
	if err := validateMode(mode); err != nil {
		return false, err
	}

	// In dry-run mode, assume changes are needed without running SSH commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if mode change is needed")
		return true, nil
	}

	currentMode := getMode(cfg, path)
	return currentMode != mode, nil
}

// Run changes the permissions of the path.
// Changed is true when mode was changed, false if already correct.
func (c *ChangeMode) Run() types.Result {
	path := c.GetArg(ArgPath)
	mode := c.GetArg(ArgMode)
	recursive := c.GetArg(ArgRecursive)

	if err := validatePath(path); err != nil {
		return types.Result{Changed: false, Message: "Invalid path", Error: err}
	}
	if err := validateMode(mode); err != nil {
		return types.Result{Changed: false, Message: "Invalid mode", Error: err}
	}

	cfg := c.GetNodeConfig()
	escPath := skills.ShellEscapeArg(path)

	chmodCmd := "chmod"
	if isTrue(recursive) {
		chmodCmd += " -R"
	}
	chmodCmd += " " + skills.ShellEscapeArg(mode) + " " + escPath

	cmdChmod := types.Command{
		Command:     chmodCmd,
		Description: "Change mode to " + mode + " on " + path,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdChmod.Command)
		return types.Result{
			Changed: true,
			Message: "Would change mode to " + mode + " on " + path,
		}
	}

	needsChange, err := c.Check()
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to check mode", Error: err}
	}
	if !needsChange {
		return types.Result{
			Changed: false,
			Message: "Mode already correct: " + mode + " on " + path,
		}
	}

	cfg.GetLoggerOrDefault().Info("changing mode", "path", path, "mode", mode)
	output, err := ssh.Run(cfg, cmdChmod)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to change mode",
			Error:   fmt.Errorf("failed to change mode: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("mode changed", "path", path, "mode", mode)
	return types.Result{
		Changed: true,
		Message: "Mode changed to " + mode + " on " + path,
		Details: map[string]string{"path": path, "mode": mode},
	}
}

// SetPath sets the file/directory path and returns ChangeMode for chaining.
func (c *ChangeMode) SetPath(path string) *ChangeMode {
	c.BaseSkill.SetArg(ArgPath, path)
	return c
}

// SetMode sets the permissions (octal, e.g. "755") and returns ChangeMode for chaining.
func (c *ChangeMode) SetMode(mode string) *ChangeMode {
	c.BaseSkill.SetArg(ArgMode, mode)
	return c
}

// SetRecursive sets whether to apply recursively and returns ChangeMode for chaining.
func (c *ChangeMode) SetRecursive(recursive bool) *ChangeMode {
	c.BaseSkill.SetArg(ArgRecursive, fmt.Sprintf("%v", recursive))
	return c
}

// WithNodeConfig sets the node config and returns ChangeMode for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (c *ChangeMode) WithNodeConfig(cfg types.NodeConfig) *ChangeMode {
	c.BaseSkill.SetNodeConfig(cfg)
	return c
}

// SetArgs sets the arguments and returns ChangeMode for fluent chaining.
func (c *ChangeMode) SetArgs(args map[string]string) types.RunnableInterface {
	c.BaseSkill.SetArgs(args)
	return c
}

// SetArg sets a single argument and returns ChangeMode for fluent chaining.
func (c *ChangeMode) SetArg(key, value string) types.RunnableInterface {
	c.BaseSkill.SetArg(key, value)
	return c
}

// SetID sets the ID and returns ChangeMode for fluent chaining.
func (c *ChangeMode) SetID(id string) types.RunnableInterface {
	c.BaseSkill.SetID(id)
	return c
}

// SetDescription sets the description and returns ChangeMode for fluent chaining.
func (c *ChangeMode) SetDescription(description string) types.RunnableInterface {
	c.BaseSkill.SetDescription(description)
	return c
}

// SetTimeout sets the timeout and returns ChangeMode for fluent chaining.
func (c *ChangeMode) SetTimeout(timeout time.Duration) types.RunnableInterface {
	c.BaseSkill.SetTimeout(timeout)
	return c
}

// NewChangeMode creates a new fs-change-mode skill.
func NewChangeMode() *ChangeMode {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSChangeMode)
	pb.SetDescription("Change file/directory permissions (chmod)")
	return &ChangeMode{BaseSkill: pb}
}
