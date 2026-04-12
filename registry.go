package ork

import (
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/playbooks"
)

// defaultRegistry is the global playbook registry that holds all built-in
// and user-registered playbooks. It is initialized at package load time
// with all 11 built-in playbooks pre-registered.
//
// The registry is used by Node.Playbook() to look up and execute playbooks.
var defaultRegistry *playbook.Registry

// RegisterPlaybook registers a custom playbook on the default registry.
// This allows external packages to add their own playbooks that can then
// be executed via node.RunPlaybookByID("custom-id").
// This function is thread-safe and can be called from multiple goroutines.
func RegisterPlaybook(p playbook.PlaybookInterface) {
	if defaultRegistry == nil {
		panic("ork: defaultRegistry not initialized - RegisterPlaybook called before init()")
	}
	if p == nil {
		panic("ork: cannot register nil playbook")
	}
	defaultRegistry.PlaybookRegister(p)
}

func init() {
	defaultRegistry = playbook.NewRegistry()

	// Register all 11 built-in playbooks
	defaultRegistry.PlaybookRegister(playbooks.NewPing())
	defaultRegistry.PlaybookRegister(playbooks.NewAptUpdate())
	defaultRegistry.PlaybookRegister(playbooks.NewAptUpgrade())
	defaultRegistry.PlaybookRegister(playbooks.NewAptStatus())
	defaultRegistry.PlaybookRegister(playbooks.NewReboot())
	defaultRegistry.PlaybookRegister(playbooks.NewSwapCreate())
	defaultRegistry.PlaybookRegister(playbooks.NewSwapDelete())
	defaultRegistry.PlaybookRegister(playbooks.NewSwapStatus())
	defaultRegistry.PlaybookRegister(playbooks.NewUserCreate())
	defaultRegistry.PlaybookRegister(playbooks.NewUserDelete())
	defaultRegistry.PlaybookRegister(playbooks.NewUserStatus())
}
