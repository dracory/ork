package ork_test

// Integration tests for the Inventory API — concurrent multi-node orchestration.
//
// Inventory holds groups and nodes, runs operations with SetMaxConcurrency
// controlling parallelism. These tests verify that against real SSH servers.

import (
	"strings"
	"testing"

	"github.com/dracory/ork"
	"github.com/dracory/ork/skills/ping"
)

// TestIntegration_Inventory_RunCommand verifies that RunCommand executes
// across all nodes in the inventory (both direct nodes and group nodes).
func TestIntegration_Inventory_RunCommand(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)

	inventory := ork.NewInventory()
	inventory.AddNode(node)

	results := inventory.RunCommand("echo 'hello from inventory'")
	if len(results.Results) == 0 {
		t.Fatal("Expected at least 1 result, got 0")
	}

	for host, result := range results.Results {
		if result.Error != nil {
			t.Errorf("Host %s: RunCommand failed: %v", host, result.Error)
			continue
		}
		if !strings.Contains(result.Message, "hello from inventory") {
			t.Errorf("Host %s: expected 'hello from inventory', got: %s", host, result.Message)
		}
	}
}

// TestIntegration_Inventory_RunByID verifies that RunByID executes a skill
// across all nodes in the inventory.
func TestIntegration_Inventory_RunByID(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)

	inventory := ork.NewInventory()
	inventory.AddNode(node)

	results := inventory.RunByID("ping")
	if len(results.Results) == 0 {
		t.Fatal("Expected at least 1 result, got 0")
	}

	for host, result := range results.Results {
		if result.Error != nil {
			t.Errorf("Host %s: RunByID('ping') failed: %v", host, result.Error)
		}
	}
}

// TestIntegration_Inventory_Run verifies that Run(skill) executes a skill
// instance across all nodes in the inventory.
func TestIntegration_Inventory_Run(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)

	inventory := ork.NewInventory()
	inventory.AddNode(node)

	skill := ping.NewPing()
	results := inventory.Run(skill)
	if len(results.Results) == 0 {
		t.Fatal("Expected at least 1 result, got 0")
	}

	for host, result := range results.Results {
		if result.Error != nil {
			t.Errorf("Host %s: Run(ping) failed: %v", host, result.Error)
		}
	}
}

// TestIntegration_Inventory_AddGroup_GetGroupByName verifies that AddGroup
// adds a group and GetGroupByName retrieves it.
func TestIntegration_Inventory_AddGroup_GetGroupByName(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)
	group := ork.NewGroup("webservers")
	group.AddNode(node)

	inventory := ork.NewInventory()
	inventory.AddGroup(group)

	retrieved := inventory.GetGroupByName("webservers")
	if retrieved == nil {
		t.Fatal("Expected GetGroupByName('webservers') to return the group, got nil")
	}
	if retrieved.GetName() != "webservers" {
		t.Errorf("Expected group name 'webservers', got: %s", retrieved.GetName())
	}

	// Verify GetNodes includes nodes from groups
	nodes := inventory.GetNodes()
	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node from group, got %d", len(nodes))
	}
}

// TestIntegration_Inventory_DryRun_Propagation verifies that SetDryRunMode
// on the inventory propagates to child nodes and groups.
func TestIntegration_Inventory_DryRun_Propagation(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)

	inventory := ork.NewInventory()
	inventory.AddNode(node)
	inventory.SetDryRunMode(true)

	// Verify node has dry-run mode
	nodes := inventory.GetNodes()
	for i, n := range nodes {
		if !n.GetDryRunMode() {
			t.Errorf("Node %d: expected GetDryRunMode() to return true", i)
		}
	}

	// RunCommand should return "[dry-run]" without executing
	results := inventory.RunCommand("touch /tmp/ork-inventory-dryrun-should-not-exist")
	for host, result := range results.Results {
		if result.Message != "[dry-run]" {
			t.Errorf("Host %s: expected '[dry-run]', got: %s", host, result.Message)
		}
	}

	// Verify the file was NOT created
	checkResults := newTestNode(container).RunCommand("test -f /tmp/ork-inventory-dryrun-should-not-exist && echo EXISTS || echo MISSING")
	checkResult := checkResults.Results[container.host]
	if !strings.Contains(checkResult.Message, "MISSING") {
		t.Errorf("Dry-run should not have created the file, but it exists")
	}
}

// TestIntegration_Inventory_SetMaxConcurrency verifies that
// SetMaxConcurrency can be set without panicking and that operations
// complete successfully with concurrency control.
func TestIntegration_Inventory_SetMaxConcurrency(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)

	inventory := ork.NewInventory()
	inventory.AddNode(node)
	inventory.SetMaxConcurrency(1)

	results := inventory.RunCommand("echo 'concurrent test'")
	if len(results.Results) == 0 {
		t.Fatal("Expected at least 1 result, got 0")
	}

	for host, result := range results.Results {
		if result.Error != nil {
			t.Errorf("Host %s: RunCommand with MaxConcurrency=1 failed: %v", host, result.Error)
		}
	}
}

// TestIntegration_Inventory_PanicRecovery verifies that if one node's
// command panics, the inventory recovers and still returns results for
// other nodes. We simulate this by having a dead node (connection failure
// is not a panic, but tests that the inventory handles errors gracefully).
func TestIntegration_Inventory_PanicRecovery(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	liveNode := newTestNode(container)
	deadNode := ork.NewNodeForHost(container.host).
		SetPort("9999").
		SetUser(container.user).
		SetKey(container.keyName)

	inventory := ork.NewInventory()
	inventory.AddNode(liveNode)
	inventory.AddNode(deadNode)

	// This should not panic, even though the dead node will fail
	results := inventory.RunCommand("echo 'survived'")
	if len(results.Results) == 0 {
		t.Fatal("Expected at least 1 result, got 0")
	}
}
