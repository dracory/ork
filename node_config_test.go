package ork

import (
	"log/slog"
	"testing"
)

func TestNewNodeConfig(t *testing.T) {
	cfg := NewNodeConfig()

	if cfg == nil {
		t.Fatal("NewNodeConfig returned nil")
	}

	// Verify default values
	if cfg.SSHHost != "" {
		t.Errorf("Expected empty SSHHost, got %s", cfg.SSHHost)
	}

	if cfg.SSHPort != "22" {
		t.Errorf("Expected SSHPort to be '22', got %s", cfg.SSHPort)
	}

	if cfg.IsDryRunMode {
		t.Error("Expected IsDryRunMode to be false by default")
	}

	if cfg.Args == nil {
		t.Error("Expected Args to be initialized")
	}
}

func TestNewNodeConfig_WithHost(t *testing.T) {
	cfg := NewNodeConfig().
		WithHost("example.com")

	if cfg.SSHHost != "example.com" {
		t.Errorf("Expected SSHHost to be 'example.com', got %s", cfg.SSHHost)
	}
}

func TestNewNodeConfig_WithPort(t *testing.T) {
	cfg := NewNodeConfig().
		WithPort("2222")

	if cfg.SSHPort != "2222" {
		t.Errorf("Expected SSHPort to be '2222', got %s", cfg.SSHPort)
	}
}

func TestNewNodeConfig_WithLogin(t *testing.T) {
	cfg := NewNodeConfig().
		WithLogin("ubuntu")

	if cfg.SSHLogin != "ubuntu" {
		t.Errorf("Expected SSHLogin to be 'ubuntu', got %s", cfg.SSHLogin)
	}
}

func TestNewNodeConfig_WithKey(t *testing.T) {
	cfg := NewNodeConfig().
		WithKey("/home/user/.ssh/id_rsa")

	if cfg.SSHKey != "/home/user/.ssh/id_rsa" {
		t.Errorf("Expected SSHKey to be '/home/user/.ssh/id_rsa', got %s", cfg.SSHKey)
	}
}

func TestNewNodeConfig_WithDryRun(t *testing.T) {
	cfg := NewNodeConfig().
		WithDryRun(true)

	if !cfg.IsDryRunMode {
		t.Error("Expected IsDryRunMode to be true")
	}
}

func TestNewNodeConfig_WithArg(t *testing.T) {
	cfg := NewNodeConfig().
		WithArg("key", "value")

	if cfg.GetArg("key") != "value" {
		t.Errorf("Expected arg 'key' to be 'value', got %s", cfg.GetArg("key"))
	}
}

func TestNewNodeConfig_WithArgs(t *testing.T) {
	args := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	cfg := NewNodeConfig().
		WithArgs(args)

	if cfg.GetArg("key1") != "value1" {
		t.Errorf("Expected arg 'key1' to be 'value1', got %s", cfg.GetArg("key1"))
	}

	if cfg.GetArg("key2") != "value2" {
		t.Errorf("Expected arg 'key2' to be 'value2', got %s", cfg.GetArg("key2"))
	}
}

func TestNewNodeConfig_WithLogger(t *testing.T) {
	logger := slog.Default()
	cfg := NewNodeConfig().
		WithLogger(logger)

	if cfg.Logger != logger {
		t.Error("Expected Logger to be set")
	}
}

func TestNewNodeConfig_WithBecomeUser(t *testing.T) {
	cfg := NewNodeConfig().
		WithBecomeUser("testuser")

	if cfg.BecomeUser != "testuser" {
		t.Errorf("Expected BecomeUser to be 'testuser', got %s", cfg.BecomeUser)
	}
}

func TestNewNodeConfig_WithChdir(t *testing.T) {
	cfg := NewNodeConfig().
		WithChdir("/var/www")

	if cfg.Chdir != "/var/www" {
		t.Errorf("Expected Chdir to be '/var/www', got %s", cfg.Chdir)
	}
}

func TestNewNodeConfig_FluentChaining(t *testing.T) {
	args := map[string]string{"key": "value"}
	logger := slog.Default()

	cfg := NewNodeConfig().
		WithHost("example.com").
		WithPort("2222").
		WithLogin("ubuntu").
		WithKey("/home/user/.ssh/id_rsa").
		WithDryRun(true).
		WithArgs(args).
		WithLogger(logger).
		WithBecomeUser("testuser").
		WithChdir("/var/www")

	// Verify all values were set correctly
	if cfg.SSHHost != "example.com" {
		t.Errorf("Expected SSHHost to be 'example.com', got %s", cfg.SSHHost)
	}

	if cfg.SSHPort != "2222" {
		t.Errorf("Expected SSHPort to be '2222', got %s", cfg.SSHPort)
	}

	if cfg.SSHLogin != "ubuntu" {
		t.Errorf("Expected SSHLogin to be 'ubuntu', got %s", cfg.SSHLogin)
	}

	if cfg.SSHKey != "/home/user/.ssh/id_rsa" {
		t.Errorf("Expected SSHKey to be '/home/user/.ssh/id_rsa', got %s", cfg.SSHKey)
	}

	if !cfg.IsDryRunMode {
		t.Error("Expected IsDryRunMode to be true")
	}

	if cfg.GetArg("key") != "value" {
		t.Errorf("Expected arg 'key' to be 'value', got %s", cfg.GetArg("key"))
	}

	if cfg.Logger != logger {
		t.Error("Expected Logger to be set")
	}

	if cfg.BecomeUser != "testuser" {
		t.Errorf("Expected BecomeUser to be 'testuser', got %s", cfg.BecomeUser)
	}

	if cfg.Chdir != "/var/www" {
		t.Errorf("Expected Chdir to be '/var/www', got %s", cfg.Chdir)
	}
}
