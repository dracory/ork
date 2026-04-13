// Package ssh provides SSH connectivity utilities for remote server automation.
// It wraps github.com/sfreiberg/simplessh with a simplified API for playbook-style
// operations where you connect, run commands, and disconnect.
package ssh

import (
	"fmt"

	"github.com/sfreiberg/simplessh"
)

// Client wraps an SSH connection with convenient methods for running commands.
type Client struct {
	host    string
	port    string
	user    string
	keyPath string
	client  *simplessh.Client
}

// NewClient creates a new SSH client configuration.
// The host parameter should be just the hostname or IP (e.g., "db3.sinevia.com").
// The port parameter is the SSH port (e.g., "22" or "40022").
// The key parameter is just the filename (e.g., "2024_sinevia.prv"),
// which gets resolved to ~/.ssh/<key>.
func NewClient(host, port, user, key string) *Client {
	if port == "" {
		port = "22"
	}
	return &Client{
		host:    host,
		port:    port,
		user:    user,
		keyPath: PrivateKeyPath(key),
	}
}

// Connect establishes the SSH connection.
// Must be called before Run or Close.
// Returns an error if the host is empty.
func (c *Client) Connect() error {
	if c.host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	addr := c.host + ":" + c.port
	client, err := simplessh.ConnectWithKeyFile(addr, c.user, c.keyPath)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	c.client = client
	return nil
}

// Run executes a command on the remote server.
// Returns combined stdout/stderr output and any error.
func (c *Client) Run(cmd string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("not connected, call Connect() first")
	}
	output, err := c.client.Exec(cmd)
	return string(output), err
}

// Close closes the SSH connection.
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}
