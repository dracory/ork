package ork

import (
	"fmt"
	"log/slog"

	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// NewNode creates a new Node with default configuration values.
// The host parameter specifies the remote server (hostname or IP address).
//
// Default values:
//   - Port: "22"
//   - User: "root"
//   - Key: "id_rsa"
//   - Args: empty map
//
// The returned NodeInterface can be configured using setter methods
// (SetPort, SetUser, SetKey, SetArg, SetArgs) before connecting.
//
// Example:
//
//	node := ork.NewNode("server.example.com")
//	// Equivalent to:
//	// Node{
//	//     cfg: types.NodeConfig{
//	//         SSHHost: "server.example.com",
//	//         SSHPort: "22",
//	//         RootUser: "root",
//	//         SSHKey: "id_rsa",
//	//         Args: map[string]string{},
//	//     },
//	//     connected: false,
//	// }
//
// Example with configuration:
//
//	node := ork.NewNode("server.example.com").
//	    SetPort("2222").
//	    SetUser("deploy").
//	    SetKey("production.prv")
func NewNodeForHost(host string) NodeInterface {
	return &nodeImplementation{
		cfg: types.NodeConfig{
			SSHHost:  host,
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
		connected: false,
	}
}

// NewNode creates a new Node with default configuration values.
// Unlike NewNodeForHost, this function takes no arguments and creates
// a node with an empty host. Use SetArg or SetArgs to configure the node.
//
// Default values:
//   - Host: "" (empty - must be set before connecting)
//   - Port: "22"
//   - User: "root"
//   - Key: "id_rsa"
//   - Args: empty map
//
// Example:
//
//	node := ork.NewNode().
//	    SetHost("server.example.com").
//	    SetPort("2222").
//	    SetUser("deploy")
//
//	if err := node.Connect(); err != nil {
//	    log.Fatal(err)
//	}
func NewNode() NodeInterface {
	return &nodeImplementation{
		cfg: types.NodeConfig{
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
		connected: false,
	}
}

// NewNodeFromConfig creates a new Node from an existing types.NodeConfig.
// This is useful when you have a pre-built configuration and want to
// create a Node from it directly.
//
// The config is copied internally, so modifications to the original config
// after calling this function will not affect the Node.
//
// Example:
//
//	cfg := types.NodeConfig{
//	    SSHHost:  "server.example.com",
//	    SSHPort:  "2222",
//	    RootUser: "deploy",
//	    SSHKey:   "production.prv",
//	    Args: map[string]string{"env": "production"},
//	}
//	node := ork.NewNodeFromConfig(cfg)
//
//	if err := node.Connect(); err != nil {
//	    log.Fatal(err)
//	}
func NewNodeFromConfig(cfg types.NodeConfig) NodeInterface {
	// Create a deep copy of the config to prevent external modifications
	cfgCopy := cfg
	if cfg.Args != nil {
		cfgCopy.Args = make(map[string]string, len(cfg.Args))
		for k, v := range cfg.Args {
			cfgCopy.Args[k] = v
		}
	} else {
		cfgCopy.Args = make(map[string]string)
	}

	return &nodeImplementation{
		cfg:       cfgCopy,
		connected: false,
	}
}

// nodeImplementation is the default implementation of NodeInterface.
// It wraps types.NodeConfig and optionally maintains a persistent SSH connection.
//
// The nodeImplementation struct stores all configuration in a types.NodeConfig and tracks
// connection state. When Connect() is called, it establishes a persistent
// SSH connection that is reused for subsequent operations. When not connected,
// operations create one-time connections.
//
// Example:
//
//	node := ork.NewNode("server.example.com")
//	// node.cfg contains: SSHHost="server.example.com", SSHPort="22",
//	// RootUser="root", SSHKey="id_rsa", Args={}
//	// node.connected is false
//	// node.sshClient is nil
type nodeImplementation struct {
	cfg       types.NodeConfig
	sshClient *ssh.Client
	connected bool
}

// SetPort sets the SSH port for the connection.
// Returns the NodeInterface to enable method chaining.
// Default is "22" if not set.
//
// Example:
//
//	node := ork.NewNode("server.example.com").SetPort("2222")
func (n *nodeImplementation) SetPort(port string) NodeInterface {
	n.cfg.SSHPort = port
	return n
}

// WithPort sets the SSH port and returns NodeInterface for chaining.
// Shortcut alias to SetPort for fluent interface convenience.
func (n *nodeImplementation) WithPort(port string) NodeInterface {
	return n.SetPort(port)
}

// SetUser sets the SSH user for the connection.
// Returns the NodeInterface to enable method chaining.
// Default is "root" if not set.
//
// Example:
//
//	node := ork.NewNode("server.example.com").SetUser("deploy")
func (n *nodeImplementation) SetUser(user string) NodeInterface {
	n.cfg.RootUser = user
	return n
}

// WithUser sets the SSH user and returns NodeInterface for chaining.
// Shortcut alias to SetUser for fluent interface convenience.
func (n *nodeImplementation) WithUser(user string) NodeInterface {
	return n.SetUser(user)
}

// SetKey sets the SSH private key filename for authentication.
// The key is resolved to ~/.ssh/<keyname>.
// Returns the NodeInterface to enable method chaining.
// Default is "id_rsa" if not set.
//
// Example:
//
//	node := ork.NewNode("server.example.com").SetKey("production.prv")
func (n *nodeImplementation) SetKey(key string) NodeInterface {
	n.cfg.SSHKey = key
	return n
}

// WithKey sets the SSH key and returns NodeInterface for chaining.
// Shortcut alias to SetKey for fluent interface convenience.
func (n *nodeImplementation) WithKey(key string) NodeInterface {
	return n.SetKey(key)
}

// SetArg adds a single argument to the arguments map.
// This adds to existing arguments without replacing them.
// Arguments are passed to playbooks for configuration.
// Returns the NodeInterface to enable method chaining.
//
// Example:
//
//	node := ork.NewNode("server.example.com").
//	    SetArg("username", "alice").
//	    SetArg("shell", "/bin/bash")
func (n *nodeImplementation) SetArg(key, value string) NodeInterface {
	if n.cfg.Args == nil {
		n.cfg.Args = make(map[string]string)
	}
	n.cfg.Args[key] = value
	return n
}

// WithArg sets a single argument and returns NodeInterface for chaining.
// Shortcut alias to SetArg for fluent interface convenience.
func (n *nodeImplementation) WithArg(key, value string) NodeInterface {
	return n.SetArg(key, value)
}

// SetArgs replaces the entire arguments map with the provided map.
// Any existing arguments are discarded.
// Arguments are passed to playbooks for configuration.
// Returns the NodeInterface to enable method chaining.
//
// Example:
//
//	args := map[string]string{
//	    "username": "alice",
//	    "shell": "/bin/bash",
//	}
//	node := ork.NewNode("server.example.com").SetArgs(args)
func (n *nodeImplementation) SetArgs(args map[string]string) NodeInterface {
	n.cfg.Args = args
	return n
}

// WithArgs replaces the arguments map and returns NodeInterface for chaining.
// Shortcut alias to SetArgs for fluent interface convenience.
func (n *nodeImplementation) WithArgs(args map[string]string) NodeInterface {
	return n.SetArgs(args)
}

// GetHost returns the configured SSH host (hostname or IP address).
//
// Example:
//
//	node := ork.NewNode("server.example.com")
//	fmt.Println(node.GetHost())  // Output: server.example.com
func (n *nodeImplementation) GetHost() string {
	return n.cfg.SSHHost
}

// GetPort returns the configured SSH port.
// Returns "22" if not explicitly set.
//
// Example:
//
//	node := ork.NewNode("server.example.com").SetPort("2222")
//	fmt.Println(node.GetPort())  // Output: 2222
func (n *nodeImplementation) GetPort() string {
	return n.cfg.SSHPort
}

// GetUser returns the configured SSH user.
// Returns "root" if not explicitly set.
//
// Example:
//
//	node := ork.NewNode("server.example.com").SetUser("deploy")
//	fmt.Println(node.GetUser())  // Output: deploy
func (n *nodeImplementation) GetUser() string {
	return n.cfg.RootUser
}

// GetKey returns the configured SSH private key filename.
// Returns "id_rsa" if not explicitly set.
//
// Example:
//
//	node := ork.NewNode("server.example.com").SetKey("production.prv")
//	fmt.Println(node.GetKey())  // Output: production.prv
func (n *nodeImplementation) GetKey() string {
	return n.cfg.SSHKey
}

// GetArg retrieves a single argument value by key.
// Returns empty string if the argument is not set.
//
// Example:
//
//	node := ork.NewNode("server.example.com").SetArg("username", "alice")
//	fmt.Println(node.GetArg("username"))  // Output: alice
func (n *nodeImplementation) GetArg(key string) string {
	return n.cfg.GetArg(key)
}

// GetArgs returns a copy of the entire arguments map.
// Modifying the returned map will not affect the node's internal state.
//
// Example:
//
//	node := ork.NewNode("server.example.com").SetArg("username", "alice")
//	args := node.GetArgs()
//	fmt.Println(args["username"])  // Output: alice
func (n *nodeImplementation) GetArgs() map[string]string {
	if n.cfg.Args == nil {
		return make(map[string]string)
	}
	argsCopy := make(map[string]string, len(n.cfg.Args))
	for k, v := range n.cfg.Args {
		argsCopy[k] = v
	}
	return argsCopy
}

// GetNodeConfig returns a copy of the underlying types.NodeConfig.
// This allows integration with code that uses the config package directly.
// The returned configuration includes all accumulated settings (host, port, user, key, args).
//
// The returned config is a deep copy to prevent external modification of internal state.
// Modifying the returned config will not affect the Node's internal configuration.
//
// Example:
//
//	node := ork.NewNode("server.example.com").
//	    SetPort("2222").
//	    SetUser("deploy")
//	cfg := node.GetNodeConfig()
//	fmt.Printf("Connecting to %s\n", cfg.SSHAddr())
func (n *nodeImplementation) GetNodeConfig() types.NodeConfig {
	cfgCopy := n.cfg

	if n.cfg.Args != nil {
		cfgCopy.Args = make(map[string]string, len(n.cfg.Args))
		for k, v := range n.cfg.Args {
			cfgCopy.Args[k] = v
		}
	}

	return cfgCopy
}

// Connect establishes a persistent SSH connection to the remote server.
// The connection is maintained until Close() is called.
// Subsequent Run() and Playbook() calls will reuse this connection.
//
// Returns an error if the connection fails, with a descriptive message
// including the host and port.
//
// Example:
//
//	node := ork.NewNode("server.example.com")
//	if err := node.Connect(); err != nil {
//	    log.Fatalf("Failed to connect: %v", err)
//	}
//	defer node.Close()
func (n *nodeImplementation) Connect() error {
	client := ssh.NewClient(n.cfg.SSHHost, n.cfg.SSHPort, n.cfg.RootUser, n.cfg.SSHKey)
	if err := client.Connect(); err != nil {
		return err
	}
	n.sshClient = client
	n.connected = true
	return nil
}

// Close terminates the persistent SSH connection and releases resources.
// After calling Close(), IsConnected() will return false.
// It is safe to call Close() multiple times or on a non-connected node.
//
// Returns an error if closing the connection fails.
//
// Example:
//
//	node := ork.NewNode("server.example.com")
//	node.Connect()
//	defer node.Close()
func (n *nodeImplementation) Close() error {
	if n.sshClient == nil {
		n.connected = false
		return nil
	}
	err := n.sshClient.Close()
	n.sshClient = nil
	n.connected = false
	return err
}

// IsConnected returns true if a persistent SSH connection is currently active.
// Returns false if Connect() has not been called or if Close() was called.
//
// Example:
//
//	node := ork.NewNode("server.example.com")
//	fmt.Println(node.IsConnected())  // Output: false
//	node.Connect()
//	fmt.Println(node.IsConnected())  // Output: true
//	node.Close()
//	fmt.Println(node.IsConnected())  // Output: false
func (n *nodeImplementation) IsConnected() bool {
	return n.connected
}

// Run executes a shell command on the remote server.
// If a persistent connection is active (via Connect()), it is reused.
// Otherwise, a one-time connection is created for this command.
//
// Returns the command output as a string and any error that occurred.
// If the command execution fails, the error message includes the command
// and failure reason.
//
// Example with persistent connection:
//
//	node := ork.NewNode("server.example.com")
//	node.Connect()
//	defer node.Close()
//
//	output1, _ := node.RunCommand("uptime")
//	output2, _ := node.RunCommand("df -h")  // Reuses same connection
//
// Example without persistent connection:
//
//	node := ork.NewNode("server.example.com")
//	output, err := node.RunCommand("uptime")  // Creates one-time connection
func (n *nodeImplementation) RunCommand(cmd string) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	// Check dry-run mode before executing
	if n.cfg.IsDryRunMode {
		n.cfg.GetLoggerOrDefault().Info("dry-run: would run command", "host", n.GetHost(), "command", cmd)
		results.Results[n.GetHost()] = types.Result{
			Changed: true,
			Message: "[dry-run]",
			Error:   nil,
		}
		return results
	}

	var output string
	var err error

	if n.sshClient != nil && n.connected {
		output, err = n.sshClient.Run(cmd)
		if err != nil {
			err = fmt.Errorf("failed to execute command '%s': %w", cmd, err)
		}
	} else {
		output, err = ssh.Run(n.cfg, types.Command{Command: cmd})
		if err != nil {
			err = fmt.Errorf("failed to execute command '%s': %w", cmd, err)
		}
	}

	results.Results[n.GetHost()] = types.Result{
		Changed: true,
		Message: output,
		Error:   err,
	}
	return results
}

// Run executes a skill instance directly and returns detailed result information.
// This is the preferred method for executing skills.
//
// The skill is configured with the node's settings and executed immediately.
// This method allows running custom or programmatically created skills without registry lookup.
func (n *nodeImplementation) Run(skill types.RunnableInterface) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	skill.SetNodeConfig(n.cfg)
	// Propagate node's dry-run mode to skill
	skill.SetDryRun(n.cfg.IsDryRunMode)
	// Propagate node's become user to skill only if skill doesn't already have one
	if skill.GetBecomeUser() == "" {
		skill.SetBecomeUser(n.cfg.BecomeUser)
	}
	result := skill.Run()

	results.Results[n.GetHost()] = types.Result{
		Changed: result.Changed,
		Message: result.Message,
		Details: result.Details,
		Error:   result.Error,
	}
	return results
}

// RunByID executes a skill by ID from the registry.
// This is useful when you want to run skills by string identifier.
//
// Optional RunnableOptions can be provided to override node-level arguments for this
// specific execution. Skill-level args take precedence over node-level args.
func (n *nodeImplementation) RunByID(id string, opts ...types.RunnableOptions) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}

	registry, err := GetGlobalSkillRegistry()
	if err != nil {
		results.Results[n.GetHost()] = types.Result{
			Changed: false,
			Message: fmt.Sprintf("failed to get skill registry: %v", err),
			Error:   fmt.Errorf("failed to get skill registry: %w", err),
		}
		return results
	}

	skill, ok := registry.FindByID(id)
	if !ok {
		results.Results[n.GetHost()] = types.Result{
			Changed: false,
			Message: fmt.Sprintf("skill '%s' not found in registry", id),
			Error:   fmt.Errorf("skill '%s' not found in registry", id),
		}
		return results
	}

	skill.SetNodeConfig(n.cfg)
	// Start with node's dry-run mode, allow opts to override
	skill.SetDryRun(n.cfg.IsDryRunMode)
	// Start with node's become user, only if skill doesn't already have one
	if skill.GetBecomeUser() == "" {
		skill.SetBecomeUser(n.cfg.BecomeUser)
	}
	if len(opts) > 0 {
		skill.SetArgs(opts[0].Args)
		skill.SetDryRun(opts[0].DryRun)
		skill.SetTimeout(opts[0].Timeout)
	}

	result := skill.Run()
	results.Results[n.GetHost()] = types.Result{
		Changed: result.Changed,
		Message: result.Message,
		Details: result.Details,
		Error:   result.Error,
	}
	return results
}

// Check implements RunnerInterface.
// Runs skill in dry-run mode to check if changes are needed.
func (n *nodeImplementation) Check(skill types.RunnableInterface) types.Results {
	results := types.Results{
		Results: make(map[string]types.Result),
	}
	// Use node's dry-run mode setting (may be already set via SetDryRunMode)
	skill.SetDryRun(n.cfg.IsDryRunMode)
	result := skill.Run()
	results.Results[n.GetHost()] = types.Result{
		Changed: result.Changed,
		Message: result.Message,
		Details: result.Details,
		Error:   result.Error,
	}
	return results
}

// GetLogger returns the logger. Returns slog.Default() if not set.
func (n *nodeImplementation) GetLogger() *slog.Logger {
	if n.cfg.Logger == nil {
		return slog.Default()
	}
	return n.cfg.Logger
}

// SetLogger sets a custom logger. Returns RunnerInterface for chaining.
func (n *nodeImplementation) SetLogger(logger *slog.Logger) RunnerInterface {
	n.cfg.Logger = logger
	return n
}

// SetDryRunMode sets whether to simulate execution without making changes.
// When true, ssh.Run() will log commands and return "[dry-run]" marker instead of executing.
// Returns RunnerInterface for fluent method chaining.
func (n *nodeImplementation) SetDryRunMode(dryRun bool) RunnerInterface {
	n.cfg.IsDryRunMode = dryRun
	return n
}

// GetDryRunMode returns true if dry-run mode is enabled.
// When true, commands are logged but not executed on the server.
func (n *nodeImplementation) GetDryRunMode() bool {
	return n.cfg.IsDryRunMode
}

// SetBecomeUser sets the user to become when executing commands via sudo.
// Returns RunnerInterface for fluent method chaining.
func (n *nodeImplementation) SetBecomeUser(user string) types.BecomeInterface {
	n.cfg.BecomeUser = user
	return n
}

// GetBecomeUser returns the configured become user.
// Returns empty string if not set.
func (n *nodeImplementation) GetBecomeUser() string {
	return n.cfg.BecomeUser
}
