package types

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dracory/omni"
)

// BaseSkill provides default implementations of the RunnableInterface.
// Embed this in your skill struct to get boilerplate getter/setter methods.
// Only implement Check() and Run() for the specific skill logic.
//
// State is stored in an omni.Atom (thread-safe via sync.RWMutex) for simple
// properties (id, description, args, dryRun, timeout, becomeUser). NodeConfig
// is stored as a separate typed field because it contains a *slog.Logger
// pointer that cannot be serialized to the Atom's map[string]string.
//
// The framework clones the skill via ToMap()/FromMap() before mutating it
// for concurrency safety. Each goroutine gets its own clone — the original
// shared instance is never mutated during parallel execution.
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
	atom    omni.AtomInterface
	nodeCfg NodeConfig
	mu      sync.RWMutex // protects nodeCfg
}

// NewBaseSkill creates a new BaseSkill with default values.
// Use the setter methods to configure it before returning from your constructor.
func NewBaseSkill() *BaseSkill {
	return &BaseSkill{
		atom: omni.NewAtom(atomTypeSkill),
	}
}

// GetID returns the unique identifier for this skill.
func (b *BaseSkill) GetID() string {
	return b.atom.Get(propID)
}

// SetID sets the unique identifier for this skill.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BaseSkill) SetID(id string) RunnableInterface {
	b.atom.Set(propID, id)
	return b
}

// GetDescription returns a short description of what the skill does.
func (b *BaseSkill) GetDescription() string {
	return b.atom.Get(propDescription)
}

// SetDescription sets a short description of what the skill does.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BaseSkill) SetDescription(description string) RunnableInterface {
	b.atom.Set(propDescription, description)
	return b
}

// GetNodeConfig returns the current node configuration for this skill.
func (b *BaseSkill) GetNodeConfig() NodeConfig {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.nodeCfg
}

// SetNodeConfig sets the node configuration for this skill execution.
// Returns RunnableInterface for fluent method chaining with embedding types.
func (b *BaseSkill) SetNodeConfig(cfg NodeConfig) RunnableInterface {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nodeCfg = cfg
	return b
}

// GetArg retrieves a single argument value by key.
func (b *BaseSkill) GetArg(key string) string {
	return b.atom.Get(argPrefix + key)
}

// SetArg sets a single argument value.
// Returns RunnableInterface for fluent method chaining.
func (b *BaseSkill) SetArg(key, value string) RunnableInterface {
	b.atom.Set(argPrefix+key, value)
	return b
}

// GetArgs returns the entire arguments map.
func (b *BaseSkill) GetArgs() map[string]string {
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
func (b *BaseSkill) SetArgs(args map[string]string) RunnableInterface {
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
func (b *BaseSkill) IsDryRun() bool {
	return b.atom.Get(propDryRun) == boolTrue
}

// SetDryRun sets whether to simulate execution without making changes.
// Returns RunnableInterface for fluent method chaining.
func (b *BaseSkill) SetDryRun(dryRun bool) RunnableInterface {
	b.atom.Set(propDryRun, strconv.FormatBool(dryRun))
	return b
}

// GetTimeout returns the maximum duration for skill execution.
func (b *BaseSkill) GetTimeout() time.Duration {
	s := b.atom.Get(propTimeout)
	if s == "" {
		return 0
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return time.Duration(n)
}

// SetTimeout sets the maximum duration for skill execution.
// Returns RunnableInterface for fluent method chaining.
func (b *BaseSkill) SetTimeout(timeout time.Duration) RunnableInterface {
	b.atom.Set(propTimeout, strconv.FormatInt(int64(timeout), 10))
	return b
}

// GetBecomeUser returns the configured become user.
// Returns empty string if not set.
func (b *BaseSkill) GetBecomeUser() string {
	return b.atom.Get(propBecomeUser)
}

// SetBecomeUser sets the user to become when executing commands.
// Returns BecomeInterface for fluent method chaining.
func (b *BaseSkill) SetBecomeUser(user string) BecomeInterface {
	b.atom.Set(propBecomeUser, user)
	return b
}

// ToMap returns the skill's state as a map for cloning.
// The map contains the Atom's properties plus the NodeConfig as a struct value.
// This is used by the framework to clone the skill before mutation.
//
// Note: atom.ToMap() strips "id" from properties and puts the Atom's auto-generated
// id field in the result. We override m["id"] with the actual ID stored as a property
// so that FromMap can restore it correctly.
func (b *BaseSkill) ToMap() map[string]any {
	m := b.atom.ToMap()
	m[propID] = b.atom.Get(propID)
	b.mu.RLock()
	m[mapKeyNodeConfig] = b.nodeCfg
	b.mu.RUnlock()
	return m
}

// FromMap populates the skill's state from a map (inverse of ToMap).
// Creates a fresh Atom and extracts the NodeConfig.
func (b *BaseSkill) FromMap(m map[string]any) {
	atomType, _ := m[mapKeyType].(string)
	if atomType == "" {
		atomType = atomTypeSkill
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
	b.SetBecomeUser(user)
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
