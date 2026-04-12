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

// PlaybookInterface is the interface that all playbooks must implement.
// A playbook is a self-contained automation task that runs on a remote server.
// All playbooks support idempotency through the Check() method and Result return value.
//
// Usage:
//
//	pb := playbooks.NewUserCreate().
//	    SetConfig(cfg).
//	    SetOptions(&playbook.PlaybookOptions{Args: map[string]string{"username": "alice"}})
//
//	needsRun, _ := pb.Check()
//	result := pb.Run()
type PlaybookInterface interface {
	// GetID returns the unique identifier for this playbook (e.g., "apt-update")
	GetID() string

	// SetID sets the unique identifier for this playbook.
	SetID(id string) PlaybookInterface

	// GetDescription returns a short description of what the playbook does
	GetDescription() string

	// SetDescription sets a short description of what the playbook does.
	SetDescription(description string) PlaybookInterface

	// GetConfig returns the current node configuration for this playbook.
	GetConfig() config.Config

	// SetConfig sets the node configuration for this playbook execution.
	// Returns the PlaybookInterface for fluent method chaining.
	SetConfig(cfg config.Config) PlaybookInterface

	// GetArg retrieves a single argument value by key.
	GetArg(key string) string

	// SetArg sets a single argument value.
	// Returns the PlaybookInterface for fluent method chaining.
	SetArg(key, value string) PlaybookInterface

	// GetArgs returns the entire arguments map.
	GetArgs() map[string]string

	// SetArgs replaces the entire arguments map.
	// Returns the PlaybookInterface for fluent method chaining.
	SetArgs(args map[string]string) PlaybookInterface

	// IsDryRun returns true if this is a dry-run execution.
	IsDryRun() bool

	// SetDryRun sets whether to simulate execution without making changes.
	// Returns the PlaybookInterface for fluent method chaining.
	SetDryRun(dryRun bool) PlaybookInterface

	// GetTimeout returns the maximum duration for playbook execution.
	GetTimeout() time.Duration

	// SetTimeout sets the maximum duration for playbook execution.
	// Returns the PlaybookInterface for fluent method chaining.
	SetTimeout(timeout time.Duration) PlaybookInterface

	// Check determines if the playbook needs to make any changes.
	// Uses the config and options set via SetConfig/SetOptions.
	// Returns true if changes are needed, false if the system is already in the desired state.
	// Returns an error if the check itself fails.
	Check() (bool, error)

	// Run executes the playbook and returns a detailed result.
	// Uses the config and options set via SetConfig/SetOptions.
	// The Result.Changed field indicates whether any changes were made.
	Run() Result
}
