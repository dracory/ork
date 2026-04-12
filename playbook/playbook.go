// Package playbook provides the base types and interfaces for creating
// automation playbooks using SSH-based remote execution.
package playbook

import (
	"github.com/dracory/ork/config"
)

// Playbook name constants for use with RunPlaybook.
// These constants provide compile-time safety and IDE autocomplete for playbook names.
//
// Example:
//
//	node := ork.NewNodeForHost("server.example.com")
//	err := node.RunPlaybook(playbook.NamePing)
const (
	// NamePing checks SSH connectivity
	NamePing = "ping"

	// NameAptUpdate refreshes the package database
	NameAptUpdate = "apt-update"

	// NameAptUpgrade installs available updates
	NameAptUpgrade = "apt-upgrade"

	// NameAptStatus shows available updates
	NameAptStatus = "apt-status"

	// NameReboot reboots the server
	NameReboot = "reboot"

	// NameSwapCreate creates a swap file (requires "size" arg in GB)
	NameSwapCreate = "swap-create"

	// NameSwapDelete removes the swap file
	NameSwapDelete = "swap-delete"

	// NameSwapStatus shows swap status
	NameSwapStatus = "swap-status"

	// NameUserCreate creates a user with sudo (requires "username" arg)
	NameUserCreate = "user-create"

	// NameUserDelete deletes a user (requires "username" arg)
	NameUserDelete = "user-delete"

	// NameUserStatus shows user info (accepts optional "username" arg)
	NameUserStatus = "user-status"
)

// Playbook is the interface that all playbooks must implement.
// A playbook is a self-contained automation task that runs on a remote server.
type Playbook interface {
	// Name returns the unique identifier for this playbook (e.g., "apt-update")
	Name() string

	// Description returns a short description of what the playbook does
	Description() string

	// Run executes the playbook with the given configuration
	Run(config config.Config) error
}

// SimplePlaybook is a function-based playbook implementation.
// Use this for simple playbooks that don't need complex state.
type SimplePlaybook struct {
	name        string
	description string
	runFn       func(config.Config) error
}

// NewSimplePlaybook creates a new simple playbook from a function.
func NewSimplePlaybook(name, description string, runFn func(config.Config) error) *SimplePlaybook {
	return &SimplePlaybook{
		name:        name,
		description: description,
		runFn:       runFn,
	}
}

// Name returns the playbook name.
func (p *SimplePlaybook) Name() string {
	return p.name
}

// Description returns the playbook description.
func (p *SimplePlaybook) Description() string {
	return p.description
}

// Run executes the playbook.
func (p *SimplePlaybook) Run(cfg config.Config) error {
	return p.runFn(cfg)
}

// Registry holds a collection of playbooks.
type Registry struct {
	playbooks map[string]Playbook
}

// NewRegistry creates a new playbook registry.
func NewRegistry() *Registry {
	return &Registry{
		playbooks: make(map[string]Playbook),
	}
}

// Register adds a playbook to the registry.
func (r *Registry) Register(p Playbook) {
	r.playbooks[p.Name()] = p
}

// Get retrieves a playbook by name.
func (r *Registry) Get(name string) (Playbook, bool) {
	p, ok := r.playbooks[name]
	return p, ok
}

// List returns all registered playbooks.
func (r *Registry) List() []Playbook {
	list := make([]Playbook, 0, len(r.playbooks))
	for _, p := range r.playbooks {
		list = append(list, p)
	}
	return list
}

// Names returns all registered playbook names.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.playbooks))
	for name := range r.playbooks {
		names = append(names, name)
	}
	return names
}
