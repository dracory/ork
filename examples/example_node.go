package examples

import (
	"fmt"

	"github.com/dracory/ork"
)

// ExampleNode demonstrates using nodes with fluent configuration.
// This shows how to configure and use nodes for remote operations.
func ExampleNode() {
	// Create a node with fluent chaining
	node := ork.NewNodeForHost("server.example.com").
		WithPort("2222").
		WithUser("deploy").
		WithKey("/home/user/.ssh/id_rsa").
		WithArg("app-name", "myapp").
		WithArg("environment", "production")

	// Run a command on the node
	command := ork.NewCommand().
		WithDescription("Check server uptime").
		WithCommand("uptime").
		WithRequired(true)

	result := node.Run(command).FirstResult()

	if result.Error != nil {
		fmt.Printf("Command failed: %v\n", result.Error)
	} else {
		fmt.Printf("Command succeeded: %s\n", result.Message)
	}
}

// ExampleNodeWithConfig demonstrates using node configuration.
func ExampleNodeWithConfig() {
	// Create node configuration with fluent chaining
	cfg := ork.NewNodeConfig().
		WithHost("server.example.com").
		WithPort("22").
		WithLogin("ubuntu").
		WithKey("/home/user/.ssh/id_rsa").
		WithArg("app-name", "myapp").
		WithDryRun(true)

	// Create a skill with fluent chaining
	skill := ork.NewSkill().
		WithID("check-config").
		WithDescription("Check server configuration").
		WithNodeConfig(*cfg)

	// In dry-run mode, the skill would log what it would do
	// without actually executing
	fmt.Printf("Skill configured with ID: %s\n", skill.GetID())
	fmt.Printf("Dry-run mode: %v\n", skill.IsDryRun())
}

// ExampleNodeConnect demonstrates persistent SSH connections.
func ExampleNodeConnect() {
	// Create a node
	node := ork.NewNodeForHost("server.example.com").
		SetUser("ubuntu").
		SetKey("/home/user/.ssh/id_rsa")

	// Establish persistent connection
	err := node.Connect()
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer node.Close()

	// Run multiple commands on the same connection
	command1 := ork.NewCommand().
		WithDescription("Check uptime").
		WithCommand("uptime").
		WithRequired(true)

	result1 := node.Run(command1).FirstResult()
	fmt.Printf("Command 1: %s\n", result1.Message)

	command2 := ork.NewCommand().
		WithDescription("Check disk space").
		WithCommand("df -h").
		WithRequired(true)

	result2 := node.Run(command2).FirstResult()
	fmt.Printf("Command 2: %s\n", result2.Message)
}
