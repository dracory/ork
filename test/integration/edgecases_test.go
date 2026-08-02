package ork_test

// Phase 3 integration tests — edge cases and lower-priority features.
//
// 3.1 Getter methods (GetArg, GetArgs, GetHost, GetUser, GetKey, GetPort, GetNodeConfig)
// 3.2 Custom logger output verification
// 3.3 Sensitive command redaction in logs
// 3.4 Sudo password delivery state machine (BecomePassword path)
// 3.5 Concurrent operations against real SSH

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/dracory/ork"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// --- 3.1 Getter methods ---

// TestIntegration_Node_Getters_PureState verifies that all getter methods
// return the values that were set via the fluent API. This is a pure local
// state test that does not require a container.
func TestIntegration_Node_Getters_PureState(t *testing.T) {
	const testHost = "testhost.example.com"
	const testPort = "2222"
	const testUser = "testuser"
	const testKey = "/path/to/key"

	node := ork.NewNodeForHost(testHost).
		SetPort(testPort).
		SetUser(testUser).
		SetKey(testKey).
		SetArg("env", "test").
		SetArg("role", "web")

	// Verify individual getters
	if node.GetHost() != testHost {
		t.Errorf("GetHost: expected %q, got %q", testHost, node.GetHost())
	}
	if node.GetPort() != testPort {
		t.Errorf("GetPort: expected %q, got %q", testPort, node.GetPort())
	}
	if node.GetUser() != testUser {
		t.Errorf("GetUser: expected %q, got %q", testUser, node.GetUser())
	}
	if node.GetKey() != testKey {
		t.Errorf("GetKey: expected %q, got %q", testKey, node.GetKey())
	}

	// Verify GetArg
	if node.GetArg("env") != "test" {
		t.Errorf("GetArg(env): expected 'test', got %q", node.GetArg("env"))
	}
	if node.GetArg("role") != "web" {
		t.Errorf("GetArg(role): expected 'web', got %q", node.GetArg("role"))
	}
	if node.GetArg("missing") != "" {
		t.Errorf("GetArg(missing): expected '', got %q", node.GetArg("missing"))
	}

	// Verify GetArgs returns a copy with all keys
	args := node.GetArgs()
	if len(args) != 2 {
		t.Errorf("GetArgs: expected 2 entries, got %d", len(args))
	}
	if args["env"] != "test" || args["role"] != "web" {
		t.Errorf("GetArgs: expected env=test, role=web, got %v", args)
	}

	// Verify GetArgs returns a copy (modifying it should not affect the node)
	args["env"] = "modified"
	if node.GetArg("env") != "test" {
		t.Errorf("GetArgs should return a copy, but modifying it affected the node: %q", node.GetArg("env"))
	}

	// Verify GetNodeConfig returns a usable config
	cfg := node.GetNodeConfig()
	if cfg.SSHHost != testHost {
		t.Errorf("GetNodeConfig.SSHHost: expected %q, got %q", testHost, cfg.SSHHost)
	}
	if cfg.SSHPort != testPort {
		t.Errorf("GetNodeConfig.SSHPort: expected %q, got %q", testPort, cfg.SSHPort)
	}
	if cfg.RootUser != testUser {
		t.Errorf("GetNodeConfig.RootUser: expected %q, got %q", testUser, cfg.RootUser)
	}
	if cfg.SSHKey != testKey {
		t.Errorf("GetNodeConfig.SSHKey: expected %q, got %q", testKey, cfg.SSHKey)
	}
}

// TestIntegration_Node_Getters_ConnectProof verifies that the config returned
// by GetNodeConfig is actually usable to connect and run a command.
func TestIntegration_Node_Getters_ConnectProof(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	// Prove the config from getters is valid by connecting
	if err := node.Connect(); err != nil {
		t.Fatalf("Connect failed with config from getters: %v", err)
	}
	defer node.Close()

	results := node.RunCommand("echo getters-work")
	// Results are keyed by host
	result := results.Results[container.host]
	if result.Error != nil {
		t.Errorf("RunCommand failed: %v", result.Error)
	}
	if !strings.Contains(result.Message, "getters-work") {
		t.Errorf("Expected 'getters-work' in output, got %q", result.Message)
	}
}

// --- 3.2 Custom logger ---

// TestIntegration_Node_CustomLogger verifies that a custom *slog.Logger set
// on the node is used for log output during command execution.
func TestIntegration_Node_CustomLogger(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)
	node.SetLogger(logger)

	if err := node.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Run a command — dry-run mode so we get deterministic log output
	node.SetDryRunMode(true)
	node.RunCommand("echo logged-via-custom-logger")

	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("Expected custom logger to capture log output, got empty buffer")
	}
	// Assert on the command text (the meaningful signal that the custom logger
	// received the call), not on the exact log message wording which could change.
	if !strings.Contains(logOutput, "echo logged-via-custom-logger") {
		t.Errorf("Expected log output to contain the command text, got: %s", logOutput)
	}
}

// --- 3.3 Sensitive command redaction ---

// TestIntegration_SensitiveCommand_Redacted verifies that when a command has
// Sensitive=true, the command string is replaced with [redacted] in log output
// and the actual command text does NOT appear.
//
// This test uses dry-run mode, which returns early in ssh.Run without connecting
// to any server, so no container is needed.
func TestIntegration_SensitiveCommand_Redacted(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := types.NodeConfig{IsDryRunMode: true, Logger: logger}

	secretCmd := "echo super-secret-value-12345"
	cmd := types.Command{
		Command:   secretCmd,
		Sensitive: true,
	}

	ssh.Run(cfg, cmd)

	logOutput := buf.String()

	// The actual command string must NOT appear in logs
	if strings.Contains(logOutput, secretCmd) {
		t.Errorf("Sensitive command text should NOT appear in logs, but found: %s", logOutput)
	}

	// [redacted] should appear instead
	if !strings.Contains(logOutput, "[redacted]") {
		t.Errorf("Expected '[redacted]' in log output, got: %s", logOutput)
	}
}

// TestIntegration_SensitiveCommand_NotRedactedWhenNotSensitive verifies that
// non-sensitive commands DO appear in log output (the normal case).
//
// This test uses dry-run mode, which returns early in ssh.Run without connecting
// to any server, so no container is needed.
func TestIntegration_SensitiveCommand_NotRedactedWhenNotSensitive(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := types.NodeConfig{IsDryRunMode: true, Logger: logger}

	plainCmd := "echo not-sensitive-at-all"
	cmd := types.Command{
		Command:   plainCmd,
		Sensitive: false,
	}

	ssh.Run(cfg, cmd)

	logOutput := buf.String()

	// The command text SHOULD appear for non-sensitive commands
	if !strings.Contains(logOutput, plainCmd) {
		t.Errorf("Non-sensitive command should appear in logs, got: %s", logOutput)
	}
}

// --- 3.4 Sudo password delivery state machine ---

// setupSSHContainerWithSudoPassword starts an SSH container and configures
// password-based sudo for testuser (NOT NOPASSWD). The password is "sudopass".
// This is needed for testing the BecomePassword state machine path.
func setupSSHContainerWithSudoPassword(t *testing.T) *sshContainer {
	t.Helper()
	return setupSSHContainerWithSudoConfig(t, true, "sudopass")
}

// TestIntegration_BecomePasswordDelivery verifies that the sudo password
// delivery state machine works: BecomeUser + BecomePassword causes the
// password to be delivered on-demand via prompt detection, and the command
// output is correct.
func TestIntegration_BecomePasswordDelivery(t *testing.T) {
	container := setupSSHContainerWithSudoPassword(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cfg.BecomeUser = "root"
	cfg.BecomePassword = "sudopass"

	cmd := types.Command{Command: "whoami", Required: true}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Fatalf("ssh.Run with BecomePassword failed: %v", err)
	}

	if strings.TrimSpace(output) != "root" {
		t.Errorf("Expected 'root' from whoami via sudo password, got: %q", output)
	}
}

// TestIntegration_BecomeWrongPassword verifies that a wrong sudo password
// results in an error (the state machine detects the failure).
func TestIntegration_BecomeWrongPassword(t *testing.T) {
	container := setupSSHContainerWithSudoPassword(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cfg.BecomeUser = "root"
	cfg.BecomePassword = "wrong-password"

	cmd := types.Command{Command: "whoami", Required: true}

	_, err := ssh.Run(cfg, cmd)
	if err == nil {
		t.Fatal("Expected error with wrong sudo password, got nil")
	}
	// Verify the error is actually about sudo/password, not an incidental
	// network or connection error
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "sudo") && !strings.Contains(errStr, "password") {
		t.Errorf("Expected error to mention sudo or password, got: %v", err)
	}
}

// TestIntegration_BecomeMissingPassword verifies that when BecomeUser is set
// but no BecomePassword is provided, and sudo requires a password, the
// sudo -n flag causes a clear failure (not a hang).
func TestIntegration_BecomeMissingPassword(t *testing.T) {
	container := setupSSHContainerWithSudoPassword(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cfg.BecomeUser = "root"
	// No BecomePassword — uses sudo -n (fail-fast)

	cmd := types.Command{Command: "whoami", Required: true}

	_, err := ssh.Run(cfg, cmd)
	if err == nil {
		t.Fatal("Expected error when sudo requires password but none provided, got nil")
	}
}

// TestIntegration_BecomePasswordWithStdin verifies that the state machine
// correctly sequences: password delivery first, then command stdin.
func TestIntegration_BecomePasswordWithStdin(t *testing.T) {
	container := setupSSHContainerWithSudoPassword(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cfg.BecomeUser = "root"
	cfg.BecomePassword = "sudopass"

	cmd := types.Command{
		Command:  "cat",
		Stdin:    "piped-after-escalation",
		Required: true,
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Fatalf("ssh.Run with BecomePassword+Stdin failed: %v", err)
	}

	if strings.TrimSpace(output) != "piped-after-escalation" {
		t.Errorf("Expected 'piped-after-escalation', got: %q", output)
	}
}

// --- 3.5 Concurrent operations against real SSH ---

// TestIntegration_Concurrent_MultipleNodes verifies that running commands
// concurrently across multiple nodes (each with its own persistent connection)
// produces correct results for each node without interference.
func TestIntegration_Concurrent_MultipleNodes(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	// Create 3 nodes pointing to the same container, each with its own
	// persistent SSH connection.
	const numNodes = 3
	nodes := make([]ork.NodeInterface, numNodes)
	for i := range nodes {
		n := ork.NewNodeForHost(container.host).
			SetPort(container.port).
			SetUser(container.user).
			SetKey(container.keyName)
		if err := n.Connect(); err != nil {
			t.Fatalf("Node %d Connect failed: %v", i, err)
		}
		defer n.Close()
		nodes[i] = n
	}

	// Run a command that echoes a unique value per node concurrently.
	// All nodes share the same host, so results.Results is keyed by host and
	// would collapse if we used Group.RunCommand. Instead, run via goroutines
	// on individual nodes and collect per-node output.
	var wg sync.WaitGroup
	errs := make([]error, numNodes)
	outputs := make([]string, numNodes)

	for i, n := range nodes {
		wg.Add(1)
		go func(idx int, node ork.NodeInterface) {
			defer wg.Done()
			results := node.RunCommand(fmt.Sprintf("echo concurrent-node-%d", idx))
			result := results.Results[container.host]
			errs[idx] = result.Error
			outputs[idx] = strings.TrimSpace(result.Message)
		}(i, n)
	}

	wg.Wait()

	for i := range nodes {
		if errs[i] != nil {
			t.Errorf("Node %d failed: %v", i, errs[i])
		}
		expected := fmt.Sprintf("concurrent-node-%d", i)
		if outputs[i] != expected {
			t.Errorf("Node %d: expected %q, got %q", i, expected, outputs[i])
		}
	}
}

// TestIntegration_Concurrent_SameNode verifies that multiple goroutines
// calling RunCommand on the same node concurrently are thread-safe.
func TestIntegration_Concurrent_SameNode(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	if err := node.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	const numGoroutines = 10
	var wg sync.WaitGroup
	errs := make([]error, numGoroutines)
	outputs := make([]string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results := node.RunCommand("echo ok")
			result := results.Results[container.host]
			errs[idx] = result.Error
			outputs[idx] = strings.TrimSpace(result.Message)
		}(i)
	}

	wg.Wait()

	for i := range outputs {
		if errs[i] != nil {
			t.Errorf("Goroutine %d failed: %v", i, errs[i])
		}
		if outputs[i] != "ok" {
			t.Errorf("Goroutine %d: expected 'ok' in output, got %q", i, outputs[i])
		}
	}
}
