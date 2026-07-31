package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// ChangeOwner changes ownership of a file or directory (chown).
// Optionally recursive with chown -R.
//
// Usage:
//
//	node.Run(fs.NewChangeOwner().SetArgs(map[string]string{
//	    fs.ArgPath:      "/var/www/myapp",
//	    fs.ArgOwner:     "www-data:www-data",
//	    fs.ArgRecursive: "true",
//	}))
//
// Args:
//   - path: File or directory path (required, must be absolute)
//   - owner: Owner in user:group format (required, e.g. "www-data:www-data")
//   - recursive: Apply recursively (optional, default "false")
//
// Idempotency:
//   - Check() returns false if current owner already matches desired owner
//   - Check() returns true if ownership mismatch
type ChangeOwner struct {
	*types.BaseSkill
}

// Check determines if ownership needs to be changed.
// Returns true if current owner doesn't match desired owner.
func (c *ChangeOwner) Check() (bool, error) {
	cfg := c.GetNodeConfig()
	path := c.GetArg(ArgPath)
	owner := c.GetArg(ArgOwner)

	if err := validatePath(path); err != nil {
		return false, err
	}
	if err := validateOwner(owner); err != nil {
		return false, err
	}

	// In dry-run mode, assume changes are needed without running SSH commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if owner change is needed")
		return true, nil
	}

	// Get current owner
	currentOwner := getOwner(cfg, path)
	return currentOwner != owner, nil
}

// Run changes the ownership of the path.
// Changed is true when ownership was changed, false if already correct.
func (c *ChangeOwner) Run() types.Result {
	path := c.GetArg(ArgPath)
	owner := c.GetArg(ArgOwner)
	recursive := c.GetArg(ArgRecursive)

	if err := validatePath(path); err != nil {
		return types.Result{Changed: false, Message: "Invalid path", Error: err}
	}
	if err := validateOwner(owner); err != nil {
		return types.Result{Changed: false, Message: "Invalid owner", Error: err}
	}

	cfg := c.GetNodeConfig()
	escPath := skills.ShellEscapeArg(path)

	chownCmd := "chown"
	if isTrue(recursive) {
		chownCmd += " -R"
	}
	chownCmd += " " + skills.ShellEscapeArg(owner) + " " + escPath

	cmdChown := types.Command{
		Command:     chownCmd,
		Description: "Change owner to " + owner + " on " + path,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdChown.Command)
		return types.Result{
			Changed: true,
			Message: "Would change owner to " + owner + " on " + path,
		}
	}

	// Check if already correct
	needsChange, err := c.Check()
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to check owner", Error: err}
	}
	if !needsChange {
		return types.Result{
			Changed: false,
			Message: "Owner already correct: " + owner + " on " + path,
		}
	}

	cfg.GetLoggerOrDefault().Info("changing owner", "path", path, "owner", owner)
	output, err := ssh.Run(cfg, cmdChown)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to change owner",
			Error:   fmt.Errorf("failed to change owner: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("owner changed", "path", path, "owner", owner)
	return types.Result{
		Changed: true,
		Message: "Owner changed to " + owner + " on " + path,
		Details: map[string]string{"path": path, "owner": owner},
	}
}

// WithNodeConfig sets the node config and returns ChangeOwner for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (c *ChangeOwner) WithNodeConfig(cfg types.NodeConfig) *ChangeOwner {
	c.BaseSkill.SetNodeConfig(cfg)
	return c
}

// SetArgs sets the arguments and returns ChangeOwner for fluent chaining.
func (c *ChangeOwner) SetArgs(args map[string]string) types.RunnableInterface {
	c.BaseSkill.SetArgs(args)
	return c
}

// SetArg sets a single argument and returns ChangeOwner for fluent chaining.
func (c *ChangeOwner) SetArg(key, value string) types.RunnableInterface {
	c.BaseSkill.SetArg(key, value)
	return c
}

// SetID sets the ID and returns ChangeOwner for fluent chaining.
func (c *ChangeOwner) SetID(id string) types.RunnableInterface {
	c.BaseSkill.SetID(id)
	return c
}

// SetDescription sets the description and returns ChangeOwner for fluent chaining.
func (c *ChangeOwner) SetDescription(description string) types.RunnableInterface {
	c.BaseSkill.SetDescription(description)
	return c
}

// SetTimeout sets the timeout and returns ChangeOwner for fluent chaining.
func (c *ChangeOwner) SetTimeout(timeout time.Duration) types.RunnableInterface {
	c.BaseSkill.SetTimeout(timeout)
	return c
}

// NewChangeOwner creates a new fs-change-owner skill.
func NewChangeOwner() *ChangeOwner {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSChangeOwner)
	pb.SetDescription("Change file/directory ownership (chown)")
	return &ChangeOwner{BaseSkill: pb}
}
