package ork_test

// Integration tests for BecomeUser (sudo privilege escalation).
//
// These tests verify that SetBecomeUser("root") causes commands to be
// wrapped with sudo and executed as the target user. The container is
// configured with NOPASSWD sudo for testuser.

import (
	"strings"
	"testing"

	"github.com/dracory/ork"
)

// TestIntegration_Node_BecomeUser_Nopasswd verifies that setting BecomeUser
// causes the command to run as the target user via sudo. With NOPASSWD
// configured, sudo -H -n -u root <cmd> should succeed.
func TestIntegration_Node_BecomeUser_Nopasswd(t *testing.T) {
	container := setupSSHContainerWithSudo(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)
	node.SetBecomeUser("root")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// whoami should return "root" when running as root via sudo
	results := node.RunCommand("whoami")
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("RunCommand with BecomeUser failed: %v", result.Error)
	}

	output := strings.TrimSpace(result.Message)
	if output != "root" {
		t.Errorf("Expected 'root' from whoami with BecomeUser=root, got: %q", output)
	}
}

// TestIntegration_Node_BecomeUser_GetBecomeUser verifies the getter returns
// what was set.
func TestIntegration_Node_BecomeUser_GetBecomeUser(t *testing.T) {
	container := setupSSHContainerWithSudo(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	if node.GetBecomeUser() != "" {
		t.Error("Expected GetBecomeUser() to return empty string by default")
	}

	node.SetBecomeUser("root")
	if node.GetBecomeUser() != "root" {
		t.Errorf("Expected GetBecomeUser() to return 'root', got: %q", node.GetBecomeUser())
	}
}

// TestIntegration_Node_BecomeUser_Idempotent verifies that BecomeUser works
// for multiple commands on the same persistent connection.
func TestIntegration_Node_BecomeUser_Idempotent(t *testing.T) {
	container := setupSSHContainerWithSudo(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)
	node.SetBecomeUser("root")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Run multiple commands as root.
	// Note: we avoid echo $HOME because $HOME expands in the current shell
	// before sudo switches users. Use sh -c to evaluate in the sudo context.
	commands := []struct {
		cmd      string
		expected string
	}{
		{"whoami", "root"},
		{"id -u", "0"},
		{"sh -c 'echo $HOME'", "/root"},
	}

	for i, tc := range commands {
		results := node.RunCommand(tc.cmd)
		result := results.Results[container.host]
		if result.Error != nil {
			t.Errorf("Command %d (%s) failed: %v", i+1, tc.cmd, result.Error)
			continue
		}
		output := strings.TrimSpace(result.Message)
		if output != tc.expected {
			t.Errorf("Command %d (%s): expected %q, got %q", i+1, tc.cmd, tc.expected, output)
		}
	}
}

// TestIntegration_Node_BecomeUser_WithoutNopasswd verifies that BecomeUser
// with sudo -n (no password) fails when NOPASSWD is NOT configured. The
// standard container has sudo installed but testuser is not in sudoers,
// so sudo -n should return an error.
func TestIntegration_Node_BecomeUser_WithoutNopasswd(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)
	node.SetBecomeUser("root")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// sudo -n should fail because testuser is not in sudoers (no NOPASSWD configured)
	results := node.RunCommand("whoami")
	result := results.Results[container.host]
	if result.Error == nil {
		t.Error("Expected error when BecomeUser is set but testuser is not in sudoers, got nil")
	}
}
