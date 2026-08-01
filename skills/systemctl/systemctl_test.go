package systemctl

import (
	"fmt"
	"log/slog"
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

// ---------------------------------------------------------------------------
// Status
// ---------------------------------------------------------------------------

func TestStatus_Run_DryRun(t *testing.T) {
	pb := NewStatus().SetService("caddy")
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false for read-only status in dry-run")
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run, got: %v", result.Error)
	}
	expected := "Would show status of unit: caddy"
	if result.Message != expected {
		t.Errorf("Expected message %q, got %q", expected, result.Message)
	}
}

func TestStatus_Run_NoService(t *testing.T) {
	pb := NewStatus()
	pb.SetNodeConfig(nonDryRunCfg())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when no service specified")
	}
	if result.Error == nil {
		t.Error("Expected an error when no service is specified")
	}
}

func TestStatus_Check_AlwaysFalse(t *testing.T) {
	pb := NewStatus().SetService("caddy")
	pb.SetNodeConfig(dryRunCfg())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}
	if needs {
		t.Error("Expected Check to return false for read-only status")
	}
}

func TestStatus_NewStatus(t *testing.T) {
	pb := NewStatus()
	if pb.GetID() != "systemctl-status" {
		t.Errorf("Expected ID 'systemctl-status', got '%s'", pb.GetID())
	}
	if pb.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestStatus_SetService(t *testing.T) {
	pb := NewStatus().SetService("mariadb")
	if pb.GetArg(ArgService) != "mariadb" {
		t.Errorf("Expected service 'mariadb', got '%s'", pb.GetArg(ArgService))
	}
}

// ---------------------------------------------------------------------------
// IsActive
// ---------------------------------------------------------------------------

func TestIsActive_Run_DryRun(t *testing.T) {
	pb := NewIsActive().SetService("caddy")
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false for read-only is-active in dry-run")
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run, got: %v", result.Error)
	}
}

func TestIsActive_Run_NoService(t *testing.T) {
	pb := NewIsActive()
	pb.SetNodeConfig(nonDryRunCfg())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when no service specified")
	}
	if result.Error == nil {
		t.Error("Expected an error when no service is specified")
	}
}

func TestIsActive_Check_AlwaysFalse(t *testing.T) {
	pb := NewIsActive().SetService("caddy")
	pb.SetNodeConfig(dryRunCfg())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}
	if needs {
		t.Error("Expected Check to return false for read-only is-active")
	}
}

func TestIsActive_NewIsActive(t *testing.T) {
	pb := NewIsActive()
	if pb.GetID() != "systemctl-is-active" {
		t.Errorf("Expected ID 'systemctl-is-active', got '%s'", pb.GetID())
	}
}

// ---------------------------------------------------------------------------
// DaemonReload
// ---------------------------------------------------------------------------

func TestDaemonReload_Run_DryRun(t *testing.T) {
	pb := NewDaemonReload()
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true for daemon-reload in dry-run")
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run, got: %v", result.Error)
	}
}

func TestDaemonReload_Check_AlwaysTrue(t *testing.T) {
	pb := NewDaemonReload()
	pb.SetNodeConfig(dryRunCfg())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true for daemon-reload")
	}
}

func TestDaemonReload_NewDaemonReload(t *testing.T) {
	pb := NewDaemonReload()
	if pb.GetID() != "systemctl-daemon-reload" {
		t.Errorf("Expected ID 'systemctl-daemon-reload', got '%s'", pb.GetID())
	}
}

// ---------------------------------------------------------------------------
// Restart
// ---------------------------------------------------------------------------

func TestRestart_Run_DryRun(t *testing.T) {
	pb := NewRestart().SetService("caddy")
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true for restart in dry-run")
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run, got: %v", result.Error)
	}
	expected := "Would restart unit: caddy"
	if result.Message != expected {
		t.Errorf("Expected message %q, got %q", expected, result.Message)
	}
}

func TestRestart_Run_NoService(t *testing.T) {
	pb := NewRestart()
	pb.SetNodeConfig(nonDryRunCfg())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when no service specified")
	}
	if result.Error == nil {
		t.Error("Expected an error when no service is specified")
	}
}

func TestRestart_Check_AlwaysTrue(t *testing.T) {
	pb := NewRestart().SetService("caddy")
	pb.SetNodeConfig(dryRunCfg())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true for restart")
	}
}

func TestRestart_NewRestart(t *testing.T) {
	pb := NewRestart()
	if pb.GetID() != "systemctl-restart" {
		t.Errorf("Expected ID 'systemctl-restart', got '%s'", pb.GetID())
	}
}

func TestRestart_SetService(t *testing.T) {
	pb := NewRestart().SetService("php8.5-fpm")
	if pb.GetArg(ArgService) != "php8.5-fpm" {
		t.Errorf("Expected service 'php8.5-fpm', got '%s'", pb.GetArg(ArgService))
	}
}

// ---------------------------------------------------------------------------
// Reload
// ---------------------------------------------------------------------------

func TestReload_Run_DryRun(t *testing.T) {
	pb := NewReload().SetService("caddy")
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true for reload in dry-run")
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run, got: %v", result.Error)
	}
}

func TestReload_Run_NoService(t *testing.T) {
	pb := NewReload()
	pb.SetNodeConfig(nonDryRunCfg())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when no service specified")
	}
	if result.Error == nil {
		t.Error("Expected an error when no service is specified")
	}
}

func TestReload_Check_AlwaysTrue(t *testing.T) {
	pb := NewReload().SetService("caddy")
	pb.SetNodeConfig(dryRunCfg())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true for reload")
	}
}

func TestReload_NewReload(t *testing.T) {
	pb := NewReload()
	if pb.GetID() != "systemctl-reload" {
		t.Errorf("Expected ID 'systemctl-reload', got '%s'", pb.GetID())
	}
}

// ---------------------------------------------------------------------------
// Enable
// ---------------------------------------------------------------------------

func TestEnable_Run_DryRun(t *testing.T) {
	pb := NewEnable().SetService("mariadb-backup.timer")
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true for enable in dry-run")
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run, got: %v", result.Error)
	}
}

func TestEnable_Run_DryRun_WithStart(t *testing.T) {
	pb := NewEnable().SetService("mariadb-backup.timer").SetStart(true)
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true for enable+start in dry-run")
	}
}

func TestEnable_Run_NoService(t *testing.T) {
	pb := NewEnable()
	pb.SetNodeConfig(nonDryRunCfg())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when no service specified")
	}
	if result.Error == nil {
		t.Error("Expected an error when no service is specified")
	}
}

func TestEnable_Check_NoService(t *testing.T) {
	pb := NewEnable()
	pb.SetNodeConfig(nonDryRunCfg())

	_, err := pb.Check()
	if err == nil {
		t.Error("Expected an error from Check when no service is specified")
	}
}

func TestEnable_Check_DryRun(t *testing.T) {
	pb := NewEnable().SetService("caddy")
	pb.SetNodeConfig(dryRunCfg())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error from Check in dry-run, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true in dry-run")
	}
}

func TestEnable_NewEnable(t *testing.T) {
	pb := NewEnable()
	if pb.GetID() != "systemctl-enable" {
		t.Errorf("Expected ID 'systemctl-enable', got '%s'", pb.GetID())
	}
}

func TestEnable_SetStart(t *testing.T) {
	pb := NewEnable().SetStart(true)
	if pb.GetArg(ArgStart) != "true" {
		t.Errorf("Expected ArgStart 'true', got '%s'", pb.GetArg(ArgStart))
	}
}

func TestEnable_ShouldStart(t *testing.T) {
	pb := NewEnable()
	if pb.shouldStart() {
		t.Error("Expected shouldStart=false by default")
	}
	pb.SetStart(true)
	if !pb.shouldStart() {
		t.Error("Expected shouldStart=true after SetStart(true)")
	}
}

// ---------------------------------------------------------------------------
// Disable
// ---------------------------------------------------------------------------

func TestDisable_Run_DryRun(t *testing.T) {
	pb := NewDisable().SetService("old-service")
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true for disable in dry-run")
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run, got: %v", result.Error)
	}
}

func TestDisable_Run_DryRun_WithStop(t *testing.T) {
	pb := NewDisable().SetService("old-service").SetStop(true)
	pb.SetNodeConfig(dryRunCfg())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true for disable+stop in dry-run")
	}
}

func TestDisable_Run_NoService(t *testing.T) {
	pb := NewDisable()
	pb.SetNodeConfig(nonDryRunCfg())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when no service specified")
	}
	if result.Error == nil {
		t.Error("Expected an error when no service is specified")
	}
}

func TestDisable_Check_NoService(t *testing.T) {
	pb := NewDisable()
	pb.SetNodeConfig(nonDryRunCfg())

	_, err := pb.Check()
	if err == nil {
		t.Error("Expected an error from Check when no service is specified")
	}
}

func TestDisable_Check_DryRun(t *testing.T) {
	pb := NewDisable().SetService("old-service")
	pb.SetNodeConfig(dryRunCfg())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error from Check in dry-run, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true in dry-run")
	}
}

func TestDisable_NewDisable(t *testing.T) {
	pb := NewDisable()
	if pb.GetID() != "systemctl-disable" {
		t.Errorf("Expected ID 'systemctl-disable', got '%s'", pb.GetID())
	}
}

func TestDisable_SetStop(t *testing.T) {
	pb := NewDisable().SetStop(true)
	if pb.GetArg(ArgStop) != "true" {
		t.Errorf("Expected ArgStop 'true', got '%s'", pb.GetArg(ArgStop))
	}
}

func TestDisable_ShouldStop(t *testing.T) {
	pb := NewDisable()
	if pb.shouldStop() {
		t.Error("Expected shouldStop=false by default")
	}
	pb.SetStop(true)
	if !pb.shouldStop() {
		t.Error("Expected shouldStop=true after SetStop(true)")
	}
}

// ---------------------------------------------------------------------------
// Method chaining preserves concrete type (shared across all skills)
// ---------------------------------------------------------------------------

func TestMethodChaining_PreservesType(t *testing.T) {
	// Status
	s := NewStatus().
		SetService("caddy").
		SetID("custom").
		SetDescription("desc").
		SetTimeout(30 * time.Second)
	if _, ok := s.(*Status); !ok {
		t.Error("SetService should preserve *Status type")
	}
	if s.GetID() != "custom" {
		t.Error("Chaining should set ID")
	}

	// Restart
	r := NewRestart().
		SetService("caddy").
		SetID("custom").
		SetDescription("desc").
		SetTimeout(30 * time.Second)
	if _, ok := r.(*Restart); !ok {
		t.Error("SetService should preserve *Restart type")
	}

	// Reload
	l := NewReload().
		SetService("caddy").
		SetID("custom").
		SetDescription("desc").
		SetTimeout(30 * time.Second)
	if _, ok := l.(*Reload); !ok {
		t.Error("SetService should preserve *Reload type")
	}

	// Enable
	e := NewEnable().
		SetService("caddy").
		SetStart(true).
		SetID("custom").
		SetDescription("desc").
		SetTimeout(30 * time.Second)
	if _, ok := e.(*Enable); !ok {
		t.Error("SetService should preserve *Enable type")
	}

	// Disable
	d := NewDisable().
		SetService("caddy").
		SetStop(true).
		SetID("custom").
		SetDescription("desc").
		SetTimeout(30 * time.Second)
	if _, ok := d.(*Disable); !ok {
		t.Error("SetService should preserve *Disable type")
	}

	// IsActive
	a := NewIsActive().
		SetService("caddy").
		SetID("custom").
		SetDescription("desc").
		SetTimeout(30 * time.Second)
	if _, ok := a.(*IsActive); !ok {
		t.Error("SetService should preserve *IsActive type")
	}

	// DaemonReload
	dr := NewDaemonReload().
		SetID("custom").
		SetDescription("desc").
		SetTimeout(30 * time.Second)
	if _, ok := dr.(*DaemonReload); !ok {
		t.Error("Chaining should preserve *DaemonReload type")
	}
}

// ---------------------------------------------------------------------------
// SetArg returns RunnableInterface (per the interface contract), not the
// concrete type. Verify it returns a non-nil value and that the arg is set.
// SetService / SetStart / SetStop DO return the concrete type — those are
// tested separately below.
// ---------------------------------------------------------------------------

func TestSetArg_ReturnsNonNilAndSetsArg(t *testing.T) {
	cases := []struct {
		name string
		call func() types.RunnableInterface
	}{
		{"Status", func() types.RunnableInterface { return NewStatus().SetArg(ArgService, "caddy") }},
		{"IsActive", func() types.RunnableInterface { return NewIsActive().SetArg(ArgService, "caddy") }},
		{"DaemonReload", func() types.RunnableInterface { return NewDaemonReload().SetArg("x", "y") }},
		{"Restart", func() types.RunnableInterface { return NewRestart().SetArg(ArgService, "caddy") }},
		{"Reload", func() types.RunnableInterface { return NewReload().SetArg(ArgService, "caddy") }},
		{"Enable", func() types.RunnableInterface { return NewEnable().SetArg(ArgService, "caddy") }},
		{"Disable", func() types.RunnableInterface { return NewDisable().SetArg(ArgService, "caddy") }},
	}

	for _, c := range cases {
		result := c.call()
		if result == nil {
			t.Errorf("%s: SetArg returned nil", c.name)
		}
	}
}

// TestSetService_ReturnsConcreteType verifies that SetService (the package-
// specific fluent setter) returns the concrete struct type, not just
// RunnableInterface. This is what enables .SetService("caddy").SetID("x")
// chaining without type assertions.
func TestSetService_ReturnsConcreteType(t *testing.T) {
	if _, ok := any(NewStatus().SetService("caddy")).(*Status); !ok {
		t.Error("Status.SetService should return *Status")
	}
	if _, ok := any(NewIsActive().SetService("caddy")).(*IsActive); !ok {
		t.Error("IsActive.SetService should return *IsActive")
	}
	if _, ok := any(NewRestart().SetService("caddy")).(*Restart); !ok {
		t.Error("Restart.SetService should return *Restart")
	}
	if _, ok := any(NewReload().SetService("caddy")).(*Reload); !ok {
		t.Error("Reload.SetService should return *Reload")
	}
	if _, ok := any(NewEnable().SetService("caddy")).(*Enable); !ok {
		t.Error("Enable.SetService should return *Enable")
	}
	if _, ok := any(NewDisable().SetService("caddy")).(*Disable); !ok {
		t.Error("Disable.SetService should return *Disable")
	}
}

// TestSetStart_SetStop_ReturnConcreteType verifies the bool setters return
// the concrete type for chaining.
func TestSetStart_SetStop_ReturnConcreteType(t *testing.T) {
	if _, ok := any(NewEnable().SetStart(true)).(*Enable); !ok {
		t.Error("Enable.SetStart should return *Enable")
	}
	if _, ok := any(NewDisable().SetStop(true)).(*Disable); !ok {
		t.Error("Disable.SetStop should return *Disable")
	}
}

// ---------------------------------------------------------------------------
// Mock-based tests (using skilltest helpers)
// ---------------------------------------------------------------------------

// exitErr creates a *ssh.ExitError for testing command exit failures.
// The exit code doesn't matter — what matters is that errors.As matches
// the *ssh.ExitError type, distinguishing command exits from SSH failures.
func exitErr() error {
	return ssh.NewExitError()
}

// connErr creates a connection error for testing SSH failures.
func connErr() error {
	return fmt.Errorf("connection refused")
}

// --- Enable.Check idempotency ---

func TestEnable_Check_AlreadyEnabled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectCommand("systemctl is-enabled 'caddy'", "enabled")

	pb := NewEnable().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if needs {
		t.Error("Expected Check to return false for already-enabled unit")
	}
}

func TestEnable_Check_AlreadyEnabled_NotActive(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectCommand("systemctl is-enabled 'caddy'", "enabled")
	test.ExpectError("systemctl is-active 'caddy'", exitErr())

	pb := NewEnable().SetService("caddy").SetStart(true)
	pb.SetNodeConfig(test.Config())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true for enabled-but-not-active unit with start")
	}
}

func TestEnable_Check_NotEnabled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", exitErr())

	pb := NewEnable().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true for not-enabled unit")
	}
}

func TestEnable_Check_SSHErrors_IsEnabled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", connErr())

	pb := NewEnable().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	_, err := pb.Check()
	if err == nil {
		t.Error("Expected SSH connection error to be propagated")
	}
}

func TestEnable_Check_SSHErrors_IsActive(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectCommand("systemctl is-enabled 'caddy'", "enabled")
	test.ExpectError("systemctl is-active 'caddy'", connErr())

	pb := NewEnable().SetService("caddy").SetStart(true)
	pb.SetNodeConfig(test.Config())

	_, err := pb.Check()
	if err == nil {
		t.Error("Expected SSH connection error from is-active to be propagated")
	}
}

// --- Enable.Run idempotency and command construction ---

func TestEnable_Run_AlreadyEnabled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectCommand("systemctl is-enabled 'caddy'", "enabled")

	pb := NewEnable().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false for already-enabled unit")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	test.AssertCommandNotRun("systemctl enable 'caddy'")
}

func TestEnable_Run_WithStart_CommandConstruction(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", exitErr())
	test.ExpectCommand("systemctl enable 'caddy' && systemctl start 'caddy'", "")

	pb := NewEnable().SetService("caddy").SetStart(true)
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	test.AssertCommandRun("systemctl enable 'caddy' && systemctl start 'caddy'")
}

func TestEnable_Run_CommandConstruction_NoStart(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", exitErr())
	test.ExpectCommand("systemctl enable 'caddy'", "")

	pb := NewEnable().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	test.AssertCommandRun("systemctl enable 'caddy'")
	test.AssertCommandNotRun("systemctl enable 'caddy' && systemctl start 'caddy'")
}

// --- Disable.Check idempotency ---

func TestDisable_Check_AlreadyDisabled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", exitErr())

	pb := NewDisable().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if needs {
		t.Error("Expected Check to return false for already-disabled unit")
	}
}

func TestDisable_Check_AlreadyDisabled_StillActive(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", exitErr())
	test.ExpectCommand("systemctl is-active 'caddy'", "active")

	pb := NewDisable().SetService("caddy").SetStop(true)
	pb.SetNodeConfig(test.Config())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !needs {
		t.Error("Expected Check to return true for disabled-but-still-active unit with stop")
	}
}

func TestDisable_Check_AlreadyDisabled_NotActive(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", exitErr())
	test.ExpectError("systemctl is-active 'caddy'", exitErr())

	pb := NewDisable().SetService("caddy").SetStop(true)
	pb.SetNodeConfig(test.Config())

	needs, err := pb.Check()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if needs {
		t.Error("Expected Check to return false for disabled-and-not-active unit")
	}
}

func TestDisable_Check_SSHErrors_IsEnabled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", connErr())

	pb := NewDisable().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	_, err := pb.Check()
	if err == nil {
		t.Error("Expected SSH connection error to be propagated")
	}
}

func TestDisable_Check_SSHErrors_IsActive(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", exitErr())
	test.ExpectError("systemctl is-active 'caddy'", connErr())

	pb := NewDisable().SetService("caddy").SetStop(true)
	pb.SetNodeConfig(test.Config())

	_, err := pb.Check()
	if err == nil {
		t.Error("Expected SSH connection error from is-active to be propagated")
	}
}

// --- Disable.Run idempotency and command construction ---

func TestDisable_Run_AlreadyDisabled(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", exitErr())

	pb := NewDisable().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false for already-disabled unit")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	test.AssertCommandNotRun("systemctl disable 'caddy'")
}

func TestDisable_Run_WithStop_CommandConstruction(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectCommand("systemctl is-enabled 'caddy'", "enabled")
	test.ExpectCommand("systemctl disable 'caddy' && systemctl stop 'caddy'", "")

	pb := NewDisable().SetService("caddy").SetStop(true)
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	test.AssertCommandRun("systemctl disable 'caddy' && systemctl stop 'caddy'")
}

func TestDisable_Run_CommandConstruction_NoStop(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectCommand("systemctl is-enabled 'caddy'", "enabled")
	test.ExpectCommand("systemctl disable 'caddy'", "")

	pb := NewDisable().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	test.AssertCommandRun("systemctl disable 'caddy'")
	test.AssertCommandNotRun("systemctl disable 'caddy' && systemctl stop 'caddy'")
}

// --- Reload.Run fallback ---

func TestReload_Run_ReloadSucceeds(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectCommand("systemctl reload 'caddy'", "")

	pb := NewReload().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.Details["method"] != "reload" {
		t.Errorf("Expected method 'reload', got '%s'", result.Details["method"])
	}
	test.AssertCommandNotRun("systemctl restart 'caddy'")
}

func TestReload_Run_ReloadFails_RestartSucceeds(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl reload 'caddy'", exitErr())
	test.ExpectCommand("systemctl restart 'caddy'", "")

	pb := NewReload().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed=true")
	}
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
	if result.Details["method"] != "restart" {
		t.Errorf("Expected method 'restart', got '%s'", result.Details["method"])
	}
	test.AssertCommandRun("systemctl restart 'caddy'")
}

func TestReload_Run_BothFail(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl reload 'caddy'", exitErr())
	test.ExpectError("systemctl restart 'caddy'", exitErr())

	pb := NewReload().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when both reload and restart fail")
	}
	if result.Error == nil {
		t.Error("Expected error when both reload and restart fail")
	}
}

func TestReload_Run_SSHErrors_NoFallback(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl reload 'caddy'", connErr())

	pb := NewReload().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false for SSH connection error")
	}
	if result.Error == nil {
		t.Error("Expected error to be propagated for SSH connection failure")
	}
	test.AssertCommandNotRun("systemctl restart 'caddy'")
}

// --- Enable.Run with SSH error during Check ---

func TestEnable_Run_CheckSSHErrors(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", connErr())

	pb := NewEnable().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when Check fails with SSH error")
	}
	if result.Error == nil {
		t.Error("Expected SSH error to be propagated from Run")
	}
	test.AssertCommandNotRun("systemctl enable 'caddy'")
}

// --- Disable.Run with SSH error during Check ---

func TestDisable_Run_CheckSSHErrors(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-enabled 'caddy'", connErr())

	pb := NewDisable().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false when Check fails with SSH error")
	}
	if result.Error == nil {
		t.Error("Expected SSH error to be propagated from Run")
	}
	test.AssertCommandNotRun("systemctl disable 'caddy'")
}

// --- IsActive.Run with SSH connection error ---

func TestIsActive_Run_SSHErrors(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-active 'caddy'", connErr())

	pb := NewIsActive().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false for SSH connection error")
	}
	if result.Error == nil {
		t.Error("Expected SSH connection error to be propagated")
	}
}

func TestIsActive_Run_ExitError_NoError(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()
	test.Setup()

	test.ExpectError("systemctl is-active 'caddy'", exitErr())

	pb := NewIsActive().SetService("caddy")
	pb.SetNodeConfig(test.Config())

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed=false for read-only is-active")
	}
	if result.Error != nil {
		t.Errorf("Expected no error for exit error (inactive unit), got: %v", result.Error)
	}
}
