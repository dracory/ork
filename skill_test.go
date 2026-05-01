package ork

import (
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

func TestNewSkill(t *testing.T) {
	skill := NewSkill()

	if skill == nil {
		t.Fatal("NewSkill returned nil")
	}

	// Verify default values
	if skill.GetID() != "" {
		t.Errorf("Expected empty ID, got %s", skill.GetID())
	}

	if skill.GetDescription() != "" {
		t.Errorf("Expected empty description, got %s", skill.GetDescription())
	}

	if skill.IsDryRun() {
		t.Error("Expected DryRun to be false by default")
	}

	if skill.GetTimeout() != 0 {
		t.Errorf("Expected timeout to be 0, got %v", skill.GetTimeout())
	}
}

func TestNewSkill_WithID(t *testing.T) {
	skill := NewSkill().
		WithID("test-skill")

	if skill.GetID() != "test-skill" {
		t.Errorf("Expected ID to be 'test-skill', got %s", skill.GetID())
	}
}

func TestNewSkill_WithDescription(t *testing.T) {
	skill := NewSkill().
		WithDescription("Test description")

	if skill.GetDescription() != "Test description" {
		t.Errorf("Expected description to be 'Test description', got %s", skill.GetDescription())
	}
}

func TestNewSkill_WithDryRun(t *testing.T) {
	skill := NewSkill().
		WithDryRun(true)

	if !skill.IsDryRun() {
		t.Error("Expected DryRun to be true")
	}
}

func TestNewSkill_WithTimeout(t *testing.T) {
	skill := NewSkill().
		WithTimeout(30 * time.Second)

	if skill.GetTimeout() != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", skill.GetTimeout())
	}
}

func TestNewSkill_WithArg(t *testing.T) {
	skill := NewSkill().
		WithArg("key", "value")

	if skill.GetArg("key") != "value" {
		t.Errorf("Expected arg 'key' to be 'value', got %s", skill.GetArg("key"))
	}
}

func TestNewSkill_WithArgs(t *testing.T) {
	args := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	skill := NewSkill().
		WithArgs(args)

	if skill.GetArg("key1") != "value1" {
		t.Errorf("Expected arg 'key1' to be 'value1', got %s", skill.GetArg("key1"))
	}

	if skill.GetArg("key2") != "value2" {
		t.Errorf("Expected arg 'key2' to be 'value2', got %s", skill.GetArg("key2"))
	}
}

func TestNewSkill_WithNodeConfig(t *testing.T) {
	cfg := types.NodeConfig{
		SSHHost:  "example.com",
		SSHPort:  "22",
		SSHLogin: "user",
	}

	skill := NewSkill().
		WithNodeConfig(cfg)

	if skill.GetNodeConfig().SSHHost != "example.com" {
		t.Errorf("Expected SSHHost to be 'example.com', got %s", skill.GetNodeConfig().SSHHost)
	}
}

func TestNewSkill_WithBecomeUser(t *testing.T) {
	skill := NewSkill().
		WithBecomeUser("testuser")

	if skill.GetBecomeUser() != "testuser" {
		t.Errorf("Expected become user to be 'testuser', got %s", skill.GetBecomeUser())
	}
}

func TestNewSkill_FluentChaining(t *testing.T) {
	args := map[string]string{"key": "value"}
	cfg := types.NodeConfig{SSHHost: "example.com"}

	skill := NewSkill().
		WithID("test-skill").
		WithDescription("Test description").
		WithDryRun(true).
		WithTimeout(30*time.Second).
		WithArgs(args).
		WithNodeConfig(cfg).
		WithBecomeUser("testuser")

	// Verify all values were set correctly
	if skill.GetID() != "test-skill" {
		t.Errorf("Expected ID to be 'test-skill', got %s", skill.GetID())
	}

	if skill.GetDescription() != "Test description" {
		t.Errorf("Expected description to be 'Test description', got %s", skill.GetDescription())
	}

	if !skill.IsDryRun() {
		t.Error("Expected DryRun to be true")
	}

	if skill.GetTimeout() != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", skill.GetTimeout())
	}

	if skill.GetArg("key") != "value" {
		t.Errorf("Expected arg 'key' to be 'value', got %s", skill.GetArg("key"))
	}

	if skill.GetNodeConfig().SSHHost != "example.com" {
		t.Errorf("Expected SSHHost to be 'example.com', got %s", skill.GetNodeConfig().SSHHost)
	}

	if skill.GetBecomeUser() != "testuser" {
		t.Errorf("Expected become user to be 'testuser', got %s", skill.GetBecomeUser())
	}
}
