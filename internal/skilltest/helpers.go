package skilltest

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/internal/sshtest"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// SkillTest provides test helpers for skill testing with mock SSH.
type SkillTest struct {
	t          *testing.T
	mockClient *sshtest.MockClient
	config     config.NodeConfig
}

// New creates a new SkillTest instance with default configuration.
func New(t *testing.T) *SkillTest {
	return &SkillTest{
		t:          t,
		mockClient: sshtest.NewMockClient(),
		config: config.NodeConfig{
			SSHHost:  "test.example.com",
			SSHPort:  "22",
			SSHLogin: "testuser",
			SSHKey:   "test_key",
			Args:     make(map[string]string),
			Logger:   slog.Default(),
		},
	}
}

// Setup configures the SSH override to use the mock client.
// Call this before running any skill that uses SSH.
// Returns the SkillTest for chaining.
func (pt *SkillTest) Setup() *SkillTest {
	// Connect the mock client
	pt.mockClient.Connect()

	// Set the SSH override function
	ssh.SetRunFunc(func(cfg config.NodeConfig, cmd types.Command) (string, error) {
		return pt.mockClient.Run(cmd.Command)
	})

	return pt
}

// Cleanup restores the default SSH behavior.
// Should be called in a defer after Setup.
func (pt *SkillTest) Cleanup() {
	ssh.SetRunFunc(nil)
	pt.mockClient.Close()
}

// ExpectCommand sets the expected output for a command.
func (pt *SkillTest) ExpectCommand(cmd, output string) *SkillTest {
	pt.mockClient.ExpectCommand(cmd, output)
	return pt
}

// ExpectError sets the expected error for a command.
func (pt *SkillTest) ExpectError(cmd string, err error) *SkillTest {
	pt.mockClient.ExpectError(cmd, err)
	return pt
}

// Config returns the test configuration.
func (pt *SkillTest) Config() config.NodeConfig {
	return pt.config
}

// SetConfig sets a custom configuration for the test.
func (pt *SkillTest) SetConfig(cfg config.NodeConfig) *SkillTest {
	pt.config = cfg
	return pt
}

// SetArg sets a single argument in the configuration.
func (pt *SkillTest) SetArg(key, value string) *SkillTest {
	if pt.config.Args == nil {
		pt.config.Args = make(map[string]string)
	}
	pt.config.Args[key] = value
	return pt
}

// SetArgs replaces the entire arguments map.
func (pt *SkillTest) SetArgs(args map[string]string) *SkillTest {
	pt.config.Args = args
	return pt
}

// MockClient returns the mock SSH client for direct manipulation.
func (pt *SkillTest) MockClient() *sshtest.MockClient {
	return pt.mockClient
}

// AssertCommandRun verifies that a command was executed.
func (pt *SkillTest) AssertCommandRun(cmd string) {
	if !pt.mockClient.AssertCommandRun(cmd) {
		pt.t.Errorf("Expected command '%s' to be run, but it was not", cmd)
	}
}

// AssertCommandNotRun verifies that a command was NOT executed.
func (pt *SkillTest) AssertCommandNotRun(cmd string) {
	if pt.mockClient.AssertCommandRun(cmd) {
		pt.t.Errorf("Expected command '%s' NOT to be run, but it was", cmd)
	}
}

// AssertNoError verifies that the error is nil.
func (pt *SkillTest) AssertNoError(err error) {
	if err != nil {
		pt.t.Errorf("Expected no error, got: %v", err)
	}
}

// AssertError verifies that the error is non-nil.
func (pt *SkillTest) AssertError(err error) {
	if err == nil {
		pt.t.Error("Expected error, got nil")
	}
}

// AssertErrorContains verifies that the error message contains the expected text.
func (pt *SkillTest) AssertErrorContains(err error, expected string) {
	if err == nil {
		pt.t.Errorf("Expected error containing '%s', got nil", expected)
		return
	}
	if !contains(err.Error(), expected) {
		pt.t.Errorf("Expected error to contain '%s', got: %v", expected, err)
	}
}

// AssertResultChanged verifies that the result indicates changes were made.
func (pt *SkillTest) AssertResultChanged(result types.Result) {
	if !result.Changed {
		pt.t.Error("Expected result.Changed to be true")
	}
}

// AssertResultUnchanged verifies that the result indicates no changes were made.
func (pt *SkillTest) AssertResultUnchanged(result types.Result) {
	if result.Changed {
		pt.t.Error("Expected result.Changed to be false")
	}
}

// AssertResultNoError verifies that the result has no error.
func (pt *SkillTest) AssertResultNoError(result types.Result) {
	if result.Error != nil {
		pt.t.Errorf("Expected result to have no error, got: %v", result.Error)
	}
}

// AssertResultError verifies that the result has an error.
func (pt *SkillTest) AssertResultError(result types.Result) {
	if result.Error == nil {
		pt.t.Error("Expected result to have an error, got nil")
	}
}

// AssertResultMessageContains verifies that the result message contains the expected text.
func (pt *SkillTest) AssertResultMessageContains(result types.Result, expected string) {
	if !contains(result.Message, expected) {
		pt.t.Errorf("Expected result message to contain '%s', got: %s", expected, result.Message)
	}
}

// GetCommands returns a copy of all executed commands.
func (pt *SkillTest) GetCommands() []string {
	return pt.mockClient.GetCommands()
}

// Reset clears all recorded commands and expectations.
func (pt *SkillTest) Reset() {
	pt.mockClient.Reset()
}

// contains is a helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

// findSubstring is a simple substring search.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
