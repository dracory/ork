package ork_test

// Integration tests for the caddy skill package.
//
// These tests use a systemd-enabled Ubuntu 24.04 container
// (geerlingguy/docker-ubuntu2404-ansible) so that apt, systemctl, and the
// Caddy package all work as they would on a real server.
//
// Running:
//   CI=true go test -v ./test/integration/... -run TestIntegration_Caddy
//   go test -v -short ./test/integration/...    # skips these

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dracory/ork/skills/caddy"
)

// --- Install ---

// TestIntegration_Caddy_Install verifies that caddy.NewInstall() installs
// Caddy via apt on a real Ubuntu container with systemd.
func TestIntegration_Caddy_Install(t *testing.T) {
	container := setupSSHContainerSystemd(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	skill := caddy.NewInstall()
	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Install failed: %v", result.Error)
	}
	if !result.Changed {
		t.Error("Expected Changed=true after installation")
	}

	// Verify caddy binary exists and is executable
	verifyResults := node.RunCommand("caddy version")
	verifyResult := verifyResults.Results[container.host]
	if verifyResult.Error != nil {
		t.Errorf("caddy version failed: %v", verifyResult.Error)
	}
	if !strings.Contains(verifyResult.Message, "v") {
		t.Errorf("Expected caddy version output, got: %q", verifyResult.Message)
	}

	// Verify caddy service is enabled and active
	activeResults := node.RunCommand("systemctl is-active caddy")
	activeResult := activeResults.Results[container.host]
	if strings.TrimSpace(activeResult.Message) != "active" {
		t.Errorf("Expected caddy service to be active, got: %q", strings.TrimSpace(activeResult.Message))
	}

	enabledResults := node.RunCommand("systemctl is-enabled caddy")
	enabledResult := enabledResults.Results[container.host]
	if strings.TrimSpace(enabledResult.Message) != "enabled" {
		t.Errorf("Expected caddy service to be enabled, got: %q", strings.TrimSpace(enabledResult.Message))
	}
}

// TestIntegration_Caddy_Install_Idempotent verifies that running Install
// a second time reports Changed=false (Caddy is already installed).
func TestIntegration_Caddy_Install_Idempotent(t *testing.T) {
	container := setupSSHContainerSystemd(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// First run: install Caddy
	skill := caddy.NewInstall()
	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("First install failed: %v", result.Error)
	}
	if !result.Changed {
		t.Error("Expected Changed=true on first install")
	}

	// Second run: should be idempotent
	results2 := node.Run(skill)
	result2 := results2.Results[container.host]
	if result2.Error != nil {
		t.Fatalf("Second install failed: %v", result2.Error)
	}
	if result2.Changed {
		t.Error("Expected Changed=false on second install (idempotency)")
	}
	expected := "Caddy is already installed"
	if result2.Message != expected {
		t.Errorf("Expected message %q, got %q", expected, result2.Message)
	}
}

// --- Status ---

// TestIntegration_Caddy_Status verifies that caddy.NewStatus() returns the
// systemd status output for the caddy service.
func TestIntegration_Caddy_Status(t *testing.T) {
	container := setupSSHContainerSystemd(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Install Caddy first
	installResults := node.Run(caddy.NewInstall())
	if installResults.Results[container.host].Error != nil {
		t.Fatalf("Install failed: %v", installResults.Results[container.host].Error)
	}

	// Now check status
	skill := caddy.NewStatus()
	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Status failed: %v", result.Error)
	}
	if result.Changed {
		t.Error("Expected Changed=false for read-only status")
	}
	output := result.Details["output"]
	if output == "" {
		t.Error("Expected non-empty status output in Details['output']")
	}
	if !strings.Contains(output, "caddy") {
		t.Errorf("Expected status output to mention 'caddy', got: %q", output)
	}
}

// --- Restart ---

// TestIntegration_Caddy_Restart verifies that caddy.NewRestart() uploads a
// Caddyfile and reloads the Caddy service.
func TestIntegration_Caddy_Restart(t *testing.T) {
	container := setupSSHContainerSystemd(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Install Caddy first
	installResults := node.Run(caddy.NewInstall())
	if installResults.Results[container.host].Error != nil {
		t.Fatalf("Install failed: %v", installResults.Results[container.host].Error)
	}

	// Create a minimal Caddyfile locally
	caddyfileDir := t.TempDir()
	caddyfilePath := filepath.Join(caddyfileDir, "Caddyfile")
	caddyfileContent := `:80 {
	respond "Hello from ork integration test"
}
`
	if err := os.WriteFile(caddyfilePath, []byte(caddyfileContent), 0644); err != nil {
		t.Fatalf("Failed to write test Caddyfile: %v", err)
	}

	// Run Restart
	skill := caddy.NewRestart().SetCaddyfilePath(caddyfilePath)
	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Restart failed: %v", result.Error)
	}
	if !result.Changed {
		t.Error("Expected Changed=true after restart")
	}

	// Verify the Caddyfile was uploaded
	verifyResults := node.RunCommand("cat /etc/caddy/Caddyfile")
	verifyResult := verifyResults.Results[container.host]
	if verifyResult.Error != nil {
		t.Errorf("cat Caddyfile failed: %v", verifyResult.Error)
	}
	if !strings.Contains(verifyResult.Message, "Hello from ork") {
		t.Errorf("Expected uploaded Caddyfile content, got: %q", verifyResult.Message)
	}

	// Verify caddy is still active after reload
	activeResults := node.RunCommand("systemctl is-active caddy")
	activeResult := activeResults.Results[container.host]
	if strings.TrimSpace(activeResult.Message) != "active" {
		t.Errorf("Expected caddy to be active after restart, got: %q", strings.TrimSpace(activeResult.Message))
	}
}

// TestIntegration_Caddy_Restart_MissingCaddyfile verifies that Restart
// returns an error when the local Caddyfile does not exist.
func TestIntegration_Caddy_Restart_MissingCaddyfile(t *testing.T) {
	container := setupSSHContainerSystemd(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	skill := caddy.NewRestart().SetCaddyfilePath("/nonexistent/path/Caddyfile")
	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error == nil {
		t.Error("Expected error when Caddyfile is missing")
	}
	if result.Changed {
		t.Error("Expected Changed=false when Caddyfile is missing")
	}
}

// --- Harden ---

// TestIntegration_Caddy_Harden verifies that caddy.NewHarden() writes a
// systemd drop-in override and restarts Caddy with sandboxing directives.
func TestIntegration_Caddy_Harden(t *testing.T) {
	container := setupSSHContainerSystemd(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Install Caddy first
	installResults := node.Run(caddy.NewInstall())
	if installResults.Results[container.host].Error != nil {
		t.Fatalf("Install failed: %v", installResults.Results[container.host].Error)
	}

	// Run Harden
	skill := caddy.NewHarden()
	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Harden failed: %v", result.Error)
	}
	if !result.Changed {
		t.Error("Expected Changed=true after hardening")
	}

	// Verify the override file exists
	verifyResults := node.RunCommand("cat /etc/systemd/system/caddy.service.d/override.conf")
	verifyResult := verifyResults.Results[container.host]
	if verifyResult.Error != nil {
		t.Errorf("cat override.conf failed: %v", verifyResult.Error)
	}
	overrideContent := verifyResult.Message
	expectedDirectives := []string{
		"ProtectSystem=strict",
		"ProtectHome=true",
		"PrivateTmp=true",
		"NoNewPrivileges=true",
		"CAP_NET_ADMIN",
		"CAP_NET_BIND_SERVICE",
		"ProtectKernelTunables=true",
		"LockPersonality=true",
		"ReadWritePaths=/var/lib/caddy /var/log/caddy",
	}
	for _, directive := range expectedDirectives {
		if !strings.Contains(overrideContent, directive) {
			t.Errorf("Expected override to contain %q, got: %q", directive, overrideContent)
		}
	}

	// Verify caddy is still active after hardening
	activeResults := node.RunCommand("systemctl is-active caddy")
	activeResult := activeResults.Results[container.host]
	if strings.TrimSpace(activeResult.Message) != "active" {
		t.Errorf("Expected caddy to be active after hardening, got: %q", strings.TrimSpace(activeResult.Message))
	}

	// Verify the sandboxing directives are actually applied to the running process
	// by checking systemctl show
	showResults := node.RunCommand("systemctl show caddy --property=ProtectSystem,ProtectHome,PrivateTmp,NoNewPrivileges")
	showResult := showResults.Results[container.host]
	if showResult.Error != nil {
		t.Errorf("systemctl show failed: %v", showResult.Error)
	}
	showOutput := showResult.Message
	if !strings.Contains(showOutput, "ProtectSystem=strict") {
		t.Errorf("Expected ProtectSystem=strict in systemctl show, got: %q", showOutput)
	}
	if !strings.Contains(showOutput, "NoNewPrivileges=yes") {
		t.Errorf("Expected NoNewPrivileges=yes in systemctl show, got: %q", showOutput)
	}
}

// TestIntegration_Caddy_Harden_CustomProtectSystem verifies that
// SetProtectSystem("full") is reflected in the override file.
func TestIntegration_Caddy_Harden_CustomProtectSystem(t *testing.T) {
	container := setupSSHContainerSystemd(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	// Install Caddy first
	installResults := node.Run(caddy.NewInstall())
	if installResults.Results[container.host].Error != nil {
		t.Fatalf("Install failed: %v", installResults.Results[container.host].Error)
	}

	// Run Harden with ProtectSystem=full
	skill := caddy.NewHarden().SetProtectSystem("full")
	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Harden failed: %v", result.Error)
	}

	// Verify the override uses ProtectSystem=full
	verifyResults := node.RunCommand("cat /etc/systemd/system/caddy.service.d/override.conf")
	verifyResult := verifyResults.Results[container.host]
	if !strings.Contains(verifyResult.Message, "ProtectSystem=full") {
		t.Errorf("Expected ProtectSystem=full in override, got: %q", verifyResult.Message)
	}
}

// --- Dry-run ---

// TestIntegration_Caddy_Install_DryRun verifies that dry-run mode does not
// install anything but reports Changed=true.
func TestIntegration_Caddy_Install_DryRun(t *testing.T) {
	container := setupSSHContainerSystemd(t)
	defer container.terminate(t)

	node := newTestNode(container)
	node.SetDryRunMode(true)

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	skill := caddy.NewInstall()
	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Dry-run install failed: %v", result.Error)
	}
	if !result.Changed {
		t.Error("Expected Changed=true in dry-run mode")
	}

	// Verify caddy was NOT actually installed
	verifyResults := node.RunCommand("which caddy 2>/dev/null || echo NOT_INSTALLED")
	verifyResult := verifyResults.Results[container.host]
	if !strings.Contains(verifyResult.Message, "NOT_INSTALLED") {
		t.Error("Expected caddy to NOT be installed in dry-run mode")
	}
}
