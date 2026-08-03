package ncdu

// Package ncdu documentation is in constants.go

import (
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/skills/apt"
	"github.com/dracory/ork/skills/dpkg"
	"github.com/dracory/ork/types"
)

// Install installs ncdu (NCurses Disk Usage) via apt-get.
//
// Usage:
//
//	node.Run(ncdu.NewInstall())
//
// Execution Flow:
//  1. Checks if ncdu is already installed via `dpkg-query`
//  2. If not present, runs apt-get install with DEBIAN_FRONTEND=noninteractive
//  3. Reports success or failure
//
// Prerequisites:
//   - Root SSH access required
//   - Internet connectivity for package installation
//
// Idempotency:
//   - Check() uses `dpkg-query` to skip installation if already present
type Install struct {
	*types.BaseSkill
}

// Compile-time assertion that Install implements types.RunnableInterface.
var _ types.RunnableInterface = (*Install)(nil)

// Check determines if ncdu needs to be installed.
// Returns true if ncdu is not present (installation needed), false if already installed.
func (s *Install) Check() (bool, error) {
	cfg := s.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if ncdu is installed")
		return true, nil
	}

	result := dpkg.NewIsPkgInstalled().
		SetPackage("ncdu").
		WithNodeConfig(cfg).
		Run()

	if result.Error != nil {
		return false, result.Error
	}

	return result.Details["installed"] != "true", nil
}

// Run installs ncdu.
// Changed is true when ncdu was installed, false when it was already present.
func (s *Install) Run() types.Result {
	cfg := s.GetNodeConfig()

	// Check if ncdu needs to be installed (idempotency)
	needsInstall, err := s.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check if ncdu is installed",
			Error:   err,
		}
	}

	if !needsInstall {
		return types.Result{
			Changed: false,
			Message: "ncdu is already installed",
		}
	}

	result := apt.NewPkgInstall().
		SetPackages("ncdu").
		WithNodeConfig(cfg).
		Run()

	if result.Error != nil {
		return result
	}

	return types.Result{
		Changed: result.Changed,
		Message: result.Message,
		Details: result.Details,
	}
}

// SetArgs sets the arguments for ncdu install.
// Returns Install for fluent method chaining.
func (s *Install) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetArg sets a single argument for ncdu install.
// Returns Install for fluent method chaining.
func (s *Install) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for ncdu install.
// Returns Install for fluent method chaining.
func (s *Install) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for ncdu install.
// Returns Install for fluent method chaining.
func (s *Install) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for ncdu install.
// Returns Install for fluent method chaining.
func (s *Install) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// WithNodeConfig sets the node config and returns Install for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (s *Install) WithNodeConfig(cfg types.NodeConfig) *Install {
	s.BaseSkill.SetNodeConfig(cfg)
	return s
}

// NewInstall creates a new ncdu-install skill.
//
// Returns:
//
//	An Install skill configured with skills.IDNcduInstall identifier
//	and description "Install ncdu (NCurses Disk Usage)".
//
// Example:
//
//	node.Run(ncdu.NewInstall())
func NewInstall() *Install {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDNcduInstall)
	pb.SetDescription("Install ncdu (NCurses Disk Usage)")
	return &Install{BaseSkill: pb}
}
