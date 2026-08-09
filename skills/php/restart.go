package php

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/skills/systemctl"
	"github.com/dracory/ork/types"
)

// Restart restarts the PHP-FPM systemd service for a given PHP version.
// It delegates to systemctl.NewRestart() with the service name derived from
// the version (e.g. "8.5" -> "php8.5-fpm").
//
// Usage:
//
//	node.Run(php.NewRestart().SetVersion("8.5"))
//	node.Run(php.NewRestart()) // uses DefaultVersion ("8.5")
//
// Execution Flow:
//  1. Resolves the systemd service name from the version arg (default "8.5")
//  2. Runs `systemctl restart php<version>-fpm` via the systemctl.Restart skill
//  3. Reports success or failure
//
// Args:
//   - version: PHP version (default: "8.5")
//
// Prerequisites:
//   - PHP-FPM must be installed (see Install)
//
// Related Skills:
//   - php.Install: Install PHP with extensions and configure FPM
//   - php.Uninstall: Remove PHP packages and FPM configuration
type Restart struct {
	*types.BaseSkill
}

// Compile-time assertion that Restart implements types.RunnableInterface.
var _ types.RunnableInterface = (*Restart)(nil)

// Check always returns true — restart is an explicit action requested by the user.
func (r *Restart) Check() (bool, error) {
	return true, nil
}

// Run restarts the PHP-FPM service via systemctl.
func (r *Restart) Run() types.Result {
	cfg := r.GetNodeConfig()

	version := r.GetArg(ArgVersion)
	if version == "" {
		version = DefaultVersion
	}

	service := fmt.Sprintf("php%s-fpm", version)

	result := types.RunSub(systemctl.NewRestart().SetService(service), cfg)
	if result.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to restart PHP-FPM: " + service,
			Error:   result.Error,
		}
	}

	return types.Result{
		Changed: true,
		Message: "PHP-FPM restarted: " + service,
	}
}

// SetArgs sets the arguments for the PHP-FPM restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetArgs(args map[string]string) types.RunnableInterface {
	r.BaseSkill.SetArgs(args)
	return r
}

// SetArg sets a single argument for the PHP-FPM restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetArg(key, value string) types.RunnableInterface {
	r.BaseSkill.SetArg(key, value)
	return r
}

// SetVersion sets the PHP version and returns Restart for chaining.
// Example: SetVersion("8.5")
func (r *Restart) SetVersion(version string) *Restart {
	r.BaseSkill.SetArg(ArgVersion, version)
	return r
}

// SetID sets the ID for the PHP-FPM restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetID(id string) types.RunnableInterface {
	r.BaseSkill.SetID(id)
	return r
}

// SetDescription sets the description for the PHP-FPM restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetDescription(description string) types.RunnableInterface {
	r.BaseSkill.SetDescription(description)
	return r
}

// SetTimeout sets the timeout for the PHP-FPM restart.
// Returns Restart for fluent method chaining.
func (r *Restart) SetTimeout(timeout time.Duration) types.RunnableInterface {
	r.BaseSkill.SetTimeout(timeout)
	return r
}

// NewRestart creates a new php-fpm-restart skill.
//
// Returns a Restart skill configured with skills.IDPhpFpmRestart identifier and
// description "Restart PHP-FPM via systemd".
//
// Example:
//
//	node.Run(php.NewRestart().SetVersion("8.5"))
func NewRestart() *Restart {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDPhpFpmRestart)
	pb.SetDescription("Restart PHP-FPM via systemd")
	return &Restart{BaseSkill: pb}
}
