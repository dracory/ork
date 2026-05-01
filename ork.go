// Package ork provides a simple, intuitive API for SSH-based server automation.
//
// The core concept is the Node - a representation of a remote server that you
// can connect to and run commands, skills, or playbooks against.
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
// Running skills:
//
//	node := ork.NewNode("server.example.com").
//	    SetArg("username", "alice")
//	err := node.Run(skills.NewUserCreate())
//
// Running commands on nodes/inventories (fluent interface):
//
//	command := ork.NewCommand().
//	    WithDescription("Restart application").
//	    WithCommand("pm2 restart app").
//	    WithRequired(true)
//
//	node := ork.NewNodeForHost("server.example.com")
//	result := node.Run(command)
//
//	inventory := ork.NewInventory()
//	result := inventory.Run(command)
//
// For advanced use cases, the internal packages remain accessible:
//   - ssh - SSH client
//   - types - Shared types (RunnableInterface, Registry, Result, Command)
//   - skills - Built-in skill implementations
package ork

// This file intentionally minimal. All functionality is in:
//   - node_interface.go - NodeInterface and NewNode
//   - node_implementation.go - Node implementation
//   - command_interface.go - CommandInterface for shell commands (Runnable)
//   - command_implementation.go - Command implementation
//   - group_interface.go - GroupInterface and NewGroup
//   - group_implementation.go - Group implementation
//   - inventory_interface.go - InventoryInterface and NewInventory
//   - inventory_implementation.go - Inventory implementation
//   - registry.go - Skill registry

// Note: Skill registration and discovery is handled automatically
// at package init time in node_interface.go
