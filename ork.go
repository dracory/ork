// Package ork provides a simple, intuitive API for SSH-based server automation.
//
// The core concept is the Node - a representation of a remote server that you
// can connect to and run commands or playbooks against.
//
// Basic usage:
//
//	node := ork.NewNode("server.example.com")
//	output, err := node.Run("uptime")
//
// With configuration:
//
//	node := ork.NewNode("server.example.com").
//	    SetPort("2222").
//	    SetUser("deploy")
//	output, err := node.Run("uptime")
//
// Persistent connections for multiple operations:
//
//	node := ork.NewNode("server.example.com")
//	if err := node.Connect(); err != nil {
//	    log.Fatal(err)
//	}
//	defer node.Close()
//
//	output1, _ := node.Run("uptime")
//	output2, _ := node.Run("df -h")
//
// Running playbooks:
//
//	node := ork.NewNode("server.example.com").
//	    SetArg("username", "alice")
//	err := node.Playbook("user-create")
//
// For advanced use cases, the internal packages remain accessible:
//   - config - Configuration types
//   - ssh - SSH client
//   - playbook - Playbook interface and registry
//   - playbooks - Built-in playbook implementations
package ork

// This file intentionally minimal. All functionality is in:
//   - node_interface.go - NodeInterface and NewNode
//   - node_implementation.go - Node implementation
//   - options.go - Functional options (to be removed)

// Note: Playbook registration and discovery is handled automatically
// at package init time in node_interface.go
