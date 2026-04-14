package ork

import (
	"testing"

	"github.com/dracory/ork/playbooks"
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

func TestNewDefaultRegistry_AllBuiltInPlaybooksRegistered(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}
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
		pb, ok := reg.PlaybookFindByID(id)
		if !ok {
			t.Errorf("expected playbook '%s' to be registered, but it was not found", id)
			continue
		}
		if pb.GetID() != id {
			t.Errorf("playbook ID mismatch: expected '%s', got '%s'", id, pb.GetID())
		}
	}
}

func TestNewDefaultRegistry_ContainsExpectedPlaybookIDs(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}
	ids := reg.GetPlaybookIDs()

	// Verify all expected built-in playbook IDs are present
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

func TestNewDefaultRegistry_PlaybooksHaveDescriptions(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}
	playbooks := reg.PlaybookList()

	for _, pb := range playbooks {
		if pb.GetDescription() == "" {
			t.Errorf("playbook '%s' has empty description", pb.GetID())
		}
	}
}

func TestGetGlobalPlaybookRegistry(t *testing.T) {
	// Create a fresh registry for this test to avoid polluting the global registry
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}
	if reg == nil {
		t.Fatal("NewDefaultRegistry() returned nil")
	}

	// Test that we can use it to register a playbook
	customPb := playbooks.NewBasePlaybook()
	customPb.SetID("test-get-registry-playbook")
	customPb.SetDescription("Test playbook via NewDefaultRegistry")

	err = reg.PlaybookRegister(customPb)
	if err != nil {
		t.Fatalf("failed to register playbook: %v", err)
	}

	// Verify it can be found
	foundPb, ok := reg.PlaybookFindByID("test-get-registry-playbook")
	if !ok {
		t.Fatal("custom playbook not found after registration")
	}
	if foundPb.GetID() != "test-get-registry-playbook" {
		t.Errorf("expected ID 'test-get-registry-playbook', got '%s'", foundPb.GetID())
	}
}

func TestGetGlobalPlaybookRegistry_LazyInitialization(t *testing.T) {
	// Test that GetGlobalPlaybookRegistry() initializes the global registry on first call
	reg, err := GetGlobalPlaybookRegistry()
	if err != nil {
		t.Fatalf("GetGlobalPlaybookRegistry() failed: %v", err)
	}
	if reg == nil {
		t.Fatal("GetGlobalPlaybookRegistry() returned nil")
	}

	// Verify it has built-in playbooks
	pb, ok := reg.PlaybookFindByID("ping")
	if !ok {
		t.Fatal("expected 'ping' playbook in global registry")
	}
	if pb.GetID() != "ping" {
		t.Errorf("expected ID 'ping', got '%s'", pb.GetID())
	}

	// Test that subsequent calls return the same instance
	reg2, err := GetGlobalPlaybookRegistry()
	if err != nil {
		t.Fatalf("GetGlobalPlaybookRegistry() failed on second call: %v", err)
	}
	if reg != reg2 {
		t.Error("GetGlobalPlaybookRegistry() should return the same instance on subsequent calls")
	}
}

func TestNewDefaultRegistry_DuplicateID(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("NewDefaultRegistry() failed: %v", err)
	}

	// Try to register a playbook with a duplicate ID
	duplicatePb := playbooks.NewBasePlaybook()
	duplicatePb.SetID("ping") // "ping" is already registered
	duplicatePb.SetDescription("Duplicate ping playbook")

	err = reg.PlaybookRegister(duplicatePb)
	if err == nil {
		t.Error("expected error when registering duplicate playbook ID, got nil")
	}
}
