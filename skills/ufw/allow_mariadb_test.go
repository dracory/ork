package ufw

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
)

// TestAllowMariaDB_Run_DryRun_AnyIP verifies that dry-run mode correctly handles allowing MariaDB from any IP.
func TestAllowMariaDB_Run_DryRun_AnyIP(t *testing.T) {
	pb := NewAllowMariaDB()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would configure UFW for MariaDB port 3306"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Details == nil {
		t.Fatal("Expected Details to be non-nil")
	}

	if result.Details["allowed_ips"] != "any" {
		t.Errorf("Expected allowed_ips to be 'any', got '%s'", result.Details["allowed_ips"])
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAllowMariaDB_Run_DryRun_SpecificIP verifies that dry-run mode correctly handles allowing MariaDB from specific IPs.
func TestAllowMariaDB_Run_DryRun_SpecificIP(t *testing.T) {
	pb := NewAllowMariaDB()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgIP: "192.168.1.10",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would configure UFW for MariaDB port 3306"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Details == nil {
		t.Fatal("Expected Details to be non-nil")
	}

	if result.Details["allowed_ips"] != "192.168.1.10" {
		t.Errorf("Expected allowed_ips to be '192.168.1.10', got '%s'", result.Details["allowed_ips"])
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAllowMariaDB_Run_DryRun_MultipleIPs verifies that dry-run mode correctly handles allowing MariaDB from multiple IPs.
func TestAllowMariaDB_Run_DryRun_MultipleIPs(t *testing.T) {
	pb := NewAllowMariaDB()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgIP: "192.168.1.10,192.168.1.11,192.168.1.12",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would configure UFW for MariaDB port 3306"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Details == nil {
		t.Fatal("Expected Details to be non-nil")
	}

	expectedIPs := "192.168.1.10,192.168.1.11,192.168.1.12"
	if result.Details["allowed_ips"] != expectedIPs {
		t.Errorf("Expected allowed_ips to be '%s', got '%s'", expectedIPs, result.Details["allowed_ips"])
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAllowMariaDB_Run_DryRun_CustomPort verifies that dry-run mode correctly handles custom port.
func TestAllowMariaDB_Run_DryRun_CustomPort(t *testing.T) {
	pb := NewAllowMariaDB()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgPort: "3307",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would configure UFW for MariaDB port 3307"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Details == nil {
		t.Fatal("Expected Details to be non-nil")
	}

	if result.Details["allowed_ips"] != "any" {
		t.Errorf("Expected allowed_ips to be 'any', got '%s'", result.Details["allowed_ips"])
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAllowMariaDB_Run_DryRun_SpecificIPAndCustomPort verifies dry-run with both specific IP and custom port.
func TestAllowMariaDB_Run_DryRun_SpecificIPAndCustomPort(t *testing.T) {
	pb := NewAllowMariaDB()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgIP:   "10.0.0.5",
			ArgPort: "3307",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would configure UFW for MariaDB port 3307"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Details == nil {
		t.Fatal("Expected Details to be non-nil")
	}

	if result.Details["allowed_ips"] != "10.0.0.5" {
		t.Errorf("Expected allowed_ips to be '10.0.0.5', got '%s'", result.Details["allowed_ips"])
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAllowMariaDB_Run_DryRun_IPWithWhitespace verifies that dry-run mode correctly trims whitespace from IP list.
func TestAllowMariaDB_Run_DryRun_IPWithWhitespace(t *testing.T) {
	pb := NewAllowMariaDB()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgIP: " 192.168.1.10 , 192.168.1.11 ",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	if result.Details == nil {
		t.Fatal("Expected Details to be non-nil")
	}

	expectedIPs := "192.168.1.10,192.168.1.11"
	if result.Details["allowed_ips"] != expectedIPs {
		t.Errorf("Expected allowed_ips to be '%s', got '%s'", expectedIPs, result.Details["allowed_ips"])
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestAllowMariaDB_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestAllowMariaDB_Run_NotDryRun(t *testing.T) {
	pb := NewAllowMariaDB()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgIP: "192.168.1.10",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would configure UFW for MariaDB port 3306" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestAllowMariaDB_Check verifies that Check always returns true.
func TestAllowMariaDB_Check(t *testing.T) {
	pb := NewAllowMariaDB()

	cfg := types.NodeConfig{
		Logger: slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	if !needsChange {
		t.Error("Expected Check to return true")
	}
}

// TestAllowMariaDB_NewAllowMariaDB verifies that NewAllowMariaDB creates a properly configured skill.
func TestAllowMariaDB_NewAllowMariaDB(t *testing.T) {
	pb := NewAllowMariaDB()

	if pb.GetID() != "ufw-allow-mariadb" {
		t.Errorf("Expected ID to be 'ufw-allow-mariadb', got '%s'", pb.GetID())
	}

	expectedDescription := "Configure UFW firewall rules for MariaDB access"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}
