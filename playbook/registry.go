package playbook

import (
	"errors"
	"sync"
)

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
		return errors.New("playbook.Registry: cannot register nil playbook")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.GetID()
	if _, exists := r.playbooks[id]; exists {
		return errors.New("playbook.Registry: playbook with ID '" + id + "' already exists")
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
