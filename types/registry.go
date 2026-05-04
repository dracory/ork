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

func (r *Registry) Set(runnable RunnableInterface) error {
	if runnable == nil {
		return errors.New("types.Registry: cannot set nil runnable")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.set(runnable)
	return nil
}

// SetAll adds multiple runnables to the registry at once.
// Returns an error if any runnable is nil or if setting any runnable fails.
func (r *Registry) SetAll(runnables []RunnableInterface) error {
	if len(runnables) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, runnable := range runnables {
		if runnable == nil {
			return errors.New("types.Registry: cannot set nil runnable")
		}
		r.set(runnable)
	}
	return nil
}

func (r *Registry) set(runnable RunnableInterface) {
	r.runnables[runnable.GetID()] = runnable
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
