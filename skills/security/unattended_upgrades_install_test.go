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

// expectedFileCreateWriteCommand builds the printf command that fs.FileCreate
// issues to write content to a path.
func expectedFileCreateWriteCommand(path, content string) string {
	return fmt.Sprintf("printf '%%s' %s > %s",
		skills.ShellEscapeContent(content),
		skills.ShellEscapeArg(path))
}

// expectedFileCreateChmodCommand builds the chmod command that fs.FileCreate
// issues after writing a file.
func expectedFileCreateChmodCommand(path string) string {
	return fmt.Sprintf("chmod %s %s",
		skills.ShellEscapeArg("644"),
		skills.ShellEscapeArg(path))
}

// expectedFileCreateTestFCommand builds the test -f command that fs.FileCreate
// issues to check if a file exists.
func expectedFileCreateTestFCommand(path string) string {
	return fmt.Sprintf("test -f %s", skills.ShellEscapeArg(path))
}

// expectedFileCreateCatCommand builds the cat command that fs.FileCreate issues
// to read existing file content for idempotency comparison.
func expectedFileCreateCatCommand(path string) string {
	return fmt.Sprintf("cat %s", skills.ShellEscapeArg(path))
}

// expectedFileCreateStatModeCommand builds the stat command that fs.FileCreate
// issues to check the current file mode.
func expectedFileCreateStatModeCommand(path string) string {
	return fmt.Sprintf("stat -c '%%a' %s", skills.ShellEscapeArg(path))
}

// expectedValidateCommand is the apt-config dump command that validates the config.
const expectedValidateCommand = "apt-config dump APT::Periodic::Unattended-Upgrade"

// expectedValidateOutput is the expected output from apt-config dump when
// unattended-upgrades is correctly enabled.
const expectedValidateOutput = `APT::Periodic::Unattended-Upgrade "1";
`

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
	// Config writes must NOT have been attempted
	test.AssertCommandNotRun(expectedFileCreateWriteCommand(pathAutoUpgrades, autoUpgradesContent))
	test.AssertCommandNotRun(expectedFileCreateWriteCommand(pathUnattendedUpgrades, unattendedUpgradesContent))
}

// ---------------------------------------------------------------------------
// Run — fresh install (full command sequence with fs.FileCreate + validation)
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_Run_NotInstalled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query fails → package not installed → proceed with install
	test.ExpectError(expectedCheckCommand, exitErr())

	// apt-get install succeeds
	test.ExpectCommand(expectedInstallCommand(), "")

	// fs.FileCreate for 20auto-upgrades: the mock returns empty string for
	// test -f (file "exists"), cat (content "" ≠ desired), stat (mode "" ≠ "644"),
	// so FileCreate proceeds to write. We just need printf and chmod to succeed.
	// The mock returns ("", nil) for unexpressed commands, which is fine.

	// fs.FileCreate for 50unattended-upgrades: same pattern.

	// apt-config dump validation succeeds
	test.ExpectCommand(expectedValidateCommand, expectedValidateOutput)

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

	// Verify the install command was issued
	test.AssertCommandRun(expectedInstallCommand())

	// Verify fs.FileCreate wrote both config files
	test.AssertCommandRun(expectedFileCreateWriteCommand(pathAutoUpgrades, autoUpgradesContent))
	test.AssertCommandRun(expectedFileCreateChmodCommand(pathAutoUpgrades))
	test.AssertCommandRun(expectedFileCreateWriteCommand(pathUnattendedUpgrades, unattendedUpgradesContent))
	test.AssertCommandRun(expectedFileCreateChmodCommand(pathUnattendedUpgrades))

	// Verify config validation was run
	test.AssertCommandRun(expectedValidateCommand)
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
	test.AssertCommandNotRun(expectedFileCreateWriteCommand(pathAutoUpgrades, autoUpgradesContent))
	test.AssertCommandNotRun(expectedFileCreateWriteCommand(pathUnattendedUpgrades, unattendedUpgradesContent))
	// Validation must NOT have been run
	test.AssertCommandNotRun(expectedValidateCommand)
}

// ---------------------------------------------------------------------------
// Run — config write failure (first config file)
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_Run_ConfigWriteFails(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query fails → package not installed
	test.ExpectError(expectedCheckCommand, exitErr())

	// Install succeeds
	test.ExpectCommand(expectedInstallCommand(), "")

	// 20auto-upgrades write (printf) fails
	test.ExpectError(expectedFileCreateWriteCommand(pathAutoUpgrades, autoUpgradesContent), connErr())

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
	test.AssertCommandNotRun(expectedFileCreateWriteCommand(pathUnattendedUpgrades, unattendedUpgradesContent))
	// Validation must NOT have been run
	test.AssertCommandNotRun(expectedValidateCommand)
}

// ---------------------------------------------------------------------------
// Run — config validation failure
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_Run_ValidationFails(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query fails → package not installed
	test.ExpectError(expectedCheckCommand, exitErr())

	// Install succeeds
	test.ExpectCommand(expectedInstallCommand(), "")

	// fs.FileCreate commands succeed (mock returns empty string for unexpressed)

	// apt-config dump returns empty output → validation fails because
	// Unattended-Upgrade is not confirmed as enabled
	test.ExpectCommand(expectedValidateCommand, "")

	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	// Changed is true because the install and config writes did happen,
	// but the validation failed — the operator needs to investigate.
	if !result.Changed {
		t.Error("Expected Changed=true because install + config writes did happen")
	}
	if result.Error == nil {
		t.Error("Expected error when validation fails")
	}

	test.AssertResultMessageContains(result, "config validation failed")
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
// Run — config already correct (fs.FileCreate idempotency)
// ---------------------------------------------------------------------------

func TestUnattendedUpgradesInstall_Run_ConfigAlreadyCorrect(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query fails → package not installed → proceed with install
	test.ExpectError(expectedCheckCommand, exitErr())

	// Install succeeds
	test.ExpectCommand(expectedInstallCommand(), "")

	// fs.FileCreate for 20auto-upgrades: simulate that the file already exists
	// with the correct content and mode. FileCreate.Check() will:
	//   1. test -f → succeeds (file exists)
	//   2. cat → returns the exact desired content
	//   3. stat -c '%a' → returns "644"
	// → Check returns false → FileCreate.Run() skips the write
	test.ExpectCommand(expectedFileCreateTestFCommand(pathAutoUpgrades), "")
	test.ExpectCommand(expectedFileCreateCatCommand(pathAutoUpgrades), autoUpgradesContent)
	test.ExpectCommand(expectedFileCreateStatModeCommand(pathAutoUpgrades), "644")

	// fs.FileCreate for 50unattended-upgrades: same — file already correct
	test.ExpectCommand(expectedFileCreateTestFCommand(pathUnattendedUpgrades), "")
	test.ExpectCommand(expectedFileCreateCatCommand(pathUnattendedUpgrades), unattendedUpgradesContent)
	test.ExpectCommand(expectedFileCreateStatModeCommand(pathUnattendedUpgrades), "644")

	// apt-config dump validation succeeds
	test.ExpectCommand(expectedValidateCommand, expectedValidateOutput)

	pb := NewUnattendedUpgradesInstall()
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true because the package was installed")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}

	// The printf write commands must NOT have been issued because the
	// content already matched — this is the fs.FileCreate idempotency.
	test.AssertCommandNotRun(expectedFileCreateWriteCommand(pathAutoUpgrades, autoUpgradesContent))
	test.AssertCommandNotRun(expectedFileCreateWriteCommand(pathUnattendedUpgrades, unattendedUpgradesContent))
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
