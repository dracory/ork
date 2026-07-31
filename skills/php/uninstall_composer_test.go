package php

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestUninstallComposer_Run_DryRun verifies that dry-run mode correctly handles composer uninstall.
func TestUninstallComposer_Run_DryRun(t *testing.T) {
	skill := NewUninstallComposer()

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

	expectedMessage := "Would uninstall Composer"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUninstallComposer_NewUninstallComposer verifies that NewUninstallComposer creates a properly configured skill.
func TestUninstallComposer_NewUninstallComposer(t *testing.T) {
	skill := NewUninstallComposer()

	if skill.GetID() != "php-uninstall-composer" {
		t.Errorf("Expected ID to be 'php-uninstall-composer', got '%s'", skill.GetID())
	}

	if skill.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}
