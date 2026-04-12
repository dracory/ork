// Package config provides configuration types for SSH-based automation.
package config

// Config holds all configuration variables for remote server operations.
type Config struct {
	// SSH connection settings
	SSHHost    string
	SSHPort    string
	SSHLogin   string
	SSHKey     string

	// User settings
	RootUser    string
	NonRootUser string

	// Database settings (when applicable)
	DBPort         string
	DBRootPassword string

	// Extra arguments passed via command line
	Args map[string]string
}

// SSHAddr returns the full SSH address as host:port.
// Defaults to port 22 if SSHPort is not set.
func (c Config) SSHAddr() string {
	port := c.SSHPort
	if port == "" {
		port = "22"
	}
	return c.SSHHost + ":" + port
}

// GetArg retrieves an argument from the Args map.
// Returns empty string if not found.
func (c Config) GetArg(key string) string {
	if c.Args == nil {
		return ""
	}
	return c.Args[key]
}

// GetArgOr retrieves an argument from the Args map with a default value.
func (c Config) GetArgOr(key, defaultValue string) string {
	if val := c.GetArg(key); val != "" {
		return val
	}
	return defaultValue
}
