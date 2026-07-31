package php

// Package php documentation is in constants.go

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// UninstallComposer removes the Composer binary.
//
// It checks if the Composer binary exists and removes it from /usr/local/bin/composer.
//
// No args needed.
//
// Execution Flow:
//  1. Checks if Composer binary exists
//  2. Removes /usr/local/bin/composer
//
// Usage:
//
//	node.Run(php.NewUninstallComposer())
//
// Idempotency:
//   - Check() verifies that /usr/local/bin/composer does not exist. When it
//     doesn't, Run() is a no-op.
type UninstallComposer struct {
	*types.BaseSkill
}

// Compile-time assertion that UninstallComposer implements types.RunnableInterface.
var _ types.RunnableInterface = (*UninstallComposer)(nil)

// Check determines if Composer needs to be uninstalled.
// Returns true if /usr/local/bin/composer exists, false if it does not.
func (s *UninstallComposer) Check() (bool, error) {
	cfg := s.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if Composer is installed")
		return true, nil
	}

	// Check if Composer binary exists.
	cmdCheck := types.Command{
		Command:     fmt.Sprintf("test -f %s", skills.ShellEscapeArg(ComposerBinaryPath)),
		Description: "Check if Composer is installed",
		Required:    false,
	}
	_, err := ssh.Run(cfg, cmdCheck)
	if err != nil {
		// Binary missing — nothing to uninstall.
		return false, nil
	}
	return true, nil // binary exists, need to uninstall
}

// Run uninstalls Composer by removing the binary.
// Changed is true when Composer was removed, false when it was not present.
//
// Result.Details contains:
//   - binary: The path of the removed Composer binary
func (s *UninstallComposer) Run() types.Result {
	cfg := s.GetNodeConfig()

	// Check if Composer needs to be uninstalled (idempotency).
	needsChange, err := s.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check Composer installation",
			Error:   err,
		}
	}

	if !needsChange {
		return types.Result{
			Changed: false,
			Message: "Composer is not installed",
		}
	}

	// Check for dry-run mode.
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would uninstall Composer")
		return types.Result{
			Changed: true,
			Message: "Would uninstall Composer",
		}
	}

	// Remove the Composer binary.
	cmdRemove := types.Command{
		Command:     fmt.Sprintf("rm -f %s", skills.ShellEscapeArg(ComposerBinaryPath)),
		Description: "Remove Composer binary",
	}
	if _, err := ssh.Run(cfg, cmdRemove); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to remove Composer",
			Error:   fmt.Errorf("failed to remove %s: %w", ComposerBinaryPath, err),
		}
	}

	cfg.GetLoggerOrDefault().Info("Composer uninstalled")
	return types.Result{
		Changed: true,
		Message: "Composer uninstalled",
		Details: map[string]string{
			"binary": ComposerBinaryPath,
		},
	}
}

// SetArgs sets the arguments for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *UninstallComposer) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// WithNodeConfig sets the node config and returns UninstallComposer for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (s *UninstallComposer) WithNodeConfig(cfg types.NodeConfig) *UninstallComposer {
	s.BaseSkill.SetNodeConfig(cfg)
	return s
}

// SetArg sets a single argument for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *UninstallComposer) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *UninstallComposer) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *UninstallComposer) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *UninstallComposer) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewUninstallComposer creates a new php-uninstall-composer skill.
//
// Returns:
//
//	An UninstallComposer skill configured with skills.IDPhpUninstallComposer identifier
//	and description "Uninstall Composer (remove binary)".
//
// Example:
//
//	node.Run(php.NewUninstallComposer())
func NewUninstallComposer() *UninstallComposer {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDPhpUninstallComposer)
	pb.SetDescription("Uninstall Composer (remove binary)")
	return &UninstallComposer{BaseSkill: pb}
}
