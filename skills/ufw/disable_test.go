package ufw

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestDisable_Run_DryRun verifies that dry-run mode correctly handles disable.
func TestDisable_Run_DryRun(t *testing.T) {
	skill := NewDisable()

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
