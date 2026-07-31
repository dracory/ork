package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Remove removes a file or directory (rm).
// For type-specific deletion with validation, use FileDelete or DirDelete.
//
// Usage:
//
//	node.Run(fs.NewRemove().SetArgs(map[string]string{
//	    fs.ArgPath:      "/tmp/old-data",
//	    fs.ArgRecursive: "true",
//	    fs.ArgForce:     "true",
//	}))
//
// Args:
//   - path: File or directory path to remove (required, must be absolute)
//   - recursive: Remove recursively, rm -r (optional, default "false")
//   - force: Force remove, rm -f, ignore errors (optional, default "false")
//
// Idempotency:
//   - Check() returns false if path doesn't exist
//   - Check() returns true if path exists and needs removal
type Remove struct {
	*types.BaseSkill
}

// Check determines if the path needs to be removed.
// Returns true if the path exists, false if it doesn't.
func (r *Remove) Check() (bool, error) {
	cfg := r.GetNodeConfig()
	path := r.GetArg(ArgPath)

	if err := validatePath(path); err != nil {
		return false, err
	}

	// In dry-run mode, assume changes are needed without running SSH commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if removal is needed")
		return true, nil
	}

	return pathExists(cfg, path), nil
}

// Run removes the path if it exists.
// Changed is true when the path was removed, false if it didn't exist.
func (r *Remove) Run() types.Result {
	path := r.GetArg(ArgPath)
	recursive := r.GetArg(ArgRecursive)
	force := r.GetArg(ArgForce)

	if err := validatePath(path); err != nil {
		return types.Result{Changed: false, Message: "Invalid path", Error: err}
	}

	cfg := r.GetNodeConfig()

	needsRemove, err := r.Check()
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to check path", Error: err}
	}

	if !needsRemove {
		return types.Result{
			Changed: false,
			Message: "Path does not exist: " + path,
		}
	}

	escPath := skills.ShellEscapeArg(path)
	rmCmd := "rm"
	if isTrue(recursive) {
		rmCmd += " -r"
	}
	if isTrue(force) {
		rmCmd += " -f"
	}
	rmCmd += " " + escPath

	cmdRm := types.Command{
		Command:     rmCmd,
		Description: "Remove: " + path,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdRm.Command)
		return types.Result{
			Changed: true,
			Message: "Would remove: " + path,
		}
	}

	cfg.GetLoggerOrDefault().Info("removing path", "path", path)
	output, err := ssh.Run(cfg, cmdRm)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to remove path",
			Error:   fmt.Errorf("failed to remove: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("path removed", "path", path)
	return types.Result{
		Changed: true,
		Message: "Removed: " + path,
		Details: map[string]string{"path": path},
	}
}

// SetPath sets the path to remove and returns Remove for chaining.
func (r *Remove) SetPath(path string) *Remove {
	r.BaseSkill.SetArg(ArgPath, path)
	return r
}

// SetRecursive sets whether to remove recursively and returns Remove for chaining.
func (r *Remove) SetRecursive(recursive bool) *Remove {
	r.BaseSkill.SetArg(ArgRecursive, fmt.Sprintf("%v", recursive))
	return r
}

// SetForce sets whether to force removal and returns Remove for chaining.
func (r *Remove) SetForce(force bool) *Remove {
	r.BaseSkill.SetArg(ArgForce, fmt.Sprintf("%v", force))
	return r
}

// WithNodeConfig sets the node config and returns Remove for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (r *Remove) WithNodeConfig(cfg types.NodeConfig) *Remove {
	r.BaseSkill.SetNodeConfig(cfg)
	return r
}

// SetArgs sets the arguments and returns Remove for fluent chaining.
func (r *Remove) SetArgs(args map[string]string) types.RunnableInterface {
	r.BaseSkill.SetArgs(args)
	return r
}

// SetArg sets a single argument and returns Remove for fluent chaining.
func (r *Remove) SetArg(key, value string) types.RunnableInterface {
	r.BaseSkill.SetArg(key, value)
	return r
}

// SetID sets the ID and returns Remove for fluent chaining.
func (r *Remove) SetID(id string) types.RunnableInterface {
	r.BaseSkill.SetID(id)
	return r
}

// SetDescription sets the description and returns Remove for fluent chaining.
func (r *Remove) SetDescription(description string) types.RunnableInterface {
	r.BaseSkill.SetDescription(description)
	return r
}

// SetTimeout sets the timeout and returns Remove for fluent chaining.
func (r *Remove) SetTimeout(timeout time.Duration) types.RunnableInterface {
	r.BaseSkill.SetTimeout(timeout)
	return r
}

// NewRemove creates a new fs-remove skill.
func NewRemove() *Remove {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSRemove)
	pb.SetDescription("Remove file or directory (rm)")
	return &Remove{BaseSkill: pb}
}
