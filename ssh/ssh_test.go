package ssh

import (
	"errors"
	"testing"
)

// TestClient_Connect_EmptyHost verifies that Connect returns an error when host is empty.
func TestClient_Connect_EmptyHost(t *testing.T) {
	client := NewClient("", "22", "root", "id_rsa")

	err := client.Connect()
	if err == nil {
		t.Fatal("Expected Connect to return error for empty host, got nil")
	}

	if err.Error() != "host cannot be empty" {
		t.Errorf("Expected error message 'host cannot be empty', got: %v", err)
	}
}

// TestClient_Connect_ValidHost verifies that Connect proceeds with valid host.
// Note: This test will fail to actually connect since there's no real SSH server,
// but it verifies the empty host check doesn't trigger for valid hosts.
func TestClient_Connect_ValidHost(t *testing.T) {
	client := NewClient("localhost", "22", "root", "id_rsa")

	// This will fail to connect (no real server), but should NOT fail with "host cannot be empty"
	err := client.Connect()
	if err != nil {
		// Expected to fail (no SSH server running), but should NOT be "host cannot be empty"
		if err.Error() == "host cannot be empty" {
			t.Error("Connect should not fail with 'host cannot be empty' when host is provided")
		}
	}
}

// TestNewClient_DefaultPort verifies that NewClient defaults port to "22" when empty.
func TestNewClient_DefaultPort(t *testing.T) {
	client := NewClient("localhost", "", "root", "id_rsa")

	if client.port != "22" {
		t.Errorf("Expected port to default to '22', got %q", client.port)
	}
}

// TestNewClient_CustomPort verifies that NewClient uses provided port.
func TestNewClient_CustomPort(t *testing.T) {
	client := NewClient("localhost", "2222", "root", "id_rsa")

	if client.port != "2222" {
		t.Errorf("Expected port to be '2222', got %q", client.port)
	}
}

// TestNewClient_StoresValues verifies that NewClient stores all provided values.
func TestNewClient_StoresValues(t *testing.T) {
	client := NewClient("server.example.com", "2222", "deploy", "production.prv")

	if client.host != "server.example.com" {
		t.Errorf("Expected host to be 'server.example.com', got %q", client.host)
	}

	if client.user != "deploy" {
		t.Errorf("Expected user to be 'deploy', got %q", client.user)
	}

	// keyPath should be resolved to full path
	if client.keyPath == "" {
		t.Error("Expected keyPath to be non-empty")
	}

	if client.keyPath == "production.prv" {
		t.Error("Expected keyPath to be resolved to full path, not just filename")
	}
}

// TestClassifySSHError_HostKeyUnknown verifies detection of unknown host key errors.
func TestClassifySSHError_HostKeyUnknown(t *testing.T) {
	err := errors.New("ssh: handshake failed: knownhosts: key is unknown")
	classified := classifySSHError(err)

	if classified == nil {
		t.Fatal("Expected classified error, got nil")
	}

	errStr := classified.Error()
	if !contains(errStr, "host key verification failed") {
		t.Errorf("Expected error to contain 'host key verification failed', got: %v", errStr)
	}
	if !contains(errStr, "known_hosts") {
		t.Errorf("Expected error to contain 'known_hosts', got: %v", errStr)
	}
}

// TestClassifySSHError_HostKeyMismatch verifies detection of host key mismatch errors.
func TestClassifySSHError_HostKeyMismatch(t *testing.T) {
	err := errors.New("ssh: handshake failed: knownhosts: key mismatch")
	classified := classifySSHError(err)

	if classified == nil {
		t.Fatal("Expected classified error, got nil")
	}

	errStr := classified.Error()
	if !contains(errStr, "host key verification failed") {
		t.Errorf("Expected error to contain 'host key verification failed', got: %v", errStr)
	}
	if !contains(errStr, "man-in-the-middle") {
		t.Errorf("Expected error to contain 'man-in-the-middle', got: %v", errStr)
	}
}

// TestClassifySSHError_HostKeyRevoked verifies detection of revoked host key errors.
func TestClassifySSHError_HostKeyRevoked(t *testing.T) {
	err := errors.New("ssh: handshake failed: knownhosts: key is revoked")
	classified := classifySSHError(err)

	if classified == nil {
		t.Fatal("Expected classified error, got nil")
	}

	errStr := classified.Error()
	if !contains(errStr, "host key verification failed") {
		t.Errorf("Expected error to contain 'host key verification failed', got: %v", errStr)
	}
	if !contains(errStr, "revoked") {
		t.Errorf("Expected error to contain 'revoked', got: %v", errStr)
	}
}

// TestClassifySSHError_AuthenticationFailed verifies detection of authentication failures.
func TestClassifySSHError_AuthenticationFailed(t *testing.T) {
	err := errors.New("ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain")
	classified := classifySSHError(err)

	if classified == nil {
		t.Fatal("Expected classified error, got nil")
	}

	errStr := classified.Error()
	if !contains(errStr, "authentication failed") {
		t.Errorf("Expected error to contain 'authentication failed', got: %v", errStr)
	}
	if !contains(errStr, "SSH key") {
		t.Errorf("Expected error to contain 'SSH key', got: %v", errStr)
	}
}

// TestClassifySSHError_ConnectionRefused verifies detection of connection refused errors.
func TestClassifySSHError_ConnectionRefused(t *testing.T) {
	err := errors.New("dial tcp 127.0.0.1:22: connect: connection refused")
	classified := classifySSHError(err)

	if classified == nil {
		t.Fatal("Expected classified error, got nil")
	}

	errStr := classified.Error()
	if !contains(errStr, "connection refused") {
		t.Errorf("Expected error to contain 'connection refused', got: %v", errStr)
	}
}

// TestClassifySSHError_Timeout verifies detection of timeout errors.
func TestClassifySSHError_Timeout(t *testing.T) {
	err := errors.New("dial tcp 127.0.0.1:22: i/o timeout")
	classified := classifySSHError(err)

	if classified == nil {
		t.Fatal("Expected classified error, got nil")
	}

	errStr := classified.Error()
	if !contains(errStr, "connection timeout") {
		t.Errorf("Expected error to contain 'connection timeout', got: %v", errStr)
	}
}

// TestClassifySSHError_NetworkUnreachable verifies detection of network errors.
func TestClassifySSHError_NetworkUnreachable(t *testing.T) {
	err := errors.New("dial tcp 127.0.0.1:22: network is unreachable")
	classified := classifySSHError(err)

	if classified == nil {
		t.Fatal("Expected classified error, got nil")
	}

	errStr := classified.Error()
	if !contains(errStr, "network error") {
		t.Errorf("Expected error to contain 'network error', got: %v", errStr)
	}
}

// TestClassifySSHError_UnknownError verifies that unknown errors are returned as-is.
func TestClassifySSHError_UnknownError(t *testing.T) {
	err := errors.New("some unknown error")
	classified := classifySSHError(err)

	if classified == nil {
		t.Fatal("Expected classified error, got nil")
	}

	if classified.Error() != "some unknown error" {
		t.Errorf("Expected error to be unchanged, got: %v", classified.Error())
	}
}

// TestClassifySSHError_Nil verifies that nil error is handled.
func TestClassifySSHError_Nil(t *testing.T) {
	classified := classifySSHError(nil)

	if classified != nil {
		t.Errorf("Expected nil error to remain nil, got: %v", classified)
	}
}

// contains is a test helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

// findSubstring is a test helper function for substring search.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
