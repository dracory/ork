package examples

import (
	"testing"

	"github.com/dracory/ork"
)

func TestExampleNode(t *testing.T) {
	// Test that we can create a node with fluent configuration
	node := ork.NewNodeForHost("server.example.com").
		WithPort("2222").
		WithUser("deploy").
		WithKey("/home/user/.ssh/id_rsa").
		WithArg("app-name", "myapp").
		WithArg("environment", "production")

	// Verify the node was created
	if node == nil {
		t.Fatal("Node was not created")
	}

	// Verify arguments were set
	if node.GetArg("app-name") != "myapp" {
		t.Errorf("Expected arg 'app-name' to be 'myapp', got %s", node.GetArg("app-name"))
	}

	if node.GetArg("environment") != "production" {
		t.Errorf("Expected arg 'environment' to be 'production', got %s", node.GetArg("environment"))
	}
}

func TestExampleNodeWithConfig(t *testing.T) {
	// Test node configuration with fluent chaining
	cfg := ork.NewNodeConfig().
		WithHost("server.example.com").
		WithPort("22").
		WithLogin("ubuntu").
		WithKey("/home/user/.ssh/id_rsa").
		WithArg("app-name", "myapp").
		WithDryRun(true)

	// Create a skill with the config (dereference since WithNodeConfig takes a value)
	skill := ork.NewSkill().
		WithID("check-config").
		WithDescription("Check server configuration").
		WithNodeConfig(*cfg)

	// Verify the skill was configured correctly
	if skill.GetID() != "check-config" {
		t.Errorf("Expected ID to be 'check-config', got %s", skill.GetID())
	}

	if skill.GetDescription() != "Check server configuration" {
		t.Errorf("Expected description to be 'Check server configuration', got %s", skill.GetDescription())
	}

	// Note: The skill doesn't inherit the node config's dry-run mode
	// This test verifies the fluent API works, not the actual logic
}
