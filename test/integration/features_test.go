package ork_test

// Integration tests for Phase 2 features: stdin piping, chdir, Required flag,
// and custom SSH algorithms.
//
// These features are set on types.Command or types.NodeConfig and executed
// via ssh.Run (the lower-level API). The Node's fluent API doesn't expose
// SetChdir or Stdin directly — they're config-level settings.

import (
	"strings"
	"testing"

	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// nodeConfigFromContainer builds a NodeConfig pointing to the given container.
// This is used for tests that need to call ssh.Run directly with custom
// Command fields (Stdin, Chdir, Required) that aren't exposed on the Node API.
func nodeConfigFromContainer(container *sshContainer) types.NodeConfig {
	return types.NodeConfig{
		SSHHost:  container.host,
		SSHPort:  container.port,
		RootUser: container.user,
		SSHKey:   container.keyName,
	}
}

// --- 2.1 Stdin piping ---

// TestIntegration_Stdin_Cat verifies that Stdin data is piped to the remote
// command. `cat` reads from stdin and echoes it back.
func TestIntegration_Stdin_Cat(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cmd := types.Command{
		Command: "cat",
		Stdin:   "hello from stdin",
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Fatalf("ssh.Run with Stdin failed: %v", err)
	}

	if strings.TrimSpace(output) != "hello from stdin" {
		t.Errorf("Expected 'hello from stdin', got: %q", output)
	}
}

// TestIntegration_Stdin_Multiline verifies that multiline stdin data is
// preserved when piped to a remote command.
func TestIntegration_Stdin_Multiline(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	input := "line1\nline2\nline3"
	cmd := types.Command{
		Command: "cat",
		Stdin:   input,
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Fatalf("ssh.Run with multiline Stdin failed: %v", err)
	}

	if strings.TrimSpace(output) != input {
		t.Errorf("Expected %q, got: %q", input, strings.TrimSpace(output))
	}
}

// TestIntegration_Stdin_Empty verifies that empty Stdin doesn't break
// command execution.
func TestIntegration_Stdin_Empty(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cmd := types.Command{
		Command: "echo 'no stdin needed'",
		Stdin:   "",
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Fatalf("ssh.Run with empty Stdin failed: %v", err)
	}

	if !strings.Contains(output, "no stdin needed") {
		t.Errorf("Expected 'no stdin needed', got: %q", output)
	}
}

// TestIntegration_Stdin_WithBecomeUser verifies that Stdin works in
// combination with BecomeUser (NOPASSWD path). The command runs as root
// and reads from stdin.
func TestIntegration_Stdin_WithBecomeUser(t *testing.T) {
	container := setupSSHContainerWithSudo(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cfg.BecomeUser = "root"
	cmd := types.Command{
		Command: "cat",
		Stdin:   "piped via sudo",
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Fatalf("ssh.Run with Stdin+BecomeUser failed: %v", err)
	}

	if strings.TrimSpace(output) != "piped via sudo" {
		t.Errorf("Expected 'piped via sudo', got: %q", output)
	}
}

// --- 2.2 Chdir ---

// TestIntegration_Chdir_Pwd verifies that Chdir causes the command to run
// in the specified directory.
func TestIntegration_Chdir_Pwd(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cfg.Chdir = "/tmp"
	cmd := types.Command{
		Command: "pwd",
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Fatalf("ssh.Run with Chdir failed: %v", err)
	}

	if strings.TrimSpace(output) != "/tmp" {
		t.Errorf("Expected '/tmp', got: %q", strings.TrimSpace(output))
	}
}

// TestIntegration_Chdir_CommandLevel verifies that command-level Chdir
// takes precedence over config-level Chdir.
func TestIntegration_Chdir_CommandLevel(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cfg.Chdir = "/tmp"
	cmd := types.Command{
		Command: "pwd",
		Chdir:   "/", // command-level overrides config-level
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Fatalf("ssh.Run with command-level Chdir failed: %v", err)
	}

	if strings.TrimSpace(output) != "/" {
		t.Errorf("Expected '/', got: %q", strings.TrimSpace(output))
	}
}

// TestIntegration_Chdir_WithBecomeUser verifies that Chdir works in
// combination with BecomeUser. The cd wraps outside sudo:
//
//	cd /tmp && sudo -H -n -u root pwd
func TestIntegration_Chdir_WithBecomeUser(t *testing.T) {
	container := setupSSHContainerWithSudo(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cfg.Chdir = "/tmp"
	cfg.BecomeUser = "root"
	cmd := types.Command{
		Command: "pwd",
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Fatalf("ssh.Run with Chdir+BecomeUser failed: %v", err)
	}

	// pwd runs as root, but cd happens before sudo, so the working dir is /tmp
	if strings.TrimSpace(output) != "/tmp" {
		t.Errorf("Expected '/tmp', got: %q", strings.TrimSpace(output))
	}
}

// TestIntegration_Chdir_NonexistentDir verifies that chdir to a non-existent
// directory causes the command to fail.
func TestIntegration_Chdir_NonexistentDir(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cfg.Chdir = "/nonexistent/directory/path"
	cmd := types.Command{
		Command:     "pwd",
		Required:    true,
		Description: "pwd in a non-existent directory",
	}

	_, err := ssh.Run(cfg, cmd)
	if err == nil {
		t.Fatal("Expected error when chdir to non-existent directory, got nil")
	}
}

// --- 2.3 Required=false error suppression ---

// TestIntegration_RequiredFalse_ExitSuppressed verifies that when Required
// is false and the command exits non-zero, the error is suppressed (nil).
func TestIntegration_RequiredFalse_ExitSuppressed(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cmd := types.Command{
		Command:  "exit 1",
		Required: false,
	}

	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Errorf("Expected nil error with Required=false and exit 1, got: %v", err)
	}
	_ = output
}

// TestIntegration_RequiredTrue_ExitPropagated verifies that when Required
// is true and the command exits non-zero, the error is propagated.
func TestIntegration_RequiredTrue_ExitPropagated(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cmd := types.Command{
		Command:  "exit 1",
		Required: true,
	}

	_, err := ssh.Run(cfg, cmd)
	if err == nil {
		t.Fatal("Expected error with Required=true and exit 1, got nil")
	}

	if !ssh.IsExitError(err) {
		t.Errorf("Expected IsExitError to be true, got: %v", err)
	}
}

// TestIntegration_RequiredFalse_ConnectionErrorNotSuppressed verifies that
// connection failures are NOT suppressed even when Required=false. The
// suppression only applies to non-zero exit codes, not connection errors.
func TestIntegration_RequiredFalse_ConnectionErrorNotSuppressed(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cfg.SSHPort = "9999" // dead port — connection failure
	cmd := types.Command{
		Command:  "echo hello",
		Required: false,
	}

	_, err := ssh.Run(cfg, cmd)
	if err == nil {
		t.Fatal("Expected connection error with Required=false, got nil (connection errors should NOT be suppressed)")
	}

	if ssh.IsExitError(err) {
		t.Errorf("Connection error should NOT be an exit error, got: %v", err)
	}
}

// --- 2.4 KexAlgorithms / HostKeyAlgorithms ---

// TestIntegration_CustomKexAlgorithms verifies that setting a subset of
// KEX algorithms still allows connection to succeed.
func TestIntegration_CustomKexAlgorithms(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	cfg.KexAlgorithms = []string{"curve25519-sha256"}

	cmd := types.Command{Command: "echo 'kex test'"}
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Fatalf("ssh.Run with custom KexAlgorithms failed: %v", err)
	}

	if !strings.Contains(output, "kex test") {
		t.Errorf("Expected 'kex test', got: %q", output)
	}
}

// TestIntegration_CustomHostKeyAlgorithms verifies that setting a subset of
// host key algorithms still allows connection to succeed.
//
// Note: The container's host key is captured during setupSSHContainer's probe
// dial, which uses the default algorithm set (negotiates ecdsa typically).
// When we restrict HostKeyAlgorithms, we must include the algorithm that
// was used to capture the known_hosts entry, or known_hosts verification
// will fail with "key mismatch".
func TestIntegration_CustomHostKeyAlgorithms(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	// Use a subset that includes the common types the server offers.
	// The probe dial in setup typically negotiates ecdsa-sha2-nistp256.
	cfg.HostKeyAlgorithms = []string{
		types.HostKeyAlgoECDSA256,
		types.HostKeyAlgoED25519,
	}

	cmd := types.Command{Command: "echo 'hostkey test'"}
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		t.Fatalf("ssh.Run with custom HostKeyAlgorithms failed: %v", err)
	}

	if !strings.Contains(output, "hostkey test") {
		t.Errorf("Expected 'hostkey test', got: %q", output)
	}
}

// TestIntegration_IncompatibleKexAlgorithms verifies that setting an
// incompatible KEX algorithm causes a connection error.
func TestIntegration_IncompatibleKexAlgorithms(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	cfg := nodeConfigFromContainer(container)
	// Use an algorithm that the server definitely doesn't support
	cfg.KexAlgorithms = []string{"diffie-hellman-group1-sha1"}

	cmd := types.Command{Command: "echo 'should never run'"}
	_, err := ssh.Run(cfg, cmd)
	if err == nil {
		t.Fatal("Expected error with incompatible KexAlgorithms, got nil")
	}

	if ssh.IsExitError(err) {
		t.Errorf("Connection error should NOT be an exit error, got: %v", err)
	}
}
