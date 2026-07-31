package ufw

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestEnable_Run_DryRun verifies that dry-run mode correctly handles enable.
func TestEnable_Run_DryRun(t *testing.T) {
	skill := NewEnable()

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
