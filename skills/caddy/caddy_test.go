package caddy

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dracory/ork/internal/skilltest"
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

// writeTempCaddyfile creates a temporary Caddyfile and returns its path.
func writeTempCaddyfile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "Caddyfile")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp Caddyfile: %v", err)
	}
	return path
}

// ---------------------------------------------------------------------------
// Install — Constructor and metadata
// ---------------------------------------------------------------------------

func TestNewInstall(t *testing.T) {
	pb := NewInstall()
	if pb.GetID() != "caddy-install" {
		t.Errorf("Expected ID 'caddy-install', got '%s'", pb.GetID())
	}
	if pb.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestInstall_Check_DryRun(t *testing.T) {
	pb := NewInstall()
	pb.SetNodeConfig(dryRunCfg())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error from Check in dry-run, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true in dry-run mode")
	}
}

func TestInstall_Run_DryRun(t *testing.T) {
	pb := NewInstall()
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true in dry-run mode")
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestInstall_Run_AlreadyInstalled verifies that Run returns Changed=false
// when Caddy is already installed (Check returns false).
func TestInstall_Run_AlreadyInstalled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query succeeds → package is installed → Check returns false
	test.ExpectCommand("dpkg-query -W -- caddy 2>/dev/null", "caddy 2.8.4\n")

	pb := NewInstall()
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when Caddy is already installed")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	expected := "Caddy is already installed"
	if result.Message != expected {
		t.Errorf("Expected message %q, got %q", expected, result.Message)
	}
}

// TestInstall_Run_NotInstalled verifies that Run proceeds with installation
// when Caddy is not installed (dpkg-query fails with exit error).
func TestInstall_Run_NotInstalled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	// dpkg-query fails → package not installed → Check returns true
	test.ExpectError("dpkg-query -W -- caddy 2>/dev/null", exitErr())

	// Prerequisite apt install check: dpkg-query for prereqs fails (not installed)
	test.ExpectError("dpkg-query -W -- 'debian-keyring' 'debian-archive-keyring' 'apt-transport-https' 'curl' 'gnupg' 2>/dev/null", exitErr())
	// Prerequisite apt install succeeds
	test.ExpectCommand("DEBIAN_FRONTEND=noninteractive apt-get install -y -o Dpkg::Options::='--force-confdef' -o Dpkg::Options::='--force-confold' -- 'debian-keyring' 'debian-archive-keyring' 'apt-transport-https' 'curl' 'gnupg'", "")

	// GPG key add
	test.ExpectCommand("curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --batch --yes --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg", "")
	// Repo add
	test.ExpectCommand("curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list", "")
	// chmod keyring + sources list world-readable
	test.ExpectCommand("chmod o+r /usr/share/keyrings/caddy-stable-archive-keyring.gpg /etc/apt/sources.list.d/caddy-stable.list", "")

	// apt update check: dpkg-query for apt itself (apt-update's Check) — it doesn't use dpkg-query,
	// but apt-update's Check runs `apt-get update --dry-run` or similar. We just expect it to succeed.
	// Actually apt.NewAptUpdate().Check() runs `apt-get update` in dry-run check mode.
	// Let's just expect the apt-get update command to succeed.
	test.ExpectCommand("apt-get update", "")

	// Caddy apt install check: dpkg-query for caddy fails (already set above via exitErr)
	// Caddy apt install succeeds
	test.ExpectCommand("DEBIAN_FRONTEND=noninteractive apt-get install -y -o Dpkg::Options::='--force-confdef' -o Dpkg::Options::='--force-confold' -- 'caddy'", "")

	// Log dir create: check if dir exists → not exists
	test.ExpectError("test -d '/var/log/caddy'", exitErr())
	// mkdir + chown + chmod
	test.ExpectCommand("mkdir -p '/var/log/caddy'", "")
	test.ExpectCommand("chown 'caddy:caddy' '/var/log/caddy'", "")
	test.ExpectCommand("chmod '755' '/var/log/caddy'", "")

	// User check (non-required, output doesn't matter)
	test.ExpectCommand("id 'caddy'", "uid=999(caddy)\n")

	pb := NewInstall()
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true after installation")
	}
	if result.Error != nil {
		t.Errorf("Expected no error after installation, got: %v", result.Error)
	}
}

// ---------------------------------------------------------------------------
// Install — Method chaining
// ---------------------------------------------------------------------------

func TestInstall_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewInstall()
	result := skill.SetArgs(map[string]string{"x": "y"})
	if _, ok := result.(*Install); !ok {
		t.Error("SetArgs should return *Install")
	}
}

func TestInstall_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewInstall()
	result := skill.SetArg("x", "y")
	if _, ok := result.(*Install); !ok {
		t.Error("SetArg should return *Install")
	}
}

func TestInstall_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewInstall()
	result := skill.SetID("custom")
	if _, ok := result.(*Install); !ok {
		t.Error("SetID should return *Install")
	}
	if skill.GetID() != "custom" {
		t.Error("SetID should set the ID")
	}
}

func TestInstall_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewInstall()
	result := skill.SetDescription("custom desc")
	if _, ok := result.(*Install); !ok {
		t.Error("SetDescription should return *Install")
	}
}

func TestInstall_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewInstall()
	result := skill.SetTimeout(30 * time.Second)
	if _, ok := result.(*Install); !ok {
		t.Error("SetTimeout should return *Install")
	}
}

func TestInstall_MethodChaining_PreservesType(t *testing.T) {
	skill := NewInstall().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("x", "y").
		SetArgs(map[string]string{"a": "b"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*Install); !ok {
		t.Error("Method chaining should preserve *Install type")
	}
	if skill.GetID() != "custom-id" {
		t.Error("Chaining should set ID")
	}
	if skill.GetDescription() != "custom description" {
		t.Error("Chaining should set description")
	}
}

// ---------------------------------------------------------------------------
// Restart — Constructor and metadata
// ---------------------------------------------------------------------------

func TestNewRestart(t *testing.T) {
	pb := NewRestart()
	if pb.GetID() != "caddy-restart" {
		t.Errorf("Expected ID 'caddy-restart', got '%s'", pb.GetID())
	}
	if pb.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestRestart_Check_AlwaysTrue(t *testing.T) {
	pb := NewRestart()
	pb.SetNodeConfig(dryRunCfg())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true (Restart is intentionally non-idempotent)")
	}
}

// ---------------------------------------------------------------------------
// Restart — Dry-run
// ---------------------------------------------------------------------------

func TestRestart_Run_DryRun(t *testing.T) {
	pb := NewRestart().SetCaddyfilePath("webserver/Caddyfile")
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true in dry-run mode")
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestRestart_Run_DryRun_NoLocalFile verifies that dry-run does not require
// the local Caddyfile to exist.
func TestRestart_Run_DryRun_NoLocalFile(t *testing.T) {
	pb := NewRestart().SetCaddyfilePath("/nonexistent/path/Caddyfile")
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true in dry-run mode")
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode without local file, got: %v", result.Error)
	}
}

// ---------------------------------------------------------------------------
// Restart — Mock SSH tests
// ---------------------------------------------------------------------------

// TestRestart_Run_Success verifies the full restart flow with mock SSH.
func TestRestart_Run_Success(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	caddyfileContent := ":80 {\n  respond \"hello\"\n}\n"
	localPath := writeTempCaddyfile(t, caddyfileContent)

	// FileCreate.Check: file doesn't exist on remote
	test.ExpectError("test -f '/etc/caddy/Caddyfile'", exitErr())
	// FileCreate: write file
	test.ExpectCommand("printf '%s' '"+caddyfileContent+"' > '/etc/caddy/Caddyfile'", "")
	// FileCreate: chown
	test.ExpectCommand("chown 'root:caddy' '/etc/caddy/Caddyfile'", "")
	// FileCreate: chmod
	test.ExpectCommand("chmod '644' '/etc/caddy/Caddyfile'", "")
	// Validate
	test.ExpectCommand("caddy validate --config '/etc/caddy/Caddyfile'", "")
	// systemctl reload succeeds
	test.ExpectCommand("systemctl reload 'caddy'", "")
	// systemctl is-active returns "active"
	test.ExpectCommand("systemctl is-active 'caddy'", "active\n")

	pb := NewRestart().SetCaddyfilePath(localPath)
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true after successful restart")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
}

// TestRestart_Run_MissingCaddyfile verifies that a missing local Caddyfile
// returns an error in non-dry-run mode.
func TestRestart_Run_MissingCaddyfile(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	pb := NewRestart().SetCaddyfilePath("/nonexistent/Caddyfile")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when Caddyfile is missing")
	}
	if result.Error == nil {
		t.Error("Expected error when local Caddyfile is missing")
	}
}

// TestRestart_Run_ValidationFails verifies that a failed caddy validate
// stops the restart and returns an error.
func TestRestart_Run_ValidationFails(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	caddyfileContent := "invalid config\n"
	localPath := writeTempCaddyfile(t, caddyfileContent)

	// FileCreate.Check: file exists but content differs
	test.ExpectCommand("test -f '/etc/caddy/Caddyfile'", "")
	test.ExpectCommand("cat '/etc/caddy/Caddyfile'", "old content\n")
	// FileCreate: write file
	test.ExpectCommand("printf '%s' '"+caddyfileContent+"' > '/etc/caddy/Caddyfile'", "")
	// FileCreate: chown + chmod
	test.ExpectCommand("chown 'root:caddy' '/etc/caddy/Caddyfile'", "")
	test.ExpectCommand("chmod '644' '/etc/caddy/Caddyfile'", "")
	// Validate fails
	test.ExpectError("caddy validate --config '/etc/caddy/Caddyfile'", exitErr())

	pb := NewRestart().SetCaddyfilePath(localPath)
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when validation fails")
	}
	if result.Error == nil {
		t.Error("Expected error when Caddyfile validation fails")
	}
	// Reload should not have been called
	test.AssertCommandNotRun("systemctl reload 'caddy'")
}

// TestRestart_Run_ReloadFails verifies that a failed reload returns an error.
func TestRestart_Run_ReloadFails(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	caddyfileContent := ":80 {\n  respond \"hello\"\n}\n"
	localPath := writeTempCaddyfile(t, caddyfileContent)

	// FileCreate flow
	test.ExpectError("test -f '/etc/caddy/Caddyfile'", exitErr())
	test.ExpectCommand("printf '%s' '"+caddyfileContent+"' > '/etc/caddy/Caddyfile'", "")
	test.ExpectCommand("chown 'root:caddy' '/etc/caddy/Caddyfile'", "")
	test.ExpectCommand("chmod '644' '/etc/caddy/Caddyfile'", "")
	// Validate succeeds
	test.ExpectCommand("caddy validate --config '/etc/caddy/Caddyfile'", "")
	// Reload fails with exit error, then restart also fails
	test.ExpectError("systemctl reload 'caddy'", exitErr())
	test.ExpectError("systemctl restart 'caddy'", exitErr())

	pb := NewRestart().SetCaddyfilePath(localPath)
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when reload/restart fails")
	}
	if result.Error == nil {
		t.Error("Expected error when reload and restart both fail")
	}
}

// TestRestart_Run_IsActiveError verifies that an SSH error from is-active
// is propagated (not silently ignored).
func TestRestart_Run_IsActiveError(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	caddyfileContent := ":80 {\n  respond \"hello\"\n}\n"
	localPath := writeTempCaddyfile(t, caddyfileContent)

	// FileCreate flow
	test.ExpectError("test -f '/etc/caddy/Caddyfile'", exitErr())
	test.ExpectCommand("printf '%s' '"+caddyfileContent+"' > '/etc/caddy/Caddyfile'", "")
	test.ExpectCommand("chown 'root:caddy' '/etc/caddy/Caddyfile'", "")
	test.ExpectCommand("chmod '644' '/etc/caddy/Caddyfile'", "")
	// Validate succeeds
	test.ExpectCommand("caddy validate --config '/etc/caddy/Caddyfile'", "")
	// Reload succeeds
	test.ExpectCommand("systemctl reload 'caddy'", "")
	// is-active fails with connection error (not exit error)
	test.ExpectError("systemctl is-active 'caddy'", connErr())

	pb := NewRestart().SetCaddyfilePath(localPath)
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true (reload succeeded before is-active failed)")
	}
	if result.Error == nil {
		t.Error("Expected error from is-active SSH failure to be propagated")
	}
}

// ---------------------------------------------------------------------------
// Restart — SetCaddyfilePath / SetCaddyfileRemotePath
// ---------------------------------------------------------------------------

func TestRestart_SetCaddyfilePath(t *testing.T) {
	pb := NewRestart().SetCaddyfilePath("webserver/Caddyfile")
	if pb.GetArg(ArgCaddyfilePath) != "webserver/Caddyfile" {
		t.Errorf("Expected caddyfile-path 'webserver/Caddyfile', got '%s'", pb.GetArg(ArgCaddyfilePath))
	}
}

func TestRestart_SetCaddyfileRemotePath(t *testing.T) {
	pb := NewRestart().SetCaddyfileRemotePath("/custom/path/Caddyfile")
	if pb.GetArg(ArgCaddyfileRemotePath) != "/custom/path/Caddyfile" {
		t.Errorf("Expected caddyfile-remote-path '/custom/path/Caddyfile', got '%s'", pb.GetArg(ArgCaddyfileRemotePath))
	}
}

func TestRestart_SetCaddyfilePath_ReturnsConcreteType(t *testing.T) {
	if _, ok := any(NewRestart().SetCaddyfilePath("Caddyfile")).(*Restart); !ok {
		t.Error("SetCaddyfilePath should return *Restart")
	}
}

func TestRestart_SetCaddyfileRemotePath_ReturnsConcreteType(t *testing.T) {
	if _, ok := any(NewRestart().SetCaddyfileRemotePath("/custom")).(*Restart); !ok {
		t.Error("SetCaddyfileRemotePath should return *Restart")
	}
}

// ---------------------------------------------------------------------------
// Restart — Method chaining
// ---------------------------------------------------------------------------

func TestRestart_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewRestart()
	result := skill.SetArgs(map[string]string{"x": "y"})
	if _, ok := result.(*Restart); !ok {
		t.Error("SetArgs should return *Restart")
	}
}

func TestRestart_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewRestart()
	result := skill.SetArg("x", "y")
	if _, ok := result.(*Restart); !ok {
		t.Error("SetArg should return *Restart")
	}
}

func TestRestart_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewRestart()
	result := skill.SetID("custom")
	if _, ok := result.(*Restart); !ok {
		t.Error("SetID should return *Restart")
	}
}

func TestRestart_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewRestart()
	result := skill.SetDescription("custom desc")
	if _, ok := result.(*Restart); !ok {
		t.Error("SetDescription should return *Restart")
	}
}

func TestRestart_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewRestart()
	result := skill.SetTimeout(30 * time.Second)
	if _, ok := result.(*Restart); !ok {
		t.Error("SetTimeout should return *Restart")
	}
}

func TestRestart_MethodChaining_PreservesType(t *testing.T) {
	skill := NewRestart().
		SetCaddyfilePath("webserver/Caddyfile").
		SetCaddyfileRemotePath("/etc/caddy/Caddyfile").
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("x", "y").
		SetArgs(map[string]string{"a": "b"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*Restart); !ok {
		t.Error("Method chaining should preserve *Restart type")
	}
	if skill.GetID() != "custom-id" {
		t.Error("Chaining should set ID")
	}
}

// ---------------------------------------------------------------------------
// Status — Constructor and metadata
// ---------------------------------------------------------------------------

func TestNewStatus(t *testing.T) {
	pb := NewStatus()
	if pb.GetID() != "caddy-status" {
		t.Errorf("Expected ID 'caddy-status', got '%s'", pb.GetID())
	}
	if pb.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestStatus_Check_AlwaysFalse(t *testing.T) {
	pb := NewStatus()
	pb.SetNodeConfig(dryRunCfg())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}
	if needs {
		t.Error("Expected Check to return false for read-only status")
	}
}

// ---------------------------------------------------------------------------
// Status — Dry-run
// ---------------------------------------------------------------------------

func TestStatus_Run_DryRun(t *testing.T) {
	pb := NewStatus()
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false for read-only status in dry-run")
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run, got: %v", result.Error)
	}
}

// ---------------------------------------------------------------------------
// Status — Mock SSH test
// ---------------------------------------------------------------------------

func TestStatus_Run_Success(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	expectedOutput := "● caddy.service - Caddy\n   Loaded: loaded\n   Active: active (running)\n"
	test.ExpectCommand("systemctl status 'caddy' --no-pager -l", expectedOutput)

	pb := NewStatus()
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false for read-only status")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.Details["output"] != expectedOutput {
		t.Errorf("Expected output in Details, got '%s'", result.Details["output"])
	}
	if result.Details["service"] != "caddy" {
		t.Errorf("Expected service 'caddy' in Details, got '%s'", result.Details["service"])
	}
}

// ---------------------------------------------------------------------------
// Status — Method chaining
// ---------------------------------------------------------------------------

func TestStatus_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewStatus()
	result := skill.SetArgs(map[string]string{"x": "y"})
	if _, ok := result.(*Status); !ok {
		t.Error("SetArgs should return *Status")
	}
}

func TestStatus_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewStatus()
	result := skill.SetArg("x", "y")
	if _, ok := result.(*Status); !ok {
		t.Error("SetArg should return *Status")
	}
}

func TestStatus_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewStatus()
	result := skill.SetID("custom")
	if _, ok := result.(*Status); !ok {
		t.Error("SetID should return *Status")
	}
}

func TestStatus_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewStatus()
	result := skill.SetDescription("custom desc")
	if _, ok := result.(*Status); !ok {
		t.Error("SetDescription should return *Status")
	}
}

func TestStatus_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewStatus()
	result := skill.SetTimeout(30 * time.Second)
	if _, ok := result.(*Status); !ok {
		t.Error("SetTimeout should return *Status")
	}
}

func TestStatus_MethodChaining_PreservesType(t *testing.T) {
	skill := NewStatus().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("x", "y").
		SetArgs(map[string]string{"a": "b"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*Status); !ok {
		t.Error("Method chaining should preserve *Status type")
	}
	if skill.GetID() != "custom-id" {
		t.Error("Chaining should set ID")
	}
}

// ---------------------------------------------------------------------------
// runSub — dry-run propagation regression test
// ---------------------------------------------------------------------------

// TestRunSub_PropagatesDryRun verifies that runSub sets dry-run mode on
// sub-skills. This is a regression test for the bug where runSub only
// called SetNodeConfig but not SetDryRun, causing sub-skills to execute
// real SSH commands in dry-run mode.
func TestRunSub_PropagatesDryRun(t *testing.T) {
	// Use a Status skill as a proxy — it delegates to systemctl.Status via runSub.
	// In dry-run mode, the sub-skill should return a dry-run message, not execute SSH.
	pb := NewStatus()
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	// If dry-run was not propagated, the sub-skill would try SSH and fail
	// (no mock setup), resulting in an error or empty output.
	// With dry-run propagated, systemctl.Status returns its dry-run message.
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run (dry-run should be propagated to sub-skill), got: %v", result.Error)
	}
}
