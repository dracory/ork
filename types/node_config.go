// Package types provides core types for SSH-based automation.
package types

import "log/slog"

// NodeConfig holds all configuration variables for remote server operations.
type NodeConfig struct {
	// SSH connection settings
	SSHHost  string
	SSHPort  string
	SSHLogin string
	SSHKey   string

	// User settings
	RootUser    string
	NonRootUser string

	// Database settings (when applicable)
	DBPort         string
	DBRootPassword string

	// Extra arguments passed via command line
	Args map[string]string

	// Logger for structured logging. Defaults to slog.Default() if nil.
	Logger *slog.Logger

	// IsDryRunMode indicates whether to simulate execution without making changes.
	// When true, ssh.Run() will log commands and return "[dry-run]" marker instead of executing.
	IsDryRunMode bool

	// BecomeUser is the user to become when executing commands via sudo.
	// If empty, no privilege escalation is performed.
	BecomeUser string

	// Chdir is the working directory for command execution.
	// If set, commands will be executed in this directory.
	Chdir string

	KexAlgorithms []string

	HostKeyAlgorithms []string
}

// SSHAddr returns the full SSH address as host:port.
// Defaults to port 22 if SSHPort is not set.
func (c NodeConfig) SSHAddr() string {
	port := c.SSHPort
	if port == "" {
		port = "22"
	}
	return c.SSHHost + ":" + port
}

// GetArg retrieves an argument from the Args map.
// Returns empty string if not found.
func (c NodeConfig) GetArg(key string) string {
	if c.Args == nil {
		return ""
	}
	return c.Args[key]
}

// GetArgOr retrieves an argument from the Args map with a default value.
func (c NodeConfig) GetArgOr(key, defaultValue string) string {
	if val := c.GetArg(key); val != "" {
		return val
	}
	return defaultValue
}

// GetLoggerOrDefault returns the configured logger or slog.Default() if nil.
func (c NodeConfig) GetLoggerOrDefault() *slog.Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return slog.Default()
}

// SetChdir sets the working directory for command execution.
func (c *NodeConfig) SetChdir(dir string) {
	c.Chdir = dir
}

// WithHost sets the SSH host and returns NodeConfig for chaining.
func (c *NodeConfig) WithHost(host string) *NodeConfig {
	c.SSHHost = host
	return c
}

// WithPort sets the SSH port and returns NodeConfig for chaining.
func (c *NodeConfig) WithPort(port string) *NodeConfig {
	c.SSHPort = port
	return c
}

// WithLogin sets the SSH login user and returns NodeConfig for chaining.
func (c *NodeConfig) WithLogin(login string) *NodeConfig {
	c.SSHLogin = login
	return c
}

// WithKey sets the SSH key path and returns NodeConfig for chaining.
func (c *NodeConfig) WithKey(key string) *NodeConfig {
	c.SSHKey = key
	return c
}

// WithRootUser sets the root user and returns NodeConfig for chaining.
func (c *NodeConfig) WithRootUser(user string) *NodeConfig {
	c.RootUser = user
	return c
}

// WithNonRootUser sets the non-root user and returns NodeConfig for chaining.
func (c *NodeConfig) WithNonRootUser(user string) *NodeConfig {
	c.NonRootUser = user
	return c
}

// WithDBPort sets the database port and returns NodeConfig for chaining.
func (c *NodeConfig) WithDBPort(port string) *NodeConfig {
	c.DBPort = port
	return c
}

// WithDBRootPassword sets the database root password and returns NodeConfig for chaining.
func (c *NodeConfig) WithDBRootPassword(password string) *NodeConfig {
	c.DBRootPassword = password
	return c
}

// WithArg sets a single argument and returns NodeConfig for chaining.
func (c *NodeConfig) WithArg(key, value string) *NodeConfig {
	if c.Args == nil {
		c.Args = make(map[string]string)
	}
	c.Args[key] = value
	return c
}

// WithArgs replaces the arguments map and returns NodeConfig for chaining.
func (c *NodeConfig) WithArgs(args map[string]string) *NodeConfig {
	c.Args = args
	return c
}

// WithLogger sets the logger and returns NodeConfig for chaining.
func (c *NodeConfig) WithLogger(logger *slog.Logger) *NodeConfig {
	c.Logger = logger
	return c
}

// WithDryRun sets dry-run mode and returns NodeConfig for chaining.
func (c *NodeConfig) WithDryRun(dryRun bool) *NodeConfig {
	c.IsDryRunMode = dryRun
	return c
}

// WithBecomeUser sets the become user and returns NodeConfig for chaining.
func (c *NodeConfig) WithBecomeUser(user string) *NodeConfig {
	c.BecomeUser = user
	return c
}

// WithChdir sets the working directory and returns NodeConfig for chaining.
func (c *NodeConfig) WithChdir(dir string) *NodeConfig {
	c.Chdir = dir
	return c
}

func (c *NodeConfig) WithKexAlgorithms(algorithms []string) *NodeConfig {
	c.KexAlgorithms = algorithms
	return c
}

func (c *NodeConfig) WithHostKeyAlgorithms(algorithms []string) *NodeConfig {
	c.HostKeyAlgorithms = algorithms
	return c
}
