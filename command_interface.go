package ork

import "github.com/dracory/ork/types"

// CommandInterface defines a fluent interface for executing shell commands.
// It provides a fluent builder pattern for configuring one-off commands
// that can be executed on nodes and inventories via the standard Run() methods.
//
// Unlike skills (which are reusable and idempotent), commands are for
// simple shell command execution.
//
// Fluent chaining note: Call Command-specific methods (SetCommand, SetRequired)
// first, then RunnableInterface methods (SetDescription, SetArg, etc.) to
// maintain the CommandInterface type for chaining.
//
// Example:
//
//	command := ork.NewCommand().
//	    SetCommand("pm2 restart app").
//	    SetRequired(true)
//	command.SetDescription("Restart application")
//
//	// Run on a node
//	node := ork.NewNodeForHost("server.example.com")
//	result := node.Run(command)
//
//	// Run on an inventory
//	inventory := ork.NewInventory()
//	result := inventory.Run(command)
type CommandInterface interface {
	types.RunnableInterface

	// SetCommand sets the shell command to execute.
	SetCommand(cmd string) CommandInterface

	// SetRequired sets whether the command must succeed.
	SetRequired(required bool) CommandInterface

	// WithCommand sets the shell command and returns CommandInterface for chaining.
	// Alternative to SetCommand for consistent named-method fluent chaining.
	WithCommand(cmd string) CommandInterface

	// WithRequired sets whether the command must succeed and returns CommandInterface for chaining.
	// Alternative to SetRequired for consistent named-method fluent chaining.
	WithRequired(required bool) CommandInterface

	// SetChdir sets the working directory for command execution.
	// The command will be executed as `cd <dir> && <command>`.
	// When combined with become, the order is: `cd <dir> && sudo -u <user> <command>`.
	SetChdir(dir string) CommandInterface

	// WithChdir sets the working directory and returns CommandInterface for chaining.
	WithChdir(dir string) CommandInterface

	// WithDescription sets a description and returns CommandInterface for chaining.
	// Use this instead of SetDescription when you need fluent chaining.
	WithDescription(description string) CommandInterface

	// WithID sets the ID and returns CommandInterface for chaining.
	// Use this instead of SetID when you need fluent chaining.
	WithID(id string) CommandInterface

	// WithArg sets a single argument and returns CommandInterface for chaining.
	// Use this instead of SetArg when you need fluent chaining.
	WithArg(key, value string) CommandInterface

	// WithArgs replaces the arguments map and returns CommandInterface for chaining.
	// Use this instead of SetArgs when you need fluent chaining.
	WithArgs(args map[string]string) CommandInterface

	// WithNodeConfig sets the node config and returns CommandInterface for chaining.
	// Use this instead of SetNodeConfig when you need fluent chaining.
	WithNodeConfig(cfg types.NodeConfig) CommandInterface

	// WithDryRun sets dry-run mode and returns CommandInterface for chaining.
	// Use this instead of SetDryRun when you need fluent chaining.
	WithDryRun(dryRun bool) CommandInterface

	// WithTimeout sets the timeout and returns CommandInterface for chaining.
	// Use this instead of SetTimeout when you need fluent chaining.
	WithTimeout(timeout interface{}) CommandInterface

	// WithBecomeUser sets the become user and returns CommandInterface for chaining.
	// Use this instead of SetBecomeUser when you need fluent chaining.
	WithBecomeUser(user string) CommandInterface
}
