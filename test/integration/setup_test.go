package ork_test

// Shared setup helpers for integration tests.
//
// These tests use testcontainers-go to spin up real SSH servers and test
// the Ork APIs against them. They are skipped when running with the -short
// flag, or when the CI env var is not set.
//
// Architecture (Ansible-style shared containers):
//
// Instead of each test starting its own container (which is slow and caused
// CI timeouts), TestMain starts a small set of long-lived containers ONCE
// before any test runs. All tests reuse these shared containers. This cuts
// ~88 container starts down to 5, reducing runtime from minutes to seconds.
//
// Shared containers:
//   - sharedSSH1: basic linuxserver/openssh-server (no sudo) — most tests
//   - sharedSSH2: basic linuxserver/openssh-server (no sudo) — group tests
//     that need two distinct hosts (results are keyed by host)
//   - sharedSudo: linuxserver/openssh-server with NOPASSWD sudo — become tests
//   - sharedSudoPassword: linuxserver/openssh-server with password sudo —
//     BecomePassword state machine tests
//   - sharedSystemd: geerlingguy/docker-ubuntu2404-ansible (systemd) — caddy
//     tests that need apt/systemctl
//
// A single ed25519 keypair is generated in TestMain and injected into all
// containers. The private key is written to ~/.ssh/ork-integration-test-ed25519
// and removed on cleanup. Host keys are added to ~/.ssh/known_hosts once per
// container.
//
// Tests that mutate container state (e.g., caddy install) must reset state
// if a subsequent test depends on a clean slate. See resetCaddyState.
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

	"github.com/moby/moby/api/types/container"
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
	shared    bool   // true for containers managed by TestMain — terminate is a no-op
}

// integrationKeyName is a fixed prefix for keys written by the integration
// tests, so they are easy to identify and clean up.
const integrationKeyName = "ork-integration-test-ed25519"

// Shared container instances — started once in TestMain, reused by all tests.
var (
	sharedSSH1         *sshContainer // basic SSH (no sudo)
	sharedSSH2         *sshContainer // basic SSH #2 (for group tests needing 2 hosts)
	sharedSudo         *sshContainer // SSH with NOPASSWD sudo
	sharedSudoPassword *sshContainer // SSH with password sudo
	sharedSystemd      *sshContainer // systemd Ubuntu (for caddy tests)

	integrationKeyPEM  []byte // shared private key PEM (set in TestMain)
	integrationKeyPath string // path to the private key file (set in TestMain)
)

// TestMain starts all shared containers once, runs the tests, then tears down.
// If CI is not set, containers are not started and each test skips itself via
// skipIfNotIntegration. The -short flag is also checked per-test in
// skipIfNotIntegration (testing.Short() cannot be called in TestMain before
// flags are parsed).
func TestMain(m *testing.M) {
	if os.Getenv("CI") == "" {
		os.Exit(m.Run())
	}

	if err := startSharedContainers(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start shared containers: %v\n", err)
		stopSharedContainers()
		os.Exit(1)
	}

	code := m.Run()

	stopSharedContainers()
	os.Exit(code)
}

// startSharedContainers generates a shared keypair, writes the private key,
// and starts all 5 shared containers. If any container fails to start, the
// caller is responsible for calling stopSharedContainers to clean up.
func startSharedContainers() error {
	// 1. Generate keypair (shared across all containers)
	privPEM, authorizedKey, err := generateIntegrationKeypair()
	if err != nil {
		return fmt.Errorf("generate keypair: %w", err)
	}
	integrationKeyPEM = privPEM

	// 2. Write private key to ~/.ssh/<integrationKeyName>
	keyPath, err := writeIntegrationKey(privPEM)
	if err != nil {
		return fmt.Errorf("write integration key: %w", err)
	}
	integrationKeyPath = keyPath

	// 3. Start containers
	sharedSSH1, err = startBasicSSHContainer(authorizedKey)
	if err != nil {
		return fmt.Errorf("start basic SSH #1: %w", err)
	}

	sharedSSH2, err = startBasicSSHContainer(authorizedKey)
	if err != nil {
		return fmt.Errorf("start basic SSH #2: %w", err)
	}

	sharedSudo, err = startSudoSSHContainer(authorizedKey, false, "")
	if err != nil {
		return fmt.Errorf("start sudo SSH: %w", err)
	}

	sharedSudoPassword, err = startSudoSSHContainer(authorizedKey, true, "sudopass")
	if err != nil {
		return fmt.Errorf("start sudo password SSH: %w", err)
	}

	// The systemd container (geerlingguy/docker-ubuntu2404-ansible) runs
	// systemd as PID 1 and requires cgroup v1 or a compatible cgroup v2
	// setup. On some CI runners (e.g. GitHub Actions with cgroup v2) systemd
	// fails to initialize and the container exits 255 immediately. Rather
	// than failing the entire integration suite, log a warning and leave
	// sharedSystemd nil — tests that need it (caddy) skip themselves via
	// getSystemdContainer's nil check.
	sharedSystemd, err = startSystemdSSHContainer(authorizedKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: systemd SSH container unavailable (caddy tests will be skipped): %v\n", err)
		sharedSystemd = nil
	}

	return nil
}

// stopSharedContainers terminates all started shared containers and removes
// the integration private key. Safe to call even if some containers are nil
// (e.g., after a partial startup failure).
func stopSharedContainers() {
	ctx := context.Background()
	for _, sc := range []*sshContainer{sharedSSH1, sharedSSH2, sharedSudo, sharedSudoPassword, sharedSystemd} {
		if sc != nil && sc.container != nil {
			if err := sc.container.Terminate(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to terminate container: %v\n", err)
			}
		}
	}
	if integrationKeyPath != "" {
		if err := os.Remove(integrationKeyPath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Failed to remove integration key %s: %v\n", integrationKeyPath, err)
		}
	}
}

// skipIfNotIntegration skips the test if not in CI or running in short mode.
func skipIfNotIntegration(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getenv("CI") == "" {
		t.Skip("skipping integration test: only runs in CI (set CI=true to run)")
	}
}

// --- Internal helpers (no *testing.T dependency) ---

// generateIntegrationKeypair generates an ed25519 keypair for integration tests.
// Returns the private key PEM and the authorized_keys format public key.
func generateIntegrationKeypair() (privPEM []byte, authorizedKey string, err error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("generate ed25519 key: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, "", fmt.Errorf("marshal private key: %w", err)
	}
	privPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	sshPub, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, "", fmt.Errorf("marshal public key: %w", err)
	}
	authorizedKey = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub)))

	return privPEM, authorizedKey, nil
}

// writeIntegrationKey writes the private key to ~/.ssh/<integrationKeyName>
// and returns the full path.
func writeIntegrationKey(privPEM []byte) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("resolve current user: %w", err)
	}
	sshDir := filepath.Join(usr.HomeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return "", fmt.Errorf("create ~/.ssh: %w", err)
	}
	keyPath := filepath.Join(sshDir, integrationKeyName)
	if err := os.WriteFile(keyPath, privPEM, 0o600); err != nil {
		return "", fmt.Errorf("write private key: %w", err)
	}
	return keyPath, nil
}

// startBasicSSHContainer starts a linuxserver/openssh-server container with
// the given public key injected via the PUBLIC_KEY env var.
func startBasicSSHContainer(authorizedKey string) (*sshContainer, error) {
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
	return startSSHContainer(req, "2222", "testuser", 2*time.Second)
}

// startSudoSSHContainer starts a linuxserver/openssh-server container with
// sudo configured for testuser. If passwordRequired is false, NOPASSWD is set.
// If true, the given password is set for testuser and sudo requires it.
func startSudoSSHContainer(authorizedKey string, passwordRequired bool, password string) (*sshContainer, error) {
	sc, err := startBasicSSHContainer(authorizedKey)
	if err != nil {
		return nil, err
	}
	if err := configureSudoInContainer(sc, passwordRequired, password); err != nil {
		sc.container.Terminate(context.Background())
		return nil, err
	}
	return sc, nil
}

// configureSudoInContainer configures sudo for testuser in the container.
// If passwordRequired is false, NOPASSWD is set. If true, the given password
// is set for testuser and sudo requires it. Uses a drop-in file under
// /etc/sudoers.d/ for robustness. Includes Defaults lecture="never" to
// suppress the sudo lecture that would otherwise pollute command output on
// first password-based sudo use.
func configureSudoInContainer(sc *sshContainer, passwordRequired bool, password string) error {
	ctx := context.Background()

	if passwordRequired {
		exitCode, _, err := sc.container.Exec(ctx, []string{
			"sh", "-c", "echo 'testuser:" + password + "' | chpasswd",
		})
		if err != nil || exitCode != 0 {
			return fmt.Errorf("set testuser password: err=%v exitCode=%d", err, exitCode)
		}
	}

	sudoersLine := "testuser ALL=(ALL) NOPASSWD:ALL"
	if passwordRequired {
		sudoersLine = "Defaults lecture=\"never\"\ntestuser ALL=(ALL) ALL"
	}
	exitCode, _, err := sc.container.Exec(ctx, []string{
		"sh", "-c",
		"mkdir -p /etc/sudoers.d && printf '" + sudoersLine + "\\n' > /etc/sudoers.d/testuser && chmod 440 /etc/sudoers.d/testuser && visudo -cf /etc/sudoers.d/testuser",
	})
	if err != nil {
		return fmt.Errorf("configure sudo: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("configure sudoers: exit code %d", exitCode)
	}
	return nil
}

// startSystemdSSHContainer starts a systemd-enabled Ubuntu container for
// integration tests that require apt, systemctl, and other Debian/Ubuntu
// tools. The container runs systemd as PID 1 (privileged mode required).
//
// The image (geerlingguy/docker-ubuntu2404-ansible) is purpose-built for
// Ansible/Molecule testing: it has Python, SSH, and systemd pre-installed.
//
// SSH key injection is done post-start via container.Exec (the image does
// not support PUBLIC_KEY env var like linuxserver/openssh-server).
func startSystemdSSHContainer(authorizedKey string) (*sshContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "geerlingguy/docker-ubuntu2404-ansible:latest",
		ExposedPorts: []string{"22/tcp"},
		Cmd:          []string{"/lib/systemd/systemd"},
		Privileged:   true,
		WaitingFor:   wait.ForLog("Reached target .*Multi-User System.*").WithStartupTimeout(120 * time.Second),
		Binds:        []string{"/sys/fs/cgroup:/sys/fs/cgroup:rw"},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.CgroupnsMode = "host"
		},
		Tmpfs: map[string]string{
			"/run": "",
			"/tmp": "rw,mode=1777",
		},
	}

	sc, err := startSSHContainer(req, "22", "root", 3*time.Second)
	if err != nil {
		return nil, err
	}

	// Inject SSH public key for root
	ctx := context.Background()
	exitCode, _, err := sc.container.Exec(ctx, []string{
		"sh", "-c",
		"mkdir -p /root/.ssh && chmod 700 /root/.ssh && echo '" + authorizedKey + "' > /root/.ssh/authorized_keys && chmod 600 /root/.ssh/authorized_keys",
	})
	if err != nil || exitCode != 0 {
		sc.container.Terminate(ctx)
		return nil, fmt.Errorf("inject SSH key (exitCode=%d): %w", exitCode, err)
	}

	return sc, nil
}

// startSSHContainer is the common helper that starts a container from a
// request, waits for readiness, gets host/port, and adds the host key to
// known_hosts. The readyDelay gives the service a moment to be fully ready
// after the wait strategy succeeds.
func startSSHContainer(req testcontainers.ContainerRequest, port string, user string, readyDelay time.Duration) (*sshContainer, error) {
	ctx := context.Background()

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("start container: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		c.Terminate(ctx)
		return nil, fmt.Errorf("get container host: %w", err)
	}

	mappedPort, err := c.MappedPort(ctx, port)
	if err != nil {
		c.Terminate(ctx)
		return nil, fmt.Errorf("get container port: %w", err)
	}

	time.Sleep(readyDelay)

	sc := &sshContainer{
		container: c,
		host:      host,
		port:      mappedPort.Port(),
		user:      user,
		keyName:   integrationKeyName,
		shared:    true,
	}

	// Capture the container's host key and add it to ~/.ssh/known_hosts so
	// the Node's strict known_hosts verification succeeds.
	if err := addHostKeyToKnownHosts(sc.host, sc.port, integrationKeyPEM); err != nil {
		c.Terminate(ctx)
		return nil, fmt.Errorf("add host key to known_hosts: %w", err)
	}

	return sc, nil
}

// --- Public setup functions (return shared containers) ---

// setupSSHContainer returns the shared basic SSH container (no sudo).
// The container is started once in TestMain and reused across all tests.
// terminate() is a no-op for shared containers.
func setupSSHContainer(t *testing.T) *sshContainer {
	t.Helper()
	skipIfNotIntegration(t)
	if sharedSSH1 == nil {
		t.Fatal("shared SSH container not initialized (TestMain failed to start containers)")
	}
	return sharedSSH1
}

// setupSSHContainer2 returns the second shared basic SSH container.
// Used by group tests that need two distinct hosts (results are keyed by host).
func setupSSHContainer2(t *testing.T) *sshContainer {
	t.Helper()
	skipIfNotIntegration(t)
	if sharedSSH2 == nil {
		t.Fatal("shared SSH container 2 not initialized (TestMain failed to start containers)")
	}
	return sharedSSH2
}

// setupSSHContainerWithSudo returns the shared SSH container with NOPASSWD
// sudo configured for testuser. Used by become tests.
func setupSSHContainerWithSudo(t *testing.T) *sshContainer {
	t.Helper()
	skipIfNotIntegration(t)
	if sharedSudo == nil {
		t.Fatal("shared sudo SSH container not initialized (TestMain failed to start containers)")
	}
	return sharedSudo
}

// setupSSHContainerWithSudoPassword returns the shared SSH container with
// password-based sudo configured for testuser (password is "sudopass").
// Used by BecomePassword state machine tests.
func setupSSHContainerWithSudoPassword(t *testing.T) *sshContainer {
	t.Helper()
	skipIfNotIntegration(t)
	if sharedSudoPassword == nil {
		t.Fatal("shared sudo password SSH container not initialized (TestMain failed to start containers)")
	}
	return sharedSudoPassword
}

// setupSSHContainerWithSudoConfig returns the shared sudo container matching
// the given configuration. If passwordRequired is true, returns the password
// sudo container; otherwise returns the NOPASSWD sudo container.
func setupSSHContainerWithSudoConfig(t *testing.T, passwordRequired bool, password string) *sshContainer {
	t.Helper()
	skipIfNotIntegration(t)
	if passwordRequired {
		if sharedSudoPassword == nil {
			t.Fatal("shared sudo password SSH container not initialized")
		}
		return sharedSudoPassword
	}
	if sharedSudo == nil {
		t.Fatal("shared sudo SSH container not initialized")
	}
	return sharedSudo
}

// setupSSHContainerSystemd returns the shared systemd-enabled Ubuntu container.
// Used by caddy tests that require apt, systemctl, etc.
func setupSSHContainerSystemd(t *testing.T) *sshContainer {
	t.Helper()
	skipIfNotIntegration(t)
	if sharedSystemd == nil {
		t.Skip("shared systemd SSH container unavailable (systemd not supported in this environment)")
	}
	return sharedSystemd
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

// terminate stops and removes the SSH container. For shared containers
// (managed by TestMain), this is a no-op — the actual termination happens
// in stopSharedContainers after all tests complete.
func (sc *sshContainer) terminate(t *testing.T) {
	if sc.shared {
		return // shared containers are terminated in TestMain
	}
	if sc.container != nil {
		ctx := context.Background()
		if err := sc.container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}
}

// resetCaddyState removes caddy from the container so tests that verify
// "caddy is not installed" work correctly with shared containers. This is
// the Ansible-style state reset pattern: instead of recreating the container,
// we clean up the specific state that the test depends on.
func resetCaddyState(t *testing.T, sc *sshContainer) {
	t.Helper()
	node := newTestNode(sc)
	if err := node.Connect(); err != nil {
		t.Fatalf("resetCaddyState: connect failed: %v", err)
	}
	defer node.Close()
	node.RunCommand("systemctl stop caddy 2>/dev/null; apt-get remove -y caddy 2>/dev/null; rm -f /usr/bin/caddy; rm -rf /etc/caddy /var/lib/caddy /var/log/caddy; true")
}
