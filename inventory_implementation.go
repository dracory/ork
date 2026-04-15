package ork

import (
	"fmt"
	"log/slog"
	"maps"
	"sync"

	"github.com/dracory/ork/types"
)

// inventoryImplementation is the default implementation of InventoryInterface.
type inventoryImplementation struct {
	groups         map[string]GroupInterface
	nodes          []NodeInterface
	maxConcurrency int
	logger         *slog.Logger
	dryRunMode     bool
	mu             sync.RWMutex
}

// AddGroup adds a group to the inventory.
func (i *inventoryImplementation) AddGroup(group GroupInterface) InventoryInterface {
	i.mu.Lock()
	i.groups[group.GetName()] = group
	i.mu.Unlock()
	// Propagate dry-run mode to new group for consistency
	if group.GetDryRunMode() != i.GetDryRunMode() {
		group.SetDryRunMode(i.GetDryRunMode())
	}
	return i
}

// GetGroupByName retrieves a group by name.
func (i *inventoryImplementation) GetGroupByName(name string) GroupInterface {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.groups[name]
}

// AddNode adds a node directly to the inventory.
func (i *inventoryImplementation) AddNode(node NodeInterface) InventoryInterface {
	i.mu.Lock()
	i.nodes = append(i.nodes, node)
	i.mu.Unlock()
	// Propagate dry-run mode to new node for consistency
	if node.GetDryRunMode() != i.GetDryRunMode() {
		node.SetDryRunMode(i.GetDryRunMode())
	}
	return i
}

// GetNodes returns all nodes in the inventory across all groups.
func (i *inventoryImplementation) GetNodes() []NodeInterface {
	i.mu.RLock()
	result := make([]NodeInterface, 0, len(i.nodes))
	result = append(result, i.nodes...)
	groupsCopy := make(map[string]GroupInterface, len(i.groups))
	for k, v := range i.groups {
		groupsCopy[k] = v
	}
	i.mu.RUnlock()

	// Also include nodes from groups
	for _, group := range groupsCopy {
		result = append(result, group.GetNodes()...)
	}
	return result
}

// SetMaxConcurrency sets the maximum number of concurrent operations.
func (i *inventoryImplementation) SetMaxConcurrency(max int) InventoryInterface {
	i.mu.Lock()
	i.maxConcurrency = max
	i.mu.Unlock()
	return i
}

// propagateDryRun applies the inventory's dry-run mode to all groups and nodes.
func (i *inventoryImplementation) propagateDryRun() {
	i.mu.RLock()
	mode := i.dryRunMode
	groupsCopy := make(map[string]GroupInterface, len(i.groups))
	for k, v := range i.groups {
		groupsCopy[k] = v
	}
	nodesCopy := make([]NodeInterface, len(i.nodes))
	copy(nodesCopy, i.nodes)
	i.mu.RUnlock()

	for _, group := range groupsCopy {
		if group.GetDryRunMode() != mode {
			group.SetDryRunMode(mode)
		}
	}
	for _, node := range nodesCopy {
		if node.GetDryRunMode() != mode {
			node.SetDryRunMode(mode)
		}
	}
}

// RunCommand executes a shell command across all nodes.
func (i *inventoryImplementation) RunCommand(cmd string) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	i.propagateDryRun()

	nodes := i.GetNodes()

	// Determine concurrency limit
	concurrency := i.maxConcurrency
	if concurrency == 0 {
		concurrency = len(nodes) // unlimited
	}
	if concurrency > len(nodes) {
		concurrency = len(nodes)
	}

	// Use semaphore for concurrency control
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)
		go func(n NodeInterface) {
			defer func() {
				if r := recover(); r != nil {
					// Log panic and record failure in results
					i.GetLogger().Error("panic in RunCommand goroutine", "error", r)
					i.mu.Lock()
					results.Results[n.GetHost()] = types.Result{
						Changed: false,
						Message: fmt.Sprintf("panic: %v", r),
					}
					i.mu.Unlock()
				}
				wg.Done()
			}()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			nodeResults := n.RunCommand(cmd)

			// Protect results map with mutex
			i.mu.Lock()
			maps.Copy(results.Results, nodeResults.Results)
			i.mu.Unlock()
		}(node)
	}

	wg.Wait()
	return results
}

// RunSkill executes a skill across all nodes.
func (i *inventoryImplementation) RunSkill(skill types.RunnableInterface) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	i.propagateDryRun()
	nodes := i.GetNodes()

	// Determine concurrency limit
	concurrency := i.maxConcurrency
	if concurrency == 0 {
		concurrency = len(nodes) // unlimited
	}
	if concurrency > len(nodes) {
		concurrency = len(nodes)
	}

	// Use semaphore for concurrency control
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)
		go func(n NodeInterface) {
			defer func() {
				if r := recover(); r != nil {
					// Log panic and record failure in results
					i.GetLogger().Error("panic in RunSkill goroutine", "error", r)
					i.mu.Lock()
					results.Results[n.GetHost()] = types.Result{
						Changed: false,
						Message: fmt.Sprintf("panic: %v", r),
					}
					i.mu.Unlock()
				}
				wg.Done()
			}()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			nodeResults := n.RunSkill(skill)

			// Protect results map with mutex
			i.mu.Lock()
			maps.Copy(results.Results, nodeResults.Results)
			i.mu.Unlock()
		}(node)
	}

	wg.Wait()
	return results
}

// RunSkillByID executes a skill by ID across all nodes.
func (i *inventoryImplementation) RunSkillByID(id string, opts ...types.SkillOptions) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	i.propagateDryRun()
	nodes := i.GetNodes()

	// Determine concurrency limit
	concurrency := i.maxConcurrency
	if concurrency == 0 {
		concurrency = len(nodes) // unlimited
	}
	if concurrency > len(nodes) {
		concurrency = len(nodes)
	}

	// Use semaphore for concurrency control
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)
		go func(n NodeInterface) {
			defer func() {
				if r := recover(); r != nil {
					// Log panic and record failure in results
					i.GetLogger().Error("panic in RunSkillByID goroutine", "error", r)
					i.mu.Lock()
					results.Results[n.GetHost()] = types.Result{
						Changed: false,
						Message: fmt.Sprintf("panic: %v", r),
					}
					i.mu.Unlock()
				}
				wg.Done()
			}()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			nodeResults := n.RunSkillByID(id, opts...)

			// Protect results map with mutex
			i.mu.Lock()
			maps.Copy(results.Results, nodeResults.Results)
			i.mu.Unlock()
		}(node)
	}

	wg.Wait()
	return results
}

// CheckSkill runs the skill's check mode across all nodes.
func (i *inventoryImplementation) CheckSkill(skill types.RunnableInterface) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	i.propagateDryRun()
	nodes := i.GetNodes()

	// Determine concurrency limit
	concurrency := i.maxConcurrency
	if concurrency == 0 {
		concurrency = len(nodes) // unlimited
	}
	if concurrency > len(nodes) {
		concurrency = len(nodes)
	}

	// Use semaphore for concurrency control
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)
		go func(n NodeInterface) {
			defer func() {
				if r := recover(); r != nil {
					// Log panic and record failure in results
					i.GetLogger().Error("panic in CheckSkill goroutine", "error", r)
					i.mu.Lock()
					results.Results[n.GetHost()] = types.Result{
						Changed: false,
						Message: fmt.Sprintf("panic: %v", r),
					}
					i.mu.Unlock()
				}
				wg.Done()
			}()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			nodeResults := n.CheckSkill(skill)

			// Protect results map with mutex
			i.mu.Lock()
			maps.Copy(results.Results, nodeResults.Results)
			i.mu.Unlock()
		}(node)
	}

	wg.Wait()
	return results
}

// GetLogger returns the logger. Returns slog.Default() if not set.
func (i *inventoryImplementation) GetLogger() *slog.Logger {
	if i.logger == nil {
		return slog.Default()
	}
	return i.logger
}

// SetLogger sets a custom logger. Returns RunnerInterface for chaining.
func (i *inventoryImplementation) SetLogger(logger *slog.Logger) RunnerInterface {
	i.logger = logger
	return i
}

// SetDryRunMode sets whether to simulate execution without making changes.
// When true, ssh.Run() will log commands and return "[dry-run]" marker instead of executing.
// The dry-run mode is applied to groups/nodes at execution time and when set.
// Returns RunnerInterface for fluent method chaining.
func (i *inventoryImplementation) SetDryRunMode(dryRun bool) RunnerInterface {
	i.mu.Lock()
	i.dryRunMode = dryRun
	i.mu.Unlock()
	// Also propagate immediately for consistency
	i.propagateDryRun()
	return i
}

// GetDryRunMode returns true if dry-run mode is enabled for this inventory.
func (i *inventoryImplementation) GetDryRunMode() bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.dryRunMode
}
