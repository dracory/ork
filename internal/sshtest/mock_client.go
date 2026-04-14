package sshtest

import (
	"fmt"
	"sync"
)

// MockClient is a mock SSH client for testing without real SSH connections.
type MockClient struct {
	mu        sync.Mutex
	Commands  []string
	Outputs   map[string]string
	Errors    map[string]error
	Connected bool
	closed    bool
}

// NewMockClient creates a new mock SSH client.
func NewMockClient() *MockClient {
	return &MockClient{
		Commands: make([]string, 0),
		Outputs:  make(map[string]string),
		Errors:   make(map[string]error),
	}
}

// Connect simulates connecting to an SSH server.
func (m *MockClient) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return fmt.Errorf("client was closed")
	}
	m.Connected = true
	return nil
}

// Run executes a command and returns the predefined output or error.
func (m *MockClient) Run(cmd string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.Connected {
		return "", fmt.Errorf("not connected, call Connect() first")
	}
	
	m.Commands = append(m.Commands, cmd)
	
	// Check if there's a predefined error for this command
	if err, ok := m.Errors[cmd]; ok {
		return "", err
	}
	
	// Return predefined output or empty string
	if output, ok := m.Outputs[cmd]; ok {
		return output, nil
	}
	
	return "", nil
}

// Close simulates closing the SSH connection.
func (m *MockClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Connected = false
	m.closed = true
	return nil
}

// ExpectCommand sets the expected output for a command.
func (m *MockClient) ExpectCommand(cmd, output string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Outputs[cmd] = output
}

// ExpectError sets the expected error for a command.
func (m *MockClient) ExpectError(cmd string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors[cmd] = err
}

// AssertCommandRun verifies that a command was executed.
func (m *MockClient) AssertCommandRun(cmd string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.Commands {
		if c == cmd {
			return true
		}
	}
	return false
}

// GetCommands returns a copy of all executed commands.
func (m *MockClient) GetCommands() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	commands := make([]string, len(m.Commands))
	copy(commands, m.Commands)
	return commands
}

// Reset clears all recorded commands and expectations.
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Commands = make([]string, 0)
	m.Outputs = make(map[string]string)
	m.Errors = make(map[string]error)
	m.Connected = false
	m.closed = false
}
