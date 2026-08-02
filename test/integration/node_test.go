package ork_test

// Integration tests for the Ork Node API.
//
// These tests exercise the Node API against a real SSH server spun up via
// testcontainers-go. See setup_test.go for the container setup helpers.

import (
	"strings"
	"testing"

	"github.com/dracory/ork"
)

// TestIntegration_Node_ConnectRunClose tests Node lifecycle with real SSH.
// This is the canonical end-to-end smoke test: connect, run a command,
// verify output, close, verify disconnected.
func TestIntegration_Node_ConnectRunClose(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	// Test Connect
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Verify connected state
	if !node.IsConnected() {
		t.Error("Expected IsConnected() to return true after Connect()")
	}

	// Test RunCommand with persistent connection
	results := node.RunCommand("echo 'test1'")
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Run failed: %v", result.Error)
	}
	if !strings.Contains(result.Message, "test1") {
		t.Errorf("Expected output to contain 'test1', got: %s", result.Message)
	}

	// Test Close
	err = node.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify disconnected state
	if node.IsConnected() {
		t.Error("Expected IsConnected() to return false after Close()")
	}
}

// TestIntegration_Node_PersistentConnectionReuse tests connection reuse.
// After Connect(), multiple RunCommand() calls must share a single SSH
// session and the connection must remain usable throughout.
func TestIntegration_Node_PersistentConnectionReuse(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Execute multiple commands on same connection
	commands := []string{
		"echo 'command1'",
		"echo 'command2'",
		"echo 'command3'",
		"pwd",
		"whoami",
	}

	for i, cmd := range commands {
		results := node.RunCommand(cmd)
		result := results.Results[container.host]
		if result.Error != nil {
			t.Errorf("Run %d failed: %v", i+1, result.Error)
			continue
		}
		t.Logf("Command %d output: %s", i+1, result.Message)
	}

	// Verify still connected after multiple operations
	if !node.IsConnected() {
		t.Error("Expected connection to remain active after multiple Run calls")
	}
}

// TestIntegration_Node_WithoutPersistentConnection tests one-time connections.
// RunCommand is called without a prior Connect(); the Node should open a
// one-shot connection, run the command, and close the connection. After the
// call, IsConnected() must return false.
func TestIntegration_Node_WithoutPersistentConnection(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	// Run without calling Connect() - should create one-time connection
	results := node.RunCommand("echo 'one-time'")
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Run failed: %v", result.Error)
	}

	if !strings.Contains(result.Message, "one-time") {
		t.Errorf("Expected output to contain 'one-time', got: %s", result.Message)
	}

	// Verify not connected (one-time connection was closed)
	if node.IsConnected() {
		t.Error("Expected IsConnected() to return false after one-time Run")
	}
}

// TestIntegration_Node_Playbook tests playbook execution via Node.
// The "ping" skill is registered by default in the global skill registry
// (see NewDefaultRegistry), so RunByID("ping") must succeed against a
// live SSH server.
func TestIntegration_Node_Playbook(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Test ping playbook
	results := node.RunByID("ping")
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Playbook('ping') failed: %v", result.Error)
	}
}

// TestIntegration_MultipleOperations tests complex workflows on a persistent
// connection: connect, run a command, mutate node args, run another command,
// execute a playbook, run whoami, and verify the connection stays alive
// throughout the whole sequence.
func TestIntegration_MultipleOperations(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Test 1: Run command
	results1 := node.RunCommand("echo 'step1'")
	result1 := results1.Results[container.host]
	if result1.Error != nil {
		t.Fatalf("Step 1 failed: %v", result1.Error)
	}
	if !strings.Contains(result1.Message, "step1") {
		t.Errorf("Step 1: expected 'step1' in output, got: %s", result1.Message)
	}

	// Test 2: Update configuration
	node.SetArg("test", "value")

	// Test 3: Run another command
	results2 := node.RunCommand("echo 'step2'")
	result2 := results2.Results[container.host]
	if result2.Error != nil {
		t.Fatalf("Step 2 failed: %v", result2.Error)
	}
	if !strings.Contains(result2.Message, "step2") {
		t.Errorf("Step 2: expected 'step2' in output, got: %s", result2.Message)
	}

	// Test 4: Execute playbook
	results3 := node.RunByID("ping")
	result3 := results3.Results[container.host]
	if result3.Error != nil {
		t.Fatalf("Playbook execution failed: %v", result3.Error)
	}

	// Test 5: Run final command
	results4 := node.RunCommand("whoami")
	result4 := results4.Results[container.host]
	if result4.Error != nil {
		t.Fatalf("Step 3 failed: %v", result4.Error)
	}
	if !strings.Contains(result4.Message, container.user) {
		t.Errorf("Step 3: expected '%s' in output, got: %s", container.user, result4.Message)
	}

	// Verify connection remained active throughout
	if !node.IsConnected() {
		t.Error("Expected connection to remain active throughout operations")
	}
}
