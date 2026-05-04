package ork

import (
	"fmt"
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

	err = reg.Set(customSkill)
	if err != nil {
		t.Fatalf("failed to set skill: %v", err)
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

	err = reg.Set(duplicateSkill)
	if err != nil {
		t.Errorf("failed to set skill: %v", err)
	}
}

func TestRegistry_Set(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}

	customSkill := types.NewBaseSkill()
	customSkill.SetID("ping")
	customSkill.SetDescription("Custom ping skill")

	err = reg.Set(customSkill)
	if err != nil {
		t.Fatalf("failed to set skill: %v", err)
	}

	foundSkill, ok := reg.FindByID("ping")
	if !ok {
		t.Fatal("custom skill not found")
	}
	if foundSkill.GetDescription() != "Custom ping skill" {
		t.Errorf("expected custom skill, got '%s'", foundSkill.GetDescription())
	}
}

func TestRegistry_SetAll(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}

	skills := []types.RunnableInterface{
		func() types.RunnableInterface {
			s := types.NewBaseSkill()
			s.SetID("skill-1")
			s.SetDescription("Skill 1")
			return s
		}(),
		func() types.RunnableInterface {
			s := types.NewBaseSkill()
			s.SetID("skill-2")
			s.SetDescription("Skill 2")
			return s
		}(),
		func() types.RunnableInterface {
			s := types.NewBaseSkill()
			s.SetID("skill-3")
			s.SetDescription("Skill 3")
			return s
		}(),
	}

	err = reg.SetAll(skills)
	if err != nil {
		t.Fatalf("failed to set all skills: %v", err)
	}

	// Verify all skills were added
	for i := 1; i <= 3; i++ {
		skillID := fmt.Sprintf("skill-%d", i)
		foundSkill, ok := reg.FindByID(skillID)
		if !ok {
			t.Errorf("skill '%s' not found", skillID)
			continue
		}
		expectedDesc := fmt.Sprintf("Skill %d", i)
		if foundSkill.GetDescription() != expectedDesc {
			t.Errorf("expected description '%s', got '%s'", expectedDesc, foundSkill.GetDescription())
		}
	}
}

func TestRegistry_SetAll_EmptySlice(t *testing.T) {
	reg := types.NewRegistry()

	err := reg.SetAll([]types.RunnableInterface{})
	if err != nil {
		t.Errorf("expected no error for empty slice, got: %v", err)
	}
}

func TestRegistry_SetAll_NilInSlice(t *testing.T) {
	reg := types.NewRegistry()

	skills := []types.RunnableInterface{
		func() types.RunnableInterface {
			s := types.NewBaseSkill()
			s.SetID("skill-1")
			s.SetDescription("Skill 1")
			return s
		}(),
		nil,
	}

	err := reg.SetAll(skills)
	if err == nil {
		t.Error("expected error for nil runnable in slice, got nil")
	}
}
