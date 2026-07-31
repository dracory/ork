package ufw

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestDelete_Run_DryRun verifies that dry-run mode correctly handles delete rule.
func TestDelete_Run_DryRun(t *testing.T) {
	skill := NewDelete().SetNumber("5")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	skill.SetNodeConfig(cfg)

	result := skill.Run()

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run, got: %v", result.Error)
	}
}

// TestDelete_SetNumber verifies that SetNumber sets the number arg.
func TestDelete_SetNumber(t *testing.T) {
	skill := NewDelete().SetNumber("3")

	if skill.GetArg(ArgNumber) != "3" {
		t.Errorf("Expected number '3', got '%s'", skill.GetArg(ArgNumber))
	}
}
