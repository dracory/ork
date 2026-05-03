// Package ssh provides SSH connectivity utilities for remote server automation.
// It uses golang.org/x/crypto/ssh with a simplified API for playbook-style
// operations where you connect, run commands, and disconnect.
package ssh

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Client wraps an SSH connection with convenient methods for running commands.
type Client struct {
	host              string
	port              string
	user              string
	keyPath           string
	kexAlgorithms     []string
	hostKeyAlgorithms []string
	client            *ssh.Client
}

// NewClient creates a new SSH client configuration.
// The host parameter should be just the hostname or IP (e.g., "db3.sinevia.com").
// The port parameter is the SSH port (e.g., "22" or "40022").
// The key parameter is just the filename (e.g., "2024_sinevia.prv"),
// which gets resolved to ~/.ssh/<key>.
func NewClient(host, port, user, key string) *Client {
	if port == "" {
		port = "22" // Default SSH port
	}

	keyPath := PrivateKeyPath(key)

	return &Client{
		host:    host,
		port:    port,
		user:    user,
		keyPath: keyPath,
	}
}

func (c *Client) WithKexAlgorithms(algorithms []string) *Client {
	c.kexAlgorithms = algorithms
	return c
}

func (c *Client) WithHostKeyAlgorithms(algorithms []string) *Client {
	c.hostKeyAlgorithms = algorithms
	return c
}

// Connect establishes the SSH connection.
// Must be called before Run or Close.
// Returns an error if the host is empty.
func (c *Client) Connect() error {
	if c.host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// Check if key file exists and is readable
	_, err := os.Stat(c.keyPath)
	if err != nil {
		return fmt.Errorf("SSH key file error: %w", err)
	}

	// Try to read the file to verify it's accessible
	_, err = os.ReadFile(c.keyPath)
	if err != nil {
		return fmt.Errorf("Cannot read SSH key file: %w", err)
	}

	addr := c.host + ":" + c.port
	client, err := c.connectWithKeyFile(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, classifySSHError(err))
	}
	c.client = client
	return nil
}

func (c *Client) connectWithKeyFile(addr string) (*ssh.Client, error) {
	key, err := os.ReadFile(c.keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            c.user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: knownHostsCallback(),
		Timeout:         30 * time.Second,
	}

	if len(c.kexAlgorithms) > 0 {
		config.Config.KeyExchanges = c.kexAlgorithms
	}

	if len(c.hostKeyAlgorithms) > 0 {
		config.HostKeyAlgorithms = c.hostKeyAlgorithms
	}

	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return nil, err
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}

func knownHostsCallback() ssh.HostKeyCallback {
	usr, err := user.Current()
	if err != nil {
		return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return err
		}
	}

	callback, err := knownhosts.New(filepath.Join(usr.HomeDir, ".ssh", "known_hosts"))
	if err != nil {
		return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return err
		}
	}

	return callback
}

// Run executes a command on the remote server.
// Returns combined stdout/stderr output and any error.
func (c *Client) Run(cmd string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("not connected, call Connect() first")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	return string(output), err
}

// Close closes the SSH connection.
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}

// classifySSHError analyzes SSH connection errors and returns a more specific
// error message based on the error type. This helps distinguish between:
// - Host key verification failures (unknown host, changed host key)
// - Authentication failures (wrong key, passphrase required)
// - Connection failures (network issues, wrong port)
func classifySSHError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Host key verification errors
	if strings.Contains(errStr, "knownhosts: key is unknown") {
		return fmt.Errorf("host key verification failed: %s. The host's SSH key is not in your known_hosts file. To fix this, run: ssh-keyscan HOST >> ~/.ssh/known_hosts", errStr)
	}
	if strings.Contains(errStr, "knownhosts: key mismatch") {
		return fmt.Errorf("host key verification failed: %s. The host's SSH key has changed. This could indicate a man-in-the-middle attack. To fix this, remove the old key with: ssh-keygen -R HOST", errStr)
	}
	if strings.Contains(errStr, "knownhosts: key is revoked") {
		return fmt.Errorf("host key verification failed: %s. The host's SSH key has been revoked. This is a security concern.", errStr)
	}

	// Authentication errors
	if strings.Contains(errStr, "unable to authenticate") || strings.Contains(errStr, "no supported methods remain") {
		return fmt.Errorf("authentication failed: %s. Check your SSH key, user, and that the key is authorized on the server", errStr)
	}
	if strings.Contains(errStr, "permission denied") {
		return fmt.Errorf("permission denied: %s. Check your SSH key and user credentials", errStr)
	}

	// Connection errors
	if strings.Contains(errStr, "no common host key algorithm") {
		return fmt.Errorf("SSH host key algorithm mismatch: %s. The server and client do not share a supported host key algorithm. Check the server's HostKeyAlgorithms setting or configure compatible algorithms in NodeConfig", errStr)
	}
	if strings.Contains(errStr, "no common algorithm") ||
		strings.Contains(errStr, "no common key exchange algorithm") ||
		strings.Contains(errStr, "no common kex") ||
		strings.Contains(errStr, "kex:") ||
		strings.Contains(errStr, "key exchange") {
		return fmt.Errorf("SSH key exchange algorithm mismatch: %s. The server and client do not share a supported KEX algorithm. Check the server's KexAlgorithms setting or configure compatible algorithms in NodeConfig", errStr)
	}
	if strings.Contains(errStr, "connection refused") {
		return fmt.Errorf("connection refused: %s. Check that the host is running and the port is correct", errStr)
	}
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "timed out") {
		return fmt.Errorf("connection timeout: %s. Check network connectivity and firewall settings", errStr)
	}
	if strings.Contains(errStr, "no route to host") || strings.Contains(errStr, "network is unreachable") {
		return fmt.Errorf("network error: %s. Check network connectivity and hostname resolution", errStr)
	}

	// Return original error if no specific pattern matched
	return err
}
