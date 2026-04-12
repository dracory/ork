package ork

import (
	"testing"
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

	for _, name := range expectedPlaybooks {
		pb, ok := defaultRegistry.Get(name)
		if !ok {
			t.Errorf("expected playbook '%s' to be registered, but it was not found", name)
			continue
		}
		if pb.Name() != name {
			t.Errorf("playbook name mismatch: expected '%s', got '%s'", name, pb.Name())
		}
	}
}

func TestDefaultRegistry_ContainsExpectedPlaybookNames(t *testing.T) {
	names := defaultRegistry.Names()

	// Verify all expected built-in playbook names are present
	// Note: The registry may contain additional test playbooks from other tests,
	// so we only verify that the built-in playbooks exist, not the exact count.
	expectedNames := []string{
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

	// Create a map of actual names for quick lookup
	actualNames := make(map[string]bool)
	for _, name := range names {
		actualNames[name] = true
	}

	// Check that all expected names are present
	for _, name := range expectedNames {
		if !actualNames[name] {
			t.Errorf("expected built-in playbook '%s' not found in registry", name)
		}
	}

	// Verify we have at least the expected number of built-in playbooks
	if len(names) < len(expectedNames) {
		t.Errorf("expected at least %d playbooks, got %d", len(expectedNames), len(names))
	}
}

func TestDefaultRegistry_PlaybooksHaveDescriptions(t *testing.T) {
	playbooks := defaultRegistry.List()

	for _, pb := range playbooks {
		if pb.Description() == "" {
			t.Errorf("playbook '%s' has empty description", pb.Name())
		}
	}
}
