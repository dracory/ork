package playbook

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
	r.playbooks[p.GetID()] = p
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
