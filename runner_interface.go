// Package ork provides a framework for remote server automation.
package ork

import (
	"log/slog"

	"github.com/dracory/ork/types"
)

// RunnerInterface defines operations that can be performed on either
// a single Node or an Inventory of nodes.
type RunnerInterface interface {
	// RunCommand executes a shell command and returns the output.
	// For Inventory, runs concurrently across all nodes.
	RunCommand(cmd string) types.Results

	// RunSkill executes a skill instance.
	// For Inventory, runs concurrently across all nodes.
	RunSkill(skill types.RunnableInterface) types.Results

	// RunSkillByID executes a skill by ID from the registry.
	// Deprecated: Use RunSkill() instead.
	RunSkillByID(id string, opts ...types.SkillOptions) types.Results

	// CheckSkill runs the skill's check mode to determine if changes would be made.
	// Sets Changed=true on result if changes are needed.
	CheckSkill(skill types.RunnableInterface) types.Results

	// GetLogger returns the logger. Returns slog.Default() if not set.
	GetLogger() *slog.Logger

	// SetLogger sets a custom logger. Returns self for chaining.
	SetLogger(logger *slog.Logger) RunnerInterface

	// SetDryRunMode sets whether to simulate execution without making changes.
	// When true, ssh.Run() will log commands and return "[dry-run]" marker instead of executing.
	// Returns self for chaining.
	SetDryRunMode(dryRun bool) RunnerInterface

	// GetDryRunMode returns true if dry-run mode is enabled.
	// When true, commands are logged but not executed on the server.
	GetDryRunMode() bool
}
