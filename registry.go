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
	defaultRegistry.Register(playbooks.NewPing())
	defaultRegistry.Register(playbooks.NewAptUpdate())
	defaultRegistry.Register(playbooks.NewAptUpgrade())
	defaultRegistry.Register(playbooks.NewAptStatus())
	defaultRegistry.Register(playbooks.NewReboot())
	defaultRegistry.Register(playbooks.NewSwapCreate())
	defaultRegistry.Register(playbooks.NewSwapDelete())
	defaultRegistry.Register(playbooks.NewSwapStatus())
	defaultRegistry.Register(playbooks.NewUserCreate())
	defaultRegistry.Register(playbooks.NewUserDelete())
	defaultRegistry.Register(playbooks.NewUserStatus())
}
