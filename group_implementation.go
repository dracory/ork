package ork

import (
	"log/slog"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/types"
)

// groupImplementation is the default implementation of GroupInterface.
type groupImplementation struct {
	name   string
	nodes  []NodeInterface
	args   map[string]string
	logger *slog.Logger
}

// NewGroup creates a new group with the given name.
func NewGroup(name string) GroupInterface {
	return &groupImplementation{
		name:  name,
		nodes: make([]NodeInterface, 0),
		args:  make(map[string]string),
	}
}

// GetName returns the group's name.
func (g *groupImplementation) GetName() string {
	return g.name
}

// AddNode adds a node to this group.
func (g *groupImplementation) AddNode(node NodeInterface) GroupInterface {
	g.nodes = append(g.nodes, node)
	return g
}

// GetNodes returns all nodes in this group.
func (g *groupImplementation) GetNodes() []NodeInterface {
	result := make([]NodeInterface, len(g.nodes))
	copy(result, g.nodes)
	return result
}

// SetArg sets an argument for this group.
func (g *groupImplementation) SetArg(key, value string) GroupInterface {
	g.args[key] = value
	return g
}

// GetArg retrieves an argument value by key.
func (g *groupImplementation) GetArg(key string) string {
	return g.args[key]
}

// GetArgs returns a copy of all arguments defined for this group.
func (g *groupImplementation) GetArgs() map[string]string {
	result := make(map[string]string, len(g.args))
	for k, v := range g.args {
		result[k] = v
	}
	return result
}

// RunCommand executes a shell command across all nodes in this group.
func (g *groupImplementation) RunCommand(cmd string) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	for _, node := range g.nodes {
		nodeResults := node.RunCommand(cmd)
		for host, result := range nodeResults.Results {
			results.Results[host] = result
		}
	}
	return results
}

// RunPlaybook executes a playbook across all nodes in this group.
func (g *groupImplementation) RunPlaybook(pb playbook.PlaybookInterface) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	for _, node := range g.nodes {
		nodeResults := node.RunPlaybook(pb)
		for host, result := range nodeResults.Results {
			results.Results[host] = result
		}
	}
	return results
}

// RunPlaybookByID executes a playbook by ID across all nodes in this group.
func (g *groupImplementation) RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	for _, node := range g.nodes {
		nodeResults := node.RunPlaybookByID(id, opts...)
		for host, result := range nodeResults.Results {
			results.Results[host] = result
		}
	}
	return results
}

// CheckPlaybook runs the playbook's check mode across all nodes in this group.
func (g *groupImplementation) CheckPlaybook(pb playbook.PlaybookInterface) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	for _, node := range g.nodes {
		pbCopy := pb
		pbCopy.SetDryRun(true)
		nodeResults := node.RunPlaybook(pbCopy)
		for host, result := range nodeResults.Results {
			results.Results[host] = result
		}
	}
	return results
}

// GetLogger returns the logger. Returns slog.Default() if not set.
func (g *groupImplementation) GetLogger() *slog.Logger {
	if g.logger == nil {
		return slog.Default()
	}
	return g.logger
}

// SetLogger sets a custom logger. Returns RunnableInterface for chaining.
func (g *groupImplementation) SetLogger(logger *slog.Logger) RunnableInterface {
	g.logger = logger
	return g
}
