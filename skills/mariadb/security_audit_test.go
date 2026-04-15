package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestSecurityAudit_Run_DryRun verifies that dry-run mode correctly handles security audit.
func TestSecurityAudit_Run_DryRun(t *testing.T) {
	pb := NewSecurityAudit()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Security audit is a read-only operation, so Changed should be false even in dry-run
	if result.Changed {
		t.Error("Expected Changed to be false in dry-run mode for read-only operation")
	}

	expectedMessage := "Would perform MariaDB security audit"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestSecurityAudit_Run_DryRun_NoPassword verifies dry-run without password returns error.
func TestSecurityAudit_Run_DryRun_NoPassword(t *testing.T) {
	pb := NewSecurityAudit()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for missing root-password even in dry-run
	if result.Error == nil {
		t.Error("Expected error for missing root-password")
	}

	if result.Message != "MariaDB root password not provided" {
		t.Errorf("Expected message 'MariaDB root password not provided', got '%s'", result.Message)
	}
}

// TestSecurityAudit_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestSecurityAudit_Run_NotDryRun(t *testing.T) {
	pb := NewSecurityAudit()
	pb.SetArg("root-password", "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would perform MariaDB security audit" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}

	// Security audit is a read-only operation, so Changed should always be false
	if result.Changed {
		t.Error("Expected Changed to be false for read-only operation")
	}
}

// TestSecurityAudit_Check verifies that Check returns false for read-only operation.
func TestSecurityAudit_Check(t *testing.T) {
	pb := NewSecurityAudit()

	cfg := types.NodeConfig{
		Logger: slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if needsChange {
		t.Error("Expected Check to return false for read-only operation")
	}
}

// TestSecurityAudit_NewSecurityAudit verifies that NewSecurityAudit creates a properly configured skill.
func TestSecurityAudit_NewSecurityAudit(t *testing.T) {
	pb := NewSecurityAudit()

	if pb.GetID() != "mariadb-security-audit" {
		t.Errorf("Expected ID to be 'mariadb-security-audit', got '%s'", pb.GetID())
	}

	expectedDescription := "Perform a comprehensive security audit of MariaDB (read-only)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
