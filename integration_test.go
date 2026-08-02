package ork_test

// Integration tests for the Ork Node API.
//
// These tests use testcontainers-go to spin up real SSH servers and test
// the Node API against them. They are skipped when running with
// the -short flag, or when the CI env var is not set.
//
// Requirements:
//   - Docker must be installed and running
//   - Tests use the linuxserver/openssh-server container image
//   - An ed25519 keypair is generated per test and injected into the container
//     via the PUBLIC_KEY env var; the private key is written to
//     ~/.ssh/ork-integration-<random> and removed on test cleanup.
//
// Running integration tests:
//   CI=true go test -v -run Integration   # Run only integration tests
//   go test -v -short                     # Skip integration tests
//   go test -v                            # Skips integration tests unless CI=true
//
// For manual debugging against a standing SSH server, see docker-compose.yml
// in the repository root (docker compose up, then ssh testuser@localhost -p 2222).

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dracory/ork"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/ssh"
)

// sshContainer wraps a testcontainers SSH server for integration testing
type sshContainer struct {
	container testcontainers.Container
	host      string
	port      string
	user      string
	keyName   string // filename of the private key under ~/.ssh/
}

// integrationKeyName is a fixed prefix for keys written by the integration
// tests, so they are easy to identify and clean up.
const integrationKeyName = "ork-integration-test-ed25519"

// setupSSHContainer starts an SSH test container with key-based authentication.
// It generates a fresh ed25519 keypair, writes the private key to
// ~/.ssh/<integrationKeyName> (overwriting any prior file), and injects the
// public key into the container via the PUBLIC_KEY env var. The key file is
// removed on test cleanup.
func setupSSHContainer(t *testing.T) *sshContainer {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Only run integration tests in CI (GitHub Actions) or when explicitly
	// opted in via CI=true.
	if os.Getenv("CI") == "" {
		t.Skip("skipping integration test: only runs in CI (set CI=true to run)")
	}

	// --- Generate ed25519 keypair ---
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate ed25519 key: %v", err)
	}

	// Marshal private key as PKCS8 PEM
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	// Marshal public key in OpenSSH authorized_keys format
	sshPub, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}
	authorizedKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub)))

	// --- Write private key to ~/.ssh/<integrationKeyName> ---
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to resolve current user: %v", err)
	}
	sshDir := filepath.Join(usr.HomeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatalf("Failed to create ~/.ssh: %v", err)
	}
	keyPath := filepath.Join(sshDir, integrationKeyName)
	if err := os.WriteFile(keyPath, privPEM, 0o600); err != nil {
		t.Fatalf("Failed to write private key to %s: %v", keyPath, err)
	}
	t.Cleanup(func() {
		if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove integration key %s: %v", keyPath, err)
		}
	})

	// --- Start the SSH container with the public key injected ---
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "linuxserver/openssh-server:latest",
		ExposedPorts: []string{"2222/tcp"},
		Env: map[string]string{
			"PUID":            "1000",
			"PGID":            "1000",
			"TZ":              "UTC",
			"PASSWORD_ACCESS": "false",
			"USER_NAME":       "testuser",
			"PUBLIC_KEY":      authorizedKey,
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

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "2222")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	// Give sshd a moment to be fully ready after the log line.
	time.Sleep(2 * time.Second)

	sc := &sshContainer{
		container: container,
		host:      host,
		port:      mappedPort.Port(),
		user:      "testuser",
		keyName:   integrationKeyName,
	}

	// Capture the container's host key and add it to ~/.ssh/known_hosts so
	// the Node's strict known_hosts verification succeeds. We dial once with
	// an insecure callback to grab the key, then write it in known_hosts
	// format ([host]:port algo key).
	if err := addHostKeyToKnownHosts(sc.host, sc.port, privPEM); err != nil {
		t.Fatalf("Failed to capture/add container host key: %v", err)
	}

	return sc
}

// addHostKeyToKnownHosts dials the SSH server once with an insecure
// HostKeyCallback to capture the server's public key, then appends it to
// ~/.ssh/known_hosts in the format expected by golang.org/x/crypto/ssh/knownhosts.
// privKeyPEM is used to authenticate during the probe dial (the host key is
// exchanged before auth, so even an auth failure would still capture the key,
// but using the real key lets the full handshake succeed cleanly).
func addHostKeyToKnownHosts(host, port string, privKeyPEM []byte) error {
	signer, err := ssh.ParsePrivateKey(privKeyPEM)
	if err != nil {
		return fmt.Errorf("parse probe key: %w", err)
	}

	var capturedHostKey ssh.PublicKey
	addr := host + ":" + port
	config := &ssh.ClientConfig{
		User: "testuser",
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			capturedHostKey = key
			return nil
		},
		Timeout: 10 * time.Second,
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		// Auth may have failed, but the host key is captured during kex
		// which happens before auth. If we got the key, that's enough.
		if capturedHostKey == nil {
			return fmt.Errorf("probe dial failed and no host key captured: %w", err)
		}
	} else {
		client.Close()
	}

	if capturedHostKey == nil {
		return fmt.Errorf("host key was not captured during probe dial")
	}

	// known_hosts format for non-standard ports: [host]:port algo base64key
	line := fmt.Sprintf("[%s]:%s %s\n", host, port, strings.TrimSpace(string(ssh.MarshalAuthorizedKey(capturedHostKey))))

	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("resolve user: %w", err)
	}
	knownHostsPath := filepath.Join(usr.HomeDir, ".ssh", "known_hosts")

	// Append (don't overwrite) — there may be other entries.
	f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open known_hosts: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("write known_hosts: %w", err)
	}

	return nil
}

// String returns a human-readable description for debug logging.
func (sc *sshContainer) String() string {
	return fmt.Sprintf("ssh %s@%s:%s (key=%s)", sc.user, sc.host, sc.port, sc.keyName)
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

// TestIntegration_Node_ConnectRunClose tests Node lifecycle with real SSH.
// This is the canonical end-to-end smoke test: connect, run a command,
// verify output, close, verify disconnected.
func TestIntegration_Node_ConnectRunClose(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

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
	results := node.RunCommand("echo 'test1'")
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Run failed: %v", result.Error)
	}
	if !strings.Contains(result.Message, "test1") {
		t.Errorf("Expected output to contain 'test1', got: %s", result.Message)
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
		results := node.RunCommand(cmd)
		result := results.Results[container.host]
		if result.Error != nil {
			t.Errorf("Run %d failed: %v", i+1, result.Error)
			continue
		}
		t.Logf("Command %d output: %s", i+1, result.Message)
	}

	// Verify still connected after multiple operations
	if !node.IsConnected() {
		t.Error("Expected connection to remain active after multiple Run calls")
	}
}

// TestIntegration_Node_WithoutPersistentConnection tests one-time connections.
// RunCommand is called without a prior Connect(); the Node should open a
// one-shot connection, run the command, and close the connection. After the
// call, IsConnected() must return false.
func TestIntegration_Node_WithoutPersistentConnection(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	// Run without calling Connect() - should create one-time connection
	results := node.RunCommand("echo 'one-time'")
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Run failed: %v", result.Error)
	}

	if !strings.Contains(result.Message, "one-time") {
		t.Errorf("Expected output to contain 'one-time', got: %s", result.Message)
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
	results := node.RunByID("ping")
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Playbook('ping') failed: %v", result.Error)
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
	results1 := node.RunCommand("echo 'step1'")
	result1 := results1.Results[container.host]
	if result1.Error != nil {
		t.Fatalf("Step 1 failed: %v", result1.Error)
	}
	if !strings.Contains(result1.Message, "step1") {
		t.Errorf("Step 1: expected 'step1' in output, got: %s", result1.Message)
	}

	// Test 2: Update configuration
	node.SetArg("test", "value")

	// Test 3: Run another command
	results2 := node.RunCommand("echo 'step2'")
	result2 := results2.Results[container.host]
	if result2.Error != nil {
		t.Fatalf("Step 2 failed: %v", result2.Error)
	}
	if !strings.Contains(result2.Message, "step2") {
		t.Errorf("Step 2: expected 'step2' in output, got: %s", result2.Message)
	}

	// Test 4: Execute playbook
	results3 := node.RunByID("ping")
	result3 := results3.Results[container.host]
	if result3.Error != nil {
		t.Fatalf("Playbook execution failed: %v", result3.Error)
	}

	// Test 5: Run final command
	results4 := node.RunCommand("whoami")
	result4 := results4.Results[container.host]
	if result4.Error != nil {
		t.Fatalf("Step 3 failed: %v", result4.Error)
	}
	if !strings.Contains(result4.Message, container.user) {
		t.Errorf("Step 3: expected '%s' in output, got: %s", container.user, result4.Message)
	}

	// Verify connection remained active throughout
	if !node.IsConnected() {
		t.Error("Expected connection to remain active throughout operations")
	}
}
