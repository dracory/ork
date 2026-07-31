package ork

import (
	"sync"

	"github.com/dracory/ork/skills/apt"
	"github.com/dracory/ork/skills/fail2ban"
	"github.com/dracory/ork/skills/fs"
	"github.com/dracory/ork/skills/mariadb"
	"github.com/dracory/ork/skills/php"
	"github.com/dracory/ork/skills/ping"
	"github.com/dracory/ork/skills/reboot"
	"github.com/dracory/ork/skills/security"
	"github.com/dracory/ork/skills/swap"
	"github.com/dracory/ork/skills/ufw"
	"github.com/dracory/ork/skills/user"
	"github.com/dracory/ork/types"
)

// globalSkillRegistry is the global skill registry that holds all built-in
// and user-registered skills. It is lazily initialized on first use as a singleton.
//
// The registry is used by Node.Skill() to look up and execute skills.
var (
	globalSkillRegistry     *types.Registry
	globalSkillRegistryOnce sync.Once
)

// GetGlobalSkillRegistry returns the global skill registry singleton.
// This is syntactic sugar for user convenience - it lazily initializes and returns
// the global registry with all built-in skills pre-registered.
//
// For most use cases, users should call this function. For testing or custom
// configurations, use NewDefaultRegistry() to create isolated registries.
//
// The registry is lazily initialized on first call using sync.Once to ensure
// thread-safe singleton behavior.
// Returns an error if initialization fails.
func GetGlobalSkillRegistry() (*types.Registry, error) {
	var initErr error
	globalSkillRegistryOnce.Do(func() {
		globalSkillRegistry, initErr = NewDefaultRegistry()
	})
	if initErr != nil {
		return nil, initErr
	}
	return globalSkillRegistry, nil
}

// NewSkillRegistry creates a new empty skill registry.
// This is a convenience method (sugar) for types.NewRegistry() with a more intuitive name.
// This creates a fresh empty registry instance, which is useful for:
// - Testing with isolated registries
// - Custom configurations with selective skill registration
// - Multiple independent registries in the same application
//
// Returns an empty registry ready for custom skill registration.
func NewSkillRegistry() *types.Registry {
	return types.NewRegistry()
}

// NewDefaultRegistry creates a new skill registry with all built-in skills registered.
// This creates a fresh registry instance (not a singleton), which is useful for:
// - Testing with isolated registries
// - Custom configurations without global state
// - Multiple independent registries in the same application
//
// For most production use cases, use GetGlobalSkillRegistry() instead for convenience.
// Returns an error if any skill registration fails.
func NewDefaultRegistry() (*types.Registry, error) {
	reg := NewSkillRegistry()

	skills := []types.RunnableInterface{
		ping.NewPing(),
		apt.NewAptInstall(),
		apt.NewAptStatus(),
		apt.NewAptUpdate(),
		apt.NewAptUpgrade(),
		fail2ban.NewFail2banInstall(),
		fail2ban.NewFail2banStatus(),
		mariadb.NewBackup(),
		mariadb.NewBackupEncrypt(),
		mariadb.NewChangePort(),
		mariadb.NewCreateDB(),
		mariadb.NewCreateUser(),
		mariadb.NewEnableSSL(),
		mariadb.NewEnableEncryption(),
		mariadb.NewInstall(),
		mariadb.NewListDBs(),
		mariadb.NewListUsers(),
		mariadb.NewPurge(),
		mariadb.NewSecure(),
		mariadb.NewStatus(),
		mariadb.NewSecurityAudit(),
		reboot.NewReboot(),
		security.NewSshHarden(),
		security.NewKernelHarden(),
		security.NewAideInstall(),
		security.NewAuditdInstall(),
		security.NewSshChangePort(),
		swap.NewSwapCreate(),
		swap.NewSwapDelete(),
		swap.NewSwapStatus(),
		ufw.NewAllow(),
		ufw.NewAllowMariaDB(),
		ufw.NewDeny(),
		ufw.NewDelete(),
		ufw.NewEnable(),
		ufw.NewDefault(),
		ufw.NewDisable(),
		ufw.NewReset(),
		ufw.NewUfwInstall(),
		ufw.NewUfwStatus(),
		user.NewUserAddToGroup(),
		user.NewUserCreate(),
		user.NewUserDelete(),
		user.NewUserList(),
		user.NewUserStatus(),
		fs.NewDirCreate(),
		fs.NewDirExists(),
		fs.NewDirDelete(),
		fs.NewFileCreate(),
		fs.NewFileExists(),
		fs.NewFileDelete(),
		fs.NewFileCopy(),
		fs.NewChangeOwner(),
		fs.NewChangeMode(),
		fs.NewSymlinkCreate(),
		fs.NewRename(),
		fs.NewRemove(),
		php.NewInstall(),
		php.NewUninstall(),
		php.NewInstallComposer(),
		php.NewUninstallComposer(),
		php.NewUpdateComposer(),
	}

	if err := reg.SetAll(skills); err != nil {
		return nil, err
	}

	return reg, nil
}
