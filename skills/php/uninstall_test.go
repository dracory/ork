package php

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestUninstall_Run_DryRun verifies that dry-run mode correctly handles php uninstall.
func TestUninstall_Run_DryRun(t *testing.T) {
	skill := NewUninstall().SetVersion("8.3")

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

	expectedMessage := "Would uninstall PHP 8.3"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUninstall_SetVersion verifies that SetVersion sets the version arg.
func TestUninstall_SetVersion(t *testing.T) {
	skill := NewUninstall().SetVersion("8.3")

	if skill.GetArg(ArgVersion) != "8.3" {
		t.Errorf("Expected version '8.3', got '%s'", skill.GetArg(ArgVersion))
	}
}

// TestUninstall_NewUninstall verifies that NewUninstall creates a properly configured skill.
func TestUninstall_NewUninstall(t *testing.T) {
	skill := NewUninstall()

	if skill.GetID() != "php-uninstall" {
		t.Errorf("Expected ID to be 'php-uninstall', got '%s'", skill.GetID())
	}

	if skill.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}
