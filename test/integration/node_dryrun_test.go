package ork_test

// Integration tests for dry-run mode.
//
// When SetDryRunMode(true) is called, the node must NOT execute any commands
// on the remote server. Instead, RunCommand returns "[dry-run]", and skills
// return a "Would ..." message. These tests verify that against a real SSH
// server — the key assertion is that no side effects occur on the container.

import (
	"strings"
	"testing"

	"github.com/dracory/ork"
	"github.com/dracory/ork/skills/ping"
)

// TestIntegration_Node_DryRun_RunCommand verifies that RunCommand in dry-run
// mode returns "[dry-run]" and does NOT execute the command on the server.
func TestIntegration_Node_DryRun_RunCommand(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName).
		SetDryRunMode(true)

	// Run a command that would create a file if executed
	results := node.RunCommand("touch /tmp/ork-dryrun-should-not-exist")
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("DryRun RunCommand returned error: %v", result.Error)
	}
	if result.Message != "[dry-run]" {
		t.Errorf("Expected '[dry-run]', got: %s", result.Message)
	}

	// CRITICAL: verify the file was NOT created on the container.
	// We need a non-dry-run node to check.
	checkNode := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	verifyResults := checkNode.RunCommand("test -f /tmp/ork-dryrun-should-not-exist && echo EXISTS || echo MISSING")
	verifyResult := verifyResults.Results[container.host]
	if verifyResult.Error != nil {
		t.Fatalf("Verification command failed: %v", verifyResult.Error)
	}
	if !strings.Contains(verifyResult.Message, "MISSING") {
		t.Errorf("Dry-run should not have created the file, but it exists. Output: %s", verifyResult.Message)
	}
}

// TestIntegration_Node_DryRun_RunByID verifies that RunByID in dry-run mode
// does NOT execute the skill on the server.
func TestIntegration_Node_DryRun_RunByID(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName).
		SetDryRunMode(true)

	results := node.RunByID("ping")
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("DryRun RunByID returned error: %v", result.Error)
	}
	// In dry-run mode, the ping skill returns "Would ping: <host>"
	if !strings.Contains(result.Message, "Would ping") {
		t.Errorf("Expected message to contain 'Would ping', got: %s", result.Message)
	}
}

// TestIntegration_Node_DryRun_Run verifies that Run(skill) in dry-run mode
// does NOT execute the skill on the server.
func TestIntegration_Node_DryRun_Run(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName).
		SetDryRunMode(true)

	skill := ping.NewPing()
	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("DryRun Run(skill) returned error: %v", result.Error)
	}
	// In dry-run mode, the ping skill returns "Would ping: <host>"
	if !strings.Contains(result.Message, "Would ping") {
		t.Errorf("Expected message to contain 'Would ping', got: %s", result.Message)
	}
}

// TestIntegration_Node_DryRun_GetDryRunMode verifies the getter returns
// what was set.
func TestIntegration_Node_DryRun_GetDryRunMode(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	if node.GetDryRunMode() {
		t.Error("Expected GetDryRunMode() to return false by default")
	}

	node.SetDryRunMode(true)
	if !node.GetDryRunMode() {
		t.Error("Expected GetDryRunMode() to return true after SetDryRunMode(true)")
	}

	node.SetDryRunMode(false)
	if node.GetDryRunMode() {
		t.Error("Expected GetDryRunMode() to return false after SetDryRunMode(false)")
	}
}
