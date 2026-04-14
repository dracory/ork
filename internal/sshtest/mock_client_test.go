package sshtest

import (
	"errors"
	"testing"
)

func TestNewMockClient(t *testing.T) {
	mock := NewMockClient()
	if mock == nil {
		t.Fatal("NewMockClient returned nil")
	}
	if mock.Commands == nil {
		t.Error("Commands map not initialized")
	}
	if mock.Outputs == nil {
		t.Error("Outputs map not initialized")
	}
	if mock.Errors == nil {
		t.Error("Errors map not initialized")
	}
	if mock.Connected {
		t.Error("Connected should be false initially")
	}
}

func TestMockClient_Connect(t *testing.T) {
	mock := NewMockClient()
	
	err := mock.Connect()
	if err != nil {
		t.Errorf("Connect failed: %v", err)
	}
	if !mock.Connected {
		t.Error("Connected should be true after Connect")
	}
}

func TestMockClient_Connect_AfterClose(t *testing.T) {
	mock := NewMockClient()
	mock.Connect()
	mock.Close()
	
	err := mock.Connect()
	if err == nil {
		t.Error("Connect after close should return error")
	}
}

func TestMockClient_Run(t *testing.T) {
	mock := NewMockClient()
	mock.Connect()
	
	mock.ExpectCommand("echo hello", "hello")
	
	output, err := mock.Run("echo hello")
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
	if output != "hello" {
		t.Errorf("Expected 'hello', got '%s'", output)
	}
}

func TestMockClient_Run_NotConnected(t *testing.T) {
	mock := NewMockClient()
	
	_, err := mock.Run("echo hello")
	if err == nil {
		t.Error("Run should fail when not connected")
	}
}

func TestMockClient_Run_WithExpectedError(t *testing.T) {
	mock := NewMockClient()
	mock.Connect()
	
	expectedErr := errors.New("command failed")
	mock.ExpectError("fail", expectedErr)
	
	_, err := mock.Run("fail")
	if err != expectedErr {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestMockClient_Run_WithoutExpectation(t *testing.T) {
	mock := NewMockClient()
	mock.Connect()
	
	output, err := mock.Run("unknown command")
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
	if output != "" {
		t.Errorf("Expected empty output, got '%s'", output)
	}
}

func TestMockClient_Close(t *testing.T) {
	mock := NewMockClient()
	mock.Connect()
	
	err := mock.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
	if mock.Connected {
		t.Error("Connected should be false after Close")
	}
}

func TestMockClient_Close_Twice(t *testing.T) {
	mock := NewMockClient()
	mock.Connect()
	mock.Close()
	
	err := mock.Close()
	if err != nil {
		t.Errorf("Close should not error when called twice: %v", err)
	}
}

func TestMockClient_AssertCommandRun(t *testing.T) {
	mock := NewMockClient()
	mock.Connect()
	
	mock.Run("cmd1")
	mock.Run("cmd2")
	
	if !mock.AssertCommandRun("cmd1") {
		t.Error("cmd1 should have been run")
	}
	if !mock.AssertCommandRun("cmd2") {
		t.Error("cmd2 should have been run")
	}
	if mock.AssertCommandRun("cmd3") {
		t.Error("cmd3 should not have been run")
	}
}

func TestMockClient_GetCommands(t *testing.T) {
	mock := NewMockClient()
	mock.Connect()
	
	mock.Run("cmd1")
	mock.Run("cmd2")
	
	commands := mock.GetCommands()
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
	if commands[0] != "cmd1" {
		t.Errorf("Expected first command 'cmd1', got '%s'", commands[0])
	}
	if commands[1] != "cmd2" {
		t.Errorf("Expected second command 'cmd2', got '%s'", commands[1])
	}
}

func TestMockClient_GetCommands_Copy(t *testing.T) {
	mock := NewMockClient()
	mock.Connect()
	
	mock.Run("cmd1")
	commands := mock.GetCommands()
	
	// Modify the returned slice
	commands[0] = "modified"
	
	// Get commands again - should not be affected
	commands2 := mock.GetCommands()
	if commands2[0] == "modified" {
		t.Error("GetCommands should return a copy")
	}
}

func TestMockClient_Reset(t *testing.T) {
	mock := NewMockClient()
	mock.Connect()
	mock.ExpectCommand("cmd", "output")
	mock.Run("cmd")
	
	mock.Reset()
	
	if len(mock.Commands) != 0 {
		t.Errorf("Commands should be empty after reset, got %d", len(mock.Commands))
	}
	if len(mock.Outputs) != 0 {
		t.Errorf("Outputs should be empty after reset, got %d", len(mock.Outputs))
	}
	if len(mock.Errors) != 0 {
		t.Errorf("Errors should be empty after reset, got %d", len(mock.Errors))
	}
	if mock.Connected {
		t.Error("Connected should be false after reset")
	}
}

func TestMockClient_ThreadSafety(t *testing.T) {
	mock := NewMockClient()
	mock.Connect()
	
	done := make(chan bool)
	
	// Run multiple goroutines
	for i := 0; i < 10; i++ {
		go func() {
			mock.ExpectCommand("test", "output")
			mock.Run("test")
			mock.AssertCommandRun("test")
			done <- true
		}()
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	commands := mock.GetCommands()
	if len(commands) != 10 {
		t.Errorf("Expected 10 commands, got %d", len(commands))
	}
}
