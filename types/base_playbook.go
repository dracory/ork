package types

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dracory/omni"
)

// BasePlaybook provides a foundation for playbook development.
// Embed this in your playbook struct to get boilerplate getter/setter methods.
// Only implement Run() for the specific playbook logic (Check() is optional).
//
// State is stored in an omni.Atom (thread-safe via sync.RWMutex) for simple
// properties. NodeConfig is stored as a separate typed field because it
// contains a *slog.Logger pointer that cannot be serialized.
//
// The framework clones the playbook via ToMap()/FromMap() before mutating it
// for concurrency safety.
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
//	            WithID("my-playbook").
//	            WithDescription("What this playbook does"),
//	    }
//	}
//
//	func (m *MyPlaybook) Run() Result {
//	    // Execute the playbook with complex orchestration logic
//	}
type BasePlaybook struct {
	atom    omni.AtomInterface
	nodeCfg NodeConfig
	mu      sync.RWMutex // protects nodeCfg
}

// NewBasePlaybook creates a new BasePlaybook with default values.
// Use the setter methods to configure it before returning from your constructor.
func NewBasePlaybook() *BasePlaybook {
	return &BasePlaybook{
		atom: omni.NewAtom(atomTypePlaybook),
	}
}

// GetID returns the unique identifier for this playbook.
func (b *BasePlaybook) GetID() string {
	return b.atom.Get(propID)
}

// SetID sets the unique identifier for this playbook.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BasePlaybook) SetID(id string) RunnableInterface {
	b.atom.Set(propID, id)
	return b
}

// GetDescription returns a short description of what the playbook does.
func (b *BasePlaybook) GetDescription() string {
	return b.atom.Get(propDescription)
}

// SetDescription sets a short description of what the playbook does.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BasePlaybook) SetDescription(description string) RunnableInterface {
	b.atom.Set(propDescription, description)
	return b
}

// GetNodeConfig returns the current node configuration for this playbook.
func (b *BasePlaybook) GetNodeConfig() NodeConfig {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.nodeCfg
}

// SetNodeConfig sets the node configuration for this playbook execution.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BasePlaybook) SetNodeConfig(cfg NodeConfig) RunnableInterface {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nodeCfg = cfg
	return b
}

// GetArg retrieves a single argument value by key.
func (b *BasePlaybook) GetArg(key string) string {
	return b.atom.Get(argPrefix + key)
}

// SetArg sets a single argument value.
// Returns RunnableInterface for fluent method chaining.
func (b *BasePlaybook) SetArg(key, value string) RunnableInterface {
	b.atom.Set(argPrefix+key, value)
	return b
}

// GetArgs returns the entire arguments map.
func (b *BasePlaybook) GetArgs() map[string]string {
	all := b.atom.GetAll()
	args := make(map[string]string)
	for k, v := range all {
		if len(k) > len(argPrefix) && k[:len(argPrefix)] == argPrefix {
			args[k[len(argPrefix):]] = v
		}
	}
	return args
}

// SetArgs replaces the entire arguments map.
// Returns RunnableInterface for fluent method chaining.
func (b *BasePlaybook) SetArgs(args map[string]string) RunnableInterface {
	// Remove existing args
	all := b.atom.GetAll()
	for k := range all {
		if len(k) > len(argPrefix) && k[:len(argPrefix)] == argPrefix {
			b.atom.Remove(k)
		}
	}
	// Set new args
	for k, v := range args {
		b.atom.Set(argPrefix+k, v)
	}
	return b
}

// IsDryRun returns true if this is a dry-run execution.
func (b *BasePlaybook) IsDryRun() bool {
	return b.atom.Get(propDryRun) == boolTrue
}

// SetDryRun sets whether to simulate execution without making changes.
// Returns RunnableInterface for fluent method chaining.
func (b *BasePlaybook) SetDryRun(dryRun bool) RunnableInterface {
	b.atom.Set(propDryRun, strconv.FormatBool(dryRun))
	return b
}

// GetTimeout returns the maximum duration for playbook execution.
func (b *BasePlaybook) GetTimeout() time.Duration {
	s := b.atom.Get(propTimeout)
	if s == "" {
		return 0
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return time.Duration(n)
}

// SetTimeout sets the maximum duration for playbook execution.
// Returns RunnableInterface for fluent method chaining.
func (b *BasePlaybook) SetTimeout(timeout time.Duration) RunnableInterface {
	b.atom.Set(propTimeout, strconv.FormatInt(int64(timeout), 10))
	return b
}

// GetBecomeUser returns the configured become user.
// Returns empty string if not set.
func (b *BasePlaybook) GetBecomeUser() string {
	return b.atom.Get(propBecomeUser)
}

// SetBecomeUser sets the user to become when executing commands.
// Returns BecomeInterface for fluent method chaining.
func (b *BasePlaybook) SetBecomeUser(user string) BecomeInterface {
	b.atom.Set(propBecomeUser, user)
	return b
}

// ToMap returns the playbook's state as a map for cloning.
// The map contains the Atom's properties plus the NodeConfig as a struct value.
// This is used by the framework to clone the playbook before mutation.
//
// Note: atom.ToMap() strips "id" from properties and puts the Atom's auto-generated
// id field in the result. We override m["id"] with the actual ID stored as a property
// so that FromMap can restore it correctly.
func (b *BasePlaybook) ToMap() map[string]any {
	m := b.atom.ToMap()
	m[propID] = b.atom.Get(propID)
	b.mu.RLock()
	m[mapKeyNodeConfig] = b.nodeCfg
	b.mu.RUnlock()
	return m
}

// FromMap populates the playbook's state from a map (inverse of ToMap).
// Creates a fresh Atom and extracts the NodeConfig.
func (b *BasePlaybook) FromMap(m map[string]any) {
	atomType, _ := m[mapKeyType].(string)
	if atomType == "" {
		atomType = atomTypePlaybook
	}
	b.atom = omni.NewAtom(atomType)

	// Copy properties (ToMap returns properties as map[string]string)
	if props, ok := m[mapKeyProperties].(map[string]string); ok {
		b.atom.SetAll(props)
	}

	// Restore the ID — atom.ToMap() strips "id" from properties, so we
	// set it from the top-level m["id"] key (which ToMap overrode with
	// the actual ID value).
	if id, ok := m[propID].(string); ok {
		b.atom.Set(propID, id)
	}

	// Extract NodeConfig
	if cfg, ok := m[mapKeyNodeConfig].(NodeConfig); ok {
		b.mu.Lock()
		b.nodeCfg = cfg
		b.mu.Unlock()
	}
}

// WithID sets the unique identifier and returns BasePlaybook for chaining.
// Shortcut alias to SetID for fluent interface convenience.
func (b *BasePlaybook) WithID(id string) *BasePlaybook {
	b.SetID(id)
	return b
}

// WithDescription sets a description and returns BasePlaybook for chaining.
// Shortcut alias to SetDescription for fluent interface convenience.
func (b *BasePlaybook) WithDescription(description string) *BasePlaybook {
	b.SetDescription(description)
	return b
}

// WithNodeConfig sets the node config and returns BasePlaybook for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (b *BasePlaybook) WithNodeConfig(cfg NodeConfig) *BasePlaybook {
	b.SetNodeConfig(cfg)
	return b
}

// WithArg sets a single argument and returns BasePlaybook for chaining.
// Shortcut alias to SetArg for fluent interface convenience.
func (b *BasePlaybook) WithArg(key, value string) *BasePlaybook {
	b.SetArg(key, value)
	return b
}

// WithArgs replaces the arguments map and returns BasePlaybook for chaining.
// Shortcut alias to SetArgs for fluent interface convenience.
func (b *BasePlaybook) WithArgs(args map[string]string) *BasePlaybook {
	b.SetArgs(args)
	return b
}

// WithDryRun sets dry-run mode and returns BasePlaybook for chaining.
// Shortcut alias to SetDryRun for fluent interface convenience.
func (b *BasePlaybook) WithDryRun(dryRun bool) *BasePlaybook {
	b.SetDryRun(dryRun)
	return b
}

// WithTimeout sets the timeout and returns BasePlaybook for chaining.
// Shortcut alias to SetTimeout for fluent interface convenience.
func (b *BasePlaybook) WithTimeout(timeout time.Duration) *BasePlaybook {
	b.SetTimeout(timeout)
	return b
}

// WithBecomeUser sets the become user and returns BasePlaybook for chaining.
// Shortcut alias to SetBecomeUser for fluent interface convenience.
func (b *BasePlaybook) WithBecomeUser(user string) *BasePlaybook {
	b.SetBecomeUser(user)
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
