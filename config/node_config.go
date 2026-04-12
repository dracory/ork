// Package config provides configuration types for SSH-based automation.
package config

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
