package ork

import (
	"testing"

	"github.com/dracory/ork/types"
)

// TestClonePreservesID verifies that ToMap/FromMap round-trip preserves the ID.
// This is a regression test for a bug where the ID was silently lost during
// cloning because atom.ToMap() strips "id" from properties and puts the
// Atom's auto-generated id field in the result instead.
//
// Before the fix: clone.GetID() returned "" (empty)
// After the fix:  clone.GetID() returns the original ID
func TestClonePreservesID(t *testing.T) {
	original := types.NewBaseSkill().
		WithID("my-skill").
		WithDescription("test description").
		WithBecomeUser("root")

	// Clone via ToMap/FromMap
	m := original.ToMap()
	clone := types.NewBaseSkill()
	clone.FromMap(m)

	t.Logf("Original ID: %q", original.GetID())
	t.Logf("Clone ID:    %q", clone.GetID())
	t.Logf("Original Desc: %q", original.GetDescription())
	t.Logf("Clone Desc:    %q", clone.GetDescription())
	t.Logf("Original BecomeUser: %q", original.GetBecomeUser())
	t.Logf("Clone BecomeUser:    %q", clone.GetBecomeUser())

	if clone.GetID() != original.GetID() {
		t.Errorf("ID lost during clone: original=%q, clone=%q", original.GetID(), clone.GetID())
	}
	if clone.GetDescription() != original.GetDescription() {
		t.Errorf("Description lost during clone: original=%q, clone=%q", original.GetDescription(), clone.GetDescription())
	}
	if clone.GetBecomeUser() != original.GetBecomeUser() {
		t.Errorf("BecomeUser lost during clone: original=%q, clone=%q", original.GetBecomeUser(), clone.GetBecomeUser())
	}
}

// TestClonePreservesID_Playbook verifies the same for BasePlaybook.
func TestClonePreservesID_Playbook(t *testing.T) {
	original := types.NewBasePlaybook().
		WithID("my-playbook").
		WithDescription("test playbook").
		WithBecomeUser("admin")

	// Clone via ToMap/FromMap
	m := original.ToMap()
	clone := types.NewBasePlaybook()
	clone.FromMap(m)

	t.Logf("Original ID: %q", original.GetID())
	t.Logf("Clone ID:    %q", clone.GetID())

	if clone.GetID() != original.GetID() {
		t.Errorf("ID lost during clone: original=%q, clone=%q", original.GetID(), clone.GetID())
	}
	if clone.GetDescription() != original.GetDescription() {
		t.Errorf("Description lost during clone: original=%q, clone=%q", original.GetDescription(), clone.GetDescription())
	}
	if clone.GetBecomeUser() != original.GetBecomeUser() {
		t.Errorf("BecomeUser lost during clone: original=%q, clone=%q", original.GetBecomeUser(), clone.GetBecomeUser())
	}
}

// TestClonePreservesID_Command verifies the same for commandImplementation,
// which has custom fields (command, required, chdir) beyond BaseSkill.
func TestClonePreservesID_Command(t *testing.T) {
	original := NewCommand().
		WithID("my-command").
		WithCommand("ls -la").
		WithRequired(true).
		WithChdir("/tmp").
		WithBecomeUser("root")

	// Clone via ToMap/FromMap
	m := original.ToMap()
	clone := NewCommand()
	clone.FromMap(m)

	t.Logf("Original ID: %q", original.GetID())
	t.Logf("Clone ID:    %q", clone.GetID())

	if clone.GetID() != original.GetID() {
		t.Errorf("ID lost during clone: original=%q, clone=%q", original.GetID(), clone.GetID())
	}
	if clone.GetBecomeUser() != "root" {
		t.Errorf("BecomeUser lost during clone: got %q", clone.GetBecomeUser())
	}
	// Verify command-specific fields survived the round-trip by checking
	// the ToMap output (command/required/chdir are unexported on CommandInterface)
	cloneMap := clone.ToMap()
	if cmd, ok := cloneMap[types.MapKeyCommand].(string); !ok || cmd != "ls -la" {
		t.Errorf("Command lost during clone: got %v", cloneMap[types.MapKeyCommand])
	}
	if req, ok := cloneMap[types.MapKeyRequired].(string); !ok || req != "true" {
		t.Errorf("Required flag lost during clone: got %v", cloneMap[types.MapKeyRequired])
	}
	if dir, ok := cloneMap[types.MapKeyChdir].(string); !ok || dir != "/tmp" {
		t.Errorf("Chdir lost during clone: got %v", cloneMap[types.MapKeyChdir])
	}
}
