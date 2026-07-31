package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// DirDelete deletes a directory. By default it removes recursively with force (rm -rf).
// Set recursive="false" to use rmdir instead (only works on empty directories).
// Set force="false" to use rm -r without -f (will fail on write-protected files).
//
// Usage:
//
//	node.Run(fs.NewDirDelete().SetArg(fs.ArgPath, "/tmp/old-build"))
//
// Args:
//   - path: Directory path to delete (required, must be absolute)
//   - recursive: Remove recursively, default "true" (rm -rf vs rmdir)
//   - force: Force removal (ignore non-existent files, no prompt), default "true"
//
// Idempotency:
//   - Check() returns false if directory doesn't exist
//   - Check() returns true if directory exists
type DirDelete struct {
	*types.BaseSkill
}

// Check determines if the directory needs to be deleted.
// Returns true if the directory exists, false if it doesn't.
func (d *DirDelete) Check() (bool, error) {
	cfg := d.GetNodeConfig()
	path := d.GetArg(ArgPath)

	if err := validatePath(path); err != nil {
		return false, err
	}

	// In dry-run mode, assume changes are needed without running SSH commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if directory deletion is needed")
		return true, nil
	}

	return dirExists(cfg, path), nil
}

// Run deletes the directory if it exists.
// Changed is true when a directory was deleted, false if it didn't exist.
func (d *DirDelete) Run() types.Result {
	path := d.GetArg(ArgPath)

	if err := validatePath(path); err != nil {
		return types.Result{
			Changed: false,
			Message: "Invalid path",
			Error:   err,
		}
	}

	recursive := d.GetArg(ArgRecursive)
	if recursive == "" {
		recursive = "true"
	}
	force := d.GetArg(ArgForce)
	if force == "" {
		force = "true"
	}

	cfg := d.GetNodeConfig()

	needsDelete, err := d.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check directory existence",
			Error:   err,
		}
	}

	if !needsDelete {
		return types.Result{
			Changed: false,
			Message: "Directory does not exist: " + path,
		}
	}

	escPath := skills.ShellEscapeArg(path)
	var cmdStr string
	if isTrue(recursive) {
		cmdStr = "rm -r"
		if isTrue(force) {
			cmdStr += " -f"
		}
		cmdStr += " " + escPath
	} else {
		cmdStr = fmt.Sprintf("rmdir %s", escPath)
	}

	cmdDelete := types.Command{
		Command:     cmdStr,
		Description: "Delete directory: " + path,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdDelete.Command)
		return types.Result{
			Changed: true,
			Message: "Would delete directory: " + path,
		}
	}

	cfg.GetLoggerOrDefault().Info("deleting directory", "path", path)
	output, err := ssh.Run(cfg, cmdDelete)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to delete directory",
			Error:   fmt.Errorf("failed to delete directory: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("directory deleted", "path", path)
	return types.Result{
		Changed: true,
		Message: "Directory deleted: " + path,
		Details: map[string]string{"path": path},
	}
}

// SetPath sets the directory path and returns DirDelete for chaining.
func (d *DirDelete) SetPath(path string) *DirDelete {
	d.BaseSkill.SetArg(ArgPath, path)
	return d
}

// SetRecursive sets whether to remove recursively and returns DirDelete for chaining.
func (d *DirDelete) SetRecursive(recursive bool) *DirDelete {
	d.BaseSkill.SetArg(ArgRecursive, fmt.Sprintf("%v", recursive))
	return d
}

// SetForce sets whether to force removal and returns DirDelete for chaining.
func (d *DirDelete) SetForce(force bool) *DirDelete {
	d.BaseSkill.SetArg(ArgForce, fmt.Sprintf("%v", force))
	return d
}

// WithNodeConfig sets the node config and returns DirDelete for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (d *DirDelete) WithNodeConfig(cfg types.NodeConfig) *DirDelete {
	d.BaseSkill.SetNodeConfig(cfg)
	return d
}

// SetArgs sets the arguments and returns DirDelete for fluent chaining.
func (d *DirDelete) SetArgs(args map[string]string) types.RunnableInterface {
	d.BaseSkill.SetArgs(args)
	return d
}

// SetArg sets a single argument and returns DirDelete for fluent chaining.
func (d *DirDelete) SetArg(key, value string) types.RunnableInterface {
	d.BaseSkill.SetArg(key, value)
	return d
}

// SetID sets the ID and returns DirDelete for fluent chaining.
func (d *DirDelete) SetID(id string) types.RunnableInterface {
	d.BaseSkill.SetID(id)
	return d
}

// SetDescription sets the description and returns DirDelete for fluent chaining.
func (d *DirDelete) SetDescription(description string) types.RunnableInterface {
	d.BaseSkill.SetDescription(description)
	return d
}

// SetTimeout sets the timeout and returns DirDelete for fluent chaining.
func (d *DirDelete) SetTimeout(timeout time.Duration) types.RunnableInterface {
	d.BaseSkill.SetTimeout(timeout)
	return d
}

// NewDirDelete creates a new fs-dir-delete skill.
func NewDirDelete() *DirDelete {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSDirDelete)
	pb.SetDescription("Delete a directory")
	return &DirDelete{BaseSkill: pb}
}
