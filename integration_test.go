package ork_test

// Integration tests for the Ork Node API.
//
// These tests use testcontainers-go to spin up real SSH servers and test
// the Node API against them. They are skipped when running with
// the -short flag.
//
// Requirements:
//   - Docker must be installed and running
//   - Tests use the linuxserver/openssh-server container image
//
// Running integration tests:
//   go test -v                    # Run all tests including integration
//   go test -v -short             # Skip integration tests
//   go test -v -run Integration   # Run only integration tests
//
// Note: Most integration tests are currently skipped with t.Skip() because
// they require SSH key-based authentication setup in the container. The
// container setup code is in place and can be extended to generate and
// configure SSH keys for full integration testing.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dracory/ork"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// sshContainer wraps a testcontainers SSH server for integration testing
type sshContainer struct {
	container testcontainers.Container
	host      string
	port      string
	user      string
	keyPath   string
}

// setupSSHContainer starts an SSH test container with key-based authentication
func setupSSHContainer(t *testing.T) *sshContainer {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Only run integration tests in CI (GitHub Actions)
	if os.Getenv("CI") == "" {
		t.Skip("skipping integration test: only runs in CI (set CI=true to run)")
	}

	ctx := context.Background()

	// Create temporary directory for SSH keys
	tmpDir, err := os.MkdirTemp("", "ork-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	// Generate SSH key pair for testing
	privateKeyPath := filepath.Join(tmpDir, "test_key")
	_ = filepath.Join(tmpDir, "test_key.pub") // publicKeyPath for future use

	// Use a simple test key (in real scenario, generate with ssh-keygen)
	// For testing, we'll use the linuxserver/openssh-server image which accepts password auth
	// and we can configure it with environment variables

	req := testcontainers.ContainerRequest{
		Image:        "linuxserver/openssh-server:latest",
		ExposedPorts: []string{"2222/tcp"},
		Env: map[string]string{
			"PUID":            "1000",
			"PGID":            "1000",
			"TZ":              "UTC",
			"PASSWORD_ACCESS": "true",
			"USER_PASSWORD":   "testpass",
			"USER_NAME":       "testuser",
		},
		WaitingFor: wait.ForLog("done.").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start SSH container: %v", err)
	}

	// Get container host and port
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "2222")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	// Wait a bit for SSH to be fully ready
	time.Sleep(2 * time.Second)

	return &sshContainer{
		container: container,
		host:      host,
		port:      mappedPort.Port(),
		user:      "testuser",
		keyPath:   privateKeyPath,
	}
}

// terminate stops and removes the SSH container
func (sc *sshContainer) terminate(t *testing.T) {
	if sc.container != nil {
		ctx := context.Background()
		if err := sc.container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}
}

// TestIntegration_Node_ConnectRunClose tests Node lifecycle with real SSH
func TestIntegration_Node_ConnectRunClose(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	t.Skip("Skipping: requires SSH key setup in container")

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey("test_key")

	// Test Connect
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Verify connected state
	if !node.IsConnected() {
		t.Error("Expected IsConnected() to return true after Connect()")
	}

	// Test RunCommand with persistent connection
	output, err := node.RunCommand("echo 'test1'")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(output, "test1") {
		t.Errorf("Expected output to contain 'test1', got: %s", output)
	}

	// Test Close
	err = node.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify disconnected state
	if node.IsConnected() {
		t.Error("Expected IsConnected() to return false after Close()")
	}
}

// TestIntegration_Node_PersistentConnectionReuse tests connection reuse
func TestIntegration_Node_PersistentConnectionReuse(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	t.Skip("Skipping: requires SSH key setup in container")

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey("test_key")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Execute multiple commands on same connection
	commands := []string{
		"echo 'command1'",
		"echo 'command2'",
		"echo 'command3'",
		"pwd",
		"whoami",
	}

	for i, cmd := range commands {
		output, err := node.RunCommand(cmd)
		if err != nil {
			t.Errorf("Run %d failed: %v", i+1, err)
			continue
		}
		t.Logf("Command %d output: %s", i+1, output)
	}

	// Verify still connected after multiple operations
	if !node.IsConnected() {
		t.Error("Expected connection to remain active after multiple Run calls")
	}
}

// TestIntegration_Node_WithoutPersistentConnection tests one-time connections
func TestIntegration_Node_WithoutPersistentConnection(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	t.Skip("Skipping: requires SSH key setup in container")

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey("test_key")

	// Run without calling Connect() - should create one-time connection
	output, err := node.RunCommand("echo 'one-time'")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !strings.Contains(output, "one-time") {
		t.Errorf("Expected output to contain 'one-time', got: %s", output)
	}

	// Verify not connected (one-time connection was closed)
	if node.IsConnected() {
		t.Error("Expected IsConnected() to return false after one-time Run")
	}
}

// TestIntegration_Node_Playbook tests playbook execution via Node
func TestIntegration_Node_Playbook(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	t.Skip("Skipping: requires SSH key setup in container")

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey("test_key")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Test ping playbook
	err = node.RunPlaybook("ping")
	if err != nil {
		t.Fatalf("Playbook('ping') failed: %v", err)
	}
}

// TestIntegration_MultipleOperations tests complex workflows
func TestIntegration_MultipleOperations(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	t.Skip("Skipping: requires SSH key setup in container")

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey("test_key")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Test 1: Run command
	output1, err := node.RunCommand("echo 'step1'")
	if err != nil {
		t.Fatalf("Step 1 failed: %v", err)
	}
	if !strings.Contains(output1, "step1") {
		t.Errorf("Step 1: expected 'step1' in output, got: %s", output1)
	}

	// Test 2: Update configuration
	node.SetArg("test", "value")

	// Test 3: Run another command
	output2, err := node.RunCommand("echo 'step2'")
	if err != nil {
		t.Fatalf("Step 2 failed: %v", err)
	}
	if !strings.Contains(output2, "step2") {
		t.Errorf("Step 2: expected 'step2' in output, got: %s", output2)
	}

	// Test 4: Execute playbook
	err = node.RunPlaybook("ping")
	if err != nil {
		t.Fatalf("Playbook execution failed: %v", err)
	}

	// Test 5: Run final command
	output3, err := node.RunCommand("whoami")
	if err != nil {
		t.Fatalf("Step 3 failed: %v", err)
	}
	if !strings.Contains(output3, container.user) {
		t.Errorf("Step 3: expected '%s' in output, got: %s", container.user, output3)
	}

	// Verify connection remained active throughout
	if !node.IsConnected() {
		t.Error("Expected connection to remain active throughout operations")
	}
}
