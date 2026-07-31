package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Rename renames or moves a file or directory (mv).
//
// Usage:
//
//	node.Run(fs.NewRename().SetArgs(map[string]string{
//	    fs.ArgSrc:   "/tmp/config.tmp",
//	    fs.ArgDst:   "/etc/myapp/config",
//	    fs.ArgForce: "true",
//	}))
//
// Args:
//   - src: Source path (required, must be absolute)
//   - dst: Destination path (required, must be absolute)
//   - force: Overwrite destination if it exists (optional, default "false")
//
// Idempotency:
//   - Check() returns false if src doesn't exist and dst exists (already renamed)
//   - Check() returns true if src exists (needs rename)
type Rename struct {
	*types.BaseSkill
}

// Check determines if the rename needs to happen.
// Returns true if src exists (needs rename), false if src is gone and dst exists.
func (r *Rename) Check() (bool, error) {
	cfg := r.GetNodeConfig()
	src := r.GetArg(ArgSrc)
	dst := r.GetArg(ArgDst)

	if err := validatePath(src); err != nil {
		return false, err
	}
	if err := validatePath(dst); err != nil {
		return false, err
	}

	// In dry-run mode, assume changes are needed without running SSH commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if rename is needed")
		return true, nil
	}

	// Check if src exists
	if !pathExists(cfg, src) {
		// src doesn't exist — check if dst exists (already renamed)
		if pathExists(cfg, dst) {
			// dst exists, src doesn't — already renamed
			return false, nil
		}
		// Neither exists — nothing to do
		return false, nil
	}

	return true, nil
}

// Run renames/moves the file or directory from src to dst.
// Changed is true when the rename succeeded, false if already done.
func (r *Rename) Run() types.Result {
	src := r.GetArg(ArgSrc)
	dst := r.GetArg(ArgDst)
	force := r.GetArg(ArgForce)
	if force == "" {
		force = DefaultForce
	}

	if err := validatePath(src); err != nil {
		return types.Result{Changed: false, Message: "Invalid source path", Error: err}
	}
	if err := validatePath(dst); err != nil {
		return types.Result{Changed: false, Message: "Invalid destination path", Error: err}
	}

	cfg := r.GetNodeConfig()

	needsRename, err := r.Check()
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to check rename state", Error: err}
	}
	if !needsRename {
		return types.Result{
			Changed: false,
			Message: "Nothing to rename: source does not exist: " + src,
		}
	}

	// Check if dst exists and force is false
	if pathExists(cfg, dst) && !isTrue(force) {
		return types.Result{
			Changed: false,
			Message: "Destination exists and force is false: " + dst,
			Error:   fmt.Errorf("destination exists and force is false: %s", dst),
		}
	}

	mvCmd := "mv"
	if isTrue(force) {
		mvCmd += " -f"
	}
	mvCmd += " " + skills.ShellEscapeArg(src) + " " + skills.ShellEscapeArg(dst)

	cmdMv := types.Command{
		Command:     mvCmd,
		Description: "Rename: " + src + " -> " + dst,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdMv.Command)
		return types.Result{
			Changed: true,
			Message: "Would rename: " + src + " -> " + dst,
		}
	}

	cfg.GetLoggerOrDefault().Info("renaming", "src", src, "dst", dst)
	output, err := ssh.Run(cfg, cmdMv)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to rename",
			Error:   fmt.Errorf("failed to rename: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("renamed", "src", src, "dst", dst)
	return types.Result{
		Changed: true,
		Message: "Renamed: " + src + " -> " + dst,
		Details: map[string]string{"src": src, "dst": dst},
	}
}

// SetSrc sets the source path and returns Rename for chaining.
func (r *Rename) SetSrc(src string) *Rename {
	r.BaseSkill.SetArg(ArgSrc, src)
	return r
}

// SetDst sets the destination path and returns Rename for chaining.
func (r *Rename) SetDst(dst string) *Rename {
	r.BaseSkill.SetArg(ArgDst, dst)
	return r
}

// SetForce sets whether to overwrite destination and returns Rename for chaining.
func (r *Rename) SetForce(force bool) *Rename {
	r.BaseSkill.SetArg(ArgForce, fmt.Sprintf("%v", force))
	return r
}

// WithNodeConfig sets the node config and returns Rename for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (r *Rename) WithNodeConfig(cfg types.NodeConfig) *Rename {
	r.BaseSkill.SetNodeConfig(cfg)
	return r
}

// SetArgs sets the arguments and returns Rename for fluent chaining.
func (r *Rename) SetArgs(args map[string]string) types.RunnableInterface {
	r.BaseSkill.SetArgs(args)
	return r
}

// SetArg sets a single argument and returns Rename for fluent chaining.
func (r *Rename) SetArg(key, value string) types.RunnableInterface {
	r.BaseSkill.SetArg(key, value)
	return r
}

// SetID sets the ID and returns Rename for fluent chaining.
func (r *Rename) SetID(id string) types.RunnableInterface {
	r.BaseSkill.SetID(id)
	return r
}

// SetDescription sets the description and returns Rename for fluent chaining.
func (r *Rename) SetDescription(description string) types.RunnableInterface {
	r.BaseSkill.SetDescription(description)
	return r
}

// SetTimeout sets the timeout and returns Rename for fluent chaining.
func (r *Rename) SetTimeout(timeout time.Duration) types.RunnableInterface {
	r.BaseSkill.SetTimeout(timeout)
	return r
}

// NewRename creates a new fs-rename skill.
func NewRename() *Rename {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSRename)
	pb.SetDescription("Rename/move file or directory (mv)")
	return &Rename{BaseSkill: pb}
}
