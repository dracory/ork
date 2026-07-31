package php

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestInstallComposer_Run_DryRun verifies that dry-run mode correctly handles composer install.
func TestInstallComposer_Run_DryRun(t *testing.T) {
	skill := NewInstallComposer()

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

	expectedMessage := "Would install Composer"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestInstallComposer_NewInstallComposer verifies that NewInstallComposer creates a properly configured skill.
func TestInstallComposer_NewInstallComposer(t *testing.T) {
	skill := NewInstallComposer()

	if skill.GetID() != "php-install-composer" {
		t.Errorf("Expected ID to be 'php-install-composer', got '%s'", skill.GetID())
	}

	if skill.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}
