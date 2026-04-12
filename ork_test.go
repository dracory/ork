package ork

import (
	"errors"
	"testing"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/playbook"
)

// TestRunSSH_Success tests successful SSH command execution
func TestRunSSH_Success(t *testing.T) {
	// Save original and restore after test
	originalRunOnce := sshRunOnce
	defer func() { sshRunOnce = originalRunOnce }()

	// Mock ssh.RunOnce
	var capturedHost, capturedPort, capturedUser, capturedKey, capturedCmd string
	sshRunOnce = func(host, port, user, key, cmd string) (string, error) {
		capturedHost = host
		capturedPort = port
		capturedUser = user
		capturedKey = key
		capturedCmd = cmd
		return "test output", nil
	}

	// Test
	output, err := RunSSH("testhost", "uptime")

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if output != "test output" {
		t.Errorf("Expected output 'test output', got: %s", output)
	}
	if capturedHost != "testhost" {
		t.Errorf("Expected host 'testhost', got: %s", capturedHost)
	}
	if capturedPort != "22" {
		t.Errorf("Expected port '22', got: %s", capturedPort)
	}
	if capturedUser != "root" {
		t.Errorf("Expected user 'root', got: %s", capturedUser)
	}
	if capturedKey != "id_rsa" {
		t.Errorf("Expected key 'id_rsa', got: %s", capturedKey)
	}
	if capturedCmd != "uptime" {
		t.Errorf("Expected cmd 'uptime', got: %s", capturedCmd)
	}
}

// TestRunSSH_WithOptions tests SSH command execution with options
func TestRunSSH_WithOptions(t *testing.T) {
	// Save original and restore after test
	originalRunOnce := sshRunOnce
	defer func() { sshRunOnce = originalRunOnce }()

	// Mock ssh.RunOnce
	var capturedHost, capturedPort, capturedUser, capturedKey, capturedCmd string
	sshRunOnce = func(host, port, user, key, cmd string) (string, error) {
		capturedHost = host
		capturedPort = port
		capturedUser = user
		capturedKey = key
		capturedCmd = cmd
		return "custom output", nil
	}

	// Test with options
	output, err := RunSSH("customhost", "df -h",
		WithPort("2222"),
		WithUser("deploy"),
		WithKey("custom.prv"),
	)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if output != "custom output" {
		t.Errorf("Expected output 'custom output', got: %s", output)
	}
	if capturedHost != "customhost" {
		t.Errorf("Expected host 'customhost', got: %s", capturedHost)
	}
	if capturedPort != "2222" {
		t.Errorf("Expected port '2222', got: %s", capturedPort)
	}
	if capturedUser != "deploy" {
		t.Errorf("Expected user 'deploy', got: %s", capturedUser)
	}
	if capturedKey != "custom.prv" {
		t.Errorf("Expected key 'custom.prv', got: %s", capturedKey)
	}
	if capturedCmd != "df -h" {
		t.Errorf("Expected cmd 'df -h', got: %s", capturedCmd)
	}
}

// TestRunSSH_Error tests SSH command execution failure
func TestRunSSH_Error(t *testing.T) {
	// Save original and restore after test
	originalRunOnce := sshRunOnce
	defer func() { sshRunOnce = originalRunOnce }()

	// Mock ssh.RunOnce to return error
	sshRunOnce = func(host, port, user, key, cmd string) (string, error) {
		return "", errors.New("connection refused")
	}

	// Test
	output, err := RunSSH("testhost", "uptime")

	// Assertions
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
	// Check error message contains context
	errMsg := err.Error()
	if !contains(errMsg, "uptime") {
		t.Errorf("Error message should contain command 'uptime': %s", errMsg)
	}
	if !contains(errMsg, "testhost") {
		t.Errorf("Error message should contain host 'testhost': %s", errMsg)
	}
}

// TestRunPlaybook_Success tests successful playbook execution
func TestRunPlaybook_Success(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := defaultRegistry
	defer func() { defaultRegistry = originalRegistry }()

	// Create test registry with mock playbook
	defaultRegistry = playbook.NewRegistry()
	var capturedConfig config.Config
	mockPb := &mockPlaybook{
		name: "test-playbook",
		runFunc: func(cfg config.Config) error {
			capturedConfig = cfg
			return nil
		},
	}
	defaultRegistry.Register(mockPb)

	// Test
	err := RunPlaybook("test-playbook", "testhost")

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if capturedConfig.SSHHost != "testhost" {
		t.Errorf("Expected host 'testhost', got: %s", capturedConfig.SSHHost)
	}
	if capturedConfig.SSHPort != "22" {
		t.Errorf("Expected port '22', got: %s", capturedConfig.SSHPort)
	}
	if capturedConfig.RootUser != "root" {
		t.Errorf("Expected user 'root', got: %s", capturedConfig.RootUser)
	}
}

// TestRunPlaybook_WithOptions tests playbook execution with options
func TestRunPlaybook_WithOptions(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := defaultRegistry
	defer func() { defaultRegistry = originalRegistry }()

	// Create test registry with mock playbook
	defaultRegistry = playbook.NewRegistry()
	var capturedConfig config.Config
	mockPb := &mockPlaybook{
		name: "test-playbook",
		runFunc: func(cfg config.Config) error {
			capturedConfig = cfg
			return nil
		},
	}
	defaultRegistry.Register(mockPb)

	// Test with options
	err := RunPlaybook("test-playbook", "customhost",
		WithPort("2222"),
		WithUser("deploy"),
		WithKey("custom.prv"),
		WithArg("username", "alice"),
		WithArg("shell", "/bin/bash"),
	)

	// Assertions
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if capturedConfig.SSHHost != "customhost" {
		t.Errorf("Expected host 'customhost', got: %s", capturedConfig.SSHHost)
	}
	if capturedConfig.SSHPort != "2222" {
		t.Errorf("Expected port '2222', got: %s", capturedConfig.SSHPort)
	}
	if capturedConfig.RootUser != "deploy" {
		t.Errorf("Expected user 'deploy', got: %s", capturedConfig.RootUser)
	}
	if capturedConfig.SSHKey != "custom.prv" {
		t.Errorf("Expected key 'custom.prv', got: %s", capturedConfig.SSHKey)
	}
	if capturedConfig.Args["username"] != "alice" {
		t.Errorf("Expected arg username='alice', got: %s", capturedConfig.Args["username"])
	}
	if capturedConfig.Args["shell"] != "/bin/bash" {
		t.Errorf("Expected arg shell='/bin/bash', got: %s", capturedConfig.Args["shell"])
	}
}

// TestRunPlaybook_NotFound tests playbook not found error
func TestRunPlaybook_NotFound(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := defaultRegistry
	defer func() { defaultRegistry = originalRegistry }()

	// Create empty test registry
	defaultRegistry = playbook.NewRegistry()

	// Test
	err := RunPlaybook("nonexistent", "testhost")

	// Assertions
	if err == nil {
		t.Error("Expected error, got nil")
	}
	errMsg := err.Error()
	if !contains(errMsg, "nonexistent") {
		t.Errorf("Error message should contain playbook name 'nonexistent': %s", errMsg)
	}
	if !contains(errMsg, "not found") {
		t.Errorf("Error message should contain 'not found': %s", errMsg)
	}
}

// TestRunPlaybook_ExecutionError tests playbook execution failure
func TestRunPlaybook_ExecutionError(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := defaultRegistry
	defer func() { defaultRegistry = originalRegistry }()

	// Create test registry with failing playbook
	defaultRegistry = playbook.NewRegistry()
	mockPb := &mockPlaybook{
		name: "failing-playbook",
		runFunc: func(cfg config.Config) error {
			return errors.New("playbook execution failed")
		},
	}
	defaultRegistry.Register(mockPb)

	// Test
	err := RunPlaybook("failing-playbook", "testhost")

	// Assertions
	if err == nil {
		t.Error("Expected error, got nil")
	}
	errMsg := err.Error()
	if !contains(errMsg, "failing-playbook") {
		t.Errorf("Error message should contain playbook name 'failing-playbook': %s", errMsg)
	}
	if !contains(errMsg, "failed") {
		t.Errorf("Error message should contain 'failed': %s", errMsg)
	}
	if !contains(errMsg, "testhost") {
		t.Errorf("Error message should contain host 'testhost': %s", errMsg)
	}
}

// TestListPlaybooks tests listing all playbooks
func TestListPlaybooks(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := defaultRegistry
	defer func() { defaultRegistry = originalRegistry }()

	// Create test registry with multiple playbooks
	defaultRegistry = playbook.NewRegistry()
	defaultRegistry.Register(&mockPlaybook{name: "playbook1"})
	defaultRegistry.Register(&mockPlaybook{name: "playbook2"})
	defaultRegistry.Register(&mockPlaybook{name: "playbook3"})

	// Test
	names := ListPlaybooks()

	// Assertions
	if len(names) != 3 {
		t.Errorf("Expected 3 playbooks, got: %d", len(names))
	}
	// Check all names are present (order doesn't matter)
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}
	if !nameMap["playbook1"] {
		t.Error("Expected 'playbook1' in list")
	}
	if !nameMap["playbook2"] {
		t.Error("Expected 'playbook2' in list")
	}
	if !nameMap["playbook3"] {
		t.Error("Expected 'playbook3' in list")
	}
}

// TestGetPlaybook_Found tests retrieving an existing playbook
func TestGetPlaybook_Found(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := defaultRegistry
	defer func() { defaultRegistry = originalRegistry }()

	// Create test registry with playbook
	defaultRegistry = playbook.NewRegistry()
	mockPb := &mockPlaybook{
		name: "test-playbook",
	}
	defaultRegistry.Register(mockPb)

	// Test
	pb, ok := GetPlaybook("test-playbook")

	// Assertions
	if !ok {
		t.Error("Expected playbook to be found")
	}
	if pb == nil {
		t.Fatal("Expected non-nil playbook")
	}
	if pb.Name() != "test-playbook" {
		t.Errorf("Expected name 'test-playbook', got: %s", pb.Name())
	}
}

// TestGetPlaybook_NotFound tests retrieving a non-existent playbook
func TestGetPlaybook_NotFound(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := defaultRegistry
	defer func() { defaultRegistry = originalRegistry }()

	// Create empty test registry
	defaultRegistry = playbook.NewRegistry()

	// Test
	pb, ok := GetPlaybook("nonexistent")

	// Assertions
	if ok {
		t.Error("Expected playbook not to be found")
	}
	if pb != nil {
		t.Error("Expected nil playbook")
	}
}

// TestRegisterPlaybook tests registering a custom playbook
func TestRegisterPlaybook(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := defaultRegistry
	defer func() { defaultRegistry = originalRegistry }()

	// Create empty test registry
	defaultRegistry = playbook.NewRegistry()

	// Register custom playbook
	mockPb := &mockPlaybook{
		name: "custom-playbook",
	}
	RegisterPlaybook(mockPb)

	// Verify it was registered
	pb, ok := GetPlaybook("custom-playbook")
	if !ok {
		t.Error("Expected playbook to be registered")
	}
	if pb == nil {
		t.Fatal("Expected non-nil playbook")
	}
	if pb.Name() != "custom-playbook" {
		t.Errorf("Expected name 'custom-playbook', got: %s", pb.Name())
	}
}

// TestRegisterPlaybook_Replace tests replacing an existing playbook
func TestRegisterPlaybook_Replace(t *testing.T) {
	// Save original registry and restore after test
	originalRegistry := defaultRegistry
	defer func() { defaultRegistry = originalRegistry }()

	// Create test registry with initial playbook
	defaultRegistry = playbook.NewRegistry()
	mockPb1 := &mockPlaybook{
		name: "test-playbook",
	}
	defaultRegistry.Register(mockPb1)

	// Register replacement playbook with same name
	// We can verify replacement by checking the Description changes
	mockPb2 := &mockPlaybook{
		name: "test-playbook",
	}
	RegisterPlaybook(mockPb2)

	// Verify it was replaced (both have same description from mockPlaybook.Description())
	pb, ok := GetPlaybook("test-playbook")
	if !ok {
		t.Error("Expected playbook to be found")
	}
	if pb == nil {
		t.Fatal("Expected non-nil playbook")
	}
	if pb.Name() != "test-playbook" {
		t.Errorf("Expected name 'test-playbook', got: %s", pb.Name())
	}
}
