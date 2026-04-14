package playbooktest

import (
	"errors"
	"testing"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/types"
)

func TestNew(t *testing.T) {
	test := New(t)
	if test == nil {
		t.Fatal("New returned nil")
	}
	if test.t != t {
		t.Error("Test t not set correctly")
	}
	if test.mockClient == nil {
		t.Error("MockClient not initialized")
	}
	if test.config.SSHHost != "test.example.com" {
		t.Errorf("Expected SSHHost 'test.example.com', got '%s'", test.config.SSHHost)
	}
}

func TestPlaybookTest_Setup(t *testing.T) {
	test := New(t)
	test.Setup()

	if !test.mockClient.Connected {
		t.Error("MockClient should be connected after Setup")
	}
	test.Cleanup()
}

func TestPlaybookTest_Setup_Cleanup(t *testing.T) {
	test := New(t)
	test.Setup()
	test.Cleanup()

	// After cleanup, a new test should work fine
	test2 := New(t)
	test2.Setup()
	if !test2.mockClient.Connected {
		t.Error("New test should work after cleanup")
	}
	test2.Cleanup()
}

func TestPlaybookTest_ExpectCommand(t *testing.T) {
	test := New(t)

	test.ExpectCommand("cmd1", "output1")
	test.ExpectCommand("cmd2", "output2")

	if test.mockClient.Outputs["cmd1"] != "output1" {
		t.Error("cmd1 expectation not set")
	}
	if test.mockClient.Outputs["cmd2"] != "output2" {
		t.Error("cmd2 expectation not set")
	}
}

func TestPlaybookTest_ExpectError(t *testing.T) {
	test := New(t)
	err := errors.New("test error")

	test.ExpectError("cmd1", err)

	if test.mockClient.Errors["cmd1"] != err {
		t.Error("cmd1 error expectation not set")
	}
}

func TestPlaybookTest_Config(t *testing.T) {
	test := New(t)
	cfg := test.Config()

	if cfg.SSHHost != "test.example.com" {
		t.Errorf("Expected SSHHost 'test.example.com', got '%s'", cfg.SSHHost)
	}
}

func TestPlaybookTest_SetConfig(t *testing.T) {
	test := New(t)
	newCfg := config.NodeConfig{
		SSHHost: "new.example.com",
	}

	test.SetConfig(newCfg)

	if test.config.SSHHost != "new.example.com" {
		t.Errorf("Expected SSHHost 'new.example.com', got '%s'", test.config.SSHHost)
	}
}

func TestPlaybookTest_SetArg(t *testing.T) {
	test := New(t)

	test.SetArg("key1", "value1")

	if test.config.Args["key1"] != "value1" {
		t.Error("Arg not set correctly")
	}
}

func TestPlaybookTest_SetArgs(t *testing.T) {
	test := New(t)
	args := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	test.SetArgs(args)

	if test.config.Args["key1"] != "value1" {
		t.Error("Args not set correctly")
	}
	if test.config.Args["key2"] != "value2" {
		t.Error("Args not set correctly")
	}
}

func TestPlaybookTest_MockClient(t *testing.T) {
	test := New(t)

	mock := test.MockClient()
	if mock == nil {
		t.Error("MockClient returned nil")
	}
	if mock != test.mockClient {
		t.Error("MockClient should return the internal mock client")
	}
}

func TestPlaybookTest_AssertCommandRun(t *testing.T) {
	test := New(t)
	test.Setup()
	defer test.Cleanup()

	test.ExpectCommand("cmd1", "output")
	test.mockClient.Run("cmd1")

	// This should not panic
	test.AssertCommandRun("cmd1")
}

func TestPlaybookTest_AssertCommandNotRun(t *testing.T) {
	test := New(t)
	test.Setup()
	defer test.Cleanup()

	test.ExpectCommand("cmd1", "output")

	// Should not panic since cmd1 was not run
	test.AssertCommandNotRun("cmd1")
}

func TestPlaybookTest_AssertNoError(t *testing.T) {
	test := New(t)

	test.AssertNoError(nil)
}

func TestPlaybookTest_AssertError(t *testing.T) {
	test := New(t)
	err := errors.New("test error")

	test.AssertError(err)
}

func TestPlaybookTest_AssertErrorContains(t *testing.T) {
	test := New(t)
	err := errors.New("test error message")

	test.AssertErrorContains(err, "error")
}

func TestPlaybookTest_AssertResultChanged(t *testing.T) {
	test := New(t)
	result := types.Result{Changed: true}

	test.AssertResultChanged(result)
}

func TestPlaybookTest_AssertResultUnchanged(t *testing.T) {
	test := New(t)
	result := types.Result{Changed: false}

	test.AssertResultUnchanged(result)
}

func TestPlaybookTest_AssertResultNoError(t *testing.T) {
	test := New(t)
	result := types.Result{Error: nil}

	test.AssertResultNoError(result)
}

func TestPlaybookTest_AssertResultError(t *testing.T) {
	test := New(t)
	result := types.Result{Error: errors.New("test")}

	test.AssertResultError(result)
}

func TestPlaybookTest_AssertResultMessageContains(t *testing.T) {
	test := New(t)
	result := types.Result{Message: "test message"}

	test.AssertResultMessageContains(result, "message")
}

func TestPlaybookTest_GetCommands(t *testing.T) {
	test := New(t)
	test.Setup()
	defer test.Cleanup()

	test.ExpectCommand("cmd1", "output")
	test.mockClient.Run("cmd1")

	commands := test.GetCommands()
	if len(commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(commands))
	}
	if commands[0] != "cmd1" {
		t.Errorf("Expected 'cmd1', got '%s'", commands[0])
	}
}

func TestPlaybookTest_Reset(t *testing.T) {
	test := New(t)
	test.Setup()
	defer test.Cleanup()

	test.ExpectCommand("cmd1", "output")
	test.mockClient.Run("cmd1")

	test.Reset()

	if len(test.mockClient.Commands) != 0 {
		t.Errorf("Commands should be empty after reset")
	}
}
