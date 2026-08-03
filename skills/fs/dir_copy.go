package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// DirCopy copies a directory recursively on the remote server (cp -Rp).
// This is NOT an upload from local — both src and dst are remote paths.
// Permissions, timestamps, and symlinks are preserved (-p flag).
//
// Usage:
//
//	node.Run(fs.NewDirCopy().SetSrc("/var/www/myapp").SetDst("/var/www/myapp.bak").SetForce(true))
//
// Args:
//   - src: Source directory path (required, must be absolute)
//   - dst: Destination directory path (required, must be absolute)
//   - force: Overwrite destination if it exists (optional, default "false")
//
// Idempotency:
//   - Check() returns false if dst exists and force is false
//   - Check() returns true if dst doesn't exist or force is true
type DirCopy struct {
	*types.BaseSkill
}

// Compile-time assertion that DirCopy implements types.RunnableInterface.
var _ types.RunnableInterface = (*DirCopy)(nil)

// Check determines if the copy needs to happen.
// Returns true if dst doesn't exist or force is true.
func (d *DirCopy) Check() (bool, error) {
	cfg := d.GetNodeConfig()
	src := d.GetArg(ArgSrc)
	dst := d.GetArg(ArgDst)

	if err := validatePath(src); err != nil {
		return false, err
	}
	if err := validatePath(dst); err != nil {
		return false, err
	}

	// In dry-run mode, assume changes are needed without running SSH commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if copy is needed")
		return true, nil
	}

	// Check if src exists — if not, nothing to copy
	if !dirExists(cfg, src) {
		return false, nil
	}

	// Check if dst exists
	if !dirExists(cfg, dst) {
		// dst doesn't exist — needs copy
		return true, nil
	}

	// dst exists — check force flag
	force := d.GetArg(ArgForce)
	if !isTrue(force) {
		// Don't overwrite — no change needed
		return false, nil
	}

	// force is true — needs copy
	return true, nil
}

// Run copies the directory from src to dst recursively.
// Changed is true when the directory was copied, false if already exists and force is false.
func (d *DirCopy) Run() types.Result {
	src := d.GetArg(ArgSrc)
	dst := d.GetArg(ArgDst)
	force := d.GetArg(ArgForce)
	if force == "" {
		force = DefaultForce
	}

	if err := validatePath(src); err != nil {
		return types.Result{Changed: false, Message: "Invalid source path", Error: err}
	}
	if err := validatePath(dst); err != nil {
		return types.Result{Changed: false, Message: "Invalid destination path", Error: err}
	}

	cfg := d.GetNodeConfig()

	// In dry-run mode, skip all SSH existence checks — Check() also returns
	// true in dry-run, so we proceed directly to the dry-run command logging.
	if cfg.IsDryRunMode {
		needsCopy, err := d.Check()
		if err != nil {
			return types.Result{Changed: false, Message: "Failed to check copy state", Error: err}
		}
		if !needsCopy {
			return types.Result{
				Changed: false,
				Message: "Destination already exists and force is false: " + dst,
			}
		}

		// validateDestructivePath is a pure local check, safe to run in dry-run
		if isTrue(force) {
			if err := validateDestructivePath(dst); err != nil {
				return types.Result{
					Changed: false,
					Message: "Unsafe destination path for removal: " + dst,
					Error:   err,
				}
			}
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", "rm -rf "+skills.ShellEscapeArg(dst))
		}

		cpCmd := "cp -Rp " + skills.ShellEscapeArg(src) + " " + skills.ShellEscapeArg(dst)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cpCmd)
		return types.Result{
			Changed: true,
			Message: "Would copy directory: " + src + " -> " + dst,
		}
	}

	// Non-dry-run: check if src exists
	if !dirExists(cfg, src) {
		return types.Result{
			Changed: false,
			Message: "Source directory does not exist: " + src,
			Error:   fmt.Errorf("source directory does not exist: %s", src),
		}
	}

	// Check if already correct
	needsCopy, err := d.Check()
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to check copy state", Error: err}
	}
	if !needsCopy {
		return types.Result{
			Changed: false,
			Message: "Destination already exists and force is false: " + dst,
		}
	}

	// When force=true and dst exists, remove dst first to avoid cp copying
	// src *inside* dst (which is standard cp behavior for existing directories).
	// validateDestructivePath guards against rm -rf on root-level paths.
	if isTrue(force) && dirExists(cfg, dst) {
		if err := validateDestructivePath(dst); err != nil {
			return types.Result{
				Changed: false,
				Message: "Unsafe destination path for removal: " + dst,
				Error:   err,
			}
		}

		rmCmd := types.Command{
			Command:     "rm -rf " + skills.ShellEscapeArg(dst),
			Description: "Remove existing destination: " + dst,
		}

		cfg.GetLoggerOrDefault().Info("removing existing destination", "dst", dst)
		if _, err := ssh.Run(cfg, rmCmd); err != nil {
			return types.Result{
				Changed: false,
				Message: "Failed to remove existing destination: " + dst,
				Error:   fmt.Errorf("failed to remove existing destination %s: %w", dst, err),
			}
		}
	}

	cpCmd := "cp -Rp " + skills.ShellEscapeArg(src) + " " + skills.ShellEscapeArg(dst)

	cmdCp := types.Command{
		Command:     cpCmd,
		Description: "Copy directory: " + src + " -> " + dst,
	}

	cfg.GetLoggerOrDefault().Info("copying directory", "src", src, "dst", dst)
	output, err := ssh.Run(cfg, cmdCp)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to copy directory",
			Error:   fmt.Errorf("failed to copy directory: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("directory copied", "src", src, "dst", dst)
	return types.Result{
		Changed: true,
		Message: "Directory copied: " + src + " -> " + dst,
		Details: map[string]string{"src": src, "dst": dst},
	}
}

// SetSrc sets the source path and returns DirCopy for chaining.
func (d *DirCopy) SetSrc(src string) *DirCopy {
	d.BaseSkill.SetArg(ArgSrc, src)
	return d
}

// SetDst sets the destination path and returns DirCopy for chaining.
func (d *DirCopy) SetDst(dst string) *DirCopy {
	d.BaseSkill.SetArg(ArgDst, dst)
	return d
}

// SetForce sets whether to overwrite destination and returns DirCopy for chaining.
func (d *DirCopy) SetForce(force bool) *DirCopy {
	d.BaseSkill.SetArg(ArgForce, fmt.Sprintf("%v", force))
	return d
}

// WithNodeConfig sets the node config and returns DirCopy for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (d *DirCopy) WithNodeConfig(cfg types.NodeConfig) *DirCopy {
	d.BaseSkill.SetNodeConfig(cfg)
	return d
}

// SetArgs sets the arguments and returns DirCopy for fluent chaining.
func (d *DirCopy) SetArgs(args map[string]string) types.RunnableInterface {
	d.BaseSkill.SetArgs(args)
	return d
}

// SetArg sets a single argument and returns DirCopy for fluent chaining.
func (d *DirCopy) SetArg(key, value string) types.RunnableInterface {
	d.BaseSkill.SetArg(key, value)
	return d
}

// SetID sets the ID and returns DirCopy for fluent chaining.
func (d *DirCopy) SetID(id string) types.RunnableInterface {
	d.BaseSkill.SetID(id)
	return d
}

// SetDescription sets the description and returns DirCopy for fluent chaining.
func (d *DirCopy) SetDescription(description string) types.RunnableInterface {
	d.BaseSkill.SetDescription(description)
	return d
}

// SetTimeout sets the timeout and returns DirCopy for fluent chaining.
func (d *DirCopy) SetTimeout(timeout time.Duration) types.RunnableInterface {
	d.BaseSkill.SetTimeout(timeout)
	return d
}

// NewDirCopy creates a new fs-dir-copy skill.
func NewDirCopy() *DirCopy {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSDirCopy)
	pb.SetDescription("Copy directory recursively on remote server (cp -Rp)")
	return &DirCopy{BaseSkill: pb}
}
