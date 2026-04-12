// Package playbook provides the base types and interfaces for creating
// automation playbooks using SSH-based remote execution.
package playbook

import (
	"time"

	"github.com/dracory/ork/config"
)

// PlaybookOptions provides configuration options for playbook execution.
// This allows per-playbook variable scoping and additional execution controls.
type PlaybookOptions struct {
	// Args contains playbook-specific variables that override node-level arguments.
	// These are merged with node-level args (playbook args take precedence).
	Args map[string]string

	// DryRun indicates whether to simulate execution without making changes.
	// When true, the playbook should preview what would be done.
	DryRun bool

	// Timeout specifies the maximum duration for playbook execution.
	// Zero means no timeout.
	Timeout time.Duration
}

// Result represents the outcome of a playbook execution.
// It indicates whether any changes were made and provides details about the execution.
type Result struct {
	// Changed indicates whether the playbook made any changes to the system.
	// false means the system was already in the desired state.
	// true means the playbook modified the system.
	Changed bool

	// Message is a human-readable description of what happened.
	Message string

	// Details contains additional information about the execution.
	// Keys are field names, values are string representations.
	Details map[string]string

	// Error is non-nil if the playbook failed to execute.
	// When Error is non-nil, Changed may be true if some changes occurred before the failure.
	Error error
}

// Playbook is the interface that all playbooks must implement.
// A playbook is a self-contained automation task that runs on a remote server.
// All playbooks support idempotency through the Check() method and Result return value.
type Playbook interface {
	// Name returns the unique identifier for this playbook (e.g., "apt-update")
	Name() string

	// Description returns a short description of what the playbook does
	Description() string

	// Check determines if the playbook needs to make any changes.
	// Returns true if changes are needed, false if the system is already in the desired state.
	// Returns an error if the check itself fails.
	Check(config config.Config) (bool, error)

	// Run executes the playbook and returns a detailed result.
	// The Result.Changed field indicates whether any changes were made.
	Run(config config.Config) Result
}
