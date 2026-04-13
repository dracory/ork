package ork

import (
	"testing"

	"github.com/dracory/ork/config"
)

// TestNewNodeForHost verifies that NewNodeForHost initializes with correct default values.
func TestNewNodeForHost(t *testing.T) {
	host := "server.example.com"
	node := NewNodeForHost(host)

	// Verify the node is not nil
	if node == nil {
		t.Fatal("Expected NewNodeForHost to return non-nil NodeInterface")
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
	cfg := node.GetNodeConfig()
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

// TestNewNode verifies that NewNode (no arguments) initializes with correct default values.
func TestNewNode(t *testing.T) {
	node := NewNode()

	// Verify the node is not nil
	if node == nil {
		t.Fatal("Expected NewNode to return non-nil NodeInterface")
	}

	// Verify default values - host should be empty
	if node.GetHost() != "" {
		t.Errorf("Expected GetHost()=%q, got %q", "", node.GetHost())
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
	cfg := node.GetNodeConfig()
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

// TestNewNodeFromConfig verifies that NewNodeFromConfig creates a node from existing config.
func TestNewNodeFromConfig(t *testing.T) {
	cfg := config.NodeConfig{
		SSHHost:  "server.example.com",
		SSHPort:  "2222",
		RootUser: "deploy",
		SSHKey:   "production.prv",
		Args: map[string]string{
			"env": "production",
		},
	}

	node := NewNodeFromConfig(cfg)

	// Verify the node is not nil
	if node == nil {
		t.Fatal("Expected NewNodeFromConfig to return non-nil NodeInterface")
	}

	// Verify config values were copied
	if node.GetHost() != "server.example.com" {
		t.Errorf("Expected GetHost()=%q, got %q", "server.example.com", node.GetHost())
	}

	if node.GetPort() != "2222" {
		t.Errorf("Expected GetPort()=%q, got %q", "2222", node.GetPort())
	}

	if node.GetUser() != "deploy" {
		t.Errorf("Expected GetUser()=%q, got %q", "deploy", node.GetUser())
	}

	if node.GetKey() != "production.prv" {
		t.Errorf("Expected GetKey()=%q, got %q", "production.prv", node.GetKey())
	}

	// Verify Args is initialized
	if node.GetArg("env") != "production" {
		t.Errorf("Expected GetArg(env)=%q, got %q", "production", node.GetArg("env"))
	}

	args := node.GetArgs()
	if args["env"] != "production" {
		t.Errorf("Expected GetArgs()[env]=%q, got %q", "production", args["env"])
	}

	// Verify not connected initially
	if node.IsConnected() {
		t.Error("Expected IsConnected() to return false initially")
	}
}

// TestNewNodeFromConfig_DeepCopy verifies that NewNodeFromConfig creates a deep copy of the config.
func TestNewNodeFromConfig_DeepCopy(t *testing.T) {
	cfg := config.NodeConfig{
		SSHHost:  "server.example.com",
		SSHPort:  "22",
		RootUser: "root",
		SSHKey:   "id_rsa",
		Args: map[string]string{
			"key": "value",
		},
	}

	node := NewNodeFromConfig(cfg)

	// Modify original config
	cfg.SSHHost = "modified.example.com"
	cfg.Args["key"] = "modified"
	cfg.Args["newkey"] = "newvalue"

	// Verify node's config is unchanged
	if node.GetHost() != "server.example.com" {
		t.Errorf("Expected GetHost() unchanged, got %q", node.GetHost())
	}

	if node.GetArg("key") != "value" {
		t.Errorf("Expected GetArg(key) unchanged, got %q", node.GetArg("key"))
	}

	args := node.GetArgs()
	if _, exists := args["newkey"]; exists {
		t.Error("Expected GetArgs() not to have 'newkey'")
	}
}

// TestNodeInterface_GetArg verifies that GetArg returns the correct argument value.
func TestNodeInterface_GetArg(t *testing.T) {
	node := NewNodeForHost("server.example.com").
		SetArg("username", "alice").
		SetArg("shell", "/bin/bash")

	// Verify GetArg returns correct values
	if node.GetArg("username") != "alice" {
		t.Errorf("Expected GetArg(username)=%q, got %q", "alice", node.GetArg("username"))
	}

	if node.GetArg("shell") != "/bin/bash" {
		t.Errorf("Expected GetArg(shell)=%q, got %q", "/bin/bash", node.GetArg("shell"))
	}

	// Verify GetArg returns empty string for non-existent key
	if node.GetArg("nonexistent") != "" {
		t.Errorf("Expected GetArg(nonexistent)=%q, got %q", "", node.GetArg("nonexistent"))
	}
}

// TestNodeInterface_GetArgs verifies that GetArgs returns a copy of the arguments map.
func TestNodeInterface_GetArgs(t *testing.T) {
	node := NewNodeForHost("server.example.com").
		SetArg("username", "alice").
		SetArg("shell", "/bin/bash")

	args := node.GetArgs()

	// Verify GetArgs returns correct values
	if args["username"] != "alice" {
		t.Errorf("Expected GetArgs()[username]=%q, got %q", "alice", args["username"])
	}

	if args["shell"] != "/bin/bash" {
		t.Errorf("Expected GetArgs()[shell]=%q, got %q", "/bin/bash", args["shell"])
	}

	// Modify returned args
	args["username"] = "modified"
	args["newkey"] = "newvalue"

	// Verify node's args are unchanged
	if node.GetArg("username") != "alice" {
		t.Errorf("Expected GetArg(username) unchanged after modifying returned map, got %q", node.GetArg("username"))
	}

	if node.GetArg("newkey") != "" {
		t.Errorf("Expected GetArg(newkey) to be empty, got %q", node.GetArg("newkey"))
	}
}
