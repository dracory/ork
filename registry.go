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
