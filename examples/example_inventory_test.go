package examples

import (
	"testing"

	"github.com/dracory/ork"
)

func TestExampleInventory(t *testing.T) {
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

	// Verify groups were added
	if inventory.GetGroupByName("webservers") == nil {
		t.Error("Web servers group was not added to inventory")
	}

	if inventory.GetGroupByName("databases") == nil {
		t.Error("Databases group was not added to inventory")
	}

	// Verify total node count
	nodes := inventory.GetNodes()
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes in inventory, got %d", len(nodes))
	}

	// Verify group arguments
	if webGroup.GetArg("environment") != "production" {
		t.Errorf("Expected environment to be 'production', got %s", webGroup.GetArg("environment"))
	}

	if dbGroup.GetArg("app-type") != "database" {
		t.Errorf("Expected app-type to be 'database', got %s", dbGroup.GetArg("app-type"))
	}
}

func TestExampleInventoryWithGroups(t *testing.T) {
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

	// Verify group node counts
	webNodes := webGroup.GetNodes()
	if len(webNodes) != 2 {
		t.Errorf("Expected 2 nodes in web group, got %d", len(webNodes))
	}

	dbNodes := dbGroup.GetNodes()
	if len(dbNodes) != 1 {
		t.Errorf("Expected 1 node in database group, got %d", len(dbNodes))
	}

	// Verify group names
	if webGroup.GetName() != "webservers" {
		t.Errorf("Expected group name 'webservers', got %s", webGroup.GetName())
	}

	if dbGroup.GetName() != "databases" {
		t.Errorf("Expected group name 'databases', got %s", dbGroup.GetName())
	}
}

func TestExampleInventoryWithConcurrency(t *testing.T) {
	// Create inventory with concurrency control
	inventory := ork.NewInventory().
		SetMaxConcurrency(5)

	// Create multiple groups
	for i := 1; i <= 3; i++ {
		group := ork.NewGroup("group-" + string(rune('0'+i)))
		group.AddNode(
			ork.NewNodeForHost("server.example.com").
				WithUser("deploy").
				WithKey("/home/user/.ssh/id_rsa"),
		)
		inventory.AddGroup(group)
	}

	// Verify inventory was created
	if inventory == nil {
		t.Fatal("Inventory was not created")
	}

	// Verify groups were added
	if inventory.GetGroupByName("group-1") == nil {
		t.Error("Group 1 was not added")
	}
}

func TestExampleInventoryGroupArgs(t *testing.T) {
	// Create a group with arguments
	webGroup := ork.NewGroup("webservers").
		WithArg("environment", "production").
		WithArg("app-type", "web").
		WithArg("deploy-user", "deploy")

	// Verify arguments were set
	if webGroup.GetArg("environment") != "production" {
		t.Errorf("Expected environment to be 'production', got %s", webGroup.GetArg("environment"))
	}

	if webGroup.GetArg("app-type") != "web" {
		t.Errorf("Expected app-type to be 'web', got %s", webGroup.GetArg("app-type"))
	}

	if webGroup.GetArg("deploy-user") != "deploy" {
		t.Errorf("Expected deploy-user to be 'deploy', got %s", webGroup.GetArg("deploy-user"))
	}

	// Get all arguments
	args := webGroup.GetArgs()
	if len(args) != 3 {
		t.Errorf("Expected 3 arguments, got %d", len(args))
	}

	// Verify specific arguments in map
	if args["environment"] != "production" {
		t.Errorf("Expected environment in args map to be 'production', got %s", args["environment"])
	}
}
