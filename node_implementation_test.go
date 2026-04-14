package ork

import (
	"fmt"
	"testing"
	"time"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// TestNodeStruct verifies that the nodeImplementation struct has the correct fields.
func TestNodeStruct(t *testing.T) {
	// Create a nodeImplementation directly to test the struct definition
	host := "server.example.com"
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  host,
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
		connected: false,
	}

	// Verify fields are accessible and have correct values
	if n.cfg.SSHHost != host {
		t.Errorf("Expected SSHHost=%q, got %q", host, n.cfg.SSHHost)
	}

	if n.cfg.SSHPort != "22" {
		t.Errorf("Expected SSHPort=%q, got %q", "22", n.cfg.SSHPort)
	}

	if n.cfg.RootUser != "root" {
		t.Errorf("Expected RootUser=%q, got %q", "root", n.cfg.RootUser)
	}

	if n.cfg.SSHKey != "id_rsa" {
		t.Errorf("Expected SSHKey=%q, got %q", "id_rsa", n.cfg.SSHKey)
	}

	if n.cfg.Args == nil {
		t.Error("Expected Args to be initialized, got nil")
	}

	if len(n.cfg.Args) != 0 {
		t.Errorf("Expected Args to be empty, got %d items", len(n.cfg.Args))
	}

	if n.connected {
		t.Error("Expected connected=false, got true")
	}

	if n.sshClient != nil {
		t.Error("Expected sshClient=nil, got non-nil")
	}
}

// TestNodeImplementation_SetPort verifies that SetPort updates the SSH port configuration.
func TestNodeImplementation_SetPort(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
	}

	// Test setting a custom port
	result := n.SetPort("2222")

	// Verify the port was updated
	if n.cfg.SSHPort != "2222" {
		t.Errorf("Expected SSHPort=%q, got %q", "2222", n.cfg.SSHPort)
	}

	// Verify method returns self for chaining
	if resultNode, ok := result.(*nodeImplementation); !ok || resultNode != n {
		t.Error("Expected SetPort to return self for chaining")
	}
}

// TestNodeImplementation_SetUser verifies that SetUser updates the SSH user configuration.
func TestNodeImplementation_SetUser(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
	}

	// Test setting a custom user
	result := n.SetUser("deploy")

	// Verify the user was updated
	if n.cfg.RootUser != "deploy" {
		t.Errorf("Expected RootUser=%q, got %q", "deploy", n.cfg.RootUser)
	}

	// Verify method returns self for chaining
	if resultNode, ok := result.(*nodeImplementation); !ok || resultNode != n {
		t.Error("Expected SetUser to return self for chaining")
	}
}

// TestNodeImplementation_SetKey verifies that SetKey updates the SSH key configuration.
func TestNodeImplementation_SetKey(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
	}

	// Test setting a custom key
	result := n.SetKey("production.prv")

	// Verify the key was updated
	if n.cfg.SSHKey != "production.prv" {
		t.Errorf("Expected SSHKey=%q, got %q", "production.prv", n.cfg.SSHKey)
	}

	// Verify method returns self for chaining
	if resultNode, ok := result.(*nodeImplementation); !ok || resultNode != n {
		t.Error("Expected SetKey to return self for chaining")
	}
}

// TestNodeImplementation_SetArg verifies that SetArg adds individual arguments.
func TestNodeImplementation_SetArg(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
	}

	// Test adding first argument
	result1 := n.SetArg("username", "alice")

	// Verify the argument was added
	if n.cfg.Args["username"] != "alice" {
		t.Errorf("Expected Args[username]=%q, got %q", "alice", n.cfg.Args["username"])
	}

	// Verify method returns self for chaining
	if resultNode, ok := result1.(*nodeImplementation); !ok || resultNode != n {
		t.Error("Expected SetArg to return self for chaining")
	}

	// Test adding second argument
	result2 := n.SetArg("shell", "/bin/bash")

	// Verify both arguments exist
	if n.cfg.Args["username"] != "alice" {
		t.Errorf("Expected Args[username]=%q, got %q", "alice", n.cfg.Args["username"])
	}
	if n.cfg.Args["shell"] != "/bin/bash" {
		t.Errorf("Expected Args[shell]=%q, got %q", "/bin/bash", n.cfg.Args["shell"])
	}

	// Verify method returns self for chaining
	if resultNode, ok := result2.(*nodeImplementation); !ok || resultNode != n {
		t.Error("Expected SetArg to return self for chaining")
	}

	// Test overwriting existing argument
	n.SetArg("username", "bob")
	if n.cfg.Args["username"] != "bob" {
		t.Errorf("Expected Args[username]=%q after overwrite, got %q", "bob", n.cfg.Args["username"])
	}
}

// TestNodeImplementation_SetArg_NilArgs verifies that SetArg initializes Args if nil.
func TestNodeImplementation_SetArg_NilArgs(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     nil, // Args is nil
		},
	}

	// Test adding argument when Args is nil
	n.SetArg("key", "value")

	// Verify Args was initialized and argument was added
	if n.cfg.Args == nil {
		t.Fatal("Expected Args to be initialized, got nil")
	}
	if n.cfg.Args["key"] != "value" {
		t.Errorf("Expected Args[key]=%q, got %q", "value", n.cfg.Args["key"])
	}
}

// TestNodeImplementation_SetArgs verifies that SetArgs replaces the entire arguments map.
func TestNodeImplementation_SetArgs(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     map[string]string{"old": "value"},
		},
	}

	// Test replacing arguments
	newArgs := map[string]string{
		"username": "alice",
		"shell":    "/bin/bash",
	}
	result := n.SetArgs(newArgs)

	// Verify old arguments are gone
	if _, exists := n.cfg.Args["old"]; exists {
		t.Error("Expected old arguments to be replaced")
	}

	// Verify new arguments exist
	if n.cfg.Args["username"] != "alice" {
		t.Errorf("Expected Args[username]=%q, got %q", "alice", n.cfg.Args["username"])
	}
	if n.cfg.Args["shell"] != "/bin/bash" {
		t.Errorf("Expected Args[shell]=%q, got %q", "/bin/bash", n.cfg.Args["shell"])
	}

	// Verify method returns self for chaining
	if resultNode, ok := result.(*nodeImplementation); !ok || resultNode != n {
		t.Error("Expected SetArgs to return self for chaining")
	}
}

// TestNodeImplementation_SetterChaining verifies that all setter methods can be chained.
func TestNodeImplementation_SetterChaining(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
	}

	// Test chaining all setter methods
	result := n.SetPort("2222").
		SetUser("deploy").
		SetKey("production.prv").
		SetArg("username", "alice").
		SetArg("shell", "/bin/bash")

	// Verify all settings were applied
	if n.cfg.SSHPort != "2222" {
		t.Errorf("Expected SSHPort=%q, got %q", "2222", n.cfg.SSHPort)
	}
	if n.cfg.RootUser != "deploy" {
		t.Errorf("Expected RootUser=%q, got %q", "deploy", n.cfg.RootUser)
	}
	if n.cfg.SSHKey != "production.prv" {
		t.Errorf("Expected SSHKey=%q, got %q", "production.prv", n.cfg.SSHKey)
	}
	if n.cfg.Args["username"] != "alice" {
		t.Errorf("Expected Args[username]=%q, got %q", "alice", n.cfg.Args["username"])
	}
	if n.cfg.Args["shell"] != "/bin/bash" {
		t.Errorf("Expected Args[shell]=%q, got %q", "/bin/bash", n.cfg.Args["shell"])
	}

	// Verify final result is the same node
	if resultNode, ok := result.(*nodeImplementation); !ok || resultNode != n {
		t.Error("Expected chained methods to return self")
	}
}

// TestNodeImplementation_GetHost verifies that GetHost returns the configured SSH host.
func TestNodeImplementation_GetHost(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
	}

	host := n.GetHost()
	if host != "server.example.com" {
		t.Errorf("Expected GetHost()=%q, got %q", "server.example.com", host)
	}
}

// TestNodeImplementation_GetPort verifies that GetPort returns the configured SSH port.
func TestNodeImplementation_GetPort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected string
	}{
		{"default port", "22", "22"},
		{"custom port", "2222", "2222"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &nodeImplementation{
				cfg: config.NodeConfig{
					SSHHost:  "server.example.com",
					SSHPort:  tt.port,
					RootUser: "root",
					SSHKey:   "id_rsa",
					Args:     make(map[string]string),
				},
			}

			port := n.GetPort()
			if port != tt.expected {
				t.Errorf("Expected GetPort()=%q, got %q", tt.expected, port)
			}
		})
	}
}

// TestNodeImplementation_GetUser verifies that GetUser returns the configured SSH user.
func TestNodeImplementation_GetUser(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		expected string
	}{
		{"default user", "root", "root"},
		{"custom user", "deploy", "deploy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &nodeImplementation{
				cfg: config.NodeConfig{
					SSHHost:  "server.example.com",
					SSHPort:  "22",
					RootUser: tt.user,
					SSHKey:   "id_rsa",
					Args:     make(map[string]string),
				},
			}

			user := n.GetUser()
			if user != tt.expected {
				t.Errorf("Expected GetUser()=%q, got %q", tt.expected, user)
			}
		})
	}
}

// TestNodeImplementation_GetKey verifies that GetKey returns the configured SSH key.
func TestNodeImplementation_GetKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"default key", "id_rsa", "id_rsa"},
		{"custom key", "production.prv", "production.prv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &nodeImplementation{
				cfg: config.NodeConfig{
					SSHHost:  "server.example.com",
					SSHPort:  "22",
					RootUser: "root",
					SSHKey:   tt.key,
					Args:     make(map[string]string),
				},
			}

			key := n.GetKey()
			if key != tt.expected {
				t.Errorf("Expected GetKey()=%q, got %q", tt.expected, key)
			}
		})
	}
}

// TestNodeImplementation_GetConfig verifies that GetConfig returns a complete copy of the configuration.
func TestNodeImplementation_GetConfig(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "2222",
			RootUser: "deploy",
			SSHKey:   "production.prv",
			Args: map[string]string{
				"username": "alice",
				"shell":    "/bin/bash",
			},
		},
	}

	cfg := n.GetNodeConfig()

	// Verify all fields are copied correctly
	if cfg.SSHHost != "server.example.com" {
		t.Errorf("Expected SSHHost=%q, got %q", "server.example.com", cfg.SSHHost)
	}
	if cfg.SSHPort != "2222" {
		t.Errorf("Expected SSHPort=%q, got %q", "2222", cfg.SSHPort)
	}
	if cfg.RootUser != "deploy" {
		t.Errorf("Expected RootUser=%q, got %q", "deploy", cfg.RootUser)
	}
	if cfg.SSHKey != "production.prv" {
		t.Errorf("Expected SSHKey=%q, got %q", "production.prv", cfg.SSHKey)
	}
	if cfg.Args["username"] != "alice" {
		t.Errorf("Expected Args[username]=%q, got %q", "alice", cfg.Args["username"])
	}
	if cfg.Args["shell"] != "/bin/bash" {
		t.Errorf("Expected Args[shell]=%q, got %q", "/bin/bash", cfg.Args["shell"])
	}
}

// TestNodeImplementation_GetConfig_DeepCopy verifies that GetConfig returns a deep copy.
func TestNodeImplementation_GetConfig_DeepCopy(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args: map[string]string{
				"username": "alice",
			},
		},
	}

	// Get a copy of the config
	cfg := n.GetNodeConfig()

	// Modify the returned config
	cfg.SSHHost = "modified.example.com"
	cfg.SSHPort = "9999"
	cfg.RootUser = "modified"
	cfg.SSHKey = "modified.prv"
	cfg.Args["username"] = "modified"
	cfg.Args["newkey"] = "newvalue"

	// Verify the modifications exist in the returned copy
	if cfg.SSHHost != "modified.example.com" || cfg.SSHPort != "9999" || cfg.RootUser != "modified" || cfg.SSHKey != "modified.prv" {
		t.Error("config copy should reflect modifications")
	}

	// Verify the node's internal config is unchanged
	if n.cfg.SSHHost != "server.example.com" {
		t.Errorf("Expected internal SSHHost unchanged, got %q", n.cfg.SSHHost)
	}
	if n.cfg.SSHPort != "22" {
		t.Errorf("Expected internal SSHPort unchanged, got %q", n.cfg.SSHPort)
	}
	if n.cfg.RootUser != "root" {
		t.Errorf("Expected internal RootUser unchanged, got %q", n.cfg.RootUser)
	}
	if n.cfg.SSHKey != "id_rsa" {
		t.Errorf("Expected internal SSHKey unchanged, got %q", n.cfg.SSHKey)
	}
	if n.cfg.Args["username"] != "alice" {
		t.Errorf("Expected internal Args[username] unchanged, got %q", n.cfg.Args["username"])
	}
	if _, exists := n.cfg.Args["newkey"]; exists {
		t.Error("Expected internal Args not to have 'newkey'")
	}
}

// TestNodeImplementation_GetConfig_NilArgs verifies that GetConfig handles nil Args correctly.
func TestNodeImplementation_GetConfig_NilArgs(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     nil,
		},
	}

	cfg := n.GetNodeConfig()

	// Verify Args is nil in the copy
	if cfg.Args != nil {
		t.Errorf("Expected Args to be nil, got %v", cfg.Args)
	}
}

// TestNodeImplementation_GetConfig_EmptyArgs verifies that GetConfig handles empty Args correctly.
func TestNodeImplementation_GetConfig_EmptyArgs(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
	}

	cfg := n.GetNodeConfig()

	// Verify Args is empty in the copy
	if cfg.Args == nil {
		t.Error("Expected Args to be initialized, got nil")
	}
	if len(cfg.Args) != 0 {
		t.Errorf("Expected Args to be empty, got %d items", len(cfg.Args))
	}

	// Modify the returned config's Args
	cfg.Args["test"] = "value"

	// Verify the node's internal Args is unchanged
	if len(n.cfg.Args) != 0 {
		t.Errorf("Expected internal Args to remain empty, got %d items", len(n.cfg.Args))
	}
}

// TestNodeImplementation_GettersAfterSetters verifies that getters return updated values after setters.
func TestNodeImplementation_GettersAfterSetters(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
	}

	// Use setters to update configuration
	n.SetPort("2222").
		SetUser("deploy").
		SetKey("production.prv").
		SetArg("username", "alice")

	// Verify getters return updated values
	if n.GetHost() != "server.example.com" {
		t.Errorf("Expected GetHost()=%q, got %q", "server.example.com", n.GetHost())
	}
	if n.GetPort() != "2222" {
		t.Errorf("Expected GetPort()=%q, got %q", "2222", n.GetPort())
	}
	if n.GetUser() != "deploy" {
		t.Errorf("Expected GetUser()=%q, got %q", "deploy", n.GetUser())
	}
	if n.GetKey() != "production.prv" {
		t.Errorf("Expected GetKey()=%q, got %q", "production.prv", n.GetKey())
	}

	// Verify GetConfig returns updated values
	cfg := n.GetNodeConfig()
	if cfg.SSHPort != "2222" {
		t.Errorf("Expected config SSHPort=%q, got %q", "2222", cfg.SSHPort)
	}
	if cfg.RootUser != "deploy" {
		t.Errorf("Expected config RootUser=%q, got %q", "deploy", cfg.RootUser)
	}
	if cfg.SSHKey != "production.prv" {
		t.Errorf("Expected config SSHKey=%q, got %q", "production.prv", cfg.SSHKey)
	}
	if cfg.Args["username"] != "alice" {
		t.Errorf("Expected config Args[username]=%q, got %q", "alice", cfg.Args["username"])
	}
}

// TestNodeImplementation_IsConnected_Initial verifies that IsConnected returns false initially.
func TestNodeImplementation_IsConnected_Initial(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
		connected: false,
	}

	if n.IsConnected() {
		t.Error("Expected IsConnected() to return false initially")
	}
}

// TestNodeImplementation_Close_NotConnected verifies that Close is safe to call when not connected.
func TestNodeImplementation_Close_NotConnected(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
		connected: false,
		sshClient: nil,
	}

	// Should not panic or return error
	err := n.Close()
	if err != nil {
		t.Errorf("Expected Close() on non-connected node to succeed, got error: %v", err)
	}

	// Verify state remains consistent
	if n.IsConnected() {
		t.Error("Expected IsConnected() to return false after Close()")
	}
	if n.sshClient != nil {
		t.Error("Expected sshClient to remain nil after Close()")
	}
}

// TestNodeImplementation_Close_MultipleCalls verifies that Close can be called multiple times safely.
func TestNodeImplementation_Close_MultipleCalls(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
		connected: false,
		sshClient: nil,
	}

	// Call Close multiple times
	err1 := n.Close()
	if err1 != nil {
		t.Errorf("Expected first Close() to succeed, got error: %v", err1)
	}

	err2 := n.Close()
	if err2 != nil {
		t.Errorf("Expected second Close() to succeed, got error: %v", err2)
	}

	err3 := n.Close()
	if err3 != nil {
		t.Errorf("Expected third Close() to succeed, got error: %v", err3)
	}

	// Verify state remains consistent
	if n.IsConnected() {
		t.Error("Expected IsConnected() to return false after multiple Close() calls")
	}
}

// TestNodeImplementation_Run_WithoutPersistentConnection verifies that Run creates one-time connection when not connected.
func TestNodeImplementation_Run_WithoutPersistentConnection(t *testing.T) {
	// Mock ssh.RunSingleCommand via SetRunFunc
	var capturedHost, capturedPort, capturedUser, capturedKey, capturedCmd string
	ssh.SetRunFunc(func(cfg config.NodeConfig, cmd types.Command) (string, error) {
		capturedHost = cfg.SSHHost
		capturedPort = cfg.SSHPort
		capturedUser = cfg.RootUser
		capturedKey = cfg.SSHKey
		capturedCmd = cmd.Command
		return "output from one-time connection", nil
	})
	defer ssh.SetRunFunc(nil)

	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "2222",
			RootUser: "deploy",
			SSHKey:   "production.prv",
			Args:     make(map[string]string),
		},
		connected: false,
	}

	results := n.RunCommand("uptime")
	result := results.Results["server.example.com"]
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}

	if result.Message != "output from one-time connection" {
		t.Errorf("Expected output=%q, got %q", "output from one-time connection", result.Message)
	}

	// Verify correct parameters were passed to ssh.RunSingleCommand
	if capturedHost != "server.example.com" {
		t.Errorf("Expected host=%q, got %q", "server.example.com", capturedHost)
	}
	if capturedPort != "2222" {
		t.Errorf("Expected port=%q, got %q", "2222", capturedPort)
	}
	if capturedUser != "deploy" {
		t.Errorf("Expected user=%q, got %q", "deploy", capturedUser)
	}
	if capturedKey != "production.prv" {
		t.Errorf("Expected key=%q, got %q", "production.prv", capturedKey)
	}
	if capturedCmd != "uptime" {
		t.Errorf("Expected cmd=%q, got %q", "uptime", capturedCmd)
	}
}

// TestNodeImplementation_Run_OneTimeConnectionError verifies error handling with one-time connection.
func TestNodeImplementation_Run_OneTimeConnectionError(t *testing.T) {
	// Mock ssh.RunSingleCommand via SetRunFunc to return error
	ssh.SetRunFunc(func(cfg config.NodeConfig, cmd types.Command) (string, error) {
		return "", fmt.Errorf("connection refused")
	})
	defer ssh.SetRunFunc(nil)

	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
		connected: false,
	}

	results := n.RunCommand("uptime")
	result := results.Results["server.example.com"]
	if result.Error == nil {
		t.Error("Expected error, got nil")
	}

	if result.Message != "" {
		t.Errorf("Expected empty output on error, got %q", result.Message)
	}

	// Verify error message contains command
	if !contains(result.Error.Error(), "uptime") {
		t.Errorf("Expected error to contain command 'uptime', got: %v", result.Error)
	}

	// Verify error message contains failure reason
	if !contains(result.Error.Error(), "connection refused") {
		t.Errorf("Expected error to contain 'connection refused', got: %v", result.Error)
	}
}

// TestNodeImplementation_Playbook_Success verifies successful playbook execution.
func TestNodeImplementation_Playbook_Success(t *testing.T) {
	// Create a mock playbook
	var capturedConfig config.NodeConfig
	mockPlaybook := &mockPlaybook{
		name: "test-playbook",
		runFunc: func(cfg config.NodeConfig) error {
			capturedConfig = cfg
			return nil
		},
	}

	// Register mock playbook
	reg, err := GetGlobalPlaybookRegistry()
	if err != nil {
		t.Fatalf("GetGlobalPlaybookRegistry() failed: %v", err)
	}
	if err := reg.PlaybookRegister(mockPlaybook); err != nil {
		t.Fatalf("failed to register mock playbook: %v", err)
	}
	defer func() {
		// Clean up: remove mock playbook from registry
		// Note: Registry doesn't have Remove method, so we'll just leave it
	}()

	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "2222",
			RootUser: "deploy",
			SSHKey:   "production.prv",
			Args: map[string]string{
				"username": "alice",
				"shell":    "/bin/bash",
			},
		},
		connected: false,
	}

	results := n.RunPlaybookByID("test-playbook")
	result := results.Results["server.example.com"]
	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}

	// Verify correct config was passed to playbook
	if capturedConfig.SSHHost != "server.example.com" {
		t.Errorf("Expected SSHHost=%q, got %q", "server.example.com", capturedConfig.SSHHost)
	}
	if capturedConfig.SSHPort != "2222" {
		t.Errorf("Expected SSHPort=%q, got %q", "2222", capturedConfig.SSHPort)
	}
	if capturedConfig.RootUser != "deploy" {
		t.Errorf("Expected RootUser=%q, got %q", "deploy", capturedConfig.RootUser)
	}
	if capturedConfig.SSHKey != "production.prv" {
		t.Errorf("Expected SSHKey=%q, got %q", "production.prv", capturedConfig.SSHKey)
	}
	if capturedConfig.Args["username"] != "alice" {
		t.Errorf("Expected Args[username]=%q, got %q", "alice", capturedConfig.Args["username"])
	}
	if capturedConfig.Args["shell"] != "/bin/bash" {
		t.Errorf("Expected Args[shell]=%q, got %q", "/bin/bash", capturedConfig.Args["shell"])
	}
}

// TestNodeImplementation_Playbook_NotFound verifies error when playbook is not in registry.
func TestNodeImplementation_Playbook_NotFound(t *testing.T) {
	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
		connected: false,
	}

	results := n.RunPlaybookByID("nonexistent-playbook")
	result := results.Results["server.example.com"]
	if result.Error == nil {
		t.Error("Expected error for nonexistent playbook, got nil")
	}

	// Verify error message contains playbook name
	if !contains(result.Error.Error(), "nonexistent-playbook") {
		t.Errorf("Expected error to contain playbook name, got: %v", result.Error)
	}

	// Verify error message indicates not found
	if !contains(result.Error.Error(), "not found") {
		t.Errorf("Expected error to contain 'not found', got: %v", result.Error)
	}
}

// TestNodeImplementation_Playbook_ExecutionError verifies error handling when playbook execution fails.
func TestNodeImplementation_Playbook_ExecutionError(t *testing.T) {
	// Create a mock playbook that fails
	mockPlaybook := &mockPlaybook{
		name: "failing-playbook",
		runFunc: func(cfg config.NodeConfig) error {
			return fmt.Errorf("playbook execution failed")
		},
	}

	// Register mock playbook
	reg, err := GetGlobalPlaybookRegistry()
	if err != nil {
		t.Fatalf("GetGlobalPlaybookRegistry() failed: %v", err)
	}
	if err := reg.PlaybookRegister(mockPlaybook); err != nil {
		t.Fatalf("failed to register mock playbook: %v", err)
	}

	n := &nodeImplementation{
		cfg: config.NodeConfig{
			SSHHost:  "server.example.com",
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
		connected: false,
	}

	results := n.RunPlaybookByID("failing-playbook")
	result := results.Results["server.example.com"]
	if result.Error == nil {
		t.Error("Expected error from failing playbook, got nil")
	}

	// Verify error message contains playbook name
	if !contains(result.Error.Error(), "failing-playbook") {
		t.Errorf("Expected error to contain playbook name, got: %v", result.Error)
	}

	// Verify error message contains failure reason
	if !contains(result.Error.Error(), "playbook execution failed") {
		t.Errorf("Expected error to contain 'playbook execution failed', got: %v", result.Error)
	}
}

// Mock types for testing

// mockPlaybook is a mock implementation of playbook.Playbook for testing.
type mockPlaybook struct {
	name      string
	cfg       config.NodeConfig
	args      map[string]string
	dryRun    bool
	timeout   time.Duration
	runFunc   func(config.NodeConfig) error
	checkFunc func(config.NodeConfig) (bool, error)
}

func (m *mockPlaybook) GetID() string {
	return m.name
}

func (m *mockPlaybook) SetID(id string) types.PlaybookInterface {
	m.name = id
	return m
}

func (m *mockPlaybook) GetDescription() string {
	return "Mock playbook for testing"
}

func (m *mockPlaybook) SetDescription(description string) types.PlaybookInterface {
	return m
}

func (m *mockPlaybook) GetNodeConfig() config.NodeConfig {
	return m.cfg
}

func (m *mockPlaybook) SetNodeConfig(cfg config.NodeConfig) types.PlaybookInterface {
	m.cfg = cfg
	return m
}

func (m *mockPlaybook) GetArg(key string) string {
	return m.args[key]
}

func (m *mockPlaybook) SetArg(key, value string) types.PlaybookInterface {
	if m.args == nil {
		m.args = make(map[string]string)
	}
	m.args[key] = value
	return m
}

func (m *mockPlaybook) GetArgs() map[string]string {
	return m.args
}

func (m *mockPlaybook) SetArgs(args map[string]string) types.PlaybookInterface {
	m.args = args
	return m
}

func (m *mockPlaybook) IsDryRun() bool {
	return m.dryRun
}

func (m *mockPlaybook) SetDryRun(dryRun bool) types.PlaybookInterface {
	m.dryRun = dryRun
	return m
}

func (m *mockPlaybook) GetTimeout() time.Duration {
	return m.timeout
}

func (m *mockPlaybook) SetTimeout(timeout time.Duration) types.PlaybookInterface {
	m.timeout = timeout
	return m
}

func (m *mockPlaybook) Check() (bool, error) {
	if m.checkFunc != nil {
		return m.checkFunc(m.cfg)
	}
	return true, nil
}

func (m *mockPlaybook) Run() types.Result {
	if m.runFunc != nil {
		err := m.runFunc(m.cfg)
		if err != nil {
			return types.Result{
				Changed: false,
				Message: fmt.Sprintf("Playbook '%s' failed", m.name),
				Error:   fmt.Errorf("playbook '%s': %w", m.name, err),
			}
		}
	}
	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("Playbook '%s' executed", m.name),
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
