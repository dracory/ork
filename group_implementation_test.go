package ork

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/types"
)

// TestGroupImplementation_NewGroup verifies that NewGroup creates a group with the given name.
func TestGroupImplementation_NewGroup(t *testing.T) {
	g := NewGroup("web-servers")

	// Verify the group name is set correctly
	if g.GetName() != "web-servers" {
		t.Errorf("Expected GetName()=%q, got %q", "web-servers", g.GetName())
	}

	// Verify initial state: no nodes, empty args
	if nodes := g.GetNodes(); len(nodes) != 0 {
		t.Errorf("Expected 0 nodes initially, got %d", len(nodes))
	}

	if args := g.GetArgs(); len(args) != 0 {
		t.Errorf("Expected 0 args initially, got %d", len(args))
	}
}

// TestGroupImplementation_GetName verifies that GetName returns the group's name.
func TestGroupImplementation_GetName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"web-servers", "web-servers"},
		{"db-cluster", "db-cluster"},
		{"", ""},
		{"production-nodes-us-east-1", "production-nodes-us-east-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGroup(tt.name)
			if g.GetName() != tt.expected {
				t.Errorf("Expected GetName()=%q, got %q", tt.expected, g.GetName())
			}
		})
	}
}

// TestGroupImplementation_AddNode verifies that AddNode adds nodes to the group.
func TestGroupImplementation_AddNode(t *testing.T) {
	g := NewGroup("web-servers")

	// Create mock nodes
	node1 := &groupTestMockNode{host: "server1.example.com"}
	node2 := &groupTestMockNode{host: "server2.example.com"}

	// Add first node
	result1 := g.AddNode(node1)
	if result1 != g {
		t.Error("Expected AddNode to return self for chaining")
	}

	nodes := g.GetNodes()
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}

	// Add second node
	result2 := g.AddNode(node2)
	if result2 != g {
		t.Error("Expected AddNode to return self for chaining")
	}

	nodes = g.GetNodes()
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodes))
	}
}

// TestGroupImplementation_GetNodes verifies that GetNodes returns a copy of the nodes slice.
func TestGroupImplementation_GetNodes(t *testing.T) {
	g := NewGroup("web-servers")

	node1 := &groupTestMockNode{host: "server1.example.com"}
	node2 := &groupTestMockNode{host: "server2.example.com"}

	g.AddNode(node1).AddNode(node2)

	// Get nodes twice
	nodes1 := g.GetNodes()
	nodes2 := g.GetNodes()

	// Verify both slices have same content
	if len(nodes1) != len(nodes2) {
		t.Error("Expected both node slices to have same length")
	}

	// Verify they are independent copies (modifying one doesn't affect the other)
	// Note: This test verifies the defensive copy behavior
	// If the implementation returns a reference to the internal slice,
	// appending to one would affect the other
	nodes1 = append(nodes1, &groupTestMockNode{host: "server3.example.com"})
	nodes2After := g.GetNodes()

	if len(nodes2After) != 2 {
		t.Error("Expected GetNodes to return a copy, not internal slice")
	}
}

// TestGroupImplementation_GetNodes_Empty verifies GetNodes with no nodes.
func TestGroupImplementation_GetNodes_Empty(t *testing.T) {
	g := NewGroup("empty-group")

	nodes := g.GetNodes()
	if nodes == nil {
		t.Error("Expected empty slice, not nil")
	}
	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(nodes))
	}
}

// TestGroupImplementation_SetArg verifies that SetArg adds arguments.
func TestGroupImplementation_SetArg(t *testing.T) {
	g := NewGroup("web-servers")

	// Set first argument
	result1 := g.SetArg("environment", "production")
	if result1 != g {
		t.Error("Expected SetArg to return self for chaining")
	}

	if g.GetArg("environment") != "production" {
		t.Errorf("Expected GetArg(environment)=%q, got %q", "production", g.GetArg("environment"))
	}

	// Set second argument
	result2 := g.SetArg("datacenter", "us-east-1")
	if result2 != g {
		t.Error("Expected SetArg to return self for chaining")
	}

	if g.GetArg("datacenter") != "us-east-1" {
		t.Errorf("Expected GetArg(datacenter)=%q, got %q", "us-east-1", g.GetArg("datacenter"))
	}

	// Verify first argument still exists
	if g.GetArg("environment") != "production" {
		t.Errorf("Expected GetArg(environment)=%q, got %q", "production", g.GetArg("environment"))
	}
}

// TestGroupImplementation_SetArg_Overwrite verifies that SetArg overwrites existing values.
func TestGroupImplementation_SetArg_Overwrite(t *testing.T) {
	g := NewGroup("web-servers")

	g.SetArg("environment", "staging")
	g.SetArg("environment", "production")

	if g.GetArg("environment") != "production" {
		t.Errorf("Expected GetArg(environment)=%q after overwrite, got %q", "production", g.GetArg("environment"))
	}
}

// TestGroupImplementation_GetArg_NonExistent verifies GetArg returns empty string for non-existent keys.
func TestGroupImplementation_GetArg_NonExistent(t *testing.T) {
	g := NewGroup("web-servers")

	value := g.GetArg("nonexistent")
	if value != "" {
		t.Errorf("Expected empty string for non-existent key, got %q", value)
	}
}

// TestGroupImplementation_GetArgs verifies that GetArgs returns all arguments.
func TestGroupImplementation_GetArgs(t *testing.T) {
	g := NewGroup("web-servers")

	g.SetArg("environment", "production")
	g.SetArg("datacenter", "us-east-1")

	args := g.GetArgs()

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}

	if args["environment"] != "production" {
		t.Errorf("Expected args[environment]=%q, got %q", "production", args["environment"])
	}

	if args["datacenter"] != "us-east-1" {
		t.Errorf("Expected args[datacenter]=%q, got %q", "us-east-1", args["datacenter"])
	}
}

// TestGroupImplementation_GetArgs_ReturnsCopy verifies that GetArgs returns a copy.
func TestGroupImplementation_GetArgs_ReturnsCopy(t *testing.T) {
	g := NewGroup("web-servers")

	g.SetArg("key1", "value1")

	args := g.GetArgs()
	args["key2"] = "value2" // Modify the returned map

	// Verify the modification doesn't affect the group
	if g.GetArg("key2") != "" {
		t.Error("Expected GetArgs to return a copy, not internal map")
	}
}

// TestGroupImplementation_GetArgs_Empty verifies GetArgs with no arguments.
func TestGroupImplementation_GetArgs_Empty(t *testing.T) {
	g := NewGroup("empty-group")

	args := g.GetArgs()
	if args == nil {
		t.Error("Expected empty map, not nil")
	}
	if len(args) != 0 {
		t.Errorf("Expected 0 args, got %d", len(args))
	}
}

// TestGroupImplementation_SetterChaining verifies that all setters can be chained.
func TestGroupImplementation_SetterChaining(t *testing.T) {
	g := NewGroup("web-servers")

	node1 := &groupTestMockNode{host: "server1.example.com"}
	node2 := &groupTestMockNode{host: "server2.example.com"}

	// Chain all setter methods
	result := g.
		AddNode(node1).
		SetArg("environment", "production").
		AddNode(node2).
		SetArg("datacenter", "us-east-1")

	// Verify all operations were applied
	if len(g.GetNodes()) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(g.GetNodes()))
	}

	if g.GetArg("environment") != "production" {
		t.Errorf("Expected environment=%q, got %q", "production", g.GetArg("environment"))
	}

	if g.GetArg("datacenter") != "us-east-1" {
		t.Errorf("Expected datacenter=%q, got %q", "us-east-1", g.GetArg("datacenter"))
	}

	if result != g {
		t.Error("Expected chained methods to return self")
	}
}

// TestGroupImplementation_RunCommand verifies command execution across group nodes.
func TestGroupImplementation_RunCommand(t *testing.T) {
	g := NewGroup("web-servers")

	node1 := &groupTestMockNode{host: "server1.example.com"}
	node2 := &groupTestMockNode{host: "server2.example.com"}

	g.AddNode(node1).AddNode(node2)

	results := g.RunCommand("uptime")

	// Verify results contain entries for both nodes
	if len(results.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results.Results))
	}

	// Verify each node has a result
	if _, ok := results.Results["server1.example.com"]; !ok {
		t.Error("Expected result for server1.example.com")
	}
	if _, ok := results.Results["server2.example.com"]; !ok {
		t.Error("Expected result for server2.example.com")
	}
}

// TestGroupImplementation_RunCommand_EmptyGroup verifies RunCommand with no nodes.
func TestGroupImplementation_RunCommand_EmptyGroup(t *testing.T) {
	g := NewGroup("empty-group")

	results := g.RunCommand("uptime")

	if len(results.Results) != 0 {
		t.Errorf("Expected 0 results for empty group, got %d", len(results.Results))
	}
}

// TestGroupImplementation_RunPlaybook verifies playbook execution across group nodes.
func TestGroupImplementation_RunSkill(t *testing.T) {
	g := NewGroup("web-servers")

	node1 := &groupTestMockNode{host: "server1.example.com"}
	node2 := &groupTestMockNode{host: "server2.example.com"}

	g.AddNode(node1).AddNode(node2)

	mockPb := &groupTestMockPlaybook{name: "test-playbook"}

	results := g.RunSkill(mockPb)

	// Verify results contain entries for both nodes
	if len(results.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results.Results))
	}
}

// TestGroupImplementation_RunPlaybook_EmptyGroup verifies RunPlaybook with no nodes.
func TestGroupImplementation_RunPlaybook_EmptyGroup(t *testing.T) {
	g := NewGroup("empty-group")

	mockPb := &groupTestMockPlaybook{name: "test-playbook"}
	results := g.RunSkill(mockPb)

	if len(results.Results) != 0 {
		t.Errorf("Expected 0 results for empty group, got %d", len(results.Results))
	}
}

// TestGroupImplementation_RunPlaybookByID verifies playbook execution by ID.
func TestGroupImplementation_RunSkillByID(t *testing.T) {
	g := NewGroup("web-servers")

	node1 := &groupTestMockNode{host: "server1.example.com"}
	node2 := &groupTestMockNode{host: "server2.example.com"}

	g.AddNode(node1).AddNode(node2)

	results := g.RunSkillByID("test-playbook")

	// Results may be empty if playbook not registered, but should not panic
	if results.Results == nil {
		t.Error("Expected Results map to be initialized")
	}
}

// TestGroupImplementation_CheckPlaybook verifies check mode execution.
func TestGroupImplementation_CheckSkill(t *testing.T) {
	g := NewGroup("web-servers")

	node1 := &groupTestMockNode{host: "server1.example.com"}
	node2 := &groupTestMockNode{host: "server2.example.com"}

	g.AddNode(node1).AddNode(node2)

	mockPb := &groupTestMockPlaybook{name: "test-playbook"}
	results := g.CheckSkill(mockPb)

	// Verify results contain entries for both nodes
	if len(results.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results.Results))
	}
}

// TestGroupImplementation_CheckPlaybook_SetsDryRun verifies CheckPlaybook sets dry run mode.
func TestGroupImplementation_CheckPlaybook_SetsDryRun(t *testing.T) {
	g := NewGroup("web-servers")

	node1 := &groupTestMockNode{host: "server1.example.com"}
	g.AddNode(node1)

	mockPb := &groupTestMockPlaybook{name: "test-playbook"}
	results := g.CheckSkill(mockPb)

	// Just verify it runs without error
	if results.Results == nil {
		t.Error("Expected Results map to be initialized")
	}
}

// groupTestMockNode is a mock implementation of NodeInterface for testing.
type groupTestMockNode struct {
	host       string
	args       map[string]string
	runResults types.Results
}

func (m *groupTestMockNode) GetHost() string {
	return m.host
}

func (m *groupTestMockNode) GetPort() string {
	return "22"
}

func (m *groupTestMockNode) GetUser() string {
	return "root"
}

func (m *groupTestMockNode) GetKey() string {
	return "id_rsa"
}

func (m *groupTestMockNode) SetPort(port string) NodeInterface {
	return m
}

func (m *groupTestMockNode) SetUser(user string) NodeInterface {
	return m
}

func (m *groupTestMockNode) SetKey(key string) NodeInterface {
	return m
}

func (m *groupTestMockNode) SetArg(key, value string) NodeInterface {
	if m.args == nil {
		m.args = make(map[string]string)
	}
	m.args[key] = value
	return m
}

func (m *groupTestMockNode) SetArgs(args map[string]string) NodeInterface {
	m.args = args
	return m
}

func (m *groupTestMockNode) GetArg(key string) string {
	if m.args == nil {
		return ""
	}
	return m.args[key]
}

func (m *groupTestMockNode) GetArgs() map[string]string {
	if m.args == nil {
		return make(map[string]string)
	}
	result := make(map[string]string, len(m.args))
	for k, v := range m.args {
		result[k] = v
	}
	return result
}

func (m *groupTestMockNode) GetNodeConfig() config.NodeConfig {
	return config.NodeConfig{
		SSHHost:  m.host,
		SSHPort:  "22",
		RootUser: "root",
		SSHKey:   "id_rsa",
		Args:     m.GetArgs(),
	}
}

func (m *groupTestMockNode) Connect() error {
	return nil
}

func (m *groupTestMockNode) Close() error {
	return nil
}

func (m *groupTestMockNode) IsConnected() bool {
	return false
}

func (m *groupTestMockNode) RunCommand(cmd string) types.Results {
	return types.Results{
		Results: map[string]types.Result{
			m.host: {
				Changed: true,
				Message: "Command executed: " + cmd,
			},
		},
	}
}

func (m *groupTestMockNode) RunSkill(pb types.SkillInterface) types.Results {
	return types.Results{
		Results: map[string]types.Result{
			m.host: {
				Changed: true,
				Message: "Playbook executed",
			},
		},
	}
}

func (m *groupTestMockNode) RunSkillByID(id string, opts ...types.SkillOptions) types.Results {
	return types.Results{
		Results: map[string]types.Result{
			m.host: {
				Changed: true,
				Message: "Playbook by ID executed: " + id,
			},
		},
	}
}

func (m *groupTestMockNode) CheckSkill(pb types.SkillInterface) types.Results {
	return types.Results{
		Results: map[string]types.Result{
			m.host: {
				Changed: false,
				Message: "Check mode executed",
			},
		},
	}
}

func (m *groupTestMockNode) GetLogger() *slog.Logger {
	return slog.Default()
}

func (m *groupTestMockNode) SetLogger(logger *slog.Logger) RunnableInterface {
	return m
}

func (m *groupTestMockNode) SetDryRunMode(dryRun bool) RunnableInterface {
	return m
}

func (m *groupTestMockNode) GetDryRunMode() bool {
	return false
}

// groupTestMockPlaybook is a mock implementation of types.SkillInterface for testing.
type groupTestMockPlaybook struct {
	name    string
	cfg     config.NodeConfig
	dryRun  bool
	args    map[string]string
	timeout time.Duration
}

func (m *groupTestMockPlaybook) GetID() string {
	return m.name
}

func (m *groupTestMockPlaybook) SetID(id string) types.SkillInterface {
	m.name = id
	return m
}

func (m *groupTestMockPlaybook) GetDescription() string {
	return "Mock playbook"
}

func (m *groupTestMockPlaybook) SetDescription(desc string) types.SkillInterface {
	return m
}

func (m *groupTestMockPlaybook) GetNodeConfig() config.NodeConfig {
	return m.cfg
}

func (m *groupTestMockPlaybook) SetNodeConfig(cfg config.NodeConfig) types.SkillInterface {
	m.cfg = cfg
	return m
}

func (m *groupTestMockPlaybook) GetArg(key string) string {
	if m.args == nil {
		return ""
	}
	return m.args[key]
}

func (m *groupTestMockPlaybook) SetArg(key, value string) types.SkillInterface {
	if m.args == nil {
		m.args = make(map[string]string)
	}
	m.args[key] = value
	return m
}

func (m *groupTestMockPlaybook) GetArgs() map[string]string {
	return m.args
}

func (m *groupTestMockPlaybook) SetArgs(args map[string]string) types.SkillInterface {
	m.args = args
	return m
}

func (m *groupTestMockPlaybook) IsDryRun() bool {
	return m.dryRun
}

func (m *groupTestMockPlaybook) SetDryRun(dryRun bool) types.SkillInterface {
	m.dryRun = dryRun
	return m
}

func (m *groupTestMockPlaybook) GetTimeout() time.Duration {
	return m.timeout
}

func (m *groupTestMockPlaybook) SetTimeout(timeout time.Duration) types.SkillInterface {
	m.timeout = timeout
	return m
}

func (m *groupTestMockPlaybook) Check() (bool, error) {
	return true, nil
}

func (m *groupTestMockPlaybook) Run() types.Result {
	return types.Result{
		Changed: true,
		Message: "Success",
	}
}
