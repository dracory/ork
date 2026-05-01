package examples

import (
	"fmt"

	"github.com/dracory/ork"
)

// ExampleCommand demonstrates using CommandInterface
// which implements RunnableInterface to run on nodes and inventories.
func ExampleCommand() {
	// Create a command - use With* methods for consistency
	command := ork.NewCommand().
		WithDescription("Check server uptime").
		WithCommand("uptime").
		WithRequired(true)

	// Run on a single node - use FirstResult() for convenience
	node := ork.NewNodeForHost("server.example.com")
	result := node.Run(command).FirstResult()

	if result.Error != nil {
		fmt.Printf("Command failed: %v\n", result.Error)
		return
	}
	fmt.Printf("Command succeeded: %s\n", result.Message)
	fmt.Printf("Output: %s\n", result.Details["output"])
}

// ExampleCommandInventory demonstrates running commands
// on an inventory (multiple nodes).
func ExampleCommandInventory() {
	// Create an inventory with multiple nodes
	inventory := ork.NewInventory()
	prodGroup := ork.NewGroup("production")

	prodGroup.AddNode(ork.NewNodeForHost("server1.example.com"))
	prodGroup.AddNode(ork.NewNodeForHost("server2.example.com"))
	prodGroup.AddNode(ork.NewNodeForHost("server3.example.com"))

	inventory.AddGroup(prodGroup)

	// Create command to run on all nodes - use With* methods for consistency
	command := ork.NewCommand().
		WithDescription("Restart application").
		WithCommand("pm2 restart app").
		WithRequired(true)

	// Run on all nodes in inventory
	results := inventory.Run(command)

	// Check results
	summary := results.Summary()
	fmt.Printf("Total: %d, Changed: %d, Unchanged: %d, Failed: %d\n",
		summary.Total, summary.Changed, summary.Unchanged, summary.Failed)

	for host, result := range results.Results {
		if result.Error != nil {
			fmt.Printf("Failed on %s: %v\n", host, result.Error)
		} else {
			fmt.Printf("Success on %s: %s\n", host, result.Message)
		}
	}
}

// ExampleCommandNotRequired demonstrates using WithRequired(false)
// to allow execution to continue even if the command fails.
func ExampleCommandNotRequired() {
	node := ork.NewNodeForHost("server.example.com")

	// Create command that's not required to succeed - use With* methods for consistency
	command := ork.NewCommand().
		WithDescription("Non-critical operation").
		WithCommand("some-non-critical-command").
		WithRequired(false)

	// Run on a single node - use FirstResult() for convenience
	result := node.Run(command).FirstResult()

	if result.Error != nil {
		// Error is logged but doesn't fail the operation
		fmt.Printf("Non-required command failed (continuing): %v\n", result.Error)
	}
	fmt.Printf("Command completed: %s\n", result.Message)
}

// ExampleCommandWithArgs demonstrates that commands
// can use args like other runnables.
func ExampleCommandWithArgs() {
	node := ork.NewNodeForHost("server.example.com")
	node.SetArg("app-name", "myapp")
	node.SetArg("port", "3000")

	command := ork.NewCommand().
		WithDescription("Restart application with args").
		WithCommand("pm2 restart ${app-name} --port ${port}").
		WithRequired(true)

	// Note: This example shows arg support, but actual command
	// interpolation would need to be implemented in the command
	result := node.Run(command).FirstResult()

	if result.Error != nil {
		fmt.Printf("Failed: %v\n", result.Error)
	}
}

// ExampleCommandWithBecome demonstrates using WithBecomeUser
// to run commands as a different user (e.g., postgres, www-data).
func ExampleCommandWithBecome() {
	node := ork.NewNodeForHost("server.example.com")

	// Create command to run as postgres user - use With* methods for consistency
	command := ork.NewCommand().
		WithDescription("Backup database as postgres user").
		WithCommand("pg_dump mydb").
		WithRequired(true).
		WithBecomeUser("postgres")

	// Run on a single node - use FirstResult() for convenience
	result := node.Run(command).FirstResult()

	if result.Error != nil {
		fmt.Printf("Failed: %v\n", result.Error)
	} else {
		fmt.Printf("Success: %s\n", result.Message)
	}
}

// ExampleCommandWithChdir demonstrates using WithChdir
// to run commands in a specific working directory.
func ExampleCommandWithChdir() {
	node := ork.NewNodeForHost("server.example.com")

	// Create command to run in specific directory - use With* methods for consistency
	command := ork.NewCommand().
		WithDescription("List files in web directory").
		WithCommand("ls -la").
		WithRequired(true).
		WithChdir("/var/www")

	// Run on a single node - use FirstResult() for convenience
	result := node.Run(command).FirstResult()

	if result.Error != nil {
		fmt.Printf("Failed: %v\n", result.Error)
	} else {
		fmt.Printf("Success: %s\n", result.Message)
		fmt.Printf("Output: %s\n", result.Details["output"])
	}
}
