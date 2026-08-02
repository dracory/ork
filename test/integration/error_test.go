package ork_test

// Integration tests for error handling and IsExitError.
//
// These tests verify that the SSH layer correctly classifies errors into:
//   - Connection failures (wrong port, unreachable host)
//   - Authentication failures (wrong key, wrong user)
//   - Command exit errors (non-zero exit code)
//
// IsExitError should return true only for command exit errors, not for
// connection or auth failures.

import (
	"strings"
	"testing"

	"github.com/dracory/ork"
	"github.com/dracory/ork/ssh"
)

// TestIntegration_Node_WrongPort verifies that connecting to a wrong port
// fails with a connection error (not an exit error).
func TestIntegration_Node_WrongPort(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort("9999"). // nothing listening
		SetUser(container.user).
		SetKey(container.keyName)

	err := node.Connect()
	if err == nil {
		t.Fatal("Expected Connect to fail with wrong port, got nil")
	}

	if ssh.IsExitError(err) {
		t.Errorf("Connection error should NOT be an exit error, got: %v", err)
	}

	if !strings.Contains(err.Error(), "connect") && !strings.Contains(err.Error(), "refused") {
		t.Logf("Error message: %s", err.Error())
	}
}

// TestIntegration_Node_WrongKey verifies that connecting with a non-existent
// key fails with an auth error (not an exit error).
func TestIntegration_Node_WrongKey(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey("nonexistent-key-that-does-not-exist")

	err := node.Connect()
	if err == nil {
		t.Fatal("Expected Connect to fail with wrong key, got nil")
	}

	if ssh.IsExitError(err) {
		t.Errorf("Auth error should NOT be an exit error, got: %v", err)
	}
}

// TestIntegration_Node_WrongUser verifies that connecting with a wrong user
// fails with an auth error (not an exit error).
func TestIntegration_Node_WrongUser(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser("nonexistent-user").
		SetKey(container.keyName)

	err := node.Connect()
	if err == nil {
		t.Fatal("Expected Connect to fail with wrong user, got nil")
	}

	if ssh.IsExitError(err) {
		t.Errorf("Auth error should NOT be an exit error, got: %v", err)
	}
}

// TestIntegration_Node_CommandNonZeroExit verifies that a command that exits
// non-zero returns an error that IS an exit error.
func TestIntegration_Node_CommandNonZeroExit(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// exit 42 — a non-zero exit that is NOT a connection failure
	results := node.RunCommand("exit 42")
	result := results.Results[container.host]
	if result.Error == nil {
		t.Fatal("Expected RunCommand('exit 42') to return an error, got nil")
	}

	if !ssh.IsExitError(result.Error) {
		t.Errorf("Expected exit error to be detected by IsExitError, got: %v", result.Error)
	}
}

// TestIntegration_IsExitError_ConnectionFailure verifies that a connection
// failure is NOT classified as an exit error.
func TestIntegration_IsExitError_ConnectionFailure(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort("9999").
		SetUser(container.user).
		SetKey(container.keyName)

	// Use one-time RunCommand (no Connect) to trigger connection error
	results := node.RunCommand("echo 'should never run'")
	result := results.Results[container.host]
	if result.Error == nil {
		t.Fatal("Expected connection error, got nil")
	}

	if ssh.IsExitError(result.Error) {
		t.Errorf("Connection failure should NOT be an exit error, got: %v", result.Error)
	}
}

// TestIntegration_IsExitError_AuthFailure verifies that an auth failure
// is NOT classified as an exit error.
func TestIntegration_IsExitError_AuthFailure(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser("wrong-user").
		SetKey(container.keyName)

	// Use one-time RunCommand (no Connect) to trigger auth error
	results := node.RunCommand("echo 'should never run'")
	result := results.Results[container.host]
	if result.Error == nil {
		t.Fatal("Expected auth error, got nil")
	}

	if ssh.IsExitError(result.Error) {
		t.Errorf("Auth failure should NOT be an exit error, got: %v", result.Error)
	}
}

// TestIntegration_IsExitError_CommandExit verifies that a non-zero command
// exit IS classified as an exit error.
func TestIntegration_IsExitError_CommandExit(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	results := node.RunCommand("exit 1")
	result := results.Results[container.host]
	if result.Error == nil {
		t.Fatal("Expected error from 'exit 1', got nil")
	}

	if !ssh.IsExitError(result.Error) {
		t.Errorf("Command exit error should be detected by IsExitError, got: %v", result.Error)
	}
}
