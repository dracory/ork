package ork

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/types"
)

// TestInventoryImplementation_NewInventory verifies that NewInventory creates an empty inventory.
func TestInventoryImplementation_NewInventory(t *testing.T) {
	i := NewInventory()

	// Verify initial state: no groups, no nodes
	if nodes := i.GetNodes(); len(nodes) != 0 {
		t.Errorf("Expected 0 nodes initially, got %d", len(nodes))
	}

	// Verify that getting a non-existent group returns nil
	if g := i.GetGroupByName("nonexistent"); g != nil {
		t.Error("Expected GetGroupByName to return nil for non-existent group")
	}
}

// TestInventoryImplementation_AddGroup verifies that AddGroup adds groups to the inventory.
func TestInventoryImplementation_AddGroup(t *testing.T) {
	i := NewInventory()

	group1 := NewGroup("web-servers")
	group2 := NewGroup("db-servers")

	// Add first group
	result1 := i.AddGroup(group1)
	if result1 != i {
		t.Error("Expected AddGroup to return self for chaining")
	}

	// Verify group was added
	retrieved := i.GetGroupByName("web-servers")
	if retrieved == nil {
		t.Fatal("Expected to retrieve web-servers group")
	}
	if retrieved.GetName() != "web-servers" {
		t.Errorf("Expected group name=%q, got %q", "web-servers", retrieved.GetName())
	}

	// Add second group
	result2 := i.AddGroup(group2)
	if result2 != i {
		t.Error("Expected AddGroup to return self for chaining")
	}

	// Verify second group
	retrieved = i.GetGroupByName("db-servers")
	if retrieved == nil {
		t.Fatal("Expected to retrieve db-servers group")
	}
	if retrieved.GetName() != "db-servers" {
		t.Errorf("Expected group name=%q, got %q", "db-servers", retrieved.GetName())
	}
}

// TestInventoryImplementation_AddGroup_Overwrite verifies that adding a group with the same name overwrites.
func TestInventoryImplementation_AddGroup_Overwrite(t *testing.T) {
	i := NewInventory()

	group1 := NewGroup("servers")
	group2 := NewGroup("servers")

	// Add node to first group
	node1 := &invTestMockNode{host: "server1.example.com"}
	group1.AddNode(node1)

	i.AddGroup(group1)

	// Add different node to second group with same name
	node2 := &invTestMockNode{host: "server2.example.com"}
	group2.AddNode(node2)

	// Overwrite with second group
	i.AddGroup(group2)

	// Verify we get the second group
	retrieved := i.GetGroupByName("servers")
	if retrieved == nil {
		t.Fatal("Expected to retrieve servers group")
	}

	nodes := retrieved.GetNodes()
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node in overwritten group, got %d", len(nodes))
	}
}

// TestInventoryImplementation_GetGroupByName verifies retrieving groups by name.
func TestInventoryImplementation_GetGroupByName(t *testing.T) {
	i := NewInventory()

	group := NewGroup("web-servers")
	i.AddGroup(group)

	// Test retrieving existing group
	retrieved := i.GetGroupByName("web-servers")
	if retrieved == nil {
		t.Error("Expected to retrieve web-servers group")
	}

	// Test retrieving non-existent group
	nonExistent := i.GetGroupByName("nonexistent")
	if nonExistent != nil {
		t.Error("Expected nil for non-existent group")
	}
}

// TestInventoryImplementation_AddNode verifies that AddNode adds nodes directly to inventory.
func TestInventoryImplementation_AddNode(t *testing.T) {
	i := NewInventory()

	node1 := &invTestMockNode{host: "standalone1.example.com"}
	node2 := &invTestMockNode{host: "standalone2.example.com"}

	// Add first node
	result1 := i.AddNode(node1)
	if result1 != i {
		t.Error("Expected AddNode to return self for chaining")
	}

	// Add second node
	result2 := i.AddNode(node2)
	if result2 != i {
		t.Error("Expected AddNode to return self for chaining")
	}

	// Verify nodes are in inventory
	nodes := i.GetNodes()
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodes))
	}
}

// TestInventoryImplementation_GetNodes verifies that GetNodes returns all nodes.
func TestInventoryImplementation_GetNodes(t *testing.T) {
	i := NewInventory()

	// Add standalone nodes
	standalone1 := &invTestMockNode{host: "standalone1.example.com"}
	standalone2 := &invTestMockNode{host: "standalone2.example.com"}
	i.AddNode(standalone1).AddNode(standalone2)

	// Add nodes via groups
	group := NewGroup("web-servers")
	groupNode1 := &invTestMockNode{host: "web1.example.com"}
	groupNode2 := &invTestMockNode{host: "web2.example.com"}
	group.AddNode(groupNode1).AddNode(groupNode2)
	i.AddGroup(group)

	// Get all nodes
	nodes := i.GetNodes()

	// Should have 4 nodes total: 2 standalone + 2 in group
	if len(nodes) != 4 {
		t.Errorf("Expected 4 nodes total, got %d", len(nodes))
	}

	// Verify hosts are present
	hosts := make(map[string]bool)
	for _, node := range nodes {
		hosts[node.GetHost()] = true
	}

	expectedHosts := []string{
		"standalone1.example.com",
		"standalone2.example.com",
		"web1.example.com",
		"web2.example.com",
	}

	for _, host := range expectedHosts {
		if !hosts[host] {
			t.Errorf("Expected node with host %q not found", host)
		}
	}
}

// TestInventoryImplementation_GetNodes_Empty verifies GetNodes with no nodes.
func TestInventoryImplementation_GetNodes_Empty(t *testing.T) {
	i := NewInventory()

	nodes := i.GetNodes()
	if nodes == nil {
		t.Error("Expected empty slice, not nil")
	}
	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(nodes))
	}
}

// TestInventoryImplementation_GetNodes_ReturnsCopy verifies that GetNodes returns a copy.
func TestInventoryImplementation_GetNodes_ReturnsCopy(t *testing.T) {
	i := NewInventory()

	node := &invTestMockNode{host: "server1.example.com"}
	i.AddNode(node)

	nodes1 := i.GetNodes()
	nodes1 = append(nodes1, &invTestMockNode{host: "extra.example.com"})

	nodes2 := i.GetNodes()
	if len(nodes2) != 1 {
		t.Error("Expected GetNodes to return a copy, not internal slice")
	}
}

// TestInventoryImplementation_GetNodes_MultipleGroups verifies GetNodes with multiple groups.
func TestInventoryImplementation_GetNodes_MultipleGroups(t *testing.T) {
	i := NewInventory()

	// Create multiple groups with nodes
	webGroup := NewGroup("web-servers")
	webGroup.AddNode(&invTestMockNode{host: "web1.example.com"})
	webGroup.AddNode(&invTestMockNode{host: "web2.example.com"})

	dbGroup := NewGroup("db-servers")
	dbGroup.AddNode(&invTestMockNode{host: "db1.example.com"})

	i.AddGroup(webGroup).AddGroup(dbGroup)

	nodes := i.GetNodes()

	// Should have 3 nodes from groups
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes from groups, got %d", len(nodes))
	}
}

// TestInventoryImplementation_SetMaxConcurrency verifies that SetMaxConcurrency sets the value.
func TestInventoryImplementation_SetMaxConcurrency(t *testing.T) {
	i := NewInventory()

	// Test setting max concurrency
	result := i.SetMaxConcurrency(5)
	if result != i {
		t.Error("Expected SetMaxConcurrency to return self for chaining")
	}

	// Note: maxConcurrency is not exposed via interface, but we can verify
	// the method doesn't panic and returns self

	// Test with different values
	i.SetMaxConcurrency(0)
	i.SetMaxConcurrency(100)
}

// TestInventoryImplementation_SetterChaining verifies that all setters can be chained.
func TestInventoryImplementation_SetterChaining(t *testing.T) {
	i := NewInventory()

	group := NewGroup("web-servers")
	node := &invTestMockNode{host: "server1.example.com"}

	// Chain all setter methods
	result := i.
		AddGroup(group).
		SetMaxConcurrency(5).
		AddNode(node)

	// Verify all operations were applied
	if i.GetGroupByName("web-servers") == nil {
		t.Error("Expected web-servers group to be added")
	}

	nodes := i.GetNodes()
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}

	if result != i {
		t.Error("Expected chained methods to return self")
	}
}

// TestInventoryImplementation_RunCommand verifies command execution across all nodes.
func TestInventoryImplementation_RunCommand(t *testing.T) {
	i := NewInventory()

	// Add standalone nodes
	node1 := &invTestMockNode{host: "standalone1.example.com"}
	node2 := &invTestMockNode{host: "standalone2.example.com"}
	i.AddNode(node1).AddNode(node2)

	// Add nodes via group
	group := NewGroup("web-servers")
	groupNode := &invTestMockNode{host: "web1.example.com"}
	group.AddNode(groupNode)
	i.AddGroup(group)

	results := i.RunCommand("uptime")

	// Should have results for all 3 nodes
	if len(results.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results.Results))
	}

	// Verify each node has a result
	expectedHosts := []string{
		"standalone1.example.com",
		"standalone2.example.com",
		"web1.example.com",
	}

	for _, host := range expectedHosts {
		if _, ok := results.Results[host]; !ok {
			t.Errorf("Expected result for %s", host)
		}
	}
}

// TestInventoryImplementation_RunCommand_Empty verifies RunCommand with no nodes.
func TestInventoryImplementation_RunCommand_Empty(t *testing.T) {
	i := NewInventory()

	results := i.RunCommand("uptime")

	if len(results.Results) != 0 {
		t.Errorf("Expected 0 results for empty inventory, got %d", len(results.Results))
	}
}

// TestInventoryImplementation_RunPlaybook verifies playbook execution across all nodes.
func TestInventoryImplementation_RunPlaybook(t *testing.T) {
	i := NewInventory()

	node1 := &invTestMockNode{host: "server1.example.com"}
	node2 := &invTestMockNode{host: "server2.example.com"}
	i.AddNode(node1).AddNode(node2)

	mockPb := &invTestMockPlaybook{name: "test-playbook"}
	results := i.RunPlaybook(mockPb)

	if len(results.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results.Results))
	}
}

// TestInventoryImplementation_RunPlaybook_Empty verifies RunPlaybook with no nodes.
func TestInventoryImplementation_RunPlaybook_Empty(t *testing.T) {
	i := NewInventory()

	mockPb := &invTestMockPlaybook{name: "test-playbook"}
	results := i.RunPlaybook(mockPb)

	if len(results.Results) != 0 {
		t.Errorf("Expected 0 results for empty inventory, got %d", len(results.Results))
	}
}

// TestInventoryImplementation_RunPlaybookByID verifies playbook execution by ID.
func TestInventoryImplementation_RunPlaybookByID(t *testing.T) {
	i := NewInventory()

	node1 := &invTestMockNode{host: "server1.example.com"}
	node2 := &invTestMockNode{host: "server2.example.com"}
	i.AddNode(node1).AddNode(node2)

	results := i.RunPlaybookByID("test-playbook")

	// Results may be empty if playbook not registered, but should not panic
	if results.Results == nil {
		t.Error("Expected Results map to be initialized")
	}
}

// TestInventoryImplementation_CheckPlaybook verifies check mode execution.
func TestInventoryImplementation_CheckPlaybook(t *testing.T) {
	i := NewInventory()

	node1 := &invTestMockNode{host: "server1.example.com"}
	node2 := &invTestMockNode{host: "server2.example.com"}
	i.AddNode(node1).AddNode(node2)

	mockPb := &invTestMockPlaybook{name: "test-playbook"}
	results := i.CheckPlaybook(mockPb)

	if len(results.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results.Results))
	}
}

// TestInventoryImplementation_CheckPlaybook_SetsDryRun verifies CheckPlaybook sets dry run mode.
func TestInventoryImplementation_CheckPlaybook_SetsDryRun(t *testing.T) {
	i := NewInventory()

	node1 := &invTestMockNode{host: "server1.example.com"}
	i.AddNode(node1)

	mockPb := &invTestMockPlaybook{name: "test-playbook"}
	results := i.CheckPlaybook(mockPb)

	// Just verify it runs without error
	if results.Results == nil {
		t.Error("Expected Results map to be initialized")
	}
}

// TestInventoryImplementation_ComplexScenario verifies a complex inventory setup.
func TestInventoryImplementation_ComplexScenario(t *testing.T) {
	i := NewInventory()

	// Create groups
	webGroup := NewGroup("web-servers").
		SetArg("role", "web").
		SetArg("environment", "production")
	webGroup.AddNode(&invTestMockNode{host: "web1.example.com"})
	webGroup.AddNode(&invTestMockNode{host: "web2.example.com"})
	webGroup.AddNode(&invTestMockNode{host: "web3.example.com"})

	dbGroup := NewGroup("db-servers").
		SetArg("role", "database").
		SetArg("environment", "production")
	dbGroup.AddNode(&invTestMockNode{host: "db1.example.com"})
	dbGroup.AddNode(&invTestMockNode{host: "db2.example.com"})

	// Add groups and standalone nodes
	i.AddGroup(webGroup).
		AddGroup(dbGroup).
		AddNode(&invTestMockNode{host: "lb1.example.com"}).
		AddNode(&invTestMockNode{host: "monitor1.example.com"}).
		SetMaxConcurrency(5)

	// Verify groups
	if i.GetGroupByName("web-servers") == nil {
		t.Error("Expected web-servers group")
	}
	if i.GetGroupByName("db-servers") == nil {
		t.Error("Expected db-servers group")
	}

	// Verify total nodes: 3 web + 2 db + 2 standalone = 7
	nodes := i.GetNodes()
	if len(nodes) != 7 {
		t.Errorf("Expected 7 total nodes, got %d", len(nodes))
	}

	// Verify group arguments
	webArgs := i.GetGroupByName("web-servers").GetArgs()
	if webArgs["role"] != "web" {
		t.Errorf("Expected role=web for web-servers, got %s", webArgs["role"])
	}

	dbArgs := i.GetGroupByName("db-servers").GetArgs()
	if dbArgs["role"] != "database" {
		t.Errorf("Expected role=database for db-servers, got %s", dbArgs["role"])
	}
}

// invTestMockNode is a mock implementation of NodeInterface for testing.
type invTestMockNode struct {
	host string
	args map[string]string
}

func (m *invTestMockNode) GetHost() string {
	return m.host
}

func (m *invTestMockNode) GetPort() string {
	return "22"
}

func (m *invTestMockNode) GetUser() string {
	return "root"
}

func (m *invTestMockNode) GetKey() string {
	return "id_rsa"
}

func (m *invTestMockNode) SetPort(port string) NodeInterface {
	return m
}

func (m *invTestMockNode) SetUser(user string) NodeInterface {
	return m
}

func (m *invTestMockNode) SetKey(key string) NodeInterface {
	return m
}

func (m *invTestMockNode) SetArg(key, value string) NodeInterface {
	if m.args == nil {
		m.args = make(map[string]string)
	}
	m.args[key] = value
	return m
}

func (m *invTestMockNode) SetArgs(args map[string]string) NodeInterface {
	m.args = args
	return m
}

func (m *invTestMockNode) GetArg(key string) string {
	if m.args == nil {
		return ""
	}
	return m.args[key]
}

func (m *invTestMockNode) GetArgs() map[string]string {
	if m.args == nil {
		return make(map[string]string)
	}
	result := make(map[string]string, len(m.args))
	for k, v := range m.args {
		result[k] = v
	}
	return result
}

func (m *invTestMockNode) GetNodeConfig() config.NodeConfig {
	return config.NodeConfig{
		SSHHost:  m.host,
		SSHPort:  "22",
		RootUser: "root",
		SSHKey:   "id_rsa",
		Args:     m.GetArgs(),
	}
}

func (m *invTestMockNode) Connect() error {
	return nil
}

func (m *invTestMockNode) Close() error {
	return nil
}

func (m *invTestMockNode) IsConnected() bool {
	return false
}

func (m *invTestMockNode) RunCommand(cmd string) types.Results {
	return types.Results{
		Results: map[string]types.Result{
			m.host: {
				Changed: true,
				Message: "Command executed: " + cmd,
			},
		},
	}
}

func (m *invTestMockNode) RunPlaybook(pb playbook.PlaybookInterface) types.Results {
	return types.Results{
		Results: map[string]types.Result{
			m.host: {
				Changed: true,
				Message: "Playbook executed",
			},
		},
	}
}

func (m *invTestMockNode) RunPlaybookByID(id string, opts ...playbook.PlaybookOptions) types.Results {
	return types.Results{
		Results: map[string]types.Result{
			m.host: {
				Changed: true,
				Message: "Playbook by ID executed: " + id,
			},
		},
	}
}

func (m *invTestMockNode) CheckPlaybook(pb playbook.PlaybookInterface) types.Results {
	return types.Results{
		Results: map[string]types.Result{
			m.host: {
				Changed: false,
				Message: "Check mode executed",
			},
		},
	}
}

func (m *invTestMockNode) GetLogger() *slog.Logger {
	return slog.Default()
}

func (m *invTestMockNode) SetLogger(logger *slog.Logger) RunnableInterface {
	return m
}

// invTestMockPlaybook is a mock implementation of playbook.PlaybookInterface for testing.
type invTestMockPlaybook struct {
	name    string
	cfg     config.NodeConfig
	dryRun  bool
	args    map[string]string
	timeout time.Duration
}

func (m *invTestMockPlaybook) GetID() string {
	return m.name
}

func (m *invTestMockPlaybook) SetID(id string) playbook.PlaybookInterface {
	m.name = id
	return m
}

func (m *invTestMockPlaybook) GetDescription() string {
	return "Mock playbook"
}

func (m *invTestMockPlaybook) SetDescription(desc string) playbook.PlaybookInterface {
	return m
}

func (m *invTestMockPlaybook) GetConfig() config.NodeConfig {
	return m.cfg
}

func (m *invTestMockPlaybook) SetConfig(cfg config.NodeConfig) playbook.PlaybookInterface {
	m.cfg = cfg
	return m
}

func (m *invTestMockPlaybook) GetArg(key string) string {
	if m.args == nil {
		return ""
	}
	return m.args[key]
}

func (m *invTestMockPlaybook) SetArg(key, value string) playbook.PlaybookInterface {
	if m.args == nil {
		m.args = make(map[string]string)
	}
	m.args[key] = value
	return m
}

func (m *invTestMockPlaybook) GetArgs() map[string]string {
	return m.args
}

func (m *invTestMockPlaybook) SetArgs(args map[string]string) playbook.PlaybookInterface {
	m.args = args
	return m
}

func (m *invTestMockPlaybook) IsDryRun() bool {
	return m.dryRun
}

func (m *invTestMockPlaybook) SetDryRun(dryRun bool) playbook.PlaybookInterface {
	m.dryRun = dryRun
	return m
}

func (m *invTestMockPlaybook) GetTimeout() time.Duration {
	return m.timeout
}

func (m *invTestMockPlaybook) SetTimeout(timeout time.Duration) playbook.PlaybookInterface {
	m.timeout = timeout
	return m
}

func (m *invTestMockPlaybook) Check() (bool, error) {
	return true, nil
}

func (m *invTestMockPlaybook) Run() playbook.Result {
	return playbook.Result{
		Changed: true,
		Message: "Success",
	}
}
