package ork

import (
	"testing"

	"github.com/dracory/ork/playbook"
)

func TestDefaultRegistry_Initialized(t *testing.T) {
	if defaultRegistry == nil {
		t.Fatal("defaultRegistry should be initialized")
	}
}

func TestDefaultRegistry_AllBuiltInPlaybooksRegistered(t *testing.T) {
	expectedPlaybooks := []string{
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

	for _, id := range expectedPlaybooks {
		pb, ok := defaultRegistry.PlaybookFindByID(id)
		if !ok {
			t.Errorf("expected playbook '%s' to be registered, but it was not found", id)
			continue
		}
		if pb.GetID() != id {
			t.Errorf("playbook ID mismatch: expected '%s', got '%s'", id, pb.GetID())
		}
	}
}

func TestDefaultRegistry_ContainsExpectedPlaybookIDs(t *testing.T) {
	ids := defaultRegistry.GetPlaybookIDs()

	// Verify all expected built-in playbook IDs are present
	// Note: The registry may contain additional test playbooks from other tests,
	// so we only verify that the built-in playbooks exist, not the exact count.
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
			t.Errorf("expected built-in playbook '%s' not found in registry", id)
		}
	}

	// Verify we have at least the expected number of built-in playbooks
	if len(ids) < len(expectedIDs) {
		t.Errorf("expected at least %d playbooks, got %d", len(expectedIDs), len(ids))
	}
}

func TestDefaultRegistry_PlaybooksHaveDescriptions(t *testing.T) {
	playbooks := defaultRegistry.PlaybookList()

	for _, pb := range playbooks {
		if pb.GetDescription() == "" {
			t.Errorf("playbook '%s' has empty description", pb.GetID())
		}
	}
}

func TestRegisterPlaybook_Success(t *testing.T) {
	// Create a custom playbook
	customPb := playbook.NewBasePlaybook()
	customPb.SetID("test-custom-playbook")
	customPb.SetDescription("A test custom playbook")

	// Register it
	RegisterPlaybook(customPb)

	// Verify it can be found
	foundPb, ok := defaultRegistry.PlaybookFindByID("test-custom-playbook")
	if !ok {
		t.Fatal("custom playbook not found after registration")
	}
	if foundPb.GetID() != "test-custom-playbook" {
		t.Errorf("expected ID 'test-custom-playbook', got '%s'", foundPb.GetID())
	}
	if foundPb.GetDescription() != "A test custom playbook" {
		t.Errorf("expected description 'A test custom playbook', got '%s'", foundPb.GetDescription())
	}
}

func TestRegisterPlaybook_NilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when registering nil playbook, but none occurred")
		}
	}()

	RegisterPlaybook(nil)
}
