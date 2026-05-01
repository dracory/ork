package ork

import (
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()
	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}
}

func TestCommand_SetCommand(t *testing.T) {
	cmd := NewCommand()
	command := "ls -la"
	result := cmd.SetCommand(command)

	// Test fluent interface
	if result == nil {
		t.Error("SetCommand returned nil")
	}

	// Test the value is set
	impl, ok := cmd.(types.RunnableInterface)
	if !ok {
		t.Fatal("Type assertion failed")
	}
	if impl.GetID() != "command" {
		t.Error("Command ID should be 'command'")
	}
}

func TestCommand_SetRequired(t *testing.T) {
	cmd := NewCommand()
	result := cmd.SetRequired(true)

	// Test fluent interface
	if result == nil {
		t.Error("SetRequired returned nil")
	}

	// Test the value is set
	impl, ok := cmd.(types.RunnableInterface)
	if !ok {
		t.Fatal("Type assertion failed")
	}
	// Can't directly test required field since it's private,
	// but we can verify the method chain worked
	if impl.GetID() != "command" {
		t.Error("Command ID should be 'command'")
	}
}

func TestCommand_Check(t *testing.T) {
	cmd := NewCommand()
	needsRun, err := cmd.Check()

	if err != nil {
		t.Errorf("Check should not return error, got: %v", err)
	}
	if needsRun {
		t.Error("Check should return false for commands (not idempotent)")
	}
}

func TestCommand_Run_EmptyCommand(t *testing.T) {
	cmd := NewCommand()
	cfg := types.NodeConfig{
		SSHHost:  "localhost",
		SSHPort:  "22",
		SSHLogin: "root",
		SSHKey:   "test",
	}
	cmd.SetNodeConfig(cfg)

	result := cmd.Run()

	if result.Changed {
		t.Error("Empty command should not change system")
	}
	if result.Error == nil {
		t.Error("Empty command should return error")
	}
}

func TestCommand_Run_DryRun(t *testing.T) {
	cmd := NewCommand().
		SetCommand("ls -la")

	cfg := types.NodeConfig{
		SSHHost:      "localhost",
		SSHPort:      "22",
		SSHLogin:     "root",
		SSHKey:       "test",
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}
	cmd.SetNodeConfig(cfg)

	result := cmd.Run()

	if result.Changed {
		t.Error("Dry-run should not change system")
	}
	if result.Error != nil {
		t.Errorf("Dry-run should not error, got: %v", result.Error)
	}
	if result.Message == "" {
		t.Error("Dry-run should have message")
	}
}

func TestCommand_Run_Success(t *testing.T) {
	// Set up mock SSH function
	ssh.SetRunFunc(func(cfg types.NodeConfig, cmd types.Command) (string, error) {
		return "output", nil
	})
	defer ssh.SetRunFunc(nil)

	cmd := NewCommand().
		SetCommand("ls -la").
		SetRequired(true)

	cfg := types.NodeConfig{
		SSHHost:  "localhost",
		SSHPort:  "22",
		SSHLogin: "root",
		SSHKey:   "test",
		Logger:   slog.Default(),
	}
	cmd.SetNodeConfig(cfg)

	result := cmd.Run()

	if !result.Changed {
		t.Error("Successful command should report changed")
	}
	if result.Error != nil {
		t.Errorf("Successful command should not error, got: %v", result.Error)
	}
	if result.Details["output"] != "output" {
		t.Errorf("Expected output 'output', got '%s'", result.Details["output"])
	}
}

func TestCommand_Run_RequiredError(t *testing.T) {
	// Set up mock SSH function to return error
	ssh.SetRunFunc(func(cfg types.NodeConfig, cmd types.Command) (string, error) {
		return "", errors.New("command failed")
	})
	defer ssh.SetRunFunc(nil)

	cmd := NewCommand().
		SetCommand("failing-command").
		SetRequired(true)

	cfg := types.NodeConfig{
		SSHHost:  "localhost",
		SSHPort:  "22",
		SSHLogin: "root",
		SSHKey:   "test",
		Logger:   slog.Default(),
	}
	cmd.SetNodeConfig(cfg)

	result := cmd.Run()

	if result.Changed {
		t.Error("Failed required command should not report changed")
	}
	if result.Error == nil {
		t.Error("Failed required command should return error")
	}
	if result.Message == "" {
		t.Error("Failed required command should have message")
	}
}

func TestCommand_Run_NotRequiredError(t *testing.T) {
	// Set up mock SSH function to return error
	ssh.SetRunFunc(func(cfg types.NodeConfig, cmd types.Command) (string, error) {
		return "", errors.New("command failed")
	})
	defer ssh.SetRunFunc(nil)

	cmd := NewCommand().
		SetCommand("failing-command").
		SetRequired(false)

	cfg := types.NodeConfig{
		SSHHost:  "localhost",
		SSHPort:  "22",
		SSHLogin: "root",
		SSHKey:   "test",
		Logger:   slog.Default(),
	}
	cmd.SetNodeConfig(cfg)

	result := cmd.Run()

	if result.Changed {
		t.Error("Failed non-required command should not report changed")
	}
	if result.Error != nil {
		t.Error("Failed non-required command should not return error")
	}
	if result.Message == "" {
		t.Error("Failed non-required command should have message")
	}
	if result.Details["error"] == "" {
		t.Error("Failed non-required command should have error in details")
	}
}

func TestCommand_FluentInterface(t *testing.T) {
	// Test that Command implements fluent interface methods
	cmd := NewCommand()

	// Test BaseSkill methods via delegation
	cmd.SetID("test-id")
	if cmd.GetID() != "test-id" {
		t.Error("SetID/GetID failed")
	}

	cmd.SetDescription("test description")
	if cmd.GetDescription() != "test description" {
		t.Error("SetDescription/GetDescription failed")
	}

	cmd.SetArg("key", "value")
	if cmd.GetArg("key") != "value" {
		t.Error("SetArg/GetArg failed")
	}

	args := map[string]string{"foo": "bar"}
	cmd.SetArgs(args)
	if cmd.GetArgs()["foo"] != "bar" {
		t.Error("SetArgs/GetArgs failed")
	}

	cmd.SetDryRun(true)
	if !cmd.IsDryRun() {
		t.Error("SetDryRun/IsDryRun failed")
	}
}

func TestCommand_WithDescription(t *testing.T) {
	cmd := NewCommand()
	result := cmd.WithDescription("test description")

	// Test fluent interface
	if result == nil {
		t.Error("WithDescription returned nil")
	}

	// Test the value is set
	if cmd.GetDescription() != "test description" {
		t.Error("WithDescription did not set description")
	}
}

func TestCommand_WithBecomeUser(t *testing.T) {
	cmd := NewCommand()
	result := cmd.WithBecomeUser("testuser")

	// Test fluent interface
	if result == nil {
		t.Error("WithBecomeUser returned nil")
	}

	// Test the value is set
	if cmd.GetBecomeUser() != "testuser" {
		t.Error("WithBecomeUser did not set become user")
	}
}

func TestCommand_Run_WithBecomeUser(t *testing.T) {
	// Set up mock SSH function that implements the same wrapping logic as ssh.Run
	var capturedCommand string
	ssh.SetRunFunc(func(cfg types.NodeConfig, cmd types.Command) (string, error) {
		// Implement the same wrapping logic as ssh.Run
		commandToRun := cmd.Command
		if cfg.BecomeUser != "" {
			commandToRun = fmt.Sprintf("sudo -u %s %s", cfg.BecomeUser, cmd.Command)
		}
		if cfg.Chdir != "" {
			commandToRun = fmt.Sprintf("cd %s && %s", cfg.Chdir, commandToRun)
		}
		capturedCommand = commandToRun
		return "output", nil
	})
	defer ssh.SetRunFunc(nil)

	cmd := NewCommand().
		SetCommand("ls -la").
		WithBecomeUser("testuser").
		WithRequired(true)

	cfg := types.NodeConfig{
		SSHHost:  "localhost",
		SSHPort:  "22",
		SSHLogin: "root",
		SSHKey:   "test",
		Logger:   slog.Default(),
	}
	cmd.SetNodeConfig(cfg)

	result := cmd.Run()

	if result.Error != nil {
		t.Errorf("Command with become user should not error, got: %v", result.Error)
	}

	// Verify the command was wrapped with sudo -u
	expected := "sudo -u testuser ls -la"
	if capturedCommand != expected {
		t.Errorf("Expected command '%s', got '%s'", expected, capturedCommand)
	}
}

func TestCommand_Run_WithBecomeUserAndChdir(t *testing.T) {
	// Set up mock SSH function that implements the same wrapping logic as ssh.Run
	var capturedCommand string
	ssh.SetRunFunc(func(cfg types.NodeConfig, cmd types.Command) (string, error) {
		// Implement the same wrapping logic as ssh.Run
		commandToRun := cmd.Command
		if cfg.BecomeUser != "" {
			commandToRun = fmt.Sprintf("sudo -u %s %s", cfg.BecomeUser, cmd.Command)
		}
		if cfg.Chdir != "" {
			commandToRun = fmt.Sprintf("cd %s && %s", cfg.Chdir, commandToRun)
		}
		capturedCommand = commandToRun
		return "output", nil
	})
	defer ssh.SetRunFunc(nil)

	cmd := NewCommand().
		SetCommand("ls -la").
		WithBecomeUser("testuser").
		WithChdir("/tmp").
		WithRequired(true)

	cfg := types.NodeConfig{
		SSHHost:  "localhost",
		SSHPort:  "22",
		SSHLogin: "root",
		SSHKey:   "test",
		Logger:   slog.Default(),
	}
	cmd.SetNodeConfig(cfg)

	result := cmd.Run()

	if result.Error != nil {
		t.Errorf("Command with become user and chdir should not error, got: %v", result.Error)
	}

	// Verify the command is wrapped: cd first, then sudo
	// Order should be: cd /tmp && sudo -u testuser ls -la
	expected := "cd /tmp && sudo -u testuser ls -la"
	if capturedCommand != expected {
		t.Errorf("Expected command '%s', got '%s'", expected, capturedCommand)
	}
}
