package types

import (
	"errors"
	"strings"
	"sync"
)

// Registry holds a collection of runnables.
type Registry struct {
	runnables map[string]RunnableInterface
	// disabled maps a runnable ID to an optional human-readable comment
	// explaining why the runnable was disabled. Disabled runnables remain
	// in the registry (so they can be inspected/listed) but cannot be run.
	disabled map[string]string
	mu       sync.RWMutex
}

// NewRegistry creates a new registry.
func NewRegistry() *Registry {
	return &Registry{
		runnables: make(map[string]RunnableInterface),
		disabled:  make(map[string]string),
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

// Add is an alias for Set. It adds (or replaces) a runnable in the registry.
// Provided for callers that prefer "add to collection" semantics.
func (r *Registry) Add(runnable RunnableInterface) error {
	return r.Set(runnable)
}

// AddAll is an alias for SetAll. It adds multiple runnables to the registry at once.
// Provided for callers that prefer "add to collection" semantics.
func (r *Registry) AddAll(runnables []RunnableInterface) error {
	return r.SetAll(runnables)
}

// Remove deletes the runnable with the given ID from the registry.
// Any disable record for the ID is also cleared.
// Returns an error if no runnable with the given ID is registered.
func (r *Registry) Remove(id string) error {
	if id == "" {
		return errors.New("types.Registry: cannot remove runnable with empty ID")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.runnables[id]; !ok {
		return errors.New("types.Registry: cannot remove runnable '" + id + "': not found")
	}
	delete(r.runnables, id)
	delete(r.disabled, id)
	return nil
}

// RemoveAll deletes the runnables with the given IDs from the registry.
// Any disable records for the removed IDs are also cleared.
// IDs that are not registered are ignored (no error).
// Returns an error if any ID is empty, in which case no removals are performed.
func (r *Registry) RemoveAll(ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		if id == "" {
			return errors.New("types.Registry: cannot remove runnable with empty ID")
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, id := range ids {
		delete(r.runnables, id)
		delete(r.disabled, id)
	}
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

// Disable marks the runnable with the given ID as disabled.
// The comment is optional and may be used to record why the runnable
// was disabled (e.g. "deprecated", "broken in production", ...).
// A disabled runnable remains in the registry and can still be looked up
// via FindByID/List, but the run path will refuse to execute it.
// Returns an error if no runnable with the given ID is registered.
// Disabling an already-disabled runnable updates its comment.
func (r *Registry) Disable(id, comment string) error {
	if id == "" {
		return errors.New("types.Registry: cannot disable runnable with empty ID")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.runnables[id]; !ok {
		return errors.New("types.Registry: cannot disable runnable '" + id + "': not found")
	}
	r.disabled[id] = comment
	return nil
}

// Enable re-enables a previously disabled runnable.
// Returns an error if no runnable with the given ID is registered.
// Enabling a runnable that is not disabled is a no-op.
func (r *Registry) Enable(id string) error {
	if id == "" {
		return errors.New("types.Registry: cannot enable runnable with empty ID")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.runnables[id]; !ok {
		return errors.New("types.Registry: cannot enable runnable '" + id + "': not found")
	}
	delete(r.disabled, id)
	return nil
}

// IsDisabled reports whether the runnable with the given ID is disabled.
// It returns the disable comment (which may be empty) and a boolean
// indicating whether the runnable is currently disabled.
// Returns "", false for unknown IDs.
func (r *Registry) IsDisabled(id string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	comment, ok := r.disabled[id]
	return comment, ok
}

// ListDisabled returns the IDs of all currently disabled runnables.
// The returned slice is unordered.
func (r *Registry) ListDisabled() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.disabled))
	for id := range r.disabled {
		ids = append(ids, id)
	}
	return ids
}

// MergeReplaceOnOverlap copies all runnables and their disable state from src
// into r. When a runnable with the same ID exists in both registries (an
// "overlap"), the one from src REPLACES the one in r.
//
// The disable state from src also wins for overlapping IDs — a runnable
// disabled in src will be disabled in r after the merge, and a runnable
// enabled in src will be enabled in r after the merge (clearing any existing
// disable record in r).
//
// Runnables that exist only in r are left untouched, including their disable
// state.
//
// Returns an error if src is nil. Safe against concurrent mutation of src:
// it snapshots src under a read lock before applying changes to r, and does
// NOT hold both registry locks simultaneously (no deadlock risk).
func (r *Registry) MergeReplaceOnOverlap(src *Registry) error {
	if src == nil {
		return errors.New("types.Registry: cannot merge from nil registry")
	}

	// Snapshot src under its read lock, then release before locking r.
	// This avoids holding both locks at once (deadlock-safe).
	src.mu.RLock()
	runnables := make(map[string]RunnableInterface, len(src.runnables))
	for id, rn := range src.runnables {
		runnables[id] = rn
	}
	disabled := make(map[string]string, len(src.disabled))
	for id, comment := range src.disabled {
		disabled[id] = comment
	}
	src.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	for id, rn := range runnables {
		r.runnables[id] = rn
		// Source's disable state wins for this ID.
		if comment, isDisabled := disabled[id]; isDisabled {
			r.disabled[id] = comment
		} else {
			delete(r.disabled, id)
		}
	}
	return nil
}

// MergeKeepOnOverlap copies runnables from src into r, but when a runnable
// with the same ID exists in both registries (an "overlap"), the one already
// in r is KEPT — src's version is ignored for that ID.
//
// Only IDs that exist in src but NOT in r are added. The disable state for
// newly-added IDs is copied from src. The disable state for overlapping IDs
// is left unchanged in r (target wins).
//
// Returns an error if src is nil. Safe against concurrent mutation of src
// (snapshots src under a read lock before locking r).
func (r *Registry) MergeKeepOnOverlap(src *Registry) error {
	if src == nil {
		return errors.New("types.Registry: cannot merge from nil registry")
	}

	src.mu.RLock()
	runnables := make(map[string]RunnableInterface, len(src.runnables))
	for id, rn := range src.runnables {
		runnables[id] = rn
	}
	disabled := make(map[string]string, len(src.disabled))
	for id, comment := range src.disabled {
		disabled[id] = comment
	}
	src.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	for id, rn := range runnables {
		// Skip overlaps — target keeps its version and its disable state.
		if _, exists := r.runnables[id]; exists {
			continue
		}
		r.runnables[id] = rn
		if comment, isDisabled := disabled[id]; isDisabled {
			r.disabled[id] = comment
		}
	}
	return nil
}

// MergeNoOverlap copies runnables from src into r, but returns an error if
// any runnable ID exists in both registries (an "overlap"). No changes are
// applied when an overlap is detected — the merge is all-or-nothing.
//
// This is the strictest merge variant: it prevents accidental overwrites by
// failing loudly instead of silently replacing or skipping.
//
// Returns an error if src is nil, or if any ID in src already exists in r.
// Safe against concurrent mutation of src (snapshots src under a read lock
// before locking r).
func (r *Registry) MergeNoOverlap(src *Registry) error {
	if src == nil {
		return errors.New("types.Registry: cannot merge from nil registry")
	}

	src.mu.RLock()
	runnables := make(map[string]RunnableInterface, len(src.runnables))
	for id, rn := range src.runnables {
		runnables[id] = rn
	}
	disabled := make(map[string]string, len(src.disabled))
	for id, comment := range src.disabled {
		disabled[id] = comment
	}
	src.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Detect overlaps first — fail before mutating anything.
	var overlaps []string
	for id := range runnables {
		if _, exists := r.runnables[id]; exists {
			overlaps = append(overlaps, id)
		}
	}
	if len(overlaps) > 0 {
		return errors.New("types.Registry: merge aborted — overlapping IDs: " + strings.Join(overlaps, ", "))
	}

	for id, rn := range runnables {
		r.runnables[id] = rn
		if comment, isDisabled := disabled[id]; isDisabled {
			r.disabled[id] = comment
		}
	}
	return nil
}
