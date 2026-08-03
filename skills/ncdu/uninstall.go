package ncdu

// Package ncdu documentation is in constants.go

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/skills/dpkg"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Uninstall removes ncdu via apt-get purge.
//
// Usage:
//
//	node.Run(ncdu.NewUninstall())
//
// Execution Flow:
//  1. Checks if ncdu is installed via `dpkg-query`
//  2. If present, runs apt-get purge with DEBIAN_FRONTEND=noninteractive
//  3. Reports success or failure
//
// Prerequisites:
//   - Root SSH access required
//
// Idempotency:
//   - Check() uses `dpkg-query` to skip uninstall if ncdu is not present
type Uninstall struct {
	*types.BaseSkill
}

// Compile-time assertion that Uninstall implements types.RunnableInterface.
var _ types.RunnableInterface = (*Uninstall)(nil)

// Check determines if ncdu needs to be uninstalled.
// Returns true if ncdu is currently installed, false if it is not present.
func (s *Uninstall) Check() (bool, error) {
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

	return result.Details["installed"] == "true", nil
}

// Run uninstalls ncdu.
// Changed is true when ncdu was removed, false when it was not installed.
func (s *Uninstall) Run() types.Result {
	cfg := s.GetNodeConfig()

	// Check if ncdu is installed (idempotency)
	needsUninstall, err := s.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check if ncdu is installed",
			Error:   err,
		}
	}

	if !needsUninstall {
		return types.Result{
			Changed: false,
			Message: "ncdu is not installed",
		}
	}

	// See skills.DebianNonInteractive and skills.DpkgConfOptions for details
	cmdPurgeStr := ""
	cmdPurgeStr += skills.DebianNonInteractive // prevent interactive prompts
	cmdPurgeStr += " apt-get purge -y"         // purge ncdu, auto-confirm
	cmdPurgeStr += skills.DpkgConfOptions      // keep local config, use maintainer default if unmodified
	cmdPurgeStr += " -- ncdu"                  // -- prevents option injection

	cmdPurge := types.Command{
		Command:     cmdPurgeStr,
		Description: "Purge ncdu",
		Required:    true,
	}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdPurge.Command)
		return types.Result{
			Changed: true,
			Message: "Would uninstall ncdu",
		}
	}

	cfg.GetLoggerOrDefault().Info("uninstalling ncdu")
	output, err := ssh.Run(cfg, cmdPurge)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "ncdu uninstall failed",
			Error:   fmt.Errorf("apt-get purge failed for ncdu: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("ncdu uninstalled")
	return types.Result{
		Changed: true,
		Message: "ncdu uninstalled",
		Details: map[string]string{
			"output": output,
		},
	}
}

// SetArgs sets the arguments for ncdu uninstall.
// Returns Uninstall for fluent method chaining.
func (s *Uninstall) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetArg sets a single argument for ncdu uninstall.
// Returns Uninstall for fluent method chaining.
func (s *Uninstall) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for ncdu uninstall.
// Returns Uninstall for fluent method chaining.
func (s *Uninstall) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for ncdu uninstall.
// Returns Uninstall for fluent method chaining.
func (s *Uninstall) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for ncdu uninstall.
// Returns Uninstall for fluent method chaining.
func (s *Uninstall) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// WithNodeConfig sets the node config and returns Uninstall for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (s *Uninstall) WithNodeConfig(cfg types.NodeConfig) *Uninstall {
	s.BaseSkill.SetNodeConfig(cfg)
	return s
}

// NewUninstall creates a new ncdu-uninstall skill.
//
// Returns:
//
//	An Uninstall skill configured with skills.IDNcduUninstall identifier
//	and description "Uninstall ncdu (NCurses Disk Usage)".
//
// Example:
//
//	node.Run(ncdu.NewUninstall())
func NewUninstall() *Uninstall {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDNcduUninstall)
	pb.SetDescription("Uninstall ncdu (NCurses Disk Usage)")
	return &Uninstall{BaseSkill: pb}
}
