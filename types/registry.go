package types

import (
	"errors"
	"sync"
)

// Registry holds a collection of skills.
type Registry struct {
	skills map[string]RunnableInterface
	mu     sync.RWMutex
}

// NewRegistry creates a new skill registry.
func NewRegistry() *Registry {
	return &Registry{
		skills: make(map[string]RunnableInterface),
	}
}

// SkillRegister adds a skill to the registry.
// Returns an error if the skill is nil or if a skill with the same ID already exists.
func (r *Registry) SkillRegister(s RunnableInterface) error {
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
func (r *Registry) SkillFindByID(id string) (RunnableInterface, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.skills[id]
	return s, ok
}

// SkillList returns all registered skills.
func (r *Registry) SkillList() []RunnableInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]RunnableInterface, 0, len(r.skills))
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
