package ork_test

// Integration tests for skills beyond ping — fs and user management skills.
//
// These tests exercise filesystem skills (FileCreate, FileExists, FileDelete,
// DirCreate, DirExists) and user management skills (UserCreate, UserDelete)
// against a real SSH container.

import (
	"strings"
	"testing"

	"github.com/dracory/ork/skills/fs"
	"github.com/dracory/ork/skills/user"
)

// --- Filesystem skills ---

// TestIntegration_Skill_FileCreate verifies that fs.NewFileCreate() creates
// a file on the container with the specified content.
func TestIntegration_Skill_FileCreate(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testPath := "/tmp/ork-skill-file-create.txt"
	testContent := "created by integration test"

	skill := fs.NewFileCreate().SetArgs(map[string]string{
		fs.ArgPath:      testPath,
		fs.ArgContent:   testContent,
		fs.ArgOverwrite: "true",
		fs.ArgMode:      "644",
	})

	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("FileCreate failed: %v", result.Error)
	}

	// Verify the file exists and has the right content
	verifyResults := node.RunCommand("cat " + testPath)
	verifyResult := verifyResults.Results[container.host]
	if verifyResult.Error != nil {
		t.Fatalf("cat failed: %v", verifyResult.Error)
	}
	if strings.TrimSpace(verifyResult.Message) != testContent {
		t.Errorf("Expected content %q, got %q", testContent, strings.TrimSpace(verifyResult.Message))
	}

	// Clean up
	node.RunCommand("rm -f " + testPath)
}

// TestIntegration_Skill_FileExists_True verifies that fs.NewFileExists()
// correctly reports an existing file.
func TestIntegration_Skill_FileExists_True(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testPath := "/etc/hostname" // always exists on Linux

	skill := fs.NewFileExists().SetArgs(map[string]string{
		fs.ArgPath: testPath,
	})

	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("FileExists failed: %v", result.Error)
	}

	exists, ok := result.Details["exists"]
	if !ok {
		t.Fatal("Expected Details to contain 'exists' key")
	}
	if exists != "true" {
		t.Errorf("Expected exists=true for %s, got: %s", testPath, exists)
	}
}

// TestIntegration_Skill_FileExists_False verifies that fs.NewFileExists()
// correctly reports a non-existent file.
func TestIntegration_Skill_FileExists_False(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testPath := "/tmp/does-not-exist-" + container.port + ".txt"

	skill := fs.NewFileExists().SetArgs(map[string]string{
		fs.ArgPath: testPath,
	})

	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("FileExists failed: %v", result.Error)
	}

	exists, ok := result.Details["exists"]
	if !ok {
		t.Fatal("Expected Details to contain 'exists' key")
	}
	if exists != "false" {
		t.Errorf("Expected exists=false for %s, got: %s", testPath, exists)
	}
}

// TestIntegration_Skill_FileDelete verifies that fs.NewFileDelete() removes
// a file from the container.
func TestIntegration_Skill_FileDelete(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testPath := "/tmp/ork-skill-file-delete.txt"

	// First create the file
	node.RunCommand("echo 'to be deleted' > " + testPath)

	// Verify it exists
	checkResults := node.RunCommand("test -f " + testPath + " && echo EXISTS")
	if !strings.Contains(checkResults.Results[container.host].Message, "EXISTS") {
		t.Fatal("Setup: file should exist before delete")
	}

	// Delete it
	skill := fs.NewFileDelete().SetArgs(map[string]string{
		fs.ArgPath: testPath,
	})

	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("FileDelete failed: %v", result.Error)
	}

	// Verify it's gone
	verifyResults := node.RunCommand("test -f " + testPath + " && echo EXISTS || echo GONE")
	verifyResult := verifyResults.Results[container.host]
	if !strings.Contains(verifyResult.Message, "GONE") {
		t.Errorf("Expected file to be deleted, but it still exists")
	}
}

// TestIntegration_Skill_DirCreate verifies that fs.NewDirCreate() creates
// a directory on the container.
func TestIntegration_Skill_DirCreate(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testPath := "/tmp/ork-skill-dir-create"

	// Clean up any leftover
	node.RunCommand("rm -rf " + testPath)

	skill := fs.NewDirCreate().SetArgs(map[string]string{
		fs.ArgPath: testPath,
		fs.ArgMode: "755",
	})

	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("DirCreate failed: %v", result.Error)
	}

	// Verify the directory exists
	verifyResults := node.RunCommand("test -d " + testPath + " && echo EXISTS")
	verifyResult := verifyResults.Results[container.host]
	if !strings.Contains(verifyResult.Message, "EXISTS") {
		t.Errorf("Expected directory to exist at %s", testPath)
	}

	// Clean up
	node.RunCommand("rm -rf " + testPath)
}

// TestIntegration_Skill_DirExists_True verifies that fs.NewDirExists()
// correctly reports an existing directory.
func TestIntegration_Skill_DirExists_True(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	skill := fs.NewDirExists().SetArgs(map[string]string{
		fs.ArgPath: "/tmp", // always exists
	})

	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("DirExists failed: %v", result.Error)
	}

	exists, ok := result.Details["exists"]
	if !ok {
		t.Fatal("Expected Details to contain 'exists' key")
	}
	if exists != "true" {
		t.Errorf("Expected exists=true for /tmp, got: %s", exists)
	}
}

// TestIntegration_Skill_DirExists_False verifies that fs.NewDirExists()
// correctly reports a non-existent directory.
func TestIntegration_Skill_DirExists_False(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node := newTestNode(container)
	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testPath := "/tmp/does-not-exist-dir-" + container.port

	skill := fs.NewDirExists().SetArgs(map[string]string{
		fs.ArgPath: testPath,
	})

	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("DirExists failed: %v", result.Error)
	}

	exists, ok := result.Details["exists"]
	if !ok {
		t.Fatal("Expected Details to contain 'exists' key")
	}
	if exists != "false" {
		t.Errorf("Expected exists=false for %s, got: %s", testPath, exists)
	}
}

// --- User management skills ---

// TestIntegration_Skill_UserCreate verifies that user.NewUserCreate() creates
// a user on the container. Requires BecomeUser=root with sudo.
func TestIntegration_Skill_UserCreate(t *testing.T) {
	container := setupSSHContainerWithSudo(t)
	defer container.terminate(t)

	node := newTestNode(container)
	node.SetBecomeUser("root")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testUser := "orktestuser"

	// Clean up any leftover user
	node.RunCommand("id " + testUser + " 2>/dev/null && userdel -r " + testUser + " 2>/dev/null; true")

	skill := user.NewUserCreate().SetArgs(map[string]string{
		user.ArgUsername:  testUser,
		user.ArgShell:     "/bin/sh",
		user.ArgSudoGroup: "wheel", // Alpine uses wheel, not sudo
	})

	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("UserCreate failed: %v", result.Error)
	}

	// Verify the user exists
	verifyResults := node.RunCommand("id " + testUser)
	verifyResult := verifyResults.Results[container.host]
	if verifyResult.Error != nil {
		t.Errorf("Expected user %s to exist, but id command failed: %v", testUser, verifyResult.Error)
	}

	// Clean up
	node.RunCommand("userdel -r " + testUser + " 2>/dev/null; true")
}

// TestIntegration_Skill_UserDelete verifies that user.NewUserDelete() removes
// a user from the container. Requires BecomeUser=root with sudo.
func TestIntegration_Skill_UserDelete(t *testing.T) {
	container := setupSSHContainerWithSudo(t)
	defer container.terminate(t)

	node := newTestNode(container)
	node.SetBecomeUser("root")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testUser := "orktestuser2"

	// First create the user
	node.RunCommand("useradd -m " + testUser)

	// Verify it exists
	checkResults := node.RunCommand("id " + testUser)
	if checkResults.Results[container.host].Error != nil {
		t.Fatal("Setup: user should exist before delete")
	}

	// Delete it
	skill := user.NewUserDelete().SetArgs(map[string]string{
		user.ArgUsername: testUser,
	})

	results := node.Run(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("UserDelete failed: %v", result.Error)
	}

	// Verify the user is gone
	verifyResults := node.RunCommand("id " + testUser + " 2>/dev/null && echo EXISTS || echo GONE")
	verifyResult := verifyResults.Results[container.host]
	if !strings.Contains(verifyResult.Message, "GONE") {
		t.Errorf("Expected user %s to be deleted, but it still exists", testUser)
	}
}

// TestIntegration_Skill_UserCreate_Check_NotExists verifies that
// UserCreate.Check() returns true (needs change) when the user doesn't exist.
func TestIntegration_Skill_UserCreate_Check_NotExists(t *testing.T) {
	container := setupSSHContainerWithSudo(t)
	defer container.terminate(t)

	node := newTestNode(container)
	node.SetBecomeUser("root")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testUser := "orkchecknoexist"

	// Clean up any leftover
	node.RunCommand("id " + testUser + " 2>/dev/null && userdel -r " + testUser + " 2>/dev/null; true")

	skill := user.NewUserCreate().SetArgs(map[string]string{
		user.ArgUsername:  testUser,
		user.ArgShell:     "/bin/sh",
		user.ArgSudoGroup: "wheel",
	})

	results := node.Check(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Check failed: %v", result.Error)
	}

	// User doesn't exist → Check should report changes needed
	if !result.Changed {
		t.Error("Expected Changed=true (user doesn't exist, needs creation), got false")
	}
}

// TestIntegration_Skill_UserCreate_Check_Exists verifies that
// UserCreate.Check() returns false (no change needed) when the user exists.
func TestIntegration_Skill_UserCreate_Check_Exists(t *testing.T) {
	container := setupSSHContainerWithSudo(t)
	defer container.terminate(t)

	node := newTestNode(container)
	node.SetBecomeUser("root")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testUser := "orkcheckexists"

	// Create the user first
	node.RunCommand("useradd -m -s /bin/sh " + testUser)
	defer node.RunCommand("userdel -r " + testUser + " 2>/dev/null; true")

	skill := user.NewUserCreate().SetArgs(map[string]string{
		user.ArgUsername:  testUser,
		user.ArgShell:     "/bin/sh",
		user.ArgSudoGroup: "wheel",
	})

	results := node.Check(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Check failed: %v", result.Error)
	}

	// User exists → Check should report no changes needed
	if result.Changed {
		t.Error("Expected Changed=false (user already exists), got true")
	}
}

// TestIntegration_Skill_UserDelete_Check_Exists verifies that
// UserDelete.Check() returns true (needs change) when the user exists.
func TestIntegration_Skill_UserDelete_Check_Exists(t *testing.T) {
	container := setupSSHContainerWithSudo(t)
	defer container.terminate(t)

	node := newTestNode(container)
	node.SetBecomeUser("root")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testUser := "orkdelcheck"

	// Create the user first
	node.RunCommand("useradd -m -s /bin/sh " + testUser)
	defer node.RunCommand("userdel -r " + testUser + " 2>/dev/null; true")

	skill := user.NewUserDelete().SetArgs(map[string]string{
		user.ArgUsername: testUser,
	})

	results := node.Check(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Check failed: %v", result.Error)
	}

	// User exists → Check should report changes needed (deletion)
	if !result.Changed {
		t.Error("Expected Changed=true (user exists, needs deletion), got false")
	}
}

// TestIntegration_Skill_UserDelete_Check_NotExists verifies that
// UserDelete.Check() returns false (no change needed) when the user doesn't exist.
func TestIntegration_Skill_UserDelete_Check_NotExists(t *testing.T) {
	container := setupSSHContainerWithSudo(t)
	defer container.terminate(t)

	node := newTestNode(container)
	node.SetBecomeUser("root")

	err := node.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer node.Close()

	testUser := "orkdelchecknoexist"

	// Ensure user doesn't exist
	node.RunCommand("id " + testUser + " 2>/dev/null && userdel -r " + testUser + " 2>/dev/null; true")

	skill := user.NewUserDelete().SetArgs(map[string]string{
		user.ArgUsername: testUser,
	})

	results := node.Check(skill)
	result := results.Results[container.host]
	if result.Error != nil {
		t.Fatalf("Check failed: %v", result.Error)
	}

	// User doesn't exist → Check should report no changes needed
	if result.Changed {
		t.Error("Expected Changed=false (user doesn't exist, nothing to delete), got true")
	}
}
