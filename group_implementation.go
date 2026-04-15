package ork

import (
	"log/slog"
	"maps"
	"sync"

	"github.com/dracory/ork/types"
)

// groupImplementation is the default implementation of GroupInterface.
type groupImplementation struct {
	name       string
	nodes      []NodeInterface
	args       map[string]string
	logger     *slog.Logger
	dryRunMode bool
	mu         sync.RWMutex
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
	// Propagate dry-run mode to new node for consistency
	if node.GetDryRunMode() != g.GetDryRunMode() {
		node.SetDryRunMode(g.GetDryRunMode())
	}
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
	maps.Copy(result, g.args)
	return result
}

// propagateDryRun applies the group's dry-run mode to all nodes.
func (g *groupImplementation) propagateDryRun() {
	g.mu.RLock()
	mode := g.dryRunMode
	g.mu.RUnlock()
	for _, node := range g.nodes {
		if node.GetDryRunMode() != mode {
			node.SetDryRunMode(mode)
		}
	}
}

// RunCommand executes a shell command across all nodes in this group.
func (g *groupImplementation) RunCommand(cmd string) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	g.propagateDryRun() // !!! Important: propagate dry-run mode to nodes
	for _, node := range g.nodes {
		nodeResults := node.RunCommand(cmd)
		maps.Copy(results.Results, nodeResults.Results)
	}
	return results
}

// Run executes a skill across all nodes in this group.
func (g *groupImplementation) Run(runnable types.RunnableInterface) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	g.propagateDryRun() // !!! Important: propagate dry-run mode to nodes
	for _, node := range g.nodes {
		nodeResults := node.Run(runnable)
		maps.Copy(results.Results, nodeResults.Results)
	}
	return results
}

// RunByID executes a skill by ID across all nodes in this group.
func (g *groupImplementation) RunByID(id string, opts ...types.RunnableOptions) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	g.propagateDryRun()

	for _, node := range g.nodes {
		nodeResults := node.RunByID(id, opts...)
		maps.Copy(results.Results, nodeResults.Results)
	}
	return results
}

// Check runs the skill's check mode across all nodes in this group.
func (g *groupImplementation) Check(skill types.RunnableInterface) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	g.propagateDryRun() // !!! Important: propagate dry-run mode to nodes

	for _, node := range g.nodes {
		nodeResults := node.Check(skill)
		maps.Copy(results.Results, nodeResults.Results)
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

// SetLogger sets a custom logger. Returns RunnerInterface for chaining.
func (g *groupImplementation) SetLogger(logger *slog.Logger) RunnerInterface {
	g.logger = logger
	return g
}

// SetDryRunMode sets whether to simulate execution without making changes.
// When true, ssh.Run() will log commands and return "[dry-run]" marker instead of executing.
// The dry-run mode is applied to nodes at execution time (RunSkill, RunCommand, etc.).
// Returns RunnerInterface for fluent method chaining.
func (g *groupImplementation) SetDryRunMode(dryRun bool) RunnerInterface {
	g.mu.Lock()
	g.dryRunMode = dryRun
	g.mu.Unlock()
	// Also propagate immediately for consistency
	g.propagateDryRun()
	return g
}

// GetDryRunMode returns true if dry-run mode is enabled for this group.
func (g *groupImplementation) GetDryRunMode() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.dryRunMode
}
