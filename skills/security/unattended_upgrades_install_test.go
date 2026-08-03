package security

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/internal/skilltest"
	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// dryRunCfg returns a NodeConfig configured for dry-run mode with an empty args map.
func dryRunCfg() types.NodeConfig {
	return types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}
}

// nonDryRunCfg returns a NodeConfig with dry-run off and an empty args map.
func nonDryRunCfg() types.NodeConfig {
	return types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}
}

// exitErr creates a *ssh.ExitError for testing command exit failures.
func exitErr() error {
	return ssh.NewExitError()
}

// connErr creates a connection error for testing SSH failures.
func connErr() error {
	return fmt.Errorf("connection refused")
}

// expectedCheckCommand is the dpkg-query command Check() issues.
const expectedCheckCommand = "dpkg-query -W -- unattended-upgrades 2>/dev/null"

// expectedInstallCommand is the apt-get install command Run() issues.
// Built with the same constants as the production code so the test stays
// in sync if the constants change.
func expectedInstallCommand() string {
	cmd := ""
	cmd += skills.DebianNonInteractive
	cmd += " apt-get install -y unattended-upgrades apt-listchanges"
	cmd += skills.DpkgConfOptions
	return cmd
}

// expectedAutoUpgradesWriteCommand is the printf command that writes 20auto-upgrades.
func expectedAutoUpgradesWriteCommand() string {
	return fmt.Sprintf("printf '%%s' %s > %s",
		skills.ShellEscapeContent(autoUpgradesContent),
		skills.ShellEscapeArg(pathAutoUpgrades))
}

// expectedUnattendedWriteCommand is the printf command that writes 50unattended-upgrades.
func expectedUnattendedWriteCommand() string {
	return fmt.Sprintf("printf '%%s' %s > %s",
		skills.ShellEscapeContent(unattendedUpgradesContent),
		skills.ShellEscapeArg(pathUnattendedUpgrades))
}

// expectedChmodAutoCommand is the chmod command for 20auto-upgrades.
func expectedChmodAutoCommand() string {
	return "chmod 644 " + skills.ShellEscapeArg(pathAutoUpgrades)
}

// expectedChmodUnattendedCommand is the chmod command for 50unattended-upgrades.
func expectedChmodUnattendedCommand() string {
	return "chmod 644 " + skills.ShellEscapeArg(pathUnattendedUpgrades)
}

// ---------------------------------------------------------------------------
// Constructor and metadata
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_NewUnattendedUpgradesInstall(t *testing.T) {
	pb := NewUnattendedUpgradesInstall()

	if pb.GetID() != "unattended-upgrades-install" {
		t.Errorf("Expected ID to be 'unattended-upgrades-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install and configure unattended-upgrades for security updates"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// ---------------------------------------------------------------------------
// Check
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_Check_DryRun(t *testing.T) {
	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(dryRunCfg())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error from Check in dry-run, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true in dry-run mode")
	}
}

func TestUnattendedUpgradesInstall_Check_AlreadyInstalled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query succeeds → package is installed → Check returns false
	test.ExpectCommand(expectedCheckCommand, "unattended-upgrades 2.9\n")

	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(test.Config())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if needs {
		t.Error("Expected Check to return false when package is installed")
	}
}

func TestUnattendedUpgradesInstall_Check_NotInstalled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query exits non-zero → package not installed → Check returns true
	test.ExpectError(expectedCheckCommand, exitErr())

	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(test.Config())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error for exit-code failure, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true when package is not installed")
	}
}

func TestUnattendedUpgradesInstall_Check_ConnectionError(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// Non-exit error (e.g. SSH connection failure) must be propagated, not
	// masked as "needs install".
	test.ExpectError(expectedCheckCommand, connErr())

	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(test.Config())

	needs, err := pb.Check()
	if err == nil {
		t.Error("Expected error to be propagated for connection failure")
	}
	if needs {
		t.Error("Expected Check to return false on connection error, not mask it as needs-install")
	}
}

// ---------------------------------------------------------------------------
// Run — dry-run
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_Run_DryRun(t *testing.T) {
	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure unattended-upgrades"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// ---------------------------------------------------------------------------
// Run — already installed (idempotency)
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_Run_AlreadyInstalled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query succeeds → Check returns false → Run short-circuits
	test.ExpectCommand(expectedCheckCommand, "unattended-upgrades 2.9\n")

	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when unattended-upgrades is already installed")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}

	expected := "Unattended-upgrades is already installed"
	if result.Message != expected {
		t.Errorf("Expected message %q, got %q", expected, result.Message)
	}

	// The install command must NOT have been run
	test.AssertCommandNotRun(expectedInstallCommand())
}

// ---------------------------------------------------------------------------
// Run — fresh install (full command sequence)
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_Run_NotInstalled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query fails → package not installed → proceed with install
	test.ExpectError(expectedCheckCommand, exitErr())

	// All subsequent commands succeed
	test.ExpectCommand(expectedInstallCommand(), "")
	test.ExpectCommand(expectedAutoUpgradesWriteCommand(), "")
	test.ExpectCommand(expectedChmodAutoCommand(), "")
	test.ExpectCommand(expectedUnattendedWriteCommand(), "")
	test.ExpectCommand(expectedChmodUnattendedCommand(), "")

	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true after successful install")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}

	expected := "Unattended-upgrades installed and configured (security updates only, no auto-reboot)"
	if result.Message != expected {
		t.Errorf("Expected message %q, got %q", expected, result.Message)
	}

	// Verify every command in the sequence was actually issued
	test.AssertCommandRun(expectedInstallCommand())
	test.AssertCommandRun(expectedAutoUpgradesWriteCommand())
	test.AssertCommandRun(expectedChmodAutoCommand())
	test.AssertCommandRun(expectedUnattendedWriteCommand())
	test.AssertCommandRun(expectedChmodUnattendedCommand())
}

// ---------------------------------------------------------------------------
// Run — install command failure
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_Run_InstallFails(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query fails → package not installed
	test.ExpectError(expectedCheckCommand, exitErr())

	// apt-get install fails
	test.ExpectError(expectedInstallCommand(), connErr())

	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when install fails")
	}
	if result.Error == nil {
		t.Error("Expected error when install fails")
	}

	expected := "Failed to install unattended-upgrades"
	if result.Message != expected {
		t.Errorf("Expected message %q, got %q", expected, result.Message)
	}

	// Config writes must NOT have been attempted after install failure
	test.AssertCommandNotRun(expectedAutoUpgradesWriteCommand())
	test.AssertCommandNotRun(expectedUnattendedWriteCommand())
}

// ---------------------------------------------------------------------------
// Run — config write failure
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_Run_ConfigWriteFails(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query fails → package not installed
	test.ExpectError(expectedCheckCommand, exitErr())

	// Install succeeds
	test.ExpectCommand(expectedInstallCommand(), "")

	// 20auto-upgrades write fails
	test.ExpectError(expectedAutoUpgradesWriteCommand(), connErr())

	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when config write fails")
	}
	if result.Error == nil {
		t.Error("Expected error when config write fails")
	}

	expected := "Failed to write 20auto-upgrades config"
	if result.Message != expected {
		t.Errorf("Expected message %q, got %q", expected, result.Message)
	}

	// The second config write must NOT have been attempted
	test.AssertCommandNotRun(expectedUnattendedWriteCommand())
}

// ---------------------------------------------------------------------------
// Run — Check connection error propagates
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_Run_CheckConnectionError(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// Non-exit error from Check must propagate, not be masked as needs-install
	test.ExpectError(expectedCheckCommand, connErr())

	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when Check has a connection error")
	}
	if result.Error == nil {
		t.Error("Expected error to be propagated from Check")
	}

	// Install must NOT have been attempted
	test.AssertCommandNotRun(expectedInstallCommand())
}

// ---------------------------------------------------------------------------
// Fluent setters — return concrete type
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewUnattendedUpgradesInstall()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*UnattendedUpgradesInstall); !ok {
		t.Error("SetArgs should return *UnattendedUpgradesInstall, not just RunnableInterface")
	}
}

func TestUnattendedUpgradesInstall_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewUnattendedUpgradesInstall()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*UnattendedUpgradesInstall); !ok {
		t.Error("SetArg should return *UnattendedUpgradesInstall, not just RunnableInterface")
	}
}

func TestUnattendedUpgradesInstall_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewUnattendedUpgradesInstall()

	result := skill.SetID("custom-id")

	if _, ok := result.(*UnattendedUpgradesInstall); !ok {
		t.Error("SetID should return *UnattendedUpgradesInstall, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

func TestUnattendedUpgradesInstall_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewUnattendedUpgradesInstall()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*UnattendedUpgradesInstall); !ok {
		t.Error("SetDescription should return *UnattendedUpgradesInstall, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

func TestUnattendedUpgradesInstall_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewUnattendedUpgradesInstall()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*UnattendedUpgradesInstall); !ok {
		t.Error("SetTimeout should return *UnattendedUpgradesInstall, not just RunnableInterface")
	}
}

func TestUnattendedUpgradesInstall_MethodChaining_PreservesType(t *testing.T) {
	skill := NewUnattendedUpgradesInstall().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*UnattendedUpgradesInstall); !ok {
		t.Error("Method chaining should preserve *UnattendedUpgradesInstall type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
