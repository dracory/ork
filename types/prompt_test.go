package types

import (
	"testing"
)

func TestPromptConfig(t *testing.T) {
	cfg := PromptConfig{
		Name:     "test_var",
		Prompt:   "Enter test value",
		Private:  false,
		Default:  "default_value",
		Confirm:  false,
		Required: true,
	}

	if cfg.Name != "test_var" {
		t.Errorf("Expected Name 'test_var', got '%s'", cfg.Name)
	}
	if cfg.Prompt != "Enter test value" {
		t.Errorf("Expected Prompt 'Enter test value', got '%s'", cfg.Prompt)
	}
	if cfg.Private != false {
		t.Error("Expected Private to be false")
	}
	if cfg.Default != "default_value" {
		t.Errorf("Expected Default 'default_value', got '%s'", cfg.Default)
	}
	if cfg.Confirm != false {
		t.Error("Expected Confirm to be false")
	}
	if cfg.Required != true {
		t.Error("Expected Required to be true")
	}
}

func TestPromptResult(t *testing.T) {
	result := make(PromptResult)
	result["key1"] = "value1"
	result["key2"] = "value2"

	if len(result) != 2 {
		t.Errorf("Expected 2 items in result, got %d", len(result))
	}
	if result["key1"] != "value1" {
		t.Errorf("Expected 'value1' for key1, got '%s'", result["key1"])
	}
	if result["key2"] != "value2" {
		t.Errorf("Expected 'value2' for key2, got '%s'", result["key2"])
	}
}

func TestPromptConfig_Validation(t *testing.T) {
	cfg := PromptConfig{
		Name:   "port",
		Prompt: "Enter port",
		Validate: func(s string) error {
			if s == "" {
				return nil
			}
			// Validation function can be any custom logic
			// This is just a placeholder to verify the field can be set
			return nil
		},
	}

	// Just verify the validation function can be set
	if cfg.Validate == nil {
		t.Error("Validate function should be set")
	}
}
