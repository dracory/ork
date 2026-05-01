package types

import (
	"fmt"
	"time"
)

// BaseSkill provides default implementations of the RunnableInterface.
// Embed this in your skill struct to get boilerplate getter/setter methods.
// Only implement Check() and Run() for the specific skill logic.
//
// Example usage with fluent chaining:
//
//	type MySkill struct {
//	    *BaseSkill
//	}
//
//	func NewMySkill() *MySkill {
//	    return &MySkill{
//	        BaseSkill: types.NewBaseSkill().
//	            WithID("my-skill").
//	            WithDescription("What this skill does").
//	            WithDryRun(false),
//	    }
//	}
//
//	func (m *MySkill) Check() (bool, error) {
//	    // Check if changes are needed
//	}
//
//	func (m *MySkill) Run() Result {
//	    // Execute the skill
//	}
type BaseSkill struct {
	BaseBecome
	id          string
	description string
	nodeCfg     NodeConfig
	args        map[string]string
	dryRun      bool
	timeout     time.Duration
}

// NewBaseSkill creates a new BaseSkill with default values.
// Use the setter methods to configure it before returning from your constructor.
func NewBaseSkill() *BaseSkill {
	return &BaseSkill{
		args:   make(map[string]string),
		dryRun: false,
	}
}

// GetID returns the unique identifier for this skill.
func (b *BaseSkill) GetID() string {
	return b.id
}

// SetID sets the unique identifier for this skill.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BaseSkill) SetID(id string) RunnableInterface {
	b.id = id
	return b
}

// GetDescription returns a short description of what the skill does.
func (b *BaseSkill) GetDescription() string {
	return b.description
}

// SetDescription sets a short description of what the skill does.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BaseSkill) SetDescription(description string) RunnableInterface {
	b.description = description
	return b
}

// GetNodeConfig returns the current node configuration for this skill.
func (b *BaseSkill) GetNodeConfig() NodeConfig {
	return b.nodeCfg
}

// SetNodeConfig sets the node configuration for this skill execution.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BaseSkill) SetNodeConfig(cfg NodeConfig) RunnableInterface {
	b.nodeCfg = cfg
	return b
}

// GetArg retrieves a single argument value by key.
func (b *BaseSkill) GetArg(key string) string {
	return b.args[key]
}

// SetArg sets a single argument value.
// Returns RunnableInterface for fluent method chaining.
func (b *BaseSkill) SetArg(key, value string) RunnableInterface {
	if b.args == nil {
		b.args = make(map[string]string)
	}
	b.args[key] = value
	return b
}

// GetArgs returns the entire arguments map.
func (b *BaseSkill) GetArgs() map[string]string {
	return b.args
}

// SetArgs replaces the entire arguments map.
// Returns RunnableInterface for fluent method chaining.
func (b *BaseSkill) SetArgs(args map[string]string) RunnableInterface {
	b.args = args
	return b
}

// IsDryRun returns true if this is a dry-run execution.
func (b *BaseSkill) IsDryRun() bool {
	return b.dryRun
}

// SetDryRun sets whether to simulate execution without making changes.
// Returns RunnableInterface for fluent method chaining.
func (b *BaseSkill) SetDryRun(dryRun bool) RunnableInterface {
	b.dryRun = dryRun
	return b
}

// GetTimeout returns the maximum duration for skill execution.
func (b *BaseSkill) GetTimeout() time.Duration {
	return b.timeout
}

// SetTimeout sets the maximum duration for skill execution.
// Returns RunnableInterface for fluent method chaining.
func (b *BaseSkill) SetTimeout(timeout time.Duration) RunnableInterface {
	b.timeout = timeout
	return b
}

// WithID sets the unique identifier and returns BaseSkill for chaining.
// Shortcut alias to SetID for fluent interface convenience.
func (b *BaseSkill) WithID(id string) *BaseSkill {
	b.SetID(id)
	return b
}

// WithDescription sets a description and returns BaseSkill for chaining.
// Shortcut alias to SetDescription for fluent interface convenience.
func (b *BaseSkill) WithDescription(description string) *BaseSkill {
	b.SetDescription(description)
	return b
}

// WithNodeConfig sets the node config and returns BaseSkill for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (b *BaseSkill) WithNodeConfig(cfg NodeConfig) *BaseSkill {
	b.SetNodeConfig(cfg)
	return b
}

// WithArg sets a single argument and returns BaseSkill for chaining.
// Shortcut alias to SetArg for fluent interface convenience.
func (b *BaseSkill) WithArg(key, value string) *BaseSkill {
	b.SetArg(key, value)
	return b
}

// WithArgs replaces the arguments map and returns BaseSkill for chaining.
// Shortcut alias to SetArgs for fluent interface convenience.
func (b *BaseSkill) WithArgs(args map[string]string) *BaseSkill {
	b.SetArgs(args)
	return b
}

// WithDryRun sets dry-run mode and returns BaseSkill for chaining.
// Shortcut alias to SetDryRun for fluent interface convenience.
func (b *BaseSkill) WithDryRun(dryRun bool) *BaseSkill {
	b.SetDryRun(dryRun)
	return b
}

// WithTimeout sets the timeout and returns BaseSkill for chaining.
// Shortcut alias to SetTimeout for fluent interface convenience.
func (b *BaseSkill) WithTimeout(timeout time.Duration) *BaseSkill {
	b.SetTimeout(timeout)
	return b
}

// WithBecomeUser sets the become user and returns BaseSkill for chaining.
// Shortcut alias to SetBecomeUser for fluent interface convenience.
func (b *BaseSkill) WithBecomeUser(user string) *BaseSkill {
	b.BaseBecome.SetBecomeUser(user)
	return b
}

// Check is a stub that embedding types must override.
func (b *BaseSkill) Check() (bool, error) {
	return false, fmt.Errorf("Check() must be implemented by embedding type")
}

// Run is a stub that embedding types must override.
func (b *BaseSkill) Run() Result {
	return Result{
		Changed: false,
		Message: "Run() must be implemented by embedding type",
		Error:   fmt.Errorf("Run() must be implemented by embedding type"),
	}
}
