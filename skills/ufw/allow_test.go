package ufw

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestAllow_Run_DryRun verifies that dry-run mode correctly handles allow rule.
func TestAllow_Run_DryRun(t *testing.T) {
	skill := NewAllow().SetPort("80").SetProtocol("tcp")

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

// TestAllow_SetPort verifies that SetPort sets the port arg.
func TestAllow_SetPort(t *testing.T) {
	skill := NewAllow().SetPort("8080")

	if skill.GetArg(ArgPort) != "8080" {
		t.Errorf("Expected port '8080', got '%s'", skill.GetArg(ArgPort))
	}
}

// TestAllow_SetProtocol verifies that SetProtocol sets the protocol arg.
func TestAllow_SetProtocol(t *testing.T) {
	skill := NewAllow().SetProtocol("udp")

	if skill.GetArg(ArgProtocol) != "udp" {
		t.Errorf("Expected protocol 'udp', got '%s'", skill.GetArg(ArgProtocol))
	}
}

// TestAllow_SetComment verifies that SetComment sets the comment arg.
func TestAllow_SetComment(t *testing.T) {
	skill := NewAllow().SetComment("HTTP traffic")

	if skill.GetArg(ArgComment) != "HTTP traffic" {
		t.Errorf("Expected comment 'HTTP traffic', got '%s'", skill.GetArg(ArgComment))
	}
}

// TestAllow_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestAllow_TypedSetters_Chaining(t *testing.T) {
	skill := NewAllow().
		SetPort("443").
		SetProtocol("tcp").
		SetComment("HTTPS")

	if skill.GetArg(ArgPort) != "443" {
		t.Errorf("Expected port '443', got '%s'", skill.GetArg(ArgPort))
	}
	if skill.GetArg(ArgProtocol) != "tcp" {
		t.Errorf("Expected protocol 'tcp', got '%s'", skill.GetArg(ArgProtocol))
	}
	if skill.GetArg(ArgComment) != "HTTPS" {
		t.Errorf("Expected comment 'HTTPS', got '%s'", skill.GetArg(ArgComment))
	}
}
