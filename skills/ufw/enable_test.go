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

// TestEnable_SetAllowPorts verifies that SetAllowPorts sets the allow-ports arg.
func TestEnable_SetAllowPorts(t *testing.T) {
	skill := NewEnable().SetAllowPorts("8080", "9000")

	if skill.GetArg(ArgAllowPorts) != "8080,9000" {
		t.Errorf("Expected allow-ports '8080,9000', got '%s'", skill.GetArg(ArgAllowPorts))
	}
}

// TestEnable_SetAllowPorts_Accumulates verifies that multiple SetAllowPorts
// calls accumulate ports instead of overwriting.
func TestEnable_SetAllowPorts_Accumulates(t *testing.T) {
	skill := NewEnable().
		SetAllowPorts("8080", "9000").
		SetAllowPorts("3306")

	if skill.GetArg(ArgAllowPorts) != "8080,9000,3306" {
		t.Errorf("Expected accumulated '8080,9000,3306', got '%s'", skill.GetArg(ArgAllowPorts))
	}
}

// TestEnable_SetAllowPorts_Deduplicates verifies that duplicate ports
// are not added twice.
func TestEnable_SetAllowPorts_Deduplicates(t *testing.T) {
	skill := NewEnable().
		SetAllowPorts("8080", "9000").
		SetAllowPorts("8080", "3306")

	if skill.GetArg(ArgAllowPorts) != "8080,9000,3306" {
		t.Errorf("Expected deduplicated '8080,9000,3306', got '%s'", skill.GetArg(ArgAllowPorts))
	}
}

// TestEnable_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestEnable_TypedSetters_Chaining(t *testing.T) {
	skill := NewEnable().
		SetAllowSSH(true).
		SetAllowHTTP(true).
		SetAllowHTTPS(true).
		SetAllowPorts("8080", "9000")

	if skill.GetArg(ArgAllowSSH) != "true" {
		t.Errorf("Expected allow-ssh 'true', got '%s'", skill.GetArg(ArgAllowSSH))
	}
	if skill.GetArg(ArgAllowHTTP) != "true" {
		t.Errorf("Expected allow-http 'true', got '%s'", skill.GetArg(ArgAllowHTTP))
	}
	if skill.GetArg(ArgAllowHTTPS) != "true" {
		t.Errorf("Expected allow-https 'true', got '%s'", skill.GetArg(ArgAllowHTTPS))
	}
	if skill.GetArg(ArgAllowPorts) != "8080,9000" {
		t.Errorf("Expected allow-ports '8080,9000', got '%s'", skill.GetArg(ArgAllowPorts))
	}
}
