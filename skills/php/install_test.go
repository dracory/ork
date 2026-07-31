package php

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestInstall_Run_DryRun verifies that dry-run mode correctly handles php install.
func TestInstall_Run_DryRun(t *testing.T) {
	skill := NewInstall().SetVersion("8.3").SetUser("deploy")

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

	expectedMessage := "Would install PHP 8.3"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestInstall_SetVersion verifies that SetVersion sets the version arg.
func TestInstall_SetVersion(t *testing.T) {
	skill := NewInstall().SetVersion("8.3")

	if skill.GetArg(ArgVersion) != "8.3" {
		t.Errorf("Expected version '8.3', got '%s'", skill.GetArg(ArgVersion))
	}
}

// TestInstall_SetUser verifies that SetUser sets the user arg.
func TestInstall_SetUser(t *testing.T) {
	skill := NewInstall().SetUser("deploy")

	if skill.GetArg(ArgUser) != "deploy" {
		t.Errorf("Expected user 'deploy', got '%s'", skill.GetArg(ArgUser))
	}
}

// TestInstall_SetExtensions verifies that SetExtensions sets the extensions arg.
func TestInstall_SetExtensions(t *testing.T) {
	skill := NewInstall().SetExtensions("cli fpm mysql")

	if skill.GetArg(ArgExtensions) != "cli fpm mysql" {
		t.Errorf("Expected extensions 'cli fpm mysql', got '%s'", skill.GetArg(ArgExtensions))
	}
}

// TestInstall_SetExtensions_Variadic verifies that SetExtensions joins variadic args with spaces.
func TestInstall_SetExtensions_Variadic(t *testing.T) {
	skill := NewInstall().SetExtensions("cli", "fpm", "mysql")

	if skill.GetArg(ArgExtensions) != "cli fpm mysql" {
		t.Errorf("Expected extensions 'cli fpm mysql', got '%s'", skill.GetArg(ArgExtensions))
	}
}

// TestInstall_SetExtensions_None verifies that SetExtensions with no args means no extensions.
func TestInstall_SetExtensions_None(t *testing.T) {
	skill := NewInstall().SetExtensions()

	if skill.GetArg(ArgExtensions) != "" {
		t.Errorf("Expected empty extensions, got '%s'", skill.GetArg(ArgExtensions))
	}
}

// TestInstall_SetExtensions_Defaults verifies that DefaultExtensions can be passed explicitly.
func TestInstall_SetExtensions_Defaults(t *testing.T) {
	skill := NewInstall().SetExtensions(DefaultExtensions)

	if skill.GetArg(ArgExtensions) != DefaultExtensions {
		t.Errorf("Expected DefaultExtensions, got '%s'", skill.GetArg(ArgExtensions))
	}
}

// TestInstall_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestInstall_TypedSetters_Chaining(t *testing.T) {
	skill := NewInstall().
		SetVersion("8.3").
		SetUser("deploy").
		SetExtensions("cli fpm mysql")

	if skill.GetArg(ArgVersion) != "8.3" {
		t.Errorf("Expected version '8.3', got '%s'", skill.GetArg(ArgVersion))
	}
	if skill.GetArg(ArgUser) != "deploy" {
		t.Errorf("Expected user 'deploy', got '%s'", skill.GetArg(ArgUser))
	}
	if skill.GetArg(ArgExtensions) != "cli fpm mysql" {
		t.Errorf("Expected extensions 'cli fpm mysql', got '%s'", skill.GetArg(ArgExtensions))
	}
}

// TestInstall_NewInstall verifies that NewInstall creates a properly configured skill.
func TestInstall_NewInstall(t *testing.T) {
	skill := NewInstall()

	if skill.GetID() != "php-install" {
		t.Errorf("Expected ID to be 'php-install', got '%s'", skill.GetID())
	}

	if skill.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}
