package ork

import (
	"testing"
)

// TestNewNode verifies that NewNode initializes with correct default values.
func TestNewNode(t *testing.T) {
	host := "server.example.com"
	node := NewNode(host)

	// Verify the node is not nil
	if node == nil {
		t.Fatal("Expected NewNode to return non-nil NodeInterface")
	}

	// Verify default values using getter methods
	if node.GetHost() != host {
		t.Errorf("Expected GetHost()=%q, got %q", host, node.GetHost())
	}

	if node.GetPort() != "22" {
		t.Errorf("Expected GetPort()=%q, got %q", "22", node.GetPort())
	}

	if node.GetUser() != "root" {
		t.Errorf("Expected GetUser()=%q, got %q", "root", node.GetUser())
	}

	if node.GetKey() != "id_rsa" {
		t.Errorf("Expected GetKey()=%q, got %q", "id_rsa", node.GetKey())
	}

	// Verify Args is initialized
	cfg := node.GetConfig()
	if cfg.Args == nil {
		t.Error("Expected Args to be initialized, got nil")
	}

	if len(cfg.Args) != 0 {
		t.Errorf("Expected Args to be empty, got %d items", len(cfg.Args))
	}

	// Verify not connected initially
	if node.IsConnected() {
		t.Error("Expected IsConnected() to return false initially")
	}
}
