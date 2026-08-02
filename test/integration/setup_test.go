package ork_test

// Shared setup helpers for integration tests.
//
// These tests use testcontainers-go to spin up real SSH servers and test
// the Ork APIs against them. They are skipped when running with the -short
// flag, or when the CI env var is not set.
//
// Requirements:
//   - Docker must be installed and running
//   - Tests use the linuxserver/openssh-server container image
//   - An ed25519 keypair is generated per test and injected into the container
//     via the PUBLIC_KEY env var; the private key is written to
//     ~/.ssh/ork-integration-test-ed25519 and removed on test cleanup.
//
// Running integration tests:
//   CI=true go test -v ./test/integration/...   # Run only integration tests
//   go test -v -short ./test/integration/...    # Skip integration tests
//   go test -v ./test/integration/...           # Skips unless CI=true
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

// setupSSHContainerWithSudo starts an SSH container and configures passwordless
// sudo for the testuser. This is needed for BecomeUser (sudo escalation) tests.
// The linuxserver/openssh-server image is Alpine-based and already has sudo
// installed — we just need to add testuser to the sudoers file.
func setupSSHContainerWithSudo(t *testing.T) *sshContainer {
	t.Helper()
	sc := setupSSHContainer(t)

	ctx := context.Background()

	// Configure NOPASSWD sudo for testuser. The container runs as root during
	// init, so we can write to /etc/sudoers.
	exitCode, _, err := sc.container.Exec(ctx, []string{
		"sh", "-c",
		"echo 'testuser ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers",
	})
	if err != nil {
		t.Fatalf("Failed to configure sudo in container: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("Failed to configure sudoers in container, exit code: %d", exitCode)
	}

	return sc
}
