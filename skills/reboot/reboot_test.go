package reboot

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/internal/skilltest"
	"github.com/dracory/ork/types"
)

// TestReboot_Run_DryRun verifies that dry-run mode correctly handles reboot.
func TestReboot_Run_DryRun(t *testing.T) {
	pb := NewReboot()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
		SSHHost:      "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would reboot test.example.com"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestReboot_Run_DryRun_IgnoresWait verifies that dry-run mode short-circuits
// before reading the wait arg, so wait=true does not change the dry-run output.
func TestReboot_Run_DryRun_IgnoresWait(t *testing.T) {
	pb := NewReboot().SetWaitForReconnect(true)

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
		SSHHost:      "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would reboot test.example.com"
	if result.Message != expectedMessage {
		t.Errorf("Expected dry-run message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestReboot_Run_WithMock verifies the no-wait path via the mock SSH client.
// The reboot command is expected to error (the SSH session is killed by reboot);
// the skill should treat that as expected and report "Reboot initiated".
func TestReboot_Run_WithMock(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	// reboot kills the SSH session, so it surfaces as a connection error.
	test.ExpectError("reboot", fmt.Errorf("ssh: connection reset by peer"))

	pb := NewReboot().WithNodeConfig(test.Config())

	result := pb.Run()

	test.AssertResultChanged(result)
	test.AssertResultNoError(result)
	test.AssertCommandRun("reboot")
	test.AssertResultMessageContains(result, "Reboot initiated")
	if result.Details["wait_for_reconnect"] != "false" {
		t.Errorf("Expected wait_for_reconnect=false, got '%s'", result.Details["wait_for_reconnect"])
	}
}

// TestReboot_Run_WithMock_WaitSuccess verifies the wait-for-reconnect success
// path: reboot errors (session drop), then uptime succeeds on the first probe.
func TestReboot_Run_WithMock_WaitSuccess(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectError("reboot", fmt.Errorf("ssh: connection reset by peer"))
	test.ExpectCommand("uptime", " 10:30:01 up 0 days,  1 user,  load average: 0.01, 0.05, 0.00")

	// Keep the test fast: tiny grace period, tiny poll interval, short max wait.
	// Args must be set on the skill itself (GetArg reads the skill's atom, not
	// NodeConfig.Args), so use the typed setters rather than cfg.Args.
	pb := NewReboot().
		SetWaitForReconnect(true).
		SetMaxWaitTime(2 * time.Second).
		SetInitialWait(1 * time.Millisecond).
		SetPollInterval(1 * time.Millisecond).
		WithNodeConfig(test.Config())

	result := pb.Run()

	test.AssertResultChanged(result)
	test.AssertResultNoError(result)
	test.AssertCommandRun("reboot")
	test.AssertCommandRun("uptime")
	test.AssertResultMessageContains(result, "back online")
	if result.Details["wait_for_reconnect"] != "true" {
		t.Errorf("Expected wait_for_reconnect=true, got '%s'", result.Details["wait_for_reconnect"])
	}
	if result.Details["max_wait"] != (2 * time.Second).String() {
		t.Errorf("Expected max_wait='%s', got '%s'", (2 * time.Second).String(), result.Details["max_wait"])
	}
}

// TestReboot_Run_WithMock_WaitTimeout verifies the wait-for-reconnect timeout
// path: reboot errors and uptime never succeeds, so the skill times out.
func TestReboot_Run_WithMock_WaitTimeout(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectError("reboot", fmt.Errorf("ssh: connection reset by peer"))
	// uptime also errors -> server never comes back within the budget.
	test.ExpectError("uptime", fmt.Errorf("ssh: connect: connection refused"))

	pb := NewReboot().
		SetWaitForReconnect(true).
		SetMaxWaitTime(50 * time.Millisecond).
		SetInitialWait(1 * time.Millisecond).
		SetPollInterval(5 * time.Millisecond).
		WithNodeConfig(test.Config())

	start := time.Now()
	result := pb.Run()
	elapsed := time.Since(start)

	test.AssertResultChanged(result)
	test.AssertResultError(result)
	test.AssertCommandRun("reboot")
	test.AssertCommandRun("uptime")
	test.AssertResultMessageContains(result, "timeout")

	// Must not block significantly longer than max-wait.
	if elapsed > 500*time.Millisecond {
		t.Errorf("Expected to bail out near max-wait (50ms), took %v", elapsed)
	}
}

// TestReboot_Run_WithMock_WaitRespectsMaxWaitBudget verifies that max-wait
// bounds the TOTAL wait (initial grace + polling), not just the polling phase.
// With initial-wait larger than max-wait, no polling should occur at all.
func TestReboot_Run_WithMock_WaitRespectsMaxWaitBudget(t *testing.T) {
	test := skilltest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectError("reboot", fmt.Errorf("ssh: connection reset by peer"))
	test.ExpectCommand("uptime", "up")

	// initial-wait exceeds max-wait -> the grace period is capped to max-wait
	// and the loop body never runs (deadline already reached).
	pb := NewReboot().
		SetWaitForReconnect(true).
		SetMaxWaitTime(20 * time.Millisecond).
		SetInitialWait(1 * time.Second).
		SetPollInterval(1 * time.Millisecond).
		WithNodeConfig(test.Config())

	start := time.Now()
	result := pb.Run()
	elapsed := time.Since(start)

	test.AssertResultError(result)
	// uptime must NOT have been polled because the budget was consumed by the
	// (capped) initial grace period.
	test.AssertCommandNotRun("uptime")
	if elapsed > 250*time.Millisecond {
		t.Errorf("Expected to bail out near max-wait (20ms), took %v", elapsed)
	}
}

// TestReboot_Check verifies that Check always returns true.
func TestReboot_Check(t *testing.T) {
	pb := NewReboot()

	cfg := types.NodeConfig{
		Logger:  slog.Default(),
		SSHHost: "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if !needsChange {
		t.Error("Expected Check to return true for reboot operation")
	}
}

// TestReboot_NewReboot verifies that NewReboot creates a properly configured skill.
func TestReboot_NewReboot(t *testing.T) {
	pb := NewReboot()

	if pb.GetID() != "reboot" {
		t.Errorf("Expected ID to be 'reboot', got '%s'", pb.GetID())
	}

	expectedDescription := "Reboot the remote server"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}

	// Verify default values (stored as args)
	if pb.GetArg(ArgWait) != "false" {
		t.Errorf("Expected wait arg 'false', got '%s'", pb.GetArg(ArgWait))
	}

	if pb.GetArg(ArgMaxWait) != DefaultMaxWait {
		t.Errorf("Expected max-wait arg '%s', got '%s'", DefaultMaxWait, pb.GetArg(ArgMaxWait))
	}

	if pb.GetArg(ArgInitialWait) != DefaultInitialWait {
		t.Errorf("Expected initial-wait arg '%s', got '%s'", DefaultInitialWait, pb.GetArg(ArgInitialWait))
	}

	if pb.GetArg(ArgPollInterval) != DefaultPollInterval {
		t.Errorf("Expected poll-interval arg '%s', got '%s'", DefaultPollInterval, pb.GetArg(ArgPollInterval))
	}

	// Verify the parsed defaults resolve correctly
	if pb.getWaitForReconnect() {
		t.Error("Expected getWaitForReconnect to be false by default")
	}

	if pb.getMaxWaitTime() != 5*time.Minute {
		t.Errorf("Expected getMaxWaitTime to be 5 minutes, got %v", pb.getMaxWaitTime())
	}

	if pb.getInitialWait() != 10*time.Second {
		t.Errorf("Expected getInitialWait to be 10 seconds, got %v", pb.getInitialWait())
	}

	if pb.getPollInterval() != 5*time.Second {
		t.Errorf("Expected getPollInterval to be 5 seconds, got %v", pb.getPollInterval())
	}
}

// TestReboot_SetWaitForReconnect verifies the fluent setter stores the arg.
func TestReboot_SetWaitForReconnect(t *testing.T) {
	pb := NewReboot()

	pb.SetWaitForReconnect(true)
	if !pb.getWaitForReconnect() {
		t.Error("Expected getWaitForReconnect to be true after SetWaitForReconnect(true)")
	}
	if pb.GetArg(ArgWait) != "true" {
		t.Errorf("Expected wait arg 'true', got '%s'", pb.GetArg(ArgWait))
	}

	pb.SetWaitForReconnect(false)
	if pb.getWaitForReconnect() {
		t.Error("Expected getWaitForReconnect to be false after SetWaitForReconnect(false)")
	}
}

// TestReboot_SetMaxWaitTime verifies the fluent setter stores the arg and parses it.
func TestReboot_SetMaxWaitTime(t *testing.T) {
	pb := NewReboot()

	pb.SetMaxWaitTime(10 * time.Minute)
	if pb.getMaxWaitTime() != 10*time.Minute {
		t.Errorf("Expected getMaxWaitTime to be 10 minutes, got %v", pb.getMaxWaitTime())
	}
	if pb.GetArg(ArgMaxWait) != (10 * time.Minute).String() {
		t.Errorf("Expected max-wait arg '%s', got '%s'", (10 * time.Minute).String(), pb.GetArg(ArgMaxWait))
	}
}

// TestReboot_SetInitialWait verifies the fluent setter stores the arg and parses it.
func TestReboot_SetInitialWait(t *testing.T) {
	pb := NewReboot()

	pb.SetInitialWait(30 * time.Second)
	if pb.getInitialWait() != 30*time.Second {
		t.Errorf("Expected getInitialWait to be 30 seconds, got %v", pb.getInitialWait())
	}
	if pb.GetArg(ArgInitialWait) != (30 * time.Second).String() {
		t.Errorf("Expected initial-wait arg '%s', got '%s'", (30 * time.Second).String(), pb.GetArg(ArgInitialWait))
	}
}

// TestReboot_SetPollInterval verifies the fluent setter stores the arg and parses it.
func TestReboot_SetPollInterval(t *testing.T) {
	pb := NewReboot()

	pb.SetPollInterval(2 * time.Second)
	if pb.getPollInterval() != 2*time.Second {
		t.Errorf("Expected getPollInterval to be 2 seconds, got %v", pb.getPollInterval())
	}
	if pb.GetArg(ArgPollInterval) != (2 * time.Second).String() {
		t.Errorf("Expected poll-interval arg '%s', got '%s'", (2 * time.Second).String(), pb.GetArg(ArgPollInterval))
	}
}

// TestReboot_GetWaitForReconnect_InvalidArg verifies invalid wait arg falls back to false.
func TestReboot_GetWaitForReconnect_InvalidArg(t *testing.T) {
	pb := NewReboot()
	pb.SetNodeConfig(types.NodeConfig{Logger: slog.Default()})
	pb.SetArg(ArgWait, "not-a-bool")

	if pb.getWaitForReconnect() {
		t.Error("Expected invalid wait arg to resolve to false")
	}
}

// TestReboot_GetMaxWaitTime_InvalidArg verifies invalid max-wait arg falls back to default.
func TestReboot_GetMaxWaitTime_InvalidArg(t *testing.T) {
	pb := NewReboot()
	pb.SetNodeConfig(types.NodeConfig{Logger: slog.Default()})
	pb.SetArg(ArgMaxWait, "not-a-duration")

	if pb.getMaxWaitTime() != 5*time.Minute {
		t.Errorf("Expected invalid max-wait to fall back to 5m, got %v", pb.getMaxWaitTime())
	}
}

// TestReboot_GetInitialWait_InvalidArg verifies invalid initial-wait arg falls back to default.
func TestReboot_GetInitialWait_InvalidArg(t *testing.T) {
	pb := NewReboot()
	pb.SetNodeConfig(types.NodeConfig{Logger: slog.Default()})
	pb.SetArg(ArgInitialWait, "not-a-duration")

	if pb.getInitialWait() != 10*time.Second {
		t.Errorf("Expected invalid initial-wait to fall back to 10s, got %v", pb.getInitialWait())
	}
}

// TestReboot_GetPollInterval_InvalidArg verifies invalid poll-interval arg falls back to default.
func TestReboot_GetPollInterval_InvalidArg(t *testing.T) {
	pb := NewReboot()
	pb.SetNodeConfig(types.NodeConfig{Logger: slog.Default()})
	pb.SetArg(ArgPollInterval, "not-a-duration")

	if pb.getPollInterval() != 5*time.Second {
		t.Errorf("Expected invalid poll-interval to fall back to 5s, got %v", pb.getPollInterval())
	}
}

// TestReboot_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete Reboot type.
func TestReboot_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewReboot()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*Reboot); !ok {
		t.Error("SetArgs should return *Reboot, not just RunnableInterface")
	}
}

// TestReboot_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete Reboot type.
func TestReboot_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewReboot()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*Reboot); !ok {
		t.Error("SetArg should return *Reboot, not just RunnableInterface")
	}
}

// TestReboot_SetID_ReturnsConcreteType verifies that SetID returns the concrete Reboot type.
func TestReboot_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewReboot()

	result := skill.SetID("custom-id")

	if _, ok := result.(*Reboot); !ok {
		t.Error("SetID should return *Reboot, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestReboot_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete Reboot type.
func TestReboot_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewReboot()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*Reboot); !ok {
		t.Error("SetDescription should return *Reboot, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestReboot_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete Reboot type.
func TestReboot_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewReboot()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*Reboot); !ok {
		t.Error("SetTimeout should return *Reboot, not just RunnableInterface")
	}
}

// TestReboot_WithNodeConfig_ReturnsConcreteType verifies that WithNodeConfig
// returns the concrete Reboot type and sets the config.
func TestReboot_WithNodeConfig_ReturnsConcreteType(t *testing.T) {
	skill := NewReboot()

	cfg := types.NodeConfig{SSHHost: "host.example.com", Logger: slog.Default()}
	result := skill.WithNodeConfig(cfg)

	if result != skill {
		t.Error("WithNodeConfig should return the same *Reboot instance")
	}

	if skill.GetNodeConfig().SSHHost != "host.example.com" {
		t.Errorf("Expected SSHHost 'host.example.com', got '%s'", skill.GetNodeConfig().SSHHost)
	}
}

// TestReboot_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestReboot_MethodChaining_PreservesType(t *testing.T) {
	skill := NewReboot().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*Reboot); !ok {
		t.Error("Method chaining should preserve *Reboot type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestReboot_TypedSetters_Chain verifies that the typed fluent setters
// (SetWaitForReconnect, SetMaxWaitTime, SetInitialWait, SetPollInterval)
// chain together and all take effect.
func TestReboot_TypedSetters_Chain(t *testing.T) {
	pb := NewReboot().
		SetWaitForReconnect(true).
		SetMaxWaitTime(7 * time.Minute).
		SetInitialWait(15 * time.Second).
		SetPollInterval(3 * time.Second)

	if !pb.getWaitForReconnect() {
		t.Error("Expected wait to be true after chain")
	}
	if pb.getMaxWaitTime() != 7*time.Minute {
		t.Errorf("Expected max-wait 7m, got %v", pb.getMaxWaitTime())
	}
	if pb.getInitialWait() != 15*time.Second {
		t.Errorf("Expected initial-wait 15s, got %v", pb.getInitialWait())
	}
	if pb.getPollInterval() != 3*time.Second {
		t.Errorf("Expected poll-interval 3s, got %v", pb.getPollInterval())
	}
}
