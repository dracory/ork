// Package ork provides a framework for remote server automation.
package ork

import (
	"log/slog"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/types"
)

// RunnableInterface defines operations that can be performed on either
// a single Node or an Inventory of nodes.
type RunnableInterface interface {
	// RunCommand executes a shell command and returns the output.
	// For Inventory, runs concurrently across all nodes.
	RunCommand(cmd string) types.Results

	// RunPlaybook executes a playbook instance.
	// For Inventory, runs concurrently across all nodes.
	RunPlaybook(pb playbook.PlaybookInterface) types.Results

	// RunPlaybookByID executes a playbook by ID from the registry.
	// Deprecated: Use RunPlaybook() instead.
	RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) types.Results

	// CheckPlaybook runs the playbook's check mode to determine if changes would be made.
	// Sets Changed=true on result if changes are needed.
	CheckPlaybook(pb playbook.PlaybookInterface) types.Results

	// GetLogger returns the logger. Returns slog.Default() if not set.
	GetLogger() *slog.Logger

	// SetLogger sets a custom logger. Returns self for chaining.
	SetLogger(logger *slog.Logger) RunnableInterface

	// SetDryRunMode sets whether to simulate execution without making changes.
	// When true, ssh.Run() will log commands and return "[dry-run]" marker instead of executing.
	// Returns self for chaining.
	SetDryRunMode(dryRun bool) RunnableInterface

	// GetDryRunMode returns true if dry-run mode is enabled.
	// When true, commands are logged but not executed on the server.
	GetDryRunMode() bool
}
