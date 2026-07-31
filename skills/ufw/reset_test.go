package ufw

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestReset_Run_DryRun verifies that dry-run mode correctly handles reset.
func TestReset_Run_DryRun(t *testing.T) {
	skill := NewReset()

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
