package types

import (
	"time"
)

// RunnableOptions provides configuration options for runnable execution.
// This allows per-runnable variable scoping and additional execution controls.
type RunnableOptions struct {
	// Args contains runnable-specific variables that override node-level arguments.
	// These are merged with node-level args (runnable args take precedence).
	Args map[string]string

	// DryRun indicates whether to simulate execution without making changes.
	// When true, the runnable should preview what would be done.
	DryRun bool

	// Timeout specifies the maximum duration for runnable execution.
	// Zero means no timeout.
	Timeout time.Duration
}

// RunnableInterface defines anything that can be executed on a remote server.
// Both commands and skills implement this interface, allowing a single Run() method
// to handle any executable operation.
//
// Usage:
//
//	skill := skills.NewUserCreate().
//	    SetNodeConfig(cfg).
//	    SetOptions(&types.RunnableOptions{Args: map[string]string{"username": "alice"}})
//
//	needsRun, _ := skill.Check()
//	result := skill.Run()
type RunnableInterface interface {
	// BecomeInterface provides privilege escalation capabilities.
	BecomeInterface
	// GetID returns the unique identifier for this skill (e.g., "apt-update")
	GetID() string

	// SetID sets the unique identifier for this skill.
	SetID(id string) RunnableInterface

	// GetDescription returns a short description of what the skill does
	GetDescription() string

	// SetDescription sets a short description of what the skill does.
	SetDescription(description string) RunnableInterface

	// GetNodeConfig returns the current node configuration for this skill.
	GetNodeConfig() NodeConfig

	// SetNodeConfig sets the node configuration for this skill execution.
	// Returns the RunnableInterface for fluent method chaining.
	SetNodeConfig(cfg NodeConfig) RunnableInterface

	// GetArg retrieves a single argument value by key.
	GetArg(key string) string

	// SetArg sets a single argument value.
	// Returns the RunnableInterface for fluent method chaining.
	SetArg(key, value string) RunnableInterface

	// GetArgs returns the entire arguments map.
	GetArgs() map[string]string

	// SetArgs replaces the entire arguments map.
	// Returns the RunnableInterface for fluent method chaining.
	SetArgs(args map[string]string) RunnableInterface

	// IsDryRun returns true if this is a dry-run execution.
	IsDryRun() bool

	// SetDryRun sets whether to simulate execution without making changes.
	// Returns the RunnableInterface for fluent method chaining.
	SetDryRun(dryRun bool) RunnableInterface

	// GetTimeout returns the maximum duration for skill execution.
	GetTimeout() time.Duration

	// SetTimeout sets the maximum duration for skill execution.
	// Returns the RunnableInterface for fluent method chaining.
	SetTimeout(timeout time.Duration) RunnableInterface

	// Check determines if the skill needs to make any changes.
	// Uses the config and options set via SetConfig/SetOptions.
	// Returns true if changes are needed, false if the system is already in the desired state.
	// Returns an error if the check itself fails.
	Check() (bool, error)

	// Run executes the skill and returns a detailed result.
	// Uses the config and options set via SetConfig/SetOptions.
	// The Result.Changed field indicates whether any changes were made.
	Run() Result
}
