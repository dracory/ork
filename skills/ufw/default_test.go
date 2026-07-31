package ufw

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestDefault_Run_DryRun verifies that dry-run mode correctly handles default policy.
func TestDefault_Run_DryRun(t *testing.T) {
	skill := NewDefault().SetIncoming("deny").SetOutgoing("allow")

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

// TestDefault_SetIncoming verifies that SetIncoming sets the incoming arg.
func TestDefault_SetIncoming(t *testing.T) {
	skill := NewDefault().SetIncoming("reject")

	if skill.GetArg(ArgIncoming) != "reject" {
		t.Errorf("Expected incoming 'reject', got '%s'", skill.GetArg(ArgIncoming))
	}
}

// TestDefault_SetOutgoing verifies that SetOutgoing sets the outgoing arg.
func TestDefault_SetOutgoing(t *testing.T) {
	skill := NewDefault().SetOutgoing("deny")

	if skill.GetArg(ArgOutgoing) != "deny" {
		t.Errorf("Expected outgoing 'deny', got '%s'", skill.GetArg(ArgOutgoing))
	}
}

// TestDefault_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestDefault_TypedSetters_Chaining(t *testing.T) {
	skill := NewDefault().
		SetIncoming("deny").
		SetOutgoing("allow")

	if skill.GetArg(ArgIncoming) != "deny" {
		t.Errorf("Expected incoming 'deny', got '%s'", skill.GetArg(ArgIncoming))
	}
	if skill.GetArg(ArgOutgoing) != "allow" {
		t.Errorf("Expected outgoing 'allow', got '%s'", skill.GetArg(ArgOutgoing))
	}
}
