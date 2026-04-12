// Package playbook provides the base types and interfaces for creating
// automation playbooks using SSH-based remote execution.
package playbook

import (
	"strings"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/ssh"
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

// Playbook is the interface that all playbooks must implement.
// A playbook is a self-contained automation task that runs on a remote server.
// All playbooks support idempotency through the Check() method and Result return value.
type Playbook interface {
	// Name returns the unique identifier for this playbook (e.g., "apt-update")
	Name() string

	// Description returns a short description of what the playbook does
	Description() string

	// Check determines if the playbook needs to make any changes.
	// Returns true if changes are needed, false if the system is already in the desired state.
	// Returns an error if the check itself fails.
	Check(config config.Config) (bool, error)

	// Run executes the playbook and returns a detailed result.
	// The Result.Changed field indicates whether any changes were made.
	Run(config config.Config) Result
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

// Check always returns true for simple playbooks.
// Override this behavior by creating a custom playbook struct.
func (p *SimplePlaybook) Check(cfg config.Config) (bool, error) {
	return true, nil
}

// Run executes the playbook and returns a Result.
func (p *SimplePlaybook) Run(cfg config.Config) Result {
	err := p.runFn(cfg)
	if err != nil {
		return Result{
			Changed: false,
			Message: "Playbook execution failed",
			Error:   err,
		}
	}
	return Result{
		Changed: true,
		Message: "Playbook executed successfully",
	}
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

// CheckExists runs a check command and returns true if the command succeeds.
// This is useful for checking if a file exists, a service is running, etc.
// Returns false if the command fails or produces no output.
func CheckExists(client *ssh.Client, checkCmd string) bool {
	output, err := client.Run(checkCmd)
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) != ""
}

// EnsureState ensures a desired state by running a check command first.
// If the check fails, it runs the apply command to achieve the desired state.
// Returns true if changes were made (apply was run), false if no changes needed.
// Returns an error if either command fails.
func EnsureState(client *ssh.Client, checkCmd, applyCmd string) (bool, error) {
	// Check if already in desired state
	output, err := client.Run(checkCmd)
	if err == nil && strings.TrimSpace(output) != "" {
		// Already in desired state
		return false, nil
	}

	// Apply the change
	_, err = client.Run(applyCmd)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Execute runs a playbook and returns a Result.
// This is a convenience wrapper that calls pb.Run(cfg).
func Execute(pb Playbook, cfg config.Config) Result {
	return pb.Run(cfg)
}
