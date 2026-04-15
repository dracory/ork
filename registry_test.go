package ork

import (
	"testing"

	"github.com/dracory/ork/types"
)

func TestNewDefaultRegistry_Initialized(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}
	if reg == nil {
		t.Fatal("NewDefaultRegistry() should return a non-nil registry")
	}
}

func TestNewDefaultRegistry_AllBuiltInSkillsRegistered(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}
	expectedSkills := []string{
		"ping",
		"apt-update",
		"apt-upgrade",
		"apt-status",
		"reboot",
		"swap-create",
		"swap-delete",
		"swap-status",
		"user-create",
		"user-delete",
		"user-status",
	}

	for _, id := range expectedSkills {
		skill, ok := reg.FindByID(id)
		if !ok {
			t.Errorf("expected skill '%s' to be registered, but it was not found", id)
			continue
		}
		if skill.GetID() != id {
			t.Errorf("skill ID mismatch: expected '%s', got '%s'", id, skill.GetID())
		}
	}
}

func TestNewDefaultRegistry_ContainsExpectedSkillIDs(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}
	ids := reg.GetIDs()

	// Verify all expected built-in skill IDs are present
	expectedIDs := []string{
		"ping",
		"apt-update",
		"apt-upgrade",
		"apt-status",
		"reboot",
		"swap-create",
		"swap-delete",
		"swap-status",
		"user-create",
		"user-delete",
		"user-status",
	}

	// Create a map of actual IDs for quick lookup
	actualIDs := make(map[string]bool)
	for _, id := range ids {
		actualIDs[id] = true
	}

	// Check that all expected IDs are present
	for _, id := range expectedIDs {
		if !actualIDs[id] {
			t.Errorf("expected built-in skill '%s' not found in registry", id)
		}
	}

	// Verify we have at least the expected number of built-in skills
	if len(ids) < len(expectedIDs) {
		t.Errorf("expected at least %d skills, got %d", len(expectedIDs), len(ids))
	}
}

func TestNewDefaultRegistry_SkillsHaveDescriptions(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}
	skills := reg.List()

	for _, skill := range skills {
		if skill.GetDescription() == "" {
			t.Errorf("skill '%s' has empty description", skill.GetID())
		}
	}
}

func TestGetGlobalSkillRegistry(t *testing.T) {
	// Create a fresh registry for this test to avoid polluting the global registry
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}
	if reg == nil {
		t.Fatal("NewDefaultRegistry() returned nil")
	}

	// Test that we can use it to register a skill
	customSkill := types.NewBaseSkill()
	customSkill.SetID("test-get-registry-skill")
	customSkill.SetDescription("Test skill via NewDefaultRegistry")

	err = reg.Register(customSkill)
	if err != nil {
		t.Fatalf("failed to register skill: %v", err)
	}

	// Verify it can be found
	foundSkill, ok := reg.FindByID("test-get-registry-skill")
	if !ok {
		t.Fatal("custom skill not found after registration")
	}
	if foundSkill.GetID() != "test-get-registry-skill" {
		t.Errorf("expected ID 'test-get-registry-skill', got '%s'", foundSkill.GetID())
	}
}

func TestGetGlobalSkillRegistry_LazyInitialization(t *testing.T) {
	// Test that GetGlobalSkillRegistry() initializes the global registry on first call
	reg, err := GetGlobalSkillRegistry()
	if err != nil {
		t.Fatalf("GetGlobalSkillRegistry() failed: %v", err)
	}
	if reg == nil {
		t.Fatal("GetGlobalSkillRegistry() returned nil")
	}

	// Verify it has built-in skills
	skill, ok := reg.FindByID("ping")
	if !ok {
		t.Fatal("expected 'ping' skill in global registry")
	}
	if skill.GetID() != "ping" {
		t.Errorf("expected ID 'ping', got '%s'", skill.GetID())
	}

	// Test that subsequent calls return the same instance
	reg2, err := GetGlobalSkillRegistry()
	if err != nil {
		t.Fatalf("GetGlobalSkillRegistry() failed on second call: %v", err)
	}
	if reg != reg2 {
		t.Error("GetGlobalSkillRegistry() should return the same instance on subsequent calls")
	}
}

func TestNewDefaultRegistry_DuplicateID(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}

	// Try to register a skill with a duplicate ID
	duplicateSkill := types.NewBaseSkill()
	duplicateSkill.SetID("ping") // "ping" is already registered
	duplicateSkill.SetDescription("Duplicate ping skill")

	err = reg.Register(duplicateSkill)
	if err == nil {
		t.Error("expected error when registering duplicate skill ID, got nil")
	}
}
