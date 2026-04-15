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
