package ork

import (
	"sync"

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

// globalPlaybookRegistry is the global playbook registry that holds all built-in
// and user-registered playbooks. It is lazily initialized on first use as a singleton.
//
// The registry is used by Node.Playbook() to look up and execute playbooks.
var (
	globalPlaybookRegistry     *playbook.Registry
	globalPlaybookRegistryOnce sync.Once
)

// GetGlobalPlaybookRegistry returns the global playbook registry singleton.
// This is syntactic sugar for user convenience - it lazily initializes and returns
// the global registry with all built-in playbooks pre-registered.
//
// For most use cases, users should call this function. For testing or custom
// configurations, use NewDefaultRegistry() to create isolated registries.
//
// The registry is lazily initialized on first call using sync.Once to ensure
// thread-safe singleton behavior.
// Returns an error if initialization fails.
func GetGlobalPlaybookRegistry() (*playbook.Registry, error) {
	var initErr error
	globalPlaybookRegistryOnce.Do(func() {
		globalPlaybookRegistry, initErr = NewDefaultRegistry()
	})
	if initErr != nil {
		return nil, initErr
	}
	return globalPlaybookRegistry, nil
}

// NewDefaultRegistry creates a new playbook registry with all built-in playbooks registered.
// This creates a fresh registry instance (not a singleton), which is useful for:
// - Testing with isolated registries
// - Custom configurations without global state
// - Multiple independent registries in the same application
//
// For most production use cases, use GetGlobalPlaybookRegistry() instead for convenience.
// Returns an error if any playbook registration fails.
func NewDefaultRegistry() (*playbook.Registry, error) {
	reg := playbook.NewRegistry()

	playbooks := []playbook.PlaybookInterface{
		ping.NewPing(),
		apt.NewAptUpdate(),
		apt.NewAptUpgrade(),
		apt.NewAptStatus(),
		reboot.NewReboot(),
		swap.NewSwapCreate(),
		swap.NewSwapDelete(),
		swap.NewSwapStatus(),
		user.NewUserCreate(),
		user.NewUserDelete(),
		user.NewUserStatus(),
		fail2ban.NewFail2banInstall(),
		fail2ban.NewFail2banStatus(),
		ufw.NewUfwInstall(),
		ufw.NewUfwStatus(),
		ufw.NewAllowMariaDB(),
		security.NewSshHarden(),
		security.NewKernelHarden(),
		security.NewAideInstall(),
		security.NewAuditdInstall(),
		security.NewSshChangePort(),
		mariadb.NewInstall(),
		mariadb.NewSecure(),
		mariadb.NewCreateDB(),
		mariadb.NewCreateUser(),
		mariadb.NewStatus(),
		mariadb.NewListDBs(),
		mariadb.NewListUsers(),
		mariadb.NewBackup(),
		mariadb.NewSecurityAudit(),
		mariadb.NewChangePort(),
		mariadb.NewEnableSSL(),
		mariadb.NewEnableEncryption(),
		mariadb.NewBackupEncrypt(),
	}

	for _, pb := range playbooks {
		if err := reg.PlaybookRegister(pb); err != nil {
			return nil, err
		}
	}

	return reg, nil
}
