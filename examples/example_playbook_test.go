package examples

import (
	"testing"

	"github.com/dracory/ork/types"
)

func TestNewExamplePlaybook(t *testing.T) {
	playbook := NewExamplePlaybook()

	if playbook == nil {
		t.Fatal("NewExamplePlaybook returned nil")
	}

	if playbook.GetID() != "example-playbook" {
		t.Errorf("Expected ID to be 'example-playbook', got %s", playbook.GetID())
	}

	if playbook.GetDescription() != "Example playbook demonstrating sequential skill execution" {
		t.Errorf("Unexpected description: %s", playbook.GetDescription())
	}
}

func TestExamplePlaybook_Run(t *testing.T) {
	playbook := NewExamplePlaybook()

	cfg := types.NodeConfig{
		SSHHost:      "test-host",
		SSHPort:      "22",
		SSHLogin:     "test-user",
		SSHKey:       "test-key",
		IsDryRunMode: true,
	}

	playbook.SetNodeConfig(cfg)

	result := playbook.Run()

	// In dry-run mode, it should complete without errors
	if result.Error != nil {
		t.Errorf("Run() should not error in dry-run mode, got %v", result.Error)
	}

	if result.Message == "" {
		t.Error("Run() should return a message")
	}
}

func TestExamplePlaybook_ImplementsRunnableInterface(t *testing.T) {
	playbook := NewExamplePlaybook()

	// Verify it implements RunnableInterface
	var _ types.RunnableInterface = playbook

	// Test that all required methods exist
	_ = playbook.GetID()
	_ = playbook.GetDescription()
	_ = playbook.GetNodeConfig()
	_ = playbook.GetArg("test")
	_ = playbook.GetArgs()
	_ = playbook.IsDryRun()
	_ = playbook.GetTimeout()

	// Test setters return RunnableInterface for chaining
	var ri types.RunnableInterface
	ri = playbook.SetID("test")
	ri = playbook.SetDescription("test")
	ri = playbook.SetNodeConfig(types.NodeConfig{})
	ri = playbook.SetArg("key", "value")
	ri = playbook.SetArgs(map[string]string{"key": "value"})
	ri = playbook.SetDryRun(true)
	ri = playbook.SetTimeout(0)

	if ri == nil {
		t.Error("Setters should return RunnableInterface for chaining")
	}
}
