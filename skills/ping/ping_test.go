package ping

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestPing_Run_DryRun verifies that dry-run mode correctly handles ping.
func TestPing_Run_DryRun(t *testing.T) {
	pb := NewPing()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
		SSHHost:      "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Ping is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would ping: test.example.com"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestPing_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestPing_Run_NotDryRun(t *testing.T) {
	pb := NewPing()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
		SSHHost:      "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would ping: test.example.com" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// Ping is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestPing_Check verifies that Check returns false for read-only operation.
func TestPing_Check(t *testing.T) {
	pb := NewPing()

	cfg := types.NodeConfig{
		Logger:  slog.Default(),
		SSHHost: "test.example.com",
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	// Since there's no real SSH server, we expect an error
	if err == nil {
		t.Log("SSH connection succeeded in test environment")
	}

	// Ping is read-only, so should always return false
	if needsChange {
		t.Error("Expected Check to return false for read-only operation")
	}
}

// TestPing_NewPing verifies that NewPing creates a properly configured skill.
func TestPing_NewPing(t *testing.T) {
	pb := NewPing()

	if pb.GetID() != "ping" {
		t.Errorf("Expected ID to be 'ping', got '%s'", pb.GetID())
	}

	expectedDescription := "Check SSH connectivity and show server uptime/load"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
