package types

import (
	"errors"
	"sync"
)

// Registry holds a collection of runnables.
type Registry struct {
	runnables map[string]RunnableInterface
	mu        sync.RWMutex
}

// NewRegistry creates a new registry.
func NewRegistry() *Registry {
	return &Registry{
		runnables: make(map[string]RunnableInterface),
	}
}

// Register adds a runnable to the registry.
// Returns an error if the runnable is nil or if a runnable with the same ID already exists.
func (r *Registry) Register(runnable RunnableInterface) error {
	if runnable == nil {
		return errors.New("types.Registry: cannot register nil runnable")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := runnable.GetID()
	if _, exists := r.runnables[id]; exists {
		return errors.New("types.Registry: runnable with ID '" + id + "' already exists")
	}

	r.runnables[id] = runnable
	return nil
}

// FindByID retrieves a runnable by ID.
func (r *Registry) FindByID(id string) (RunnableInterface, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	runnable, ok := r.runnables[id]
	return runnable, ok
}

// List returns all registered runnables.
func (r *Registry) List() []RunnableInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]RunnableInterface, 0, len(r.runnables))
	for _, runnable := range r.runnables {
		list = append(list, runnable)
	}
	return list
}

// GetIDs returns all registered runnable IDs.
func (r *Registry) GetIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.runnables))
	for id := range r.runnables {
		ids = append(ids, id)
	}
	return ids
}
