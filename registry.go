package ork

import (
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/playbooks/apt"
	"github.com/dracory/ork/playbooks/ping"
	"github.com/dracory/ork/playbooks/reboot"
	"github.com/dracory/ork/playbooks/swap"
	"github.com/dracory/ork/playbooks/user"
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
	_ = defaultRegistry.PlaybookRegister(ping.NewPing())
	_ = defaultRegistry.PlaybookRegister(apt.NewAptUpdate())
	_ = defaultRegistry.PlaybookRegister(apt.NewAptUpgrade())
	_ = defaultRegistry.PlaybookRegister(apt.NewAptStatus())
	_ = defaultRegistry.PlaybookRegister(reboot.NewReboot())
	_ = defaultRegistry.PlaybookRegister(swap.NewSwapCreate())
	_ = defaultRegistry.PlaybookRegister(swap.NewSwapDelete())
	_ = defaultRegistry.PlaybookRegister(swap.NewSwapStatus())
	_ = defaultRegistry.PlaybookRegister(user.NewUserCreate())
	_ = defaultRegistry.PlaybookRegister(user.NewUserDelete())
	_ = defaultRegistry.PlaybookRegister(user.NewUserStatus())
}
