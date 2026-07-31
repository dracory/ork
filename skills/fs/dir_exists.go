package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/types"
)

// DirExists checks if a directory exists on the remote server.
// This is a read-only skill â€” it never modifies the system.
//
// Usage:
//
//	result := node.Run(fs.NewDirExists().SetArg(fs.ArgPath, "/var/www")).FirstResult()
//	if result.Details["exists"] == "true" {
//	    fmt.Println("Directory exists")
//	}
//
// Args:
//   - path: Directory path to check (required, must be absolute)
//
// Result Details:
//   - exists: "true" or "false"
type DirExists struct {
	*types.BaseSkill
}

// Compile-time assertion that DirExists implements types.RunnableInterface.
var _ types.RunnableInterface = (*DirExists)(nil)

// Check always returns false â€” this is a read-only skill that never needs changes.
func (d *DirExists) Check() (bool, error) {
	return false, nil
}

// Run checks if the directory exists and reports the result in Details.
// Changed is always false (read-only operation).
func (d *DirExists) Run() types.Result {
	path := d.GetArg(ArgPath)

	if err := validatePath(path); err != nil {
		return types.Result{
			Changed: false,
			Message: "Invalid path",
			Error:   err,
		}
	}

	cfg := d.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if directory exists", "path", path)
		return types.Result{
			Changed: false,
			Message: "Would check if directory exists: " + path,
			Details: map[string]string{"exists": "unknown"},
		}
	}

	exists := dirExists(cfg, path)

	return types.Result{
		Changed: false,
		Message: fmt.Sprintf("Directory %s exists: %v", path, exists),
		Details: map[string]string{
			"exists": fmt.Sprintf("%v", exists),
			"path":   path,
		},
	}
}

// SetPath sets the directory path and returns DirExists for chaining.
func (d *DirExists) SetPath(path string) *DirExists {
	d.BaseSkill.SetArg(ArgPath, path)
	return d
}

// WithNodeConfig sets the node config and returns DirExists for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (d *DirExists) WithNodeConfig(cfg types.NodeConfig) *DirExists {
	d.BaseSkill.SetNodeConfig(cfg)
	return d
}

// SetArgs sets the arguments and returns DirExists for fluent chaining.
func (d *DirExists) SetArgs(args map[string]string) types.RunnableInterface {
	d.BaseSkill.SetArgs(args)
	return d
}

// SetArg sets a single argument and returns DirExists for fluent chaining.
func (d *DirExists) SetArg(key, value string) types.RunnableInterface {
	d.BaseSkill.SetArg(key, value)
	return d
}

// SetID sets the ID and returns DirExists for fluent chaining.
func (d *DirExists) SetID(id string) types.RunnableInterface {
	d.BaseSkill.SetID(id)
	return d
}

// SetDescription sets the description and returns DirExists for fluent chaining.
func (d *DirExists) SetDescription(description string) types.RunnableInterface {
	d.BaseSkill.SetDescription(description)
	return d
}

// SetTimeout sets the timeout and returns DirExists for fluent chaining.
func (d *DirExists) SetTimeout(timeout time.Duration) types.RunnableInterface {
	d.BaseSkill.SetTimeout(timeout)
	return d
}

// NewDirExists creates a new fs-dir-exists skill.
func NewDirExists() *DirExists {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSDirExists)
	pb.SetDescription("Check if directory exists")
	return &DirExists{BaseSkill: pb}
}
