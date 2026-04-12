// Package ork provides a simplified, intuitive API for infrastructure automation.
// It wraps the internal packages (config, ssh, playbook, playbooks) with a clean
// interface that requires only a single import for common operations.
package ork

import (
	"time"

	"github.com/dracory/ork/config"
)

// Option is a function that modifies internal configuration options.
// Options are applied in order, with later options overriding earlier ones.
//
// Example:
//
//	ork.RunSSH("server.example.com", "uptime",
//	    ork.WithPort("2222"),
//	    ork.WithUser("deploy"),
//	    ork.WithKey("production.prv"),
//	)
type Option func(*options)

// options holds internal configuration state.
// This struct is not exported to prevent users from creating invalid configurations.
type options struct {
	port    string
	user    string
	key     string
	args    map[string]string
	dryRun  bool
	timeout time.Duration
}

// defaultOptions returns a new options struct with sensible defaults.
func defaultOptions() *options {
	return &options{
		port:    "22",
		user:    "root",
		key:     "id_rsa",
		timeout: 30 * time.Second,
		args:    make(map[string]string),
	}
}

// WithPort sets the SSH port for the connection.
// Default is "22" if not specified.
//
// Example:
//
//	ork.RunSSH("server.example.com", "uptime", ork.WithPort("2222"))
func WithPort(port string) Option {
	return func(o *options) {
		o.port = port
	}
}

// WithUser sets the SSH user for the connection.
// Default is "root" if not specified.
//
// Example:
//
//	ork.RunSSH("server.example.com", "uptime", ork.WithUser("deploy"))
func WithUser(user string) Option {
	return func(o *options) {
		o.user = user
	}
}

// WithKey sets the SSH private key filename for authentication.
// The key is resolved to ~/.ssh/<keyname>.
// Default is "id_rsa" if not specified.
//
// Example:
//
//	ork.RunSSH("server.example.com", "uptime", ork.WithKey("production.prv"))
func WithKey(key string) Option {
	return func(o *options) {
		o.key = key
	}
}

// WithArg adds a single argument to the arguments map.
// This adds to existing arguments without replacing them.
// Arguments are passed to playbooks for configuration.
//
// Example:
//
//	ork.RunPlaybook("user-create", "server.example.com",
//	    ork.WithArg("username", "alice"),
//	    ork.WithArg("shell", "/bin/bash"))
func WithArg(key, value string) Option {
	return func(o *options) {
		if o.args == nil {
			o.args = make(map[string]string)
		}
		o.args[key] = value
	}
}

// WithArgs merges the provided arguments map with existing arguments.
// If a key exists in both maps, the provided value takes precedence.
// Arguments are passed to playbooks for configuration.
//
// Example:
//
//	args := map[string]string{
//	    "username": "alice",
//	    "shell": "/bin/bash",
//	}
//	ork.RunPlaybook("user-create", "server.example.com", ork.WithArgs(args))
func WithArgs(args map[string]string) Option {
	return func(o *options) {
		if o.args == nil {
			o.args = make(map[string]string)
		}
		for k, v := range args {
			o.args[k] = v
		}
	}
}

// WithDryRun enables or disables dry-run mode.
// In dry-run mode, commands are logged but not executed.
// Default is false (disabled).
//
// Example:
//
//	ork.RunSSH("server.example.com", "rm -rf /data", ork.WithDryRun(true))
func WithDryRun(enabled bool) Option {
	return func(o *options) {
		o.dryRun = enabled
	}
}

// WithTimeout sets the timeout duration for SSH operations.
// Default is 30 seconds if not specified.
//
// Example:
//
//	ork.RunSSH("server.example.com", "long-running-task",
//	    ork.WithTimeout(5 * time.Minute))
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

// applyOptions creates a config.Config by applying functional options to default values.
// It initializes an options struct with defaults, applies all provided options in order
// (last-wins semantics), and converts the result to a config.Config.
//
// The host parameter is required and sets the SSH target host.
// Optional configuration is provided via variadic Option functions.
//
// Example:
//
//	cfg := applyOptions("server.example.com",
//	    WithPort("2222"),
//	    WithUser("deploy"),
//	    WithKey("production.prv"),
//	)
func applyOptions(host string, opts ...Option) config.Config {
	// Initialize with defaults
	o := defaultOptions()

	// Apply all options in order (last-wins semantics)
	for _, opt := range opts {
		opt(o)
	}

	// Convert to config.Config
	return config.Config{
		SSHHost:  host,
		SSHPort:  o.port,
		RootUser: o.user,
		SSHKey:   o.key,
		Args:     o.args,
	}
}
