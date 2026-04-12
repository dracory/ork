package ssh

import (
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
