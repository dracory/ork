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

func TestRegistry_Disable_Enable(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}

	// ping is a built-in skill, safe to disable for this test.
	// Re-enable it at the end so we don't leak state into other tests.
	t.Cleanup(func() {
		_ = reg.Enable("ping")
	})

	// Initially not disabled.
	if _, disabled := reg.IsDisabled("ping"); disabled {
		t.Fatal("expected 'ping' to not be disabled initially")
	}

	// Disable with a comment.
	if err := reg.Disable("ping", "too risky in tests"); err != nil {
		t.Fatalf("Disable() failed: %v", err)
	}

	comment, disabled := reg.IsDisabled("ping")
	if !disabled {
		t.Fatal("expected 'ping' to be disabled after Disable()")
	}
	if comment != "too risky in tests" {
		t.Errorf("expected disable comment 'too risky in tests', got '%s'", comment)
	}

	// Disabled runnable should still be findable/listable.
	if _, ok := reg.FindByID("ping"); !ok {
		t.Error("disabled runnable should still be findable via FindByID")
	}

	// ListDisabled should include it.
	found := false
	for _, id := range reg.ListDisabled() {
		if id == "ping" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'ping' to appear in ListDisabled()")
	}

	// Re-disabling updates the comment.
	if err := reg.Disable("ping", "new reason"); err != nil {
		t.Fatalf("Disable() on already-disabled failed: %v", err)
	}
	comment, _ = reg.IsDisabled("ping")
	if comment != "new reason" {
		t.Errorf("expected updated comment 'new reason', got '%s'", comment)
	}

	// Enable lifts the block.
	if err := reg.Enable("ping"); err != nil {
		t.Fatalf("Enable() failed: %v", err)
	}
	if _, disabled := reg.IsDisabled("ping"); disabled {
		t.Error("expected 'ping' to not be disabled after Enable()")
	}

	// ListDisabled should no longer include it.
	for _, id := range reg.ListDisabled() {
		if id == "ping" {
			t.Error("expected 'ping' to be absent from ListDisabled() after Enable()")
			break
		}
	}
}

func TestRegistry_Disable_UnknownID(t *testing.T) {
	reg := types.NewRegistry()

	if err := reg.Disable("does-not-exist", ""); err == nil {
		t.Error("expected error when disabling unknown ID, got nil")
	}
}

func TestRegistry_Disable_EmptyID(t *testing.T) {
	reg := types.NewRegistry()

	if err := reg.Disable("", "nope"); err == nil {
		t.Error("expected error when disabling empty ID, got nil")
	}
}

func TestRegistry_Enable_UnknownID(t *testing.T) {
	reg := types.NewRegistry()

	if err := reg.Enable("does-not-exist"); err == nil {
		t.Error("expected error when enabling unknown ID, got nil")
	}
}

func TestRegistry_Enable_EmptyID(t *testing.T) {
	reg := types.NewRegistry()

	if err := reg.Enable(""); err == nil {
		t.Error("expected error when enabling empty ID, got nil")
	}
}

func TestRegistry_Enable_NotDisabled(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}

	// Enabling a runnable that is not disabled should be a no-op (no error).
	if err := reg.Enable("ping"); err != nil {
		t.Errorf("Enable() on non-disabled runnable should be a no-op, got: %v", err)
	}
	if _, disabled := reg.IsDisabled("ping"); disabled {
		t.Error("expected 'ping' to remain not-disabled after no-op Enable()")
	}
}

func TestRegistry_IsDisabled_UnknownID(t *testing.T) {
	reg := types.NewRegistry()

	if _, disabled := reg.IsDisabled("does-not-exist"); disabled {
		t.Error("expected IsDisabled to return false for unknown ID")
	}
}

func TestRegistry_Add_AliasForSet(t *testing.T) {
	reg := types.NewRegistry()

	s := types.NewBaseSkill()
	s.SetID("alias-add-skill")
	s.SetDescription("added via Add")

	if err := reg.Add(s); err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	found, ok := reg.FindByID("alias-add-skill")
	if !ok {
		t.Fatal("runnable added via Add() not found")
	}
	if found.GetDescription() != "added via Add" {
		t.Errorf("expected 'added via Add', got '%s'", found.GetDescription())
	}

	// Add should replace an existing runnable, just like Set.
	replacement := types.NewBaseSkill()
	replacement.SetID("alias-add-skill")
	replacement.SetDescription("replaced via Add")
	if err := reg.Add(replacement); err != nil {
		t.Fatalf("Add() replace failed: %v", err)
	}
	found, _ = reg.FindByID("alias-add-skill")
	if found.GetDescription() != "replaced via Add" {
		t.Errorf("expected replacement, got '%s'", found.GetDescription())
	}
}

func TestRegistry_Add_NilRunnable(t *testing.T) {
	reg := types.NewRegistry()

	if err := reg.Add(nil); err == nil {
		t.Error("expected error when adding nil runnable, got nil")
	}
}

func TestRegistry_AddAll_AliasForSetAll(t *testing.T) {
	reg := types.NewRegistry()

	skills := []types.RunnableInterface{
		func() types.RunnableInterface {
			s := types.NewBaseSkill()
			s.SetID("addall-1")
			s.SetDescription("one")
			return s
		}(),
		func() types.RunnableInterface {
			s := types.NewBaseSkill()
			s.SetID("addall-2")
			s.SetDescription("two")
			return s
		}(),
	}

	if err := reg.AddAll(skills); err != nil {
		t.Fatalf("AddAll() failed: %v", err)
	}

	for _, id := range []string{"addall-1", "addall-2"} {
		if _, ok := reg.FindByID(id); !ok {
			t.Errorf("expected '%s' to be registered after AddAll()", id)
		}
	}
}

func TestRegistry_AddAll_NilInSlice(t *testing.T) {
	reg := types.NewRegistry()

	skills := []types.RunnableInterface{
		func() types.RunnableInterface {
			s := types.NewBaseSkill()
			s.SetID("addall-nil-1")
			s.SetDescription("ok")
			return s
		}(),
		nil,
	}

	if err := reg.AddAll(skills); err == nil {
		t.Error("expected error for nil runnable in AddAll() slice, got nil")
	}
}

func TestRegistry_Remove(t *testing.T) {
	reg := types.NewRegistry()

	s := types.NewBaseSkill()
	s.SetID("remove-me")
	s.SetDescription("to be removed")
	if err := reg.Set(s); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Confirm it's there.
	if _, ok := reg.FindByID("remove-me"); !ok {
		t.Fatal("expected runnable to be present before Remove()")
	}

	if err := reg.Remove("remove-me"); err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	if _, ok := reg.FindByID("remove-me"); ok {
		t.Error("expected runnable to be absent after Remove()")
	}
}

func TestRegistry_Remove_AlsoClearsDisableRecord(t *testing.T) {
	reg := types.NewRegistry()

	s := types.NewBaseSkill()
	s.SetID("remove-disabled")
	s.SetDescription("disabled then removed")
	if err := reg.Set(s); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}
	if err := reg.Disable("remove-disabled", "temp"); err != nil {
		t.Fatalf("Disable() failed: %v", err)
	}

	if err := reg.Remove("remove-disabled"); err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// Re-add a runnable with the same ID; it must NOT be disabled anymore.
	s2 := types.NewBaseSkill()
	s2.SetID("remove-disabled")
	s2.SetDescription("fresh")
	if err := reg.Set(s2); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}
	if _, disabled := reg.IsDisabled("remove-disabled"); disabled {
		t.Error("expected disable record to be cleared by Remove()")
	}
}

func TestRegistry_Remove_UnknownID(t *testing.T) {
	reg := types.NewRegistry()

	if err := reg.Remove("does-not-exist"); err == nil {
		t.Error("expected error when removing unknown ID, got nil")
	}
}

func TestRegistry_Remove_EmptyID(t *testing.T) {
	reg := types.NewRegistry()

	if err := reg.Remove(""); err == nil {
		t.Error("expected error when removing empty ID, got nil")
	}
}

func TestRegistry_RemoveAll(t *testing.T) {
	reg := types.NewRegistry()

	for _, id := range []string{"rm-1", "rm-2", "rm-3", "keep"} {
		s := types.NewBaseSkill()
		s.SetID(id)
		s.SetDescription(id)
		if err := reg.Set(s); err != nil {
			t.Fatalf("Set(%s) failed: %v", id, err)
		}
	}

	if err := reg.RemoveAll([]string{"rm-1", "rm-2", "rm-3"}); err != nil {
		t.Fatalf("RemoveAll() failed: %v", err)
	}

	for _, id := range []string{"rm-1", "rm-2", "rm-3"} {
		if _, ok := reg.FindByID(id); ok {
			t.Errorf("expected '%s' to be removed", id)
		}
	}
	if _, ok := reg.FindByID("keep"); !ok {
		t.Error("expected 'keep' to still be registered")
	}
}

func TestRegistry_RemoveAll_IgnoresUnknownIDs(t *testing.T) {
	reg := types.NewRegistry()

	s := types.NewBaseSkill()
	s.SetID("keep-2")
	s.SetDescription("kept")
	if err := reg.Set(s); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Mixed known/unknown IDs — unknown ones are ignored, no error.
	if err := reg.RemoveAll([]string{"keep-2", "ghost"}); err != nil {
		t.Errorf("expected RemoveAll to ignore unknown IDs, got: %v", err)
	}
	if _, ok := reg.FindByID("keep-2"); ok {
		t.Error("expected 'keep-2' to be removed even when mixed with unknown IDs")
	}
}

func TestRegistry_RemoveAll_EmptyID(t *testing.T) {
	reg := types.NewRegistry()

	s := types.NewBaseSkill()
	s.SetID("should-stay")
	s.SetDescription("untouched")
	if err := reg.Set(s); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Empty ID in the slice should error and perform no removals.
	if err := reg.RemoveAll([]string{"should-stay", ""}); err == nil {
		t.Error("expected error for empty ID in RemoveAll(), got nil")
	}
	if _, ok := reg.FindByID("should-stay"); !ok {
		t.Error("expected no removals to occur when RemoveAll() errors on empty ID")
	}
}

func TestRegistry_RemoveAll_EmptySlice(t *testing.T) {
	reg := types.NewRegistry()

	if err := reg.RemoveAll([]string{}); err != nil {
		t.Errorf("expected no error for empty slice, got: %v", err)
	}
}

func TestRegistry_MergeReplaceOnOverlap_AddsNewRunnables(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	s1 := types.NewBaseSkill().SetID("src-1").SetDescription("from src")
	s2 := types.NewBaseSkill().SetID("src-2").SetDescription("from src")
	if err := src.SetAll([]types.RunnableInterface{s1, s2}); err != nil {
		t.Fatalf("src.SetAll() failed: %v", err)
	}

	if err := target.MergeReplaceOnOverlap(src); err != nil {
		t.Fatalf("MergeReplaceOnOverlap() failed: %v", err)
	}

	for _, id := range []string{"src-1", "src-2"} {
		if _, ok := target.FindByID(id); !ok {
			t.Errorf("expected '%s' to be in target after merge", id)
		}
	}
}

func TestRegistry_MergeReplaceOnOverlap_SourceWinsOnConflict(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	tSkill := types.NewBaseSkill().SetID("shared").SetDescription("from target")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}

	sSkill := types.NewBaseSkill().SetID("shared").SetDescription("from src")
	if err := src.Set(sSkill); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}

	if err := target.MergeReplaceOnOverlap(src); err != nil {
		t.Fatalf("MergeReplaceOnOverlap() failed: %v", err)
	}

	found, ok := target.FindByID("shared")
	if !ok {
		t.Fatal("expected 'shared' to be present after merge")
	}
	if found.GetDescription() != "from src" {
		t.Errorf("expected source to win, got '%s'", found.GetDescription())
	}
}

func TestRegistry_MergeReplaceOnOverlap_PreservesTargetOnlyRunnables(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	tSkill := types.NewBaseSkill().SetID("target-only").SetDescription("stays")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}

	sSkill := types.NewBaseSkill().SetID("src-only").SetDescription("added")
	if err := src.Set(sSkill); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}

	if err := target.MergeReplaceOnOverlap(src); err != nil {
		t.Fatalf("MergeReplaceOnOverlap() failed: %v", err)
	}

	if _, ok := target.FindByID("target-only"); !ok {
		t.Error("expected target-only runnable to survive merge")
	}
	if _, ok := target.FindByID("src-only"); !ok {
		t.Error("expected src-only runnable to be added by merge")
	}
}

func TestRegistry_MergeReplaceOnOverlap_DisableStateCarriesOver(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	sDisabled := types.NewBaseSkill().SetID("src-disabled").SetDescription("disabled in src")
	if err := src.Set(sDisabled); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}
	if err := src.Disable("src-disabled", "broken"); err != nil {
		t.Fatalf("src.Disable() failed: %v", err)
	}

	sEnabled := types.NewBaseSkill().SetID("src-enabled").SetDescription("enabled in src")
	if err := src.Set(sEnabled); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}

	if err := target.MergeReplaceOnOverlap(src); err != nil {
		t.Fatalf("MergeReplaceOnOverlap() failed: %v", err)
	}

	if comment, disabled := target.IsDisabled("src-disabled"); !disabled {
		t.Error("expected 'src-disabled' to be disabled after merge")
	} else if comment != "broken" {
		t.Errorf("expected disable comment 'broken', got '%s'", comment)
	}

	if _, disabled := target.IsDisabled("src-enabled"); disabled {
		t.Error("expected 'src-enabled' to NOT be disabled after merge")
	}
}

func TestRegistry_MergeReplaceOnOverlap_SourceDisableWinsOverTargetEnabled(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	tSkill := types.NewBaseSkill().SetID("shared").SetDescription("from target")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}

	sSkill := types.NewBaseSkill().SetID("shared").SetDescription("from src")
	if err := src.Set(sSkill); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}
	if err := src.Disable("shared", "src says no"); err != nil {
		t.Fatalf("src.Disable() failed: %v", err)
	}

	if err := target.MergeReplaceOnOverlap(src); err != nil {
		t.Fatalf("MergeReplaceOnOverlap() failed: %v", err)
	}

	if comment, disabled := target.IsDisabled("shared"); !disabled {
		t.Error("expected source disable state to win — 'shared' should be disabled")
	} else if comment != "src says no" {
		t.Errorf("expected 'src says no', got '%s'", comment)
	}
}

func TestRegistry_MergeReplaceOnOverlap_SourceEnableWinsOverTargetDisabled(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	tSkill := types.NewBaseSkill().SetID("shared").SetDescription("from target")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}
	if err := target.Disable("shared", "target disabled it"); err != nil {
		t.Fatalf("target.Disable() failed: %v", err)
	}

	sSkill := types.NewBaseSkill().SetID("shared").SetDescription("from src")
	if err := src.Set(sSkill); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}

	if err := target.MergeReplaceOnOverlap(src); err != nil {
		t.Fatalf("MergeReplaceOnOverlap() failed: %v", err)
	}

	if _, disabled := target.IsDisabled("shared"); disabled {
		t.Error("expected source enable state to win — 'shared' should be enabled after merge")
	}
}

func TestRegistry_MergeReplaceOnOverlap_TargetOnlyDisableStatePreserved(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	tSkill := types.NewBaseSkill().SetID("target-only-disabled").SetDescription("disabled in target")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}
	if err := target.Disable("target-only-disabled", "target reason"); err != nil {
		t.Fatalf("target.Disable() failed: %v", err)
	}

	sSkill := types.NewBaseSkill().SetID("src-only").SetDescription("from src")
	if err := src.Set(sSkill); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}

	if err := target.MergeReplaceOnOverlap(src); err != nil {
		t.Fatalf("MergeReplaceOnOverlap() failed: %v", err)
	}

	if comment, disabled := target.IsDisabled("target-only-disabled"); !disabled {
		t.Error("expected target-only runnable to keep its disabled state")
	} else if comment != "target reason" {
		t.Errorf("expected 'target reason', got '%s'", comment)
	}
}

func TestRegistry_MergeReplaceOnOverlap_NilSource(t *testing.T) {
	target := types.NewRegistry()

	if err := target.MergeReplaceOnOverlap(nil); err == nil {
		t.Error("expected error when merging from nil registry, got nil")
	}
}

func TestRegistry_MergeReplaceOnOverlap_EmptySource(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	tSkill := types.NewBaseSkill().SetID("pre-existing").SetDescription("was here")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}

	if err := target.MergeReplaceOnOverlap(src); err != nil {
		t.Fatalf("MergeReplaceOnOverlap() with empty src failed: %v", err)
	}

	if _, ok := target.FindByID("pre-existing"); !ok {
		t.Error("expected pre-existing runnable to survive merge with empty src")
	}
}

// --- MergeKeepOnOverlap ---

func TestRegistry_MergeKeepOnOverlap_AddsNewRunnables(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	s := types.NewBaseSkill().SetID("src-only").SetDescription("from src")
	if err := src.Set(s); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}

	if err := target.MergeKeepOnOverlap(src); err != nil {
		t.Fatalf("MergeKeepOnOverlap() failed: %v", err)
	}

	if _, ok := target.FindByID("src-only"); !ok {
		t.Error("expected 'src-only' to be added to target")
	}
}

func TestRegistry_MergeKeepOnOverlap_KeepsTargetOnConflict(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	tSkill := types.NewBaseSkill().SetID("shared").SetDescription("from target")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}

	sSkill := types.NewBaseSkill().SetID("shared").SetDescription("from src")
	if err := src.Set(sSkill); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}

	if err := target.MergeKeepOnOverlap(src); err != nil {
		t.Fatalf("MergeKeepOnOverlap() failed: %v", err)
	}

	found, _ := target.FindByID("shared")
	if found.GetDescription() != "from target" {
		t.Errorf("expected target to win, got '%s'", found.GetDescription())
	}
}

func TestRegistry_MergeKeepOnOverlap_KeepsTargetDisableStateOnConflict(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	// Target has "shared" enabled.
	tSkill := types.NewBaseSkill().SetID("shared").SetDescription("from target")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}

	// Source has "shared" disabled — but target should win.
	sSkill := types.NewBaseSkill().SetID("shared").SetDescription("from src")
	if err := src.Set(sSkill); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}
	if err := src.Disable("shared", "src says no"); err != nil {
		t.Fatalf("src.Disable() failed: %v", err)
	}

	if err := target.MergeKeepOnOverlap(src); err != nil {
		t.Fatalf("MergeKeepOnOverlap() failed: %v", err)
	}

	if _, disabled := target.IsDisabled("shared"); disabled {
		t.Error("expected target's enabled state to win — 'shared' should NOT be disabled")
	}
}

func TestRegistry_MergeKeepOnOverlap_CopiesDisableStateForNewIDs(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	s := types.NewBaseSkill().SetID("src-disabled").SetDescription("disabled in src")
	if err := src.Set(s); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}
	if err := src.Disable("src-disabled", "broken"); err != nil {
		t.Fatalf("src.Disable() failed: %v", err)
	}

	if err := target.MergeKeepOnOverlap(src); err != nil {
		t.Fatalf("MergeKeepOnOverlap() failed: %v", err)
	}

	if comment, disabled := target.IsDisabled("src-disabled"); !disabled {
		t.Error("expected newly-added 'src-disabled' to be disabled in target")
	} else if comment != "broken" {
		t.Errorf("expected 'broken', got '%s'", comment)
	}
}

func TestRegistry_MergeKeepOnOverlap_NilSource(t *testing.T) {
	target := types.NewRegistry()

	if err := target.MergeKeepOnOverlap(nil); err == nil {
		t.Error("expected error when merging from nil registry, got nil")
	}
}

// --- MergeNoOverlap ---

func TestRegistry_MergeNoOverlap_AddsNewRunnables(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	tSkill := types.NewBaseSkill().SetID("target-only").SetDescription("target")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}

	sSkill := types.NewBaseSkill().SetID("src-only").SetDescription("src")
	if err := src.Set(sSkill); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}

	if err := target.MergeNoOverlap(src); err != nil {
		t.Fatalf("MergeNoOverlap() failed: %v", err)
	}

	if _, ok := target.FindByID("src-only"); !ok {
		t.Error("expected 'src-only' to be added")
	}
	if _, ok := target.FindByID("target-only"); !ok {
		t.Error("expected 'target-only' to survive")
	}
}

func TestRegistry_MergeNoOverlap_ErrorsOnConflict(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	tSkill := types.NewBaseSkill().SetID("shared").SetDescription("from target")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}

	sSkill := types.NewBaseSkill().SetID("shared").SetDescription("from src")
	if err := src.Set(sSkill); err != nil {
		t.Fatalf("src.Set() failed: %v", err)
	}

	err := target.MergeNoOverlap(src)
	if err == nil {
		t.Fatal("expected error on overlap, got nil")
	}

	// Target should be unchanged — no mutation on error.
	found, _ := target.FindByID("shared")
	if found.GetDescription() != "from target" {
		t.Errorf("expected target unchanged on error, got '%s'", found.GetDescription())
	}
}

func TestRegistry_MergeNoOverlap_NoMutationOnConflict(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	// Target has "keep-me".
	tSkill := types.NewBaseSkill().SetID("keep-me").SetDescription("target")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}

	// Source has "keep-me" (overlap) AND "new-one" (no overlap).
	s1 := types.NewBaseSkill().SetID("keep-me").SetDescription("src")
	s2 := types.NewBaseSkill().SetID("new-one").SetDescription("src")
	if err := src.SetAll([]types.RunnableInterface{s1, s2}); err != nil {
		t.Fatalf("src.SetAll() failed: %v", err)
	}

	err := target.MergeNoOverlap(src)
	if err == nil {
		t.Fatal("expected error on overlap, got nil")
	}

	// Neither "new-one" should be added nor "keep-me" replaced.
	if _, ok := target.FindByID("new-one"); ok {
		t.Error("expected no additions when merge aborts on overlap")
	}
	found, _ := target.FindByID("keep-me")
	if found.GetDescription() != "target" {
		t.Errorf("expected 'keep-me' unchanged, got '%s'", found.GetDescription())
	}
}

func TestRegistry_MergeNoOverlap_NilSource(t *testing.T) {
	target := types.NewRegistry()

	if err := target.MergeNoOverlap(nil); err == nil {
		t.Error("expected error when merging from nil registry, got nil")
	}
}

func TestRegistry_MergeNoOverlap_EmptySource(t *testing.T) {
	target := types.NewRegistry()
	src := types.NewRegistry()

	tSkill := types.NewBaseSkill().SetID("pre-existing").SetDescription("was here")
	if err := target.Set(tSkill); err != nil {
		t.Fatalf("target.Set() failed: %v", err)
	}

	if err := target.MergeNoOverlap(src); err != nil {
		t.Fatalf("MergeNoOverlap() with empty src failed: %v", err)
	}

	if _, ok := target.FindByID("pre-existing"); !ok {
		t.Error("expected pre-existing runnable to survive")
	}
}
