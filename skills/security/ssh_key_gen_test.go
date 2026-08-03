package security

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestSshKeyGen_Run_DryRun verifies that dry-run mode correctly handles SSH key generation.
func TestSshKeyGen_Run_DryRun(t *testing.T) {
	pb := NewSshKeyGen()
	pb.SetArg(ArgUsername, "deploy")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would generate SSH keypair for deploy"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestSshKeyGen_Run_DryRun_NoUsername verifies dry-run without username returns error.
func TestSshKeyGen_Run_DryRun_NoUsername(t *testing.T) {
	pb := NewSshKeyGen()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// Should return error for missing username even in dry-run
	if result.Error == nil {
		t.Error("Expected error for missing username")
	}

	if result.Message != "Username is required" {
		t.Errorf("Expected message 'Username is required', got '%s'", result.Message)
	}
}

// TestSshKeyGen_Run_NotDryRun verifies that non-dry-run mode returns different result structure.
func TestSshKeyGen_Run_NotDryRun(t *testing.T) {
	pb := NewSshKeyGen()
	pb.SetArg(ArgUsername, "deploy")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the dry-run message.
	if result.Message == "Would generate SSH keypair for deploy" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestSshKeyGen_NewSshKeyGen verifies that NewSshKeyGen creates a properly configured skill.
func TestSshKeyGen_NewSshKeyGen(t *testing.T) {
	pb := NewSshKeyGen()

	if pb.GetID() != "ssh-key-gen" {
		t.Errorf("Expected ID to be 'ssh-key-gen', got '%s'", pb.GetID())
	}

	if pb.GetDescription() == "" {
		t.Error("Expected non-empty description")
	}
}

// TestSshKeyGen_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete SshKeyGen type.
func TestSshKeyGen_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewSshKeyGen()
	args := map[string]string{ArgUsername: "deploy"}

	result := skill.SetArgs(args)

	if _, ok := result.(*SshKeyGen); !ok {
		t.Error("SetArgs should return *SshKeyGen, not just RunnableInterface")
	}
}

// TestSshKeyGen_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete SshKeyGen type.
func TestSshKeyGen_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewSshKeyGen()

	result := skill.SetArg(ArgUsername, "deploy")

	if _, ok := result.(*SshKeyGen); !ok {
		t.Error("SetArg should return *SshKeyGen, not just RunnableInterface")
	}
}

// TestSshKeyGen_SetID_ReturnsConcreteType verifies that SetID returns the concrete SshKeyGen type.
func TestSshKeyGen_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewSshKeyGen()

	result := skill.SetID("custom-id")

	if _, ok := result.(*SshKeyGen); !ok {
		t.Error("SetID should return *SshKeyGen, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestSshKeyGen_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete SshKeyGen type.
func TestSshKeyGen_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewSshKeyGen()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*SshKeyGen); !ok {
		t.Error("SetDescription should return *SshKeyGen, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestSshKeyGen_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete SshKeyGen type.
func TestSshKeyGen_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewSshKeyGen()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*SshKeyGen); !ok {
		t.Error("SetTimeout should return *SshKeyGen, not just RunnableInterface")
	}
}

// TestSshKeyGen_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestSshKeyGen_MethodChaining_PreservesType(t *testing.T) {
	skill := NewSshKeyGen().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgUsername, "deploy").
		SetArgs(map[string]string{ArgKeyType: "rsa"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*SshKeyGen); !ok {
		t.Error("Method chaining should preserve *SshKeyGen type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestSshKeyGen_SetUsername verifies that SetUsername sets the username arg.
func TestSshKeyGen_SetUsername(t *testing.T) {
	skill := NewSshKeyGen().SetUsername("deploy")

	if skill.GetArg(ArgUsername) != "deploy" {
		t.Errorf("Expected username 'deploy', got '%s'", skill.GetArg(ArgUsername))
	}
}

// TestSshKeyGen_SetKeyType verifies that SetKeyType sets the key-type arg.
func TestSshKeyGen_SetKeyType(t *testing.T) {
	skill := NewSshKeyGen().SetKeyType("rsa")

	if skill.GetArg(ArgKeyType) != "rsa" {
		t.Errorf("Expected key-type 'rsa', got '%s'", skill.GetArg(ArgKeyType))
	}
}

// TestSshKeyGen_SetComment verifies that SetComment sets the comment arg.
func TestSshKeyGen_SetComment(t *testing.T) {
	skill := NewSshKeyGen().SetComment("deploy@my-server")

	if skill.GetArg(ArgComment) != "deploy@my-server" {
		t.Errorf("Expected comment 'deploy@my-server', got '%s'", skill.GetArg(ArgComment))
	}
}

// TestSshKeyGen_SetKeyPath verifies that SetKeyPath sets the key-path arg.
func TestSshKeyGen_SetKeyPath(t *testing.T) {
	skill := NewSshKeyGen().SetKeyPath("/home/deploy/.ssh/id_ed25519")

	if skill.GetArg(ArgKeyPath) != "/home/deploy/.ssh/id_ed25519" {
		t.Errorf("Expected key-path '/home/deploy/.ssh/id_ed25519', got '%s'", skill.GetArg(ArgKeyPath))
	}
}

// TestSshKeyGen_SetUsername_Chaining verifies that SetUsername chains with other setters.
func TestSshKeyGen_SetUsername_Chaining(t *testing.T) {
	skill := NewSshKeyGen().
		SetUsername("deploy").
		SetKeyType("ed25519").
		SetComment("deploy@host").
		SetID("custom-id").
		SetDescription("custom description")

	if skill.GetArg(ArgUsername) != "deploy" {
		t.Errorf("Expected username 'deploy', got '%s'", skill.GetArg(ArgUsername))
	}
	if skill.GetArg(ArgKeyType) != "ed25519" {
		t.Errorf("Expected key-type 'ed25519', got '%s'", skill.GetArg(ArgKeyType))
	}
	if skill.GetArg(ArgComment) != "deploy@host" {
		t.Errorf("Expected comment 'deploy@host', got '%s'", skill.GetArg(ArgComment))
	}
	if skill.GetID() != "custom-id" {
		t.Error("Chaining should set ID")
	}
	if skill.GetDescription() != "custom description" {
		t.Error("Chaining should set description")
	}
}
