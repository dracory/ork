package ork_test

import (
	"fmt"
	"log"

	"github.com/dracory/ork"
)

// ExampleRunSSH demonstrates basic SSH command execution with default settings.
// This is the simplest way to run a command on a remote server.
func ExampleRunSSH() {
	output, err := ork.RunSSH("server.example.com", "uptime")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output)
}

// ExampleRunSSH_withOptions demonstrates SSH command execution with custom configuration.
// Functional options allow flexible configuration without verbose structs.
func ExampleRunSSH_withOptions() {
	output, err := ork.RunSSH("server.example.com", "uptime",
		ork.WithPort("2222"),
		ork.WithUser("deploy"),
		ork.WithKey("production.prv"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output)
}

// ExampleRunPlaybook demonstrates basic playbook execution with default settings.
// Playbooks are pre-registered automation tasks that can be executed by name.
func ExampleRunPlaybook() {
	err := ork.RunPlaybook("apt-update", "server.example.com")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Package database updated successfully")
}

// ExampleRunPlaybook_withOptions demonstrates playbook execution with custom configuration and arguments.
// Arguments are passed to the playbook for configuration (e.g., username for user-create).
func ExampleRunPlaybook_withOptions() {
	err := ork.RunPlaybook("user-create", "server.example.com",
		ork.WithPort("2222"),
		ork.WithUser("deploy"),
		ork.WithArg("username", "alice"),
		ork.WithArg("shell", "/bin/bash"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("User created successfully")
}

// ExampleNewNode demonstrates the fluent Node API with method chaining.
// The Node API provides a builder pattern for readable, chainable configuration.
func ExampleNewNode() {
	node := ork.NewNode("server.example.com").
		SetPort("2222").
		SetUser("deploy").
		SetKey("production.prv")

	output, err := node.Run("uptime")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output)
}

// ExampleNewNode_persistentConnection demonstrates using persistent SSH connections.
// Persistent connections are more efficient when executing multiple commands on the same server.
func ExampleNewNode_persistentConnection() {
	node := ork.NewNode("server.example.com").
		SetPort("2222").
		SetUser("deploy").
		SetKey("production.prv")

	// Establish persistent connection
	if err := node.Connect(); err != nil {
		log.Fatal(err)
	}
	defer node.Close()

	// Multiple operations reuse the same connection
	output1, err := node.Run("uptime")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output1)

	output2, err := node.Run("df -h")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output2)

	// Execute playbook using the same connection
	if err := node.Playbook("apt-status"); err != nil {
		log.Fatal(err)
	}
}

// ExampleListPlaybooks demonstrates discovering available playbooks.
// This is useful for understanding what automation tasks are available.
func ExampleListPlaybooks() {
	names := ork.ListPlaybooks()
	fmt.Println("Available playbooks:")
	for _, name := range names {
		fmt.Printf("  - %s\n", name)
	}
}

// ExampleRegisterPlaybook demonstrates registering a custom playbook.
// Users can extend Ork with custom automation tasks by implementing the Playbook interface.
func ExampleRegisterPlaybook() {
	// Note: This example shows the API usage. In practice, you would implement
	// a real playbook using playbook.NewSimplePlaybook or by implementing the
	// playbook.Playbook interface.
	//
	// customPlaybook := playbook.NewSimplePlaybook(
	//     "custom-task",
	//     "Performs a custom automation task",
	//     func(cfg config.Config) error {
	//         // Custom playbook logic here
	//         return nil
	//     },
	// )
	// ork.RegisterPlaybook(customPlaybook)

	// Now the custom playbook can be used like any built-in playbook
	// err := ork.RunPlaybook("custom-task", "server.example.com")
	// if err != nil {
	//     log.Fatal(err)
	// }

	fmt.Println("Custom playbook registered and ready to use")
}

// ExampleNodeInterface_getters demonstrates using getter methods to inspect configuration.
// Getters are useful for debugging and integration with existing code.
func ExampleNodeInterface_getters() {
	node := ork.NewNode("server.example.com").
		SetPort("2222").
		SetUser("deploy").
		SetKey("production.prv")

	fmt.Printf("Host: %s\n", node.GetHost())
	fmt.Printf("Port: %s\n", node.GetPort())
	fmt.Printf("User: %s\n", node.GetUser())
	fmt.Printf("Key: %s\n", node.GetKey())

	// Get the full configuration for integration with internal packages
	cfg := node.GetConfig()
	fmt.Printf("Full address: %s\n", cfg.SSHAddr())
}

// ExampleNodeInterface_playbook demonstrates executing playbooks with the Node API.
// The Node API allows setting arguments that are passed to playbooks.
func ExampleNodeInterface_playbook() {
	node := ork.NewNode("server.example.com").
		SetArg("username", "alice").
		SetArg("shell", "/bin/bash")

	if err := node.Playbook("user-create"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("User created successfully")
}
