package types

import (
	"errors"
	"sync"
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

// PlaybookInterface is the interface that all playbooks must implement.
// A playbook is a self-contained automation task that runs on a remote server.
// All playbooks support idempotency through the Check() method and Result return value.
//
// Usage:
//
//	pb := playbooks.NewUserCreate().
//	    SetNodeConfig(cfg).
//	    SetOptions(&types.PlaybookOptions{Args: map[string]string{"username": "alice"}})
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

	// GetNodeConfig returns the current node configuration for this playbook.
	GetNodeConfig() config.NodeConfig

	// SetNodeConfig sets the node configuration for this playbook execution.
	// Returns the PlaybookInterface for fluent method chaining.
	SetNodeConfig(cfg config.NodeConfig) PlaybookInterface

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

// Registry holds a collection of playbooks.
type Registry struct {
	playbooks map[string]PlaybookInterface
	mu        sync.RWMutex
}

// NewRegistry creates a new playbook registry.
func NewRegistry() *Registry {
	return &Registry{
		playbooks: make(map[string]PlaybookInterface),
	}
}

// PlaybookRegister adds a playbook to the registry.
// Returns an error if the playbook is nil or if a playbook with the same ID already exists.
func (r *Registry) PlaybookRegister(p PlaybookInterface) error {
	if p == nil {
		return errors.New("types.Registry: cannot register nil playbook")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.GetID()
	if _, exists := r.playbooks[id]; exists {
		return errors.New("types.Registry: playbook with ID '" + id + "' already exists")
	}

	r.playbooks[id] = p
	return nil
}

// PlaybookFindByID retrieves a playbook by ID.
func (r *Registry) PlaybookFindByID(id string) (PlaybookInterface, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.playbooks[id]
	return p, ok
}

// PlaybookList returns all registered playbooks.
func (r *Registry) PlaybookList() []PlaybookInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]PlaybookInterface, 0, len(r.playbooks))
	for _, p := range r.playbooks {
		list = append(list, p)
	}
	return list
}

// GetPlaybookIDs returns all registered playbook IDs.
func (r *Registry) GetPlaybookIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.playbooks))
	for id := range r.playbooks {
		ids = append(ids, id)
	}
	return ids
}
