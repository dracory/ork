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

// GetDefaultRegistry returns the global default playbook registry.
// This allows external packages to query and register playbooks.
func GetDefaultRegistry() *playbook.Registry {
	return defaultRegistry
}

func init() {
	defaultRegistry = playbook.NewRegistry()

	// Register all 11 built-in playbooks
	_ = defaultRegistry.PlaybookRegister(playbooks.NewPing())
	_ = defaultRegistry.PlaybookRegister(playbooks.NewAptUpdate())
	_ = defaultRegistry.PlaybookRegister(playbooks.NewAptUpgrade())
	_ = defaultRegistry.PlaybookRegister(playbooks.NewAptStatus())
	_ = defaultRegistry.PlaybookRegister(playbooks.NewReboot())
	_ = defaultRegistry.PlaybookRegister(playbooks.NewSwapCreate())
	_ = defaultRegistry.PlaybookRegister(playbooks.NewSwapDelete())
	_ = defaultRegistry.PlaybookRegister(playbooks.NewSwapStatus())
	_ = defaultRegistry.PlaybookRegister(playbooks.NewUserCreate())
	_ = defaultRegistry.PlaybookRegister(playbooks.NewUserDelete())
	_ = defaultRegistry.PlaybookRegister(playbooks.NewUserStatus())
}
