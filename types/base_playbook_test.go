package types

import (
	"testing"
	"time"
)

func TestNewBasePlaybook(t *testing.T) {
	playbook := NewBasePlaybook()

	if playbook == nil {
		t.Fatal("NewBasePlaybook returned nil")
	}

	if playbook.GetID() != "" {
		t.Errorf("Expected empty ID, got %s", playbook.GetID())
	}

	if playbook.GetDescription() != "" {
		t.Errorf("Expected empty description, got %s", playbook.GetDescription())
	}

	if playbook.IsDryRun() {
		t.Error("Expected dryRun to be false by default")
	}

	if playbook.GetTimeout() != 0 {
		t.Errorf("Expected timeout to be 0, got %v", playbook.GetTimeout())
	}

	if playbook.GetArgs() == nil {
		t.Error("Expected args map to be initialized")
	}
}

func TestBasePlaybook_Setters(t *testing.T) {
	playbook := NewBasePlaybook()

	// Test SetID
	result := playbook.SetID("test-playbook")
	if result == nil {
		t.Error("SetID should return self for chaining")
	}
	if playbook.GetID() != "test-playbook" {
		t.Errorf("Expected ID to be 'test-playbook', got %s", playbook.GetID())
	}

	// Test SetDescription
	result = playbook.SetDescription("Test playbook description")
	if result == nil {
		t.Error("SetDescription should return self for chaining")
	}
	if playbook.GetDescription() != "Test playbook description" {
		t.Errorf("Expected description to be 'Test playbook description', got %s", playbook.GetDescription())
	}

	// Test SetNodeConfig
	cfg := NodeConfig{
		SSHHost:  "test-host",
		SSHPort:  "22",
		SSHLogin: "test-user",
	}
	result = playbook.SetNodeConfig(cfg)
	if result == nil {
		t.Error("SetNodeConfig should return self for chaining")
	}
	if playbook.GetNodeConfig().SSHHost != "test-host" {
		t.Errorf("Expected SSHHost to be 'test-host', got %s", playbook.GetNodeConfig().SSHHost)
	}

	// Test SetArg
	result = playbook.SetArg("key1", "value1")
	if result == nil {
		t.Error("SetArg should return self for chaining")
	}
	if playbook.GetArg("key1") != "value1" {
		t.Errorf("Expected arg 'key1' to be 'value1', got %s", playbook.GetArg("key1"))
	}

	// Test SetArgs
	args := map[string]string{"key2": "value2", "key3": "value3"}
	result = playbook.SetArgs(args)
	if result == nil {
		t.Error("SetArgs should return self for chaining")
	}
	if playbook.GetArgs()["key2"] != "value2" {
		t.Errorf("Expected arg 'key2' to be 'value2', got %s", playbook.GetArgs()["key2"])
	}

	// Test SetDryRun
	result = playbook.SetDryRun(true)
	if result == nil {
		t.Error("SetDryRun should return self for chaining")
	}
	if !playbook.IsDryRun() {
		t.Error("Expected dryRun to be true")
	}

	// Test SetTimeout
	result = playbook.SetTimeout(30 * time.Second)
	if result == nil {
		t.Error("SetTimeout should return self for chaining")
	}
	if playbook.GetTimeout() != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", playbook.GetTimeout())
	}
}

func TestBasePlaybook_Check(t *testing.T) {
	playbook := NewBasePlaybook()

	needsRun, err := playbook.Check()
	if err != nil {
		t.Errorf("Check() should not return error, got %v", err)
	}
	if needsRun {
		t.Error("Check() should return false by default")
	}
}

func TestBasePlaybook_Run(t *testing.T) {
	playbook := NewBasePlaybook()

	result := playbook.Run()
	if result.Error == nil {
		t.Error("Run() should return an error by default")
	}
	if result.Changed {
		t.Error("Run() should return Changed=false by default")
	}
	if result.Message != "Run() must be implemented by playbook" {
		t.Errorf("Expected specific error message, got %s", result.Message)
	}
}

func TestBasePlaybook_WithEmbedding(t *testing.T) {
	type TestPlaybook struct {
		*BasePlaybook
	}

	// Create playbook with chaining
	base := NewBasePlaybook()
	base.SetID("test-playbook")
	base.SetDescription("Test playbook")

	playbook := &TestPlaybook{
		BasePlaybook: base,
	}

	if playbook.GetID() != "test-playbook" {
		t.Errorf("Expected ID to be 'test-playbook', got %s", playbook.GetID())
	}

	if playbook.GetDescription() != "Test playbook" {
		t.Errorf("Expected description to be 'Test playbook', got %s", playbook.GetDescription())
	}

	// Test that embedding type can override Run by implementing the method
	type CustomPlaybook struct {
		*BasePlaybook
	}

	customPlaybook := &CustomPlaybook{
		BasePlaybook: NewBasePlaybook(),
	}
	customPlaybook.SetID("custom-playbook")

	// Override Run by implementing the method on CustomPlaybook
	result := customPlaybook.Run()
	if result.Error == nil {
		t.Error("Default Run() should return an error")
	}
}
