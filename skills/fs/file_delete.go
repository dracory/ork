package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// FileDelete deletes a single file. Non-recursive.
// For deleting directories, use DirDelete or Remove instead.
//
// Usage:
//
//	node.Run(fs.NewFileDelete().SetArg(fs.ArgPath, "/tmp/setup.log"))
//
// Args:
//   - path: File path to delete (required, must be absolute)
//
// Idempotency:
//   - Check() returns false if file doesn't exist (already deleted)
//   - Check() returns true if file exists and needs deletion
type FileDelete struct {
	*types.BaseSkill
}

// Check determines if the file needs to be deleted.
// Returns true if the file exists, false if it doesn't.
func (f *FileDelete) Check() (bool, error) {
	cfg := f.GetNodeConfig()
	path := f.GetArg(ArgPath)

	if err := validatePath(path); err != nil {
		return false, err
	}

	// In dry-run mode, assume changes are needed without running SSH commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if file deletion is needed")
		return true, nil
	}

	return fileExists(cfg, path), nil
}

// Run deletes the file if it exists.
// Changed is true when a file was deleted, false if it didn't exist.
func (f *FileDelete) Run() types.Result {
	path := f.GetArg(ArgPath)

	if err := validatePath(path); err != nil {
		return types.Result{
			Changed: false,
			Message: "Invalid path",
			Error:   err,
		}
	}

	cfg := f.GetNodeConfig()

	// Check if file exists
	needsDelete, err := f.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check file existence",
			Error:   err,
		}
	}

	if !needsDelete {
		return types.Result{
			Changed: false,
			Message: "File does not exist: " + path,
		}
	}

	cmdDelete := types.Command{
		Command:     fmt.Sprintf("rm -f %s", skills.ShellEscapeArg(path)),
		Description: "Delete file: " + path,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdDelete.Command)
		return types.Result{
			Changed: true,
			Message: "Would delete file: " + path,
		}
	}

	cfg.GetLoggerOrDefault().Info("deleting file", "path", path)
	output, err := ssh.Run(cfg, cmdDelete)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to delete file",
			Error:   fmt.Errorf("failed to delete file: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("file deleted", "path", path)
	return types.Result{
		Changed: true,
		Message: "File deleted: " + path,
		Details: map[string]string{"path": path},
	}
}

// WithNodeConfig sets the node config and returns FileDelete for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (f *FileDelete) WithNodeConfig(cfg types.NodeConfig) *FileDelete {
	f.BaseSkill.SetNodeConfig(cfg)
	return f
}

// SetArgs sets the arguments and returns FileDelete for fluent chaining.
func (f *FileDelete) SetArgs(args map[string]string) types.RunnableInterface {
	f.BaseSkill.SetArgs(args)
	return f
}

// SetArg sets a single argument and returns FileDelete for fluent chaining.
func (f *FileDelete) SetArg(key, value string) types.RunnableInterface {
	f.BaseSkill.SetArg(key, value)
	return f
}

// SetID sets the ID and returns FileDelete for fluent chaining.
func (f *FileDelete) SetID(id string) types.RunnableInterface {
	f.BaseSkill.SetID(id)
	return f
}

// SetDescription sets the description and returns FileDelete for fluent chaining.
func (f *FileDelete) SetDescription(description string) types.RunnableInterface {
	f.BaseSkill.SetDescription(description)
	return f
}

// SetTimeout sets the timeout and returns FileDelete for fluent chaining.
func (f *FileDelete) SetTimeout(timeout time.Duration) types.RunnableInterface {
	f.BaseSkill.SetTimeout(timeout)
	return f
}

// NewFileDelete creates a new fs-file-delete skill.
func NewFileDelete() *FileDelete {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSFileDelete)
	pb.SetDescription("Delete a single file")
	return &FileDelete{BaseSkill: pb}
}
