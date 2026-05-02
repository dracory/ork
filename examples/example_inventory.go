package examples

import (
	"fmt"

	"github.com/dracory/ork"
)

// ExampleInventory demonstrates creating an inventory with multiple groups.
// This shows how to organize nodes into logical groups for bulk operations.
func ExampleInventory() {
	// Create a new inventory
	inventory := ork.NewInventory()

	// Create a web servers group
	webGroup := ork.NewGroup("webservers").
		WithArg("environment", "production").
		WithArg("app-type", "web")

	// Add nodes to the web group
	webGroup.AddNode(
		ork.NewNodeForHost("web1.example.com").
			WithUser("deploy").
			WithKey("/home/user/.ssh/id_rsa"),
	).AddNode(
		ork.NewNodeForHost("web2.example.com").
			WithUser("deploy").
			WithKey("/home/user/.ssh/id_rsa"),
	)

	// Create a database servers group
	dbGroup := ork.NewGroup("databases").
		WithArg("environment", "production").
		WithArg("app-type", "database")

	// Add nodes to the database group
	dbGroup.AddNode(
		ork.NewNodeForHost("db1.example.com").
			WithUser("deploy").
			WithKey("/home/user/.ssh/id_rsa"),
	)

	// Add groups to the inventory
	inventory.AddGroup(webGroup).AddGroup(dbGroup)

	// Run a command on all nodes across all groups
	results := inventory.RunCommand("uptime")

	// Print results
	for nodeID, result := range results.Results {
		if result.Error != nil {
			fmt.Printf("Node %s failed: %v\n", nodeID, result.Error)
		} else {
			fmt.Printf("Node %s: %s\n", nodeID, result.Message)
		}
	}
}

// ExampleInventoryWithGroups demonstrates running operations on specific groups.
func ExampleInventoryWithGroups() {
	// Create inventory with groups
	inventory := ork.NewInventory()

	// Web servers group
	webGroup := ork.NewGroup("webservers").
		WithArg("environment", "production")

	webGroup.AddNode(
		ork.NewNodeForHost("web1.example.com").
			WithUser("deploy").
			WithKey("/home/user/.ssh/id_rsa"),
	).AddNode(
		ork.NewNodeForHost("web2.example.com").
			WithUser("deploy").
			WithKey("/home/user/.ssh/id_rsa"),
	)

	// Database servers group
	dbGroup := ork.NewGroup("databases").
		WithArg("environment", "production")

	dbGroup.AddNode(
		ork.NewNodeForHost("db1.example.com").
			WithUser("deploy").
			WithKey("/home/user/.ssh/id_rsa"),
	)

	inventory.AddGroup(webGroup).AddGroup(dbGroup)

	// Run command only on web servers
	webResults := webGroup.RunCommand("systemctl restart nginx")

	fmt.Printf("Restarted nginx on %d web servers\n", len(webResults.Results))

	// Run command only on database servers
	dbResults := dbGroup.RunCommand("systemctl restart mariadb")

	fmt.Printf("Restarted mariadb on %d database servers\n", len(dbResults.Results))
}

// ExampleInventoryWithConcurrency demonstrates setting concurrency for parallel execution.
func ExampleInventoryWithConcurrency() {
	// Create inventory with concurrency control
	inventory := ork.NewInventory().
		SetMaxConcurrency(5) // Run up to 5 operations in parallel

	// Create multiple groups
	for i := 1; i <= 10; i++ {
		group := ork.NewGroup(fmt.Sprintf("group-%d", i))
		group.AddNode(
			ork.NewNodeForHost(fmt.Sprintf("server%d.example.com", i)).
				WithUser("deploy").
				WithKey("/home/user/.ssh/id_rsa"),
		)
		inventory.AddGroup(group)
	}

	// Run command on all nodes with concurrency limit
	results := inventory.RunCommand("hostname")

	fmt.Printf("Executed on %d nodes with concurrency limit of 5\n", len(results.Results))
}

// ExampleInventoryGroupArgs demonstrates group argument inheritance.
func ExampleInventoryGroupArgs() {
	// Create a group with arguments
	webGroup := ork.NewGroup("webservers").
		WithArg("environment", "production").
		WithArg("app-type", "web").
		WithArg("deploy-user", "deploy")

	// Add nodes - they inherit group arguments
	webGroup.AddNode(
		ork.NewNodeForHost("web1.example.com").
			WithKey("/home/user/.ssh/id_rsa"),
	).AddNode(
		ork.NewNodeForHost("web2.example.com").
			WithKey("/home/user/.ssh/id_rsa"),
	)

	// Retrieve group arguments
	fmt.Printf("Environment: %s\n", webGroup.GetArg("environment"))
	fmt.Printf("App type: %s\n", webGroup.GetArg("app-type"))

	// Get all arguments
	args := webGroup.GetArgs()
	fmt.Printf("All arguments: %v\n", args)
}
