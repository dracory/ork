// Package ork provides a framework for remote server automation.
package ork

// InventoryInterface defines operations for managing a collection of nodes
// organized into groups. It embeds RunnableInterface for executing operations
// across all nodes in the inventory.
type InventoryInterface interface {
	// RunnableInterface is embedded for command and playbook execution.
	// Operations run concurrently across all nodes in the inventory.
	RunnableInterface

	// AddGroup adds a group to the inventory.
	AddGroup(group GroupInterface) InventoryInterface

	// GetGroup retrieves a group by name.
	// Returns nil if the group does not exist.
	GetGroupByName(name string) GroupInterface

	// AddNode adds a node directly to the inventory (not in any specific group).
	// The node can be configured using the returned NodeInterface.
	AddNode(node NodeInterface) InventoryInterface

	// GetNodes returns all nodes in the inventory across all groups.
	GetNodes() []NodeInterface

	// SetMaxConcurrency sets the maximum number of concurrent operations.
	// Default is 1 (sequential execution). Set to 0 for unlimited.
	SetMaxConcurrency(max int) InventoryInterface
}

// GroupInterface defines operations for managing a group of nodes.
// It embeds RunnableInterface for executing operations on the group's nodes.
type GroupInterface interface {
	// RunnableInterface is embedded for command and playbook execution.
	// Operations run on all nodes in this group only.
	RunnableInterface

	// GetName returns the group's name.
	GetName() string

	// AddNode adds a node to this group.
	// The node can be configured using the returned NodeInterface.
	AddNode(node NodeInterface) GroupInterface

	// GetNodes returns all nodes in this group.
	GetNodes() []NodeInterface

	// SetArg sets an argument for this group.
	// Group arguments are inherited by all nodes in the group.
	SetArg(key, value string) GroupInterface

	// GetArg retrieves an argument value by key.
	// Returns empty string if not set.
	GetArg(key string) string

	// GetArgs returns a copy of all arguments defined for this group.
	GetArgs() map[string]string
}

// NewInventory creates a new empty inventory.
func NewInventory() InventoryInterface {
	return &inventoryImplementation{
		groups:         make(map[string]GroupInterface),
		maxConcurrency: 1, // Default to sequential execution for backward compatibility
	}
}
