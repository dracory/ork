package ork

import (
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/playbooks/apt"
	"github.com/dracory/ork/playbooks/fail2ban"
	"github.com/dracory/ork/playbooks/mariadb"
	"github.com/dracory/ork/playbooks/ping"
	"github.com/dracory/ork/playbooks/reboot"
	"github.com/dracory/ork/playbooks/security"
	"github.com/dracory/ork/playbooks/swap"
	"github.com/dracory/ork/playbooks/ufw"
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
	_ = defaultRegistry.PlaybookRegister(fail2ban.NewFail2banInstall())
	_ = defaultRegistry.PlaybookRegister(fail2ban.NewFail2banStatus())
	_ = defaultRegistry.PlaybookRegister(ufw.NewUfwInstall())
	_ = defaultRegistry.PlaybookRegister(ufw.NewUfwStatus())
	_ = defaultRegistry.PlaybookRegister(security.NewSshHarden())
	_ = defaultRegistry.PlaybookRegister(security.NewKernelHarden())
	_ = defaultRegistry.PlaybookRegister(security.NewAideInstall())
	_ = defaultRegistry.PlaybookRegister(security.NewAuditdInstall())
	_ = defaultRegistry.PlaybookRegister(security.NewSshChangePort())
	_ = defaultRegistry.PlaybookRegister(mariadb.NewInstall())
	_ = defaultRegistry.PlaybookRegister(mariadb.NewSecure())
	_ = defaultRegistry.PlaybookRegister(mariadb.NewCreateDB())
	_ = defaultRegistry.PlaybookRegister(mariadb.NewCreateUser())
	_ = defaultRegistry.PlaybookRegister(mariadb.NewStatus())
	_ = defaultRegistry.PlaybookRegister(mariadb.NewListDBs())
	_ = defaultRegistry.PlaybookRegister(mariadb.NewListUsers())
	_ = defaultRegistry.PlaybookRegister(mariadb.NewBackup())
	_ = defaultRegistry.PlaybookRegister(mariadb.NewSecurityAudit())
	_ = defaultRegistry.PlaybookRegister(mariadb.NewChangePort())
}
