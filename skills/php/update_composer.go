package php

// Package php documentation is in constants.go

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// UpdateComposer updates Composer to the latest version.
//
// It runs `composer self-update` to upgrade Composer in place. If Composer is
// not installed, Run() returns an error.
//
// No args needed.
//
// Execution Flow:
//  1. Checks if Composer is installed (returns error if not)
//  2. Runs `composer self-update`
//
// Usage:
//
//	node.Run(php.NewUpdateComposer())
//
// Idempotency:
//   - Check() always returns true because we always want to check for updates.
//     The cost of checking if an update is needed is similar to just running it.
type UpdateComposer struct {
	*types.BaseSkill
}

// Compile-time assertion that UpdateComposer implements types.RunnableInterface.
var _ types.RunnableInterface = (*UpdateComposer)(nil)

// Check always returns true for update-composer since checking for updates is
// always beneficial. Per the skill interface convention, the bool return
// indicates whether the operation would modify system state. Since we always
// want to check for updates, this always returns true.
//
// Note: The cost of checking if an update is needed is similar to just running
// it, so we skip the check and always execute.
func (s *UpdateComposer) Check() (bool, error) {
	cfg := s.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if Composer update is needed")
		return true, nil
	}

	return true, nil // Always run composer self-update
}

// Run updates Composer to the latest version.
// Changed is true when the self-update command was executed, false when
// Composer is not installed.
//
// Result.Details contains:
//   - output: Full output from the composer self-update command
func (s *UpdateComposer) Run() types.Result {
	cfg := s.GetNodeConfig()

	// Check if Composer binary exists, return error if not.
	cmdCheck := types.Command{
		Command:     fmt.Sprintf("test -f %s", skills.ShellEscapeArg(ComposerBinaryPath)),
		Description: "Check if Composer is installed",
		Required:    false,
	}
	_, err := ssh.Run(cfg, cmdCheck)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Composer is not installed",
			Error:   fmt.Errorf("composer is not installed at %s: %w", ComposerBinaryPath, err),
		}
	}

	// Check for dry-run mode.
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would update Composer")
		return types.Result{
			Changed: true,
			Message: "Would update Composer",
		}
	}

	// Run composer self-update.
	cmdUpdate := types.Command{
		Command:     "composer self-update",
		Description: "Update Composer to the latest version",
	}
	cfg.GetLoggerOrDefault().Info("running composer self-update")
	output, err := ssh.Run(cfg, cmdUpdate)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Composer self-update failed",
			Error:   fmt.Errorf("composer self-update failed: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("Composer updated")
	return types.Result{
		Changed: true,
		Message: "Composer updated",
		Details: map[string]string{
			"output": output,
		},
	}
}

// SetArgs sets the arguments for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *UpdateComposer) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// WithNodeConfig sets the node config and returns UpdateComposer for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (s *UpdateComposer) WithNodeConfig(cfg types.NodeConfig) *UpdateComposer {
	s.BaseSkill.SetNodeConfig(cfg)
	return s
}

// SetArg sets a single argument for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *UpdateComposer) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *UpdateComposer) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *UpdateComposer) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *UpdateComposer) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewUpdateComposer creates a new php-update-composer skill.
//
// Returns:
//
//	An UpdateComposer skill configured with skills.IDPhpUpdateComposer identifier
//	and description "Update Composer to the latest version".
//
// Example:
//
//	node.Run(php.NewUpdateComposer())
func NewUpdateComposer() *UpdateComposer {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDPhpUpdateComposer)
	pb.SetDescription("Update Composer to the latest version")
	return &UpdateComposer{BaseSkill: pb}
}
