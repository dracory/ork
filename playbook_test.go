package ork

import (
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

func TestNewPlaybook(t *testing.T) {
	playbook := NewPlaybook()

	if playbook == nil {
		t.Fatal("NewPlaybook returned nil")
	}

	// Verify default values
	if playbook.GetID() != "" {
		t.Errorf("Expected empty ID, got %s", playbook.GetID())
	}

	if playbook.GetDescription() != "" {
		t.Errorf("Expected empty description, got %s", playbook.GetDescription())
	}

	if playbook.IsDryRun() {
		t.Error("Expected DryRun to be false by default")
	}

	if playbook.GetTimeout() != 0 {
		t.Errorf("Expected timeout to be 0, got %v", playbook.GetTimeout())
	}
}

func TestNewPlaybook_WithID(t *testing.T) {
	playbook := NewPlaybook().
		WithID("test-playbook")

	if playbook.GetID() != "test-playbook" {
		t.Errorf("Expected ID to be 'test-playbook', got %s", playbook.GetID())
	}
}

func TestNewPlaybook_WithDescription(t *testing.T) {
	playbook := NewPlaybook().
		WithDescription("Test description")

	if playbook.GetDescription() != "Test description" {
		t.Errorf("Expected description to be 'Test description', got %s", playbook.GetDescription())
	}
}

func TestNewPlaybook_WithDryRun(t *testing.T) {
	playbook := NewPlaybook().
		WithDryRun(true)

	if !playbook.IsDryRun() {
		t.Error("Expected DryRun to be true")
	}
}

func TestNewPlaybook_WithTimeout(t *testing.T) {
	playbook := NewPlaybook().
		WithTimeout(30 * time.Second)

	if playbook.GetTimeout() != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", playbook.GetTimeout())
	}
}

func TestNewPlaybook_WithArg(t *testing.T) {
	playbook := NewPlaybook().
		WithArg("key", "value")

	if playbook.GetArg("key") != "value" {
		t.Errorf("Expected arg 'key' to be 'value', got %s", playbook.GetArg("key"))
	}
}

func TestNewPlaybook_WithArgs(t *testing.T) {
	args := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	playbook := NewPlaybook().
		WithArgs(args)

	if playbook.GetArg("key1") != "value1" {
		t.Errorf("Expected arg 'key1' to be 'value1', got %s", playbook.GetArg("key1"))
	}

	if playbook.GetArg("key2") != "value2" {
		t.Errorf("Expected arg 'key2' to be 'value2', got %s", playbook.GetArg("key2"))
	}
}

func TestNewPlaybook_WithNodeConfig(t *testing.T) {
	cfg := types.NodeConfig{
		SSHHost:  "example.com",
		SSHPort:  "22",
		SSHLogin: "user",
	}

	playbook := NewPlaybook().
		WithNodeConfig(cfg)

	if playbook.GetNodeConfig().SSHHost != "example.com" {
		t.Errorf("Expected SSHHost to be 'example.com', got %s", playbook.GetNodeConfig().SSHHost)
	}
}

func TestNewPlaybook_WithBecomeUser(t *testing.T) {
	playbook := NewPlaybook().
		WithBecomeUser("testuser")

	if playbook.GetBecomeUser() != "testuser" {
		t.Errorf("Expected become user to be 'testuser', got %s", playbook.GetBecomeUser())
	}
}

func TestNewPlaybook_FluentChaining(t *testing.T) {
	args := map[string]string{"key": "value"}
	cfg := types.NodeConfig{SSHHost: "example.com"}

	playbook := NewPlaybook().
		WithID("test-playbook").
		WithDescription("Test description").
		WithDryRun(true).
		WithTimeout(30*time.Second).
		WithArgs(args).
		WithNodeConfig(cfg).
		WithBecomeUser("testuser")

	// Verify all values were set correctly
	if playbook.GetID() != "test-playbook" {
		t.Errorf("Expected ID to be 'test-playbook', got %s", playbook.GetID())
	}

	if playbook.GetDescription() != "Test description" {
		t.Errorf("Expected description to be 'Test description', got %s", playbook.GetDescription())
	}

	if !playbook.IsDryRun() {
		t.Error("Expected DryRun to be true")
	}

	if playbook.GetTimeout() != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", playbook.GetTimeout())
	}

	if playbook.GetArg("key") != "value" {
		t.Errorf("Expected arg 'key' to be 'value', got %s", playbook.GetArg("key"))
	}

	if playbook.GetNodeConfig().SSHHost != "example.com" {
		t.Errorf("Expected SSHHost to be 'example.com', got %s", playbook.GetNodeConfig().SSHHost)
	}

	if playbook.GetBecomeUser() != "testuser" {
		t.Errorf("Expected become user to be 'testuser', got %s", playbook.GetBecomeUser())
	}
}

func TestNewPlaybook_Check(t *testing.T) {
	playbook := NewPlaybook()
	needsRun, err := playbook.Check()

	if err != nil {
		t.Errorf("Check should not return error, got %v", err)
	}

	if needsRun {
		t.Error("BasePlaybook Check should return false (no changes needed)")
	}
}
