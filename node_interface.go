package ork

import (
	"github.com/dracory/ork/config"
)

// NodeInterface defines the contract for managing a remote server via SSH.
// Implementations must support configuration via setter methods, connection
// management, command execution, and playbook execution.
//
// The interface provides two patterns for configuration:
//   - Fluent builder pattern: Chain setter methods for readable configuration
//   - Getter methods: Inspect current configuration state
//
// Connection management is explicit, allowing users to control the SSH
// connection lifecycle. Operations (RunCommand, RunPlaybook) can work with or without
// a persistent connection.
//
// Example usage with fluent builder pattern:
//
//	node := ork.NewNode("server.example.com").
//	    SetPort("2222").
//	    SetUser("deploy").
//	    SetKey("production.prv")
//
//	if err := node.Connect(); err != nil {
//	    log.Fatal(err)
//	}
//	defer node.Close()
//
//	output, err := node.RunCommand("uptime")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(output)
//
// Example usage without persistent connection:
//
//	node := ork.NewNode("server.example.com")
//	output, err := node.RunCommand("uptime")  // Creates one-time connection
//
// Example usage with skills:
//
//	node := ork.NewNode("server.example.com").
//	    SetArg("username", "alice").
//	    SetArg("shell", "/bin/bash")
//
//	if err := node.Run("user-create"); err != nil {
//	    log.Fatal(err)
//	}
type NodeInterface interface {
	// RunnerInterface defines operations that can be performed on the node.
	// NodeInterface embeds RunnerInterface for unified API with Group and Inventory.
	RunnerInterface

	// Configuration setters (fluent interface - return self for chaining)

	// Configuration getters

	// GetArg retrieves a single argument value by key.
	// Returns empty string if the argument is not set.
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").SetArg("username", "alice")
	//	fmt.Println(node.GetArg("username"))  // Output: alice
	GetArg(key string) string

	// GetArgs returns a copy of the entire arguments map.
	// Modifying the returned map will not affect the node's internal state.
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").SetArg("username", "alice")
	//	args := node.GetArgs()
	//	fmt.Println(args["username"])  // Output: alice
	GetArgs() map[string]string

	// GetNodeConfig returns a copy of the underlying config.NodeConfig.
	// This allows integration with code that uses the config package directly.
	// The returned configuration includes all accumulated settings (host, port, user, key, args).
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").
	//	    SetPort("2222").
	//	    SetUser("deploy")
	//	cfg := node.GetNodeConfig()
	//	fmt.Printf("Connecting to %s\n", cfg.SSHAddr())
	GetNodeConfig() config.NodeConfig

	// GetHost returns the configured SSH host (hostname or IP address).
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com")
	//	fmt.Println(node.GetHost())  // Output: server.example.com
	GetHost() string

	// GetUser returns the configured SSH user.
	// Returns "root" if not explicitly set.
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").SetUser("deploy")
	//	fmt.Println(node.GetUser())  // Output: deploy
	GetUser() string

	// GetKey returns the configured SSH private key filename.
	// Returns "id_rsa" if not explicitly set.
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").SetKey("production.prv")
	//	fmt.Println(node.GetKey())  // Output: production.prv
	GetKey() string

	// GetPort returns the configured SSH port.
	// Returns "22" if not explicitly set.
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").SetPort("2222")
	//	fmt.Println(node.GetPort())  // Output: 2222
	GetPort() string

	// SetPort sets the SSH port for the connection.
	// Returns the NodeInterface to enable method chaining.
	// Default is "22" if not set.
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").SetPort("2222")
	SetPort(port string) NodeInterface

	// SetUser sets the SSH user for the connection.
	// Returns the NodeInterface to enable method chaining.
	// Default is "root" if not set.
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").SetUser("deploy")
	SetUser(user string) NodeInterface

	// SetKey sets the SSH private key filename for authentication.
	// The key is resolved to ~/.ssh/<keyname>.
	// Returns the NodeInterface to enable method chaining.
	// Default is "id_rsa" if not set.
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").SetKey("production.prv")
	SetKey(key string) NodeInterface

	// SetArg adds a single argument to the arguments map.
	// This adds to existing arguments without replacing them.
	// Arguments are passed to skills for configuration.
	// Returns the NodeInterface to enable method chaining.
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").
	//	    SetArg("username", "alice").
	//	    SetArg("shell", "/bin/bash")
	SetArg(key, value string) NodeInterface

	// SetArgs replaces the entire arguments map with the provided map.
	// Any existing arguments are discarded.
	// Arguments are passed to skills for configuration.
	// Returns the NodeInterface to enable method chaining.
	//
	// Example:
	//
	//	args := map[string]string{
	//	    "username": "alice",
	//	    "shell": "/bin/bash",
	//	}
	//	node := ork.NewNode("server.example.com").SetArgs(args)
	SetArgs(args map[string]string) NodeInterface

	// Connection management

	// Connect establishes a persistent SSH connection to the remote server.
	// The connection is maintained until Close() is called.
	// Subsequent RunCommand() and RunSkill() calls will reuse this connection.
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
	Connect() error

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
	Close() error

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
	IsConnected() bool
}

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
//	//     cfg: config.NodeConfig{
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
		cfg: config.NodeConfig{
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
		cfg: config.NodeConfig{
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
		connected: false,
	}
}

// NewNodeFromConfig creates a new Node from an existing config.NodeConfig.
// This is useful when you have a pre-built configuration and want to
// create a Node from it directly.
//
// The config is copied internally, so modifications to the original config
// after calling this function will not affect the Node.
//
// Example:
//
//	cfg := config.NodeConfig{
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
func NewNodeFromConfig(cfg config.NodeConfig) NodeInterface {
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
