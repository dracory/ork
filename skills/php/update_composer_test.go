package php

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestUpdateComposer_Run_DryRun verifies that dry-run mode correctly handles composer update.
func TestUpdateComposer_Run_DryRun(t *testing.T) {
	skill := NewUpdateComposer()

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

	expectedMessage := "Would update Composer"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUpdateComposer_NewUpdateComposer verifies that NewUpdateComposer creates a properly configured skill.
func TestUpdateComposer_NewUpdateComposer(t *testing.T) {
	skill := NewUpdateComposer()

	if skill.GetID() != "php-update-composer" {
		t.Errorf("Expected ID to be 'php-update-composer', got '%s'", skill.GetID())
	}

	if skill.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}
