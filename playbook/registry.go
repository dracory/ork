package playbook

// Registry holds a collection of playbooks.
type Registry struct {
	playbooks map[string]PlaybookInterface
}

// NewRegistry creates a new playbook registry.
func NewRegistry() *Registry {
	return &Registry{
		playbooks: make(map[string]PlaybookInterface),
	}
}

// PlaybookRegister adds a playbook to the registry.
func (r *Registry) PlaybookRegister(p PlaybookInterface) {
	r.playbooks[p.GetID()] = p
}

// PlaybookFindByID retrieves a playbook by ID.
func (r *Registry) PlaybookFindByID(id string) (PlaybookInterface, bool) {
	p, ok := r.playbooks[id]
	return p, ok
}

// PlaybookList returns all registered playbooks.
func (r *Registry) PlaybookList() []PlaybookInterface {
	list := make([]PlaybookInterface, 0, len(r.playbooks))
	for _, p := range r.playbooks {
		list = append(list, p)
	}
	return list
}

// GetPlaybookIDs returns all registered playbook IDs.
func (r *Registry) GetPlaybookIDs() []string {
	ids := make([]string, 0, len(r.playbooks))
	for id := range r.playbooks {
		ids = append(ids, id)
	}
	return ids
}
