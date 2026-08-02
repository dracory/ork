package ork_test

// Integration tests for Node.Check(skill) — check mode.
//
// Check() should report whether the skill would make changes, WITHOUT
// actually making them. These tests verify that against a real SSH server.

import (
	"strings"
	"testing"

	"github.com/dracory/ork"
	"github.com/dracory/ork/skills/fs"
	"github.com/dracory/ork/skills/ping"
)

// TestIntegration_Node_Check_PingSkill calls node.Check(ping.NewPing()).
// Ping is read-only, so Check should return Changed=false and no error.
func TestIntegration_Node_Check_PingSkill(t *testing.T) {
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
	results := node.Check(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Check(ping) failed: %v", result.Error)
	}

	// Ping is read-only — Check should report no changes needed
	if result.Changed {
		t.Error("Expected Changed=false for ping (read-only skill), got true")
	}
}

// TestIntegration_Node_Check_FsFileCreate_NotExists calls node.Check() on a
// FileCreate skill for a path that doesn't exist. Check should report
// Changed=true (file needs to be created) but the file must NOT actually
// exist after Check runs.
func TestIntegration_Node_Check_FsFileCreate_NotExists(t *testing.T) {
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

	testPath := "/tmp/ork-integration-check-should-not-exist.txt"

	skill := fs.NewFileCreate().SetArgs(map[string]string{
		fs.ArgPath:    testPath,
		fs.ArgContent: "should not be created",
		fs.ArgMode:    "644",
	})

	results := node.Check(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Check(fs.FileCreate) failed: %v", result.Error)
	}

	// File doesn't exist — Check should report that changes are needed
	if !result.Changed {
		t.Error("Expected Changed=true (file doesn't exist, needs creation), got false")
	}

	// CRITICAL: the file must NOT exist after Check — Check should not make changes
	verifyResults := node.RunCommand("test -f " + testPath + " && echo EXISTS || echo MISSING")
	verifyResult := verifyResults.Results[container.host]
	if verifyResult.Error != nil {
		t.Fatalf("File existence verification failed: %v", verifyResult.Error)
	}
	if !strings.Contains(verifyResult.Message, "MISSING") {
		t.Errorf("Check should not have created the file, but it exists at %s. Output: %s", testPath, verifyResult.Message)
	}
}

// TestIntegration_Node_Check_FsFileCreate_Exists calls node.Check() on a
// FileCreate skill for a path that already exists with matching content.
// Check should report Changed=false (no changes needed).
func TestIntegration_Node_Check_FsFileCreate_Exists(t *testing.T) {
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

	testPath := "/tmp/ork-integration-check-exists.txt"
	testContent := "already here"

	// First, create the file so it exists
	createResults := node.RunCommand("echo -n '" + testContent + "' > " + testPath)
	if createResults.Results[container.host].Error != nil {
		t.Fatalf("Setup: failed to create file: %v", createResults.Results[container.host].Error)
	}
	defer node.RunCommand("rm -f " + testPath)

	// Now Check should report no changes needed (file exists, overwrite not set)
	skill := fs.NewFileCreate().SetArgs(map[string]string{
		fs.ArgPath:    testPath,
		fs.ArgContent: testContent,
		fs.ArgMode:    "644",
	})

	results := node.Check(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Check(fs.FileCreate) failed: %v", result.Error)
	}

	// File exists with matching content and overwrite is not set — no changes needed
	if result.Changed {
		t.Error("Expected Changed=false (file exists, no overwrite), got true")
	}
}
