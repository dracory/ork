package types

import (
	"fmt"
	"time"

	"github.com/dracory/ork/config"
)

// BasePlaybook provides a foundation for playbook development.
// Embed this in your playbook struct to get boilerplate getter/setter methods.
// Only implement Run() for the specific playbook logic (Check() is optional).
//
// Example usage:
//
//	type MyPlaybook struct {
//	    *BasePlaybook
//	}
//
//	func NewMyPlaybook() *MyPlaybook {
//	    return &MyPlaybook{
//	        BasePlaybook: types.NewBasePlaybook().
//	            SetID("my-playbook").
//	            SetDescription("What this playbook does"),
//	    }
//	}
//
//	func (m *MyPlaybook) Run() Result {
//	    // Execute the playbook with complex orchestration logic
//	}
type BasePlaybook struct {
	id          string
	description string
	nodeCfg     config.NodeConfig
	args        map[string]string
	dryRun      bool
	timeout     time.Duration
}

// NewBasePlaybook creates a new BasePlaybook with default values.
// Use the setter methods to configure it before returning from your constructor.
func NewBasePlaybook() *BasePlaybook {
	return &BasePlaybook{
		args:   make(map[string]string),
		dryRun: false,
	}
}

// GetID returns the unique identifier for this playbook.
func (b *BasePlaybook) GetID() string {
	return b.id
}

// SetID sets the unique identifier for this playbook.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BasePlaybook) SetID(id string) RunnableInterface {
	b.id = id
	return b
}

// GetDescription returns a short description of what the playbook does.
func (b *BasePlaybook) GetDescription() string {
	return b.description
}

// SetDescription sets a short description of what the playbook does.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BasePlaybook) SetDescription(description string) RunnableInterface {
	b.description = description
	return b
}

// GetNodeConfig returns the current node configuration for this playbook.
func (b *BasePlaybook) GetNodeConfig() config.NodeConfig {
	return b.nodeCfg
}

// SetNodeConfig sets the node configuration for this playbook execution.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BasePlaybook) SetNodeConfig(cfg config.NodeConfig) RunnableInterface {
	b.nodeCfg = cfg
	return b
}

// GetArg retrieves a single argument value by key.
func (b *BasePlaybook) GetArg(key string) string {
	return b.args[key]
}

// SetArg sets a single argument value.
// Returns RunnableInterface for fluent method chaining.
func (b *BasePlaybook) SetArg(key, value string) RunnableInterface {
	if b.args == nil {
		b.args = make(map[string]string)
	}
	b.args[key] = value
	return b
}

// GetArgs returns the entire arguments map.
func (b *BasePlaybook) GetArgs() map[string]string {
	return b.args
}

// SetArgs replaces the entire arguments map.
// Returns RunnableInterface for fluent method chaining.
func (b *BasePlaybook) SetArgs(args map[string]string) RunnableInterface {
	b.args = args
	return b
}

// IsDryRun returns true if this is a dry-run execution.
func (b *BasePlaybook) IsDryRun() bool {
	return b.dryRun
}

// SetDryRun sets whether to simulate execution without making changes.
// Returns RunnableInterface for fluent method chaining.
func (b *BasePlaybook) SetDryRun(dryRun bool) RunnableInterface {
	b.dryRun = dryRun
	return b
}

// GetTimeout returns the maximum duration for playbook execution.
func (b *BasePlaybook) GetTimeout() time.Duration {
	return b.timeout
}

// SetTimeout sets the maximum duration for playbook execution.
// Returns RunnableInterface for fluent method chaining.
func (b *BasePlaybook) SetTimeout(timeout time.Duration) RunnableInterface {
	b.timeout = timeout
	return b
}

// Check returns false (no changes needed) by default.
// Playbooks can override this to provide idempotency checks, but it's optional
// since playbooks often have complex logic that makes checking difficult.
func (b *BasePlaybook) Check() (bool, error) {
	return false, nil
}

// Run must be overridden by playbook implementations.
// Playbooks implement complex orchestration logic here.
func (b *BasePlaybook) Run() Result {
	return Result{
		Changed: false,
		Message: "Run() must be implemented by playbook",
		Error:   fmt.Errorf("Run() must be implemented by playbook"),
	}
}
