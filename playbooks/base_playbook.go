package playbooks

import (
	"fmt"
	"time"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/types"
)

// BasePlaybook provides default implementations of the types.PlaybookInterface.
// Embed this in your playbook struct to get boilerplate getter/setter methods.
// Only implement Check() and Run() for the specific playbook logic.
//
// Example usage:
//
//	type MyPlaybook struct {
//	    *BasePlaybook
//	}
//
//	func NewMyPlaybook() *MyPlaybook {
//	    return &MyPlaybook{
//	        BasePlaybook: playbooks.NewBasePlaybook().
//	            SetID("my-playbook").
//	            SetDescription("What this playbook does"),
//	    }
//	}
//
//	func (m *MyPlaybook) Check() (bool, error) {
//	    // Check if changes are needed
//	}
//
//	func (m *MyPlaybook) Run() types.Result {
//	    // Execute the playbook
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
// Returns types.PlaybookInterface for fluent method chaining with embedding types.
func (b *BasePlaybook) SetID(id string) types.PlaybookInterface {
	b.id = id
	return b
}

// GetDescription returns a short description of what the playbook does.
func (b *BasePlaybook) GetDescription() string {
	return b.description
}

// SetDescription sets a short description of what the playbook does.
// Returns types.PlaybookInterface for fluent method chaining with embedding types.
func (b *BasePlaybook) SetDescription(description string) types.PlaybookInterface {
	b.description = description
	return b
}

// GetNodeConfig returns the current node configuration for this playbook.
func (b *BasePlaybook) GetNodeConfig() config.NodeConfig {
	return b.nodeCfg
}

// SetNodeConfig sets the node configuration for this playbook execution.
// Returns types.PlaybookInterface for fluent method chaining with embedding types.
func (b *BasePlaybook) SetNodeConfig(cfg config.NodeConfig) types.PlaybookInterface {
	b.nodeCfg = cfg
	return b
}

// GetArg retrieves a single argument value by key.
func (b *BasePlaybook) GetArg(key string) string {
	return b.args[key]
}

// SetArg sets a single argument value.
// Returns types.PlaybookInterface for fluent method chaining.
func (b *BasePlaybook) SetArg(key, value string) types.PlaybookInterface {
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
// Returns types.PlaybookInterface for fluent method chaining.
func (b *BasePlaybook) SetArgs(args map[string]string) types.PlaybookInterface {
	b.args = args
	return b
}

// IsDryRun returns true if this is a dry-run execution.
func (b *BasePlaybook) IsDryRun() bool {
	return b.dryRun
}

// SetDryRun sets whether to simulate execution without making changes.
// Returns types.PlaybookInterface for fluent method chaining.
func (b *BasePlaybook) SetDryRun(dryRun bool) types.PlaybookInterface {
	b.dryRun = dryRun
	return b
}

// GetTimeout returns the maximum duration for playbook execution.
func (b *BasePlaybook) GetTimeout() time.Duration {
	return b.timeout
}

// SetTimeout sets the maximum duration for playbook execution.
// Returns types.PlaybookInterface for fluent method chaining.
func (b *BasePlaybook) SetTimeout(timeout time.Duration) types.PlaybookInterface {
	b.timeout = timeout
	return b
}

// Check is a stub that embedding types must override.
func (b *BasePlaybook) Check() (bool, error) {
	return false, fmt.Errorf("Check() must be implemented by embedding type")
}

// Run is a stub that embedding types must override.
func (b *BasePlaybook) Run() types.Result {
	return types.Result{
		Changed: false,
		Message: "Run() must be implemented by embedding type",
		Error:   fmt.Errorf("Run() must be implemented by embedding type"),
	}
}
