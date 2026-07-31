package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// FileCopy copies a file on the remote server (cp).
// This is NOT an upload from local — both src and dst are remote paths.
//
// Usage:
//
//	node.Run(fs.NewFileCopy().SetArgs(map[string]string{
//	    fs.ArgSrc:   "/etc/ssh/sshd_config",
//	    fs.ArgDst:   "/etc/ssh/sshd_config.bak",
//	    fs.ArgForce: "true",
//	}))
//
// Args:
//   - src: Source file path (required, must be absolute)
//   - dst: Destination file path (required, must be absolute)
//   - force: Overwrite destination if it exists (optional, default "false")
//
// Idempotency:
//   - Check() returns false if dst exists and content matches src
//   - Check() returns true if dst doesn't exist or content differs
type FileCopy struct {
	*types.BaseSkill
}

// Check determines if the copy needs to happen.
// Returns true if dst doesn't exist or content differs from src.
func (f *FileCopy) Check() (bool, error) {
	cfg := f.GetNodeConfig()
	src := f.GetArg(ArgSrc)
	dst := f.GetArg(ArgDst)

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
	if !fileExists(cfg, src) {
		return false, nil
	}

	// Check if dst exists
	if !fileExists(cfg, dst) {
		// dst doesn't exist — needs copy
		return true, nil
	}

	// dst exists — check force flag
	force := f.GetArg(ArgForce)
	if !isTrue(force) {
		// Don't overwrite — no change needed
		return false, nil
	}

	// force is true — check if content differs
	if !filesIdentical(cfg, src, dst) {
		// Content differs — needs copy
		return true, nil
	}

	return false, nil
}

// Run copies the file from src to dst.
// Changed is true when the file was copied, false if already identical.
func (f *FileCopy) Run() types.Result {
	src := f.GetArg(ArgSrc)
	dst := f.GetArg(ArgDst)
	force := f.GetArg(ArgForce)
	if force == "" {
		force = DefaultForce
	}

	if err := validatePath(src); err != nil {
		return types.Result{Changed: false, Message: "Invalid source path", Error: err}
	}
	if err := validatePath(dst); err != nil {
		return types.Result{Changed: false, Message: "Invalid destination path", Error: err}
	}

	cfg := f.GetNodeConfig()

	// Check if src exists
	if !fileExists(cfg, src) {
		return types.Result{
			Changed: false,
			Message: "Source file does not exist: " + src,
			Error:   fmt.Errorf("source file does not exist: %s", src),
		}
	}

	// Check if already correct
	needsCopy, err := f.Check()
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to check copy state", Error: err}
	}
	if !needsCopy {
		return types.Result{
			Changed: false,
			Message: "Destination already has identical content: " + dst,
		}
	}

	// Check if dst exists and force is false
	if fileExists(cfg, dst) && !isTrue(force) {
		return types.Result{
			Changed: false,
			Message: "Destination exists and force is false: " + dst,
		}
	}

	cpCmd := "cp"
	if isTrue(force) {
		cpCmd += " -f"
	}
	cpCmd += " " + skills.ShellEscapeArg(src) + " " + skills.ShellEscapeArg(dst)

	cmdCp := types.Command{
		Command:     cpCmd,
		Description: "Copy file: " + src + " -> " + dst,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdCp.Command)
		return types.Result{
			Changed: true,
			Message: "Would copy file: " + src + " -> " + dst,
		}
	}

	cfg.GetLoggerOrDefault().Info("copying file", "src", src, "dst", dst)
	output, err := ssh.Run(cfg, cmdCp)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to copy file",
			Error:   fmt.Errorf("failed to copy file: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("file copied", "src", src, "dst", dst)
	return types.Result{
		Changed: true,
		Message: "File copied: " + src + " -> " + dst,
		Details: map[string]string{"src": src, "dst": dst},
	}
}

// WithNodeConfig sets the node config and returns FileCopy for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (f *FileCopy) WithNodeConfig(cfg types.NodeConfig) *FileCopy {
	f.BaseSkill.SetNodeConfig(cfg)
	return f
}

// SetArgs sets the arguments and returns FileCopy for fluent chaining.
func (f *FileCopy) SetArgs(args map[string]string) types.RunnableInterface {
	f.BaseSkill.SetArgs(args)
	return f
}

// SetArg sets a single argument and returns FileCopy for fluent chaining.
func (f *FileCopy) SetArg(key, value string) types.RunnableInterface {
	f.BaseSkill.SetArg(key, value)
	return f
}

// SetID sets the ID and returns FileCopy for fluent chaining.
func (f *FileCopy) SetID(id string) types.RunnableInterface {
	f.BaseSkill.SetID(id)
	return f
}

// SetDescription sets the description and returns FileCopy for fluent chaining.
func (f *FileCopy) SetDescription(description string) types.RunnableInterface {
	f.BaseSkill.SetDescription(description)
	return f
}

// SetTimeout sets the timeout and returns FileCopy for fluent chaining.
func (f *FileCopy) SetTimeout(timeout time.Duration) types.RunnableInterface {
	f.BaseSkill.SetTimeout(timeout)
	return f
}

// NewFileCopy creates a new fs-file-copy skill.
func NewFileCopy() *FileCopy {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSFileCopy)
	pb.SetDescription("Copy file on remote server (cp)")
	return &FileCopy{BaseSkill: pb}
}
