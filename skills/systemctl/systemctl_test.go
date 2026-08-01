package systemctl

import (
	"log/slog"
	"testing"
	"time"

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
