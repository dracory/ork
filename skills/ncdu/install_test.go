package ncdu

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestInstall_Run_DryRun verifies that dry-run mode correctly handles ncdu installation.
func TestInstall_Run_DryRun(t *testing.T) {
	skill := NewInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	skill.SetNodeConfig(cfg)

	result := skill.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install packages: ncdu"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestInstall_NewInstall verifies that NewInstall creates a properly configured skill.
func TestInstall_NewInstall(t *testing.T) {
	skill := NewInstall()

	if skill.GetID() != "ncdu-install" {
		t.Errorf("Expected ID to be 'ncdu-install', got '%s'", skill.GetID())
	}

	if skill.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}
