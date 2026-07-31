package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/types"
)

// FileExists checks if a file exists on the remote server.
// This is a read-only skill â€” it never modifies the system.
//
// Usage:
//
//	result := node.Run(fs.NewFileExists().SetArg(fs.ArgPath, "/etc/hostname")).FirstResult()
//	if result.Details["exists"] == "true" {
//	    fmt.Println("File exists")
//	}
//
// Args:
//   - path: File path to check (required, must be absolute)
//
// Result Details:
//   - exists: "true" or "false"
type FileExists struct {
	*types.BaseSkill
}

// Compile-time assertion that FileExists implements types.RunnableInterface.
var _ types.RunnableInterface = (*FileExists)(nil)

// Check always returns false â€” this is a read-only skill that never needs changes.
func (f *FileExists) Check() (bool, error) {
	return false, nil
}

// Run checks if the file exists and reports the result in Details.
// Changed is always false (read-only operation).
func (f *FileExists) Run() types.Result {
	path := f.GetArg(ArgPath)

	if err := validatePath(path); err != nil {
		return types.Result{
			Changed: false,
			Message: "Invalid path",
			Error:   err,
		}
	}

	cfg := f.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if file exists", "path", path)
		return types.Result{
			Changed: false,
			Message: "Would check if file exists: " + path,
			Details: map[string]string{"exists": "unknown"},
		}
	}

	exists := fileExists(cfg, path)

	return types.Result{
		Changed: false,
		Message: fmt.Sprintf("File %s exists: %v", path, exists),
		Details: map[string]string{
			"exists": fmt.Sprintf("%v", exists),
			"path":   path,
		},
	}
}

// SetPath sets the file path and returns FileExists for chaining.
func (f *FileExists) SetPath(path string) *FileExists {
	f.BaseSkill.SetArg(ArgPath, path)
	return f
}

// WithNodeConfig sets the node config and returns FileExists for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (f *FileExists) WithNodeConfig(cfg types.NodeConfig) *FileExists {
	f.BaseSkill.SetNodeConfig(cfg)
	return f
}

// SetArgs sets the arguments and returns FileExists for fluent chaining.
func (f *FileExists) SetArgs(args map[string]string) types.RunnableInterface {
	f.BaseSkill.SetArgs(args)
	return f
}

// SetArg sets a single argument and returns FileExists for fluent chaining.
func (f *FileExists) SetArg(key, value string) types.RunnableInterface {
	f.BaseSkill.SetArg(key, value)
	return f
}

// SetID sets the ID and returns FileExists for fluent chaining.
func (f *FileExists) SetID(id string) types.RunnableInterface {
	f.BaseSkill.SetID(id)
	return f
}

// SetDescription sets the description and returns FileExists for fluent chaining.
func (f *FileExists) SetDescription(description string) types.RunnableInterface {
	f.BaseSkill.SetDescription(description)
	return f
}

// SetTimeout sets the timeout and returns FileExists for fluent chaining.
func (f *FileExists) SetTimeout(timeout time.Duration) types.RunnableInterface {
	f.BaseSkill.SetTimeout(timeout)
	return f
}

// NewFileExists creates a new fs-file-exists skill.
func NewFileExists() *FileExists {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSFileExists)
	pb.SetDescription("Check if file exists")
	return &FileExists{BaseSkill: pb}
}
