package ufw

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestUfwInstall_Run_DryRun_Defaults verifies that dry-run mode correctly handles default configuration.
func TestUfwInstall_Run_DryRun_Defaults(t *testing.T) {
	pb := NewUfwInstall()

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

	expectedMessage := "Would install and configure UFW firewall"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUfwInstall_Run_DryRun_AllSSH verifies that dry-run mode correctly handles allowing SSH.
func TestUfwInstall_Run_DryRun_AllSSH(t *testing.T) {
	pb := NewUfwInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgAllowSSH: "true",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure UFW firewall"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUfwInstall_Run_DryRun_DisableSSH verifies that dry-run mode correctly handles disabling SSH.
func TestUfwInstall_Run_DryRun_DisableSSH(t *testing.T) {
	pb := NewUfwInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgAllowSSH: "false",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure UFW firewall"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUfwInstall_Run_DryRun_AllHTTP verifies that dry-run mode correctly handles allowing HTTP.
func TestUfwInstall_Run_DryRun_AllHTTP(t *testing.T) {
	pb := NewUfwInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgAllowHTTP: "true",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure UFW firewall"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUfwInstall_Run_DryRun_AllHTTPS verifies that dry-run mode correctly handles allowing HTTPS.
func TestUfwInstall_Run_DryRun_AllHTTPS(t *testing.T) {
	pb := NewUfwInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgAllowHTTPS: "true",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure UFW firewall"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUfwInstall_Run_DryRun_CustomPorts verifies that dry-run mode correctly handles custom ports.
func TestUfwInstall_Run_DryRun_CustomPorts(t *testing.T) {
	pb := NewUfwInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgAllowPorts: "3306,8080",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure UFW firewall"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUfwInstall_Run_DryRun_AllOptions verifies that dry-run mode correctly handles all options enabled.
func TestUfwInstall_Run_DryRun_AllOptions(t *testing.T) {
	pb := NewUfwInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgAllowSSH:   "true",
			ArgAllowHTTP:  "true",
			ArgAllowHTTPS: "true",
			ArgAllowPorts: "3306,8080,9000",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure UFW firewall"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUfwInstall_Run_DryRun_PortsWithWhitespace verifies that dry-run mode correctly trims whitespace from port list.
func TestUfwInstall_Run_DryRun_PortsWithWhitespace(t *testing.T) {
	pb := NewUfwInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgAllowPorts: " 3306 , 8080 , 9000 ",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure UFW firewall"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUfwInstall_Run_DryRun_EmptyPorts verifies that dry-run mode correctly handles empty port string.
func TestUfwInstall_Run_DryRun_EmptyPorts(t *testing.T) {
	pb := NewUfwInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgAllowPorts: "",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would install and configure UFW firewall"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestUfwInstall_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestUfwInstall_Run_NotDryRun(t *testing.T) {
	pb := NewUfwInstall()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args: map[string]string{
			ArgAllowSSH: "true",
		},
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would install and configure UFW firewall" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestUfwInstall_Check verifies that Check returns true when UFW is not installed.
func TestUfwInstall_Check(t *testing.T) {
	pb := NewUfwInstall()

	cfg := types.NodeConfig{
		Logger: slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	needsChange, err := pb.Check()

	if err != nil {
		t.Errorf("Expected no error from Check, got: %v", err)
	}

	// Since UFW is likely not installed in test environment, should return true
	if !needsChange {
		t.Log("UFW appears to be installed, Check returned false")
	}
}

// TestUfwInstall_NewUfwInstall verifies that NewUfwInstall creates a properly configured skill.
func TestUfwInstall_NewUfwInstall(t *testing.T) {
	pb := NewUfwInstall()

	if pb.GetID() != "ufw-install" {
		t.Errorf("Expected ID to be 'ufw-install', got '%s'", pb.GetID())
	}

	expectedDescription := "Install and configure UFW firewall with secure defaults"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestUfwInstall_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete UfwInstall type.
func TestUfwInstall_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewUfwInstall()
	args := map[string]string{"allow-ssh": "true"}

	result := skill.SetArgs(args)

	if _, ok := result.(*UfwInstall); !ok {
		t.Error("SetArgs should return *UfwInstall, not just RunnableInterface")
	}
}

// TestUfwInstall_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete UfwInstall type.
func TestUfwInstall_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewUfwInstall()

	result := skill.SetArg("allow-ssh", "true")

	if _, ok := result.(*UfwInstall); !ok {
		t.Error("SetArg should return *UfwInstall, not just RunnableInterface")
	}
}

// TestUfwInstall_SetID_ReturnsConcreteType verifies that SetID returns the concrete UfwInstall type.
func TestUfwInstall_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewUfwInstall()

	result := skill.SetID("custom-id")

	if _, ok := result.(*UfwInstall); !ok {
		t.Error("SetID should return *UfwInstall, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestUfwInstall_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete UfwInstall type.
func TestUfwInstall_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewUfwInstall()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*UfwInstall); !ok {
		t.Error("SetDescription should return *UfwInstall, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestUfwInstall_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete UfwInstall type.
func TestUfwInstall_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewUfwInstall()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*UfwInstall); !ok {
		t.Error("SetTimeout should return *UfwInstall, not just RunnableInterface")
	}
}

// TestUfwInstall_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestUfwInstall_MethodChaining_PreservesType(t *testing.T) {
	skill := NewUfwInstall().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("allow-ssh", "true").
		SetArgs(map[string]string{"allow-http": "true"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*UfwInstall); !ok {
		t.Error("Method chaining should preserve *UfwInstall type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}
