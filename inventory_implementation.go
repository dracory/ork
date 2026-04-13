package ork

import (
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/types"
)

// inventoryImplementation is the default implementation of InventoryInterface.
type inventoryImplementation struct {
	groups         map[string]GroupInterface
	nodes          []NodeInterface
	maxConcurrency int
}

// AddGroup adds a group to the inventory.
func (i *inventoryImplementation) AddGroup(group GroupInterface) InventoryInterface {
	// Use reflection or type assertion to get group name
	// For now, store by a unique identifier
	i.groups[group.GetName()] = group
	return i
}

// GetGroupByName retrieves a group by name.
func (i *inventoryImplementation) GetGroupByName(name string) GroupInterface {
	return i.groups[name]
}

// AddNode adds a node directly to the inventory.
func (i *inventoryImplementation) AddNode(node NodeInterface) InventoryInterface {
	i.nodes = append(i.nodes, node)
	return i
}

// GetNodes returns all nodes in the inventory across all groups.
func (i *inventoryImplementation) GetNodes() []NodeInterface {
	result := make([]NodeInterface, 0, len(i.nodes))
	result = append(result, i.nodes...)

	// Also include nodes from groups
	for _, group := range i.groups {
		result = append(result, group.GetNodes()...)
	}
	return result
}

// SetMaxConcurrency sets the maximum number of concurrent operations.
func (i *inventoryImplementation) SetMaxConcurrency(max int) InventoryInterface {
	i.maxConcurrency = max
	return i
}

// RunCommand executes a shell command across all nodes.
func (i *inventoryImplementation) RunCommand(cmd string) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	nodes := i.GetNodes()
	for _, node := range nodes {
		nodeResults := node.RunCommand(cmd)
		for host, result := range nodeResults.Results {
			results.Results[host] = result
		}
	}
	return results
}

// RunPlaybook executes a playbook across all nodes.
func (i *inventoryImplementation) RunPlaybook(pb playbook.PlaybookInterface) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	nodes := i.GetNodes()
	for _, node := range nodes {
		nodeResults := node.RunPlaybook(pb)
		for host, result := range nodeResults.Results {
			results.Results[host] = result
		}
	}
	return results
}

// RunPlaybookByID executes a playbook by ID across all nodes.
func (i *inventoryImplementation) RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	nodes := i.GetNodes()
	for _, node := range nodes {
		nodeResults := node.RunPlaybookByID(id, opts...)
		for host, result := range nodeResults.Results {
			results.Results[host] = result
		}
	}
	return results
}

// CheckPlaybook runs the playbook's check mode across all nodes.
func (i *inventoryImplementation) CheckPlaybook(pb playbook.PlaybookInterface) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	nodes := i.GetNodes()
	for _, node := range nodes {
		pbCopy := pb
		pbCopy.SetDryRun(true)
		nodeResults := node.RunPlaybook(pbCopy)
		for host, result := range nodeResults.Results {
			results.Results[host] = result
		}
	}
	return results
}
