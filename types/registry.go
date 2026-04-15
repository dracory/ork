package types

import (
	"errors"
	"sync"
	"time"

	"github.com/dracory/ork/config"
)

// SkillOptions provides configuration options for skill execution.
// This allows per-skill variable scoping and additional execution controls.
type SkillOptions struct {
	// Args contains skill-specific variables that override node-level arguments.
	// These are merged with node-level args (skill args take precedence).
	Args map[string]string

	// DryRun indicates whether to simulate execution without making changes.
	// When true, the skill should preview what would be done.
	DryRun bool

	// Timeout specifies the maximum duration for skill execution.
	// Zero means no timeout.
	Timeout time.Duration
}

// SkillInterface is the interface that all skills must implement.
// A skill is a self-contained automation task that runs on a remote server.
// All skills support idempotency through the Check() method and Result return value.
//
// Usage:
//
//	skill := skills.NewUserCreate().
//	    SetNodeConfig(cfg).
//	    SetOptions(&types.SkillOptions{Args: map[string]string{"username": "alice"}})
//
//	needsRun, _ := skill.Check()
//	result := skill.Run()
type SkillInterface interface {
	// GetID returns the unique identifier for this skill (e.g., "apt-update")
	GetID() string

	// SetID sets the unique identifier for this skill.
	SetID(id string) SkillInterface

	// GetDescription returns a short description of what the skill does
	GetDescription() string

	// SetDescription sets a short description of what the skill does.
	SetDescription(description string) SkillInterface

	// GetNodeConfig returns the current node configuration for this skill.
	GetNodeConfig() config.NodeConfig

	// SetNodeConfig sets the node configuration for this skill execution.
	// Returns the SkillInterface for fluent method chaining.
	SetNodeConfig(cfg config.NodeConfig) SkillInterface

	// GetArg retrieves a single argument value by key.
	GetArg(key string) string

	// SetArg sets a single argument value.
	// Returns the SkillInterface for fluent method chaining.
	SetArg(key, value string) SkillInterface

	// GetArgs returns the entire arguments map.
	GetArgs() map[string]string

	// SetArgs replaces the entire arguments map.
	// Returns the SkillInterface for fluent method chaining.
	SetArgs(args map[string]string) SkillInterface

	// IsDryRun returns true if this is a dry-run execution.
	IsDryRun() bool

	// SetDryRun sets whether to simulate execution without making changes.
	// Returns the SkillInterface for fluent method chaining.
	SetDryRun(dryRun bool) SkillInterface

	// GetTimeout returns the maximum duration for skill execution.
	GetTimeout() time.Duration

	// SetTimeout sets the maximum duration for skill execution.
	// Returns the SkillInterface for fluent method chaining.
	SetTimeout(timeout time.Duration) SkillInterface

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

// Registry holds a collection of skills.
type Registry struct {
	skills map[string]SkillInterface
	mu     sync.RWMutex
}

// NewRegistry creates a new skill registry.
func NewRegistry() *Registry {
	return &Registry{
		skills: make(map[string]SkillInterface),
	}
}

// SkillRegister adds a skill to the registry.
// Returns an error if the skill is nil or if a skill with the same ID already exists.
func (r *Registry) SkillRegister(s SkillInterface) error {
	if s == nil {
		return errors.New("types.Registry: cannot register nil skill")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := s.GetID()
	if _, exists := r.skills[id]; exists {
		return errors.New("types.Registry: skill with ID '" + id + "' already exists")
	}

	r.skills[id] = s
	return nil
}

// SkillFindByID retrieves a skill by ID.
func (r *Registry) SkillFindByID(id string) (SkillInterface, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.skills[id]
	return s, ok
}

// SkillList returns all registered skills.
func (r *Registry) SkillList() []SkillInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]SkillInterface, 0, len(r.skills))
	for _, s := range r.skills {
		list = append(list, s)
	}
	return list
}

// GetSkillIDs returns all registered skill IDs.
func (r *Registry) GetSkillIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.skills))
	for id := range r.skills {
		ids = append(ids, id)
	}
	return ids
}
