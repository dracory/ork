package ork_test

// Integration tests for Node.Run(skill) — direct skill execution.
//
// Unlike RunByID (which looks up a skill in the global registry by string ID),
// Run(skill) takes a RunnableInterface instance directly, clones it, sets the
// node config on the clone, and calls clone.Run(). These tests verify that
// path works against a real SSH server.

import (
	"strings"
	"testing"

	"github.com/dracory/ork"
	"github.com/dracory/ork/skills/fs"
	"github.com/dracory/ork/skills/ping"
)

// TestIntegration_Node_Run_PingSkill constructs a ping.NewPing() instance
// and executes it via node.Run(skill) (not RunByID). Verifies the output
// contains the uptime and the host is reported alive.
func TestIntegration_Node_Run_PingSkill(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	skill := ping.NewPing()
	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Run(ping) failed: %v", result.Error)
	}

	if !strings.Contains(result.Message, "alive") {
		t.Errorf("Expected message to contain 'alive', got: %s", result.Message)
	}

	uptime, ok := result.Details["uptime"]
	if !ok {
		t.Fatal("Expected Details to contain 'uptime' key")
	}
	if uptime == "" {
		t.Error("Expected uptime to be non-empty")
	}
	t.Logf("Uptime: %s", uptime)
}

// TestIntegration_Node_Run_FsFileCreate creates a file on the container via
// node.Run(fs.NewFileCreate()), then verifies the file exists and its content
// matches by running a follow-up command.
func TestIntegration_Node_Run_FsFileCreate(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := ork.NewNodeForHost(container.host).
		SetPort(container.port).
		SetUser(container.user).
		SetKey(container.keyName)

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testPath := "/tmp/ork-integration-test-file.txt"
	testContent := "hello from integration test"

	// Create the file via the fs.FileCreate skill
	skill := fs.NewFileCreate().SetArgs(map[string]string{
		fs.ArgPath:      testPath,
		fs.ArgContent:   testContent,
		fs.ArgOverwrite: "true",
		fs.ArgMode:      "644",
	})

	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Run(fs.FileCreate) failed: %v", result.Error)
	}

	// Verify the file exists on the container
	checkResults := node.RunCommand("test -f " + testPath + " && echo EXISTS")
	checkResult := checkResults.Results[container.host]
	if checkResult.Error != nil {
		t.Fatalf("File existence check failed: %v", checkResult.Error)
	}
	if !strings.Contains(checkResult.Message, "EXISTS") {
		t.Errorf("Expected file to exist at %s, got: %s", testPath, checkResult.Message)
	}

	// Verify the content matches
	contentResults := node.RunCommand("cat " + testPath)
	contentResult := contentResults.Results[container.host]
	if contentResult.Error != nil {
		t.Fatalf("cat failed: %v", contentResult.Error)
	}
	if strings.TrimSpace(contentResult.Message) != testContent {
		t.Errorf("Expected content %q, got %q", testContent, strings.TrimSpace(contentResult.Message))
	}

	// Clean up
	node.RunCommand("rm -f " + testPath)
}
