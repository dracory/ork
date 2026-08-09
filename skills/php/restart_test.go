package php

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestRestart_NewRestart verifies that NewRestart creates a properly configured skill.
func TestRestart_NewRestart(t *testing.T) {
	skill := NewRestart()

	if skill.GetID() != "php-fpm-restart" {
		t.Errorf("Expected ID to be 'php-fpm-restart', got '%s'", skill.GetID())
	}

	if skill.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}

// TestRestart_Check verifies that Check always returns true.
func TestRestart_Check(t *testing.T) {
	skill := NewRestart()

	needed, err := skill.Check()
	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if !needed {
		t.Error("Expected Check to return true")
	}
}

// TestRestart_SetVersion verifies that SetVersion sets the version arg.
func TestRestart_SetVersion(t *testing.T) {
	skill := NewRestart().SetVersion("8.3")

	if skill.GetArg(ArgVersion) != "8.3" {
		t.Errorf("Expected version arg to be '8.3', got '%s'", skill.GetArg(ArgVersion))
	}
}

// TestRestart_SetVersion_Chaining verifies that SetVersion returns the same
// *Restart instance for fluent chaining.
func TestRestart_SetVersion_Chaining(t *testing.T) {
	skill := NewRestart()

	returned := skill.SetVersion("8.4")

	if returned != skill {
		t.Error("Expected SetVersion to return the same *Restart instance")
	}
}

// TestRestart_Run_DryRun verifies that dry-run mode delegates to systemctl
// restart which also respects dry-run mode.
func TestRestart_Run_DryRun(t *testing.T) {
	skill := NewRestart().SetVersion("8.5")

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

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestRestart_Run_DryRun_DefaultVersion verifies that the default version is
// used when no version arg is set.
func TestRestart_Run_DryRun_DefaultVersion(t *testing.T) {
	skill := NewRestart()

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

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestRestart_SetArgs verifies that SetArgs sets the args map.
func TestRestart_SetArgs(t *testing.T) {
	skill := NewRestart()

	skill.SetArgs(map[string]string{"version": "8.2"})

	if skill.GetArg(ArgVersion) != "8.2" {
		t.Errorf("Expected version arg to be '8.2', got '%s'", skill.GetArg(ArgVersion))
	}
}

// TestRestart_SetArg verifies that SetArg sets a single arg.
func TestRestart_SetArg(t *testing.T) {
	skill := NewRestart()

	skill.SetArg(ArgVersion, "8.1")

	if skill.GetArg(ArgVersion) != "8.1" {
		t.Errorf("Expected version arg to be '8.1', got '%s'", skill.GetArg(ArgVersion))
	}
}

// TestRestart_SetID verifies that SetID sets the ID.
func TestRestart_SetID(t *testing.T) {
	skill := NewRestart()

	skill.SetID("custom-restart")

	if skill.GetID() != "custom-restart" {
		t.Errorf("Expected ID to be 'custom-restart', got '%s'", skill.GetID())
	}
}

// TestRestart_SetDescription verifies that SetDescription sets the description.
func TestRestart_SetDescription(t *testing.T) {
	skill := NewRestart()

	skill.SetDescription("custom description")

	if skill.GetDescription() != "custom description" {
		t.Errorf("Expected description to be 'custom description', got '%s'", skill.GetDescription())
	}
}
