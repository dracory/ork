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

	// Run executes any runnable (command or skill).
	// For Inventory, runs concurrently across all nodes.
	Run(runnable types.RunnableInterface) types.Results

	// RunByID executes a skill by ID from the registry.
	// Deprecated: Use Run() instead.
	RunByID(id string, opts ...types.RunnableOptions) types.Results

	// Check runs the runnable's check mode to determine if changes would be made.
	// Sets Changed=true on result if changes are needed.
	Check(runnable types.RunnableInterface) types.Results

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
