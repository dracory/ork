package ufw

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestDeny_Run_DryRun verifies that dry-run mode correctly handles deny rule.
func TestDeny_Run_DryRun(t *testing.T) {
	skill := NewDeny().SetPort("3306").SetProtocol("tcp")

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

// TestDeny_SetPort verifies that SetPort sets the port arg.
func TestDeny_SetPort(t *testing.T) {
	skill := NewDeny().SetPort("3306")

	if skill.GetArg(ArgPort) != "3306" {
		t.Errorf("Expected port '3306', got '%s'", skill.GetArg(ArgPort))
	}
}

// TestDeny_SetProtocol verifies that SetProtocol sets the protocol arg.
func TestDeny_SetProtocol(t *testing.T) {
	skill := NewDeny().SetProtocol("udp")

	if skill.GetArg(ArgProtocol) != "udp" {
		t.Errorf("Expected protocol 'udp', got '%s'", skill.GetArg(ArgProtocol))
	}
}

// TestDeny_SetComment verifies that SetComment sets the comment arg.
func TestDeny_SetComment(t *testing.T) {
	skill := NewDeny().SetComment("Block MySQL")

	if skill.GetArg(ArgComment) != "Block MySQL" {
		t.Errorf("Expected comment 'Block MySQL', got '%s'", skill.GetArg(ArgComment))
	}
}

// TestDeny_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestDeny_TypedSetters_Chaining(t *testing.T) {
	skill := NewDeny().
		SetPort("3306").
		SetProtocol("tcp").
		SetComment("Block MySQL")

	if skill.GetArg(ArgPort) != "3306" {
		t.Errorf("Expected port '3306', got '%s'", skill.GetArg(ArgPort))
	}
	if skill.GetArg(ArgProtocol) != "tcp" {
		t.Errorf("Expected protocol 'tcp', got '%s'", skill.GetArg(ArgProtocol))
	}
	if skill.GetArg(ArgComment) != "Block MySQL" {
		t.Errorf("Expected comment 'Block MySQL', got '%s'", skill.GetArg(ArgComment))
	}
}
