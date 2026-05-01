package ork

import "github.com/dracory/ork/types"

// NewNodeConfig creates a new node configuration with default values.
// Use the fluent With* methods for configuration.
//
// Example:
//
//	cfg := ork.NewNodeConfig().
//	    WithHost("server.example.com").
//	    WithPort("22").
//	    WithLogin("ubuntu").
//	    WithKey("/home/user/.ssh/id_rsa").
//	    WithDryRun(true)
func NewNodeConfig() *types.NodeConfig {
	return &types.NodeConfig{
		Args:         make(map[string]string),
		SSHPort:      "22",
		IsDryRunMode: false,
	}
}
