package ork

import (
	"github.com/dracory/ork/config"
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/playbooks"
	"github.com/dracory/ork/ssh"
)

// sshRunOnce is a variable that points to ssh.RunOnce.
// It can be overridden in tests to mock SSH connections.
var sshRunOnce = ssh.RunOnce

// defaultRegistry is the global playbook registry that holds all built-in
// and user-registered playbooks. It is initialized at package load time
// with all 11 built-in playbooks pre-registered.
//
// Users can add custom playbooks via RegisterPlaybook() and discover
// available playbooks via ListPlaybooks() and GetPlaybook().
var defaultRegistry *playbook.Registry

func init() {
	defaultRegistry = playbook.NewRegistry()

	// Register all 11 built-in playbooks
	defaultRegistry.Register(playbooks.NewPing())
	defaultRegistry.Register(playbooks.NewAptUpdate())
	defaultRegistry.Register(playbooks.NewAptUpgrade())
	defaultRegistry.Register(playbooks.NewAptStatus())
	defaultRegistry.Register(playbooks.NewReboot())
	defaultRegistry.Register(playbooks.NewSwapCreate())
	defaultRegistry.Register(playbooks.NewSwapDelete())
	defaultRegistry.Register(playbooks.NewSwapStatus())
	defaultRegistry.Register(playbooks.NewUserCreate())
	defaultRegistry.Register(playbooks.NewUserDelete())
	defaultRegistry.Register(playbooks.NewUserStatus())
}

// NodeInterface defines the contract for managing a remote server via SSH.
// Implementations must support configuration via setter methods, connection
// management, command execution, and playbook execution.
//
// The interface provides two patterns for configuration:
//   - Fluent builder pattern: Chain setter methods for readable configuration
//   - Getter methods: Inspect current configuration state
//
// Connection management is explicit, allowing users to control the SSH
// connection lifecycle. Operations (Run, Playbook) can work with or without
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
//	output, err := node.Run("uptime")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(output)
//
// Example usage without persistent connection:
//
//	node := ork.NewNode("server.example.com")
//	output, err := node.Run("uptime")  // Creates one-time connection
//
// Example usage with playbooks:
//
//	node := ork.NewNode("server.example.com").
//	    SetArg("username", "alice").
//	    SetArg("shell", "/bin/bash")
//
//	if err := node.Playbook("user-create"); err != nil {
//	    log.Fatal(err)
//	}
type NodeInterface interface {
	// Configuration setters (fluent interface - return self for chaining)

	// Configuration getters

	// GetConfig returns a copy of the underlying config.Config.
	// This allows integration with code that uses the config package directly.
	// The returned configuration includes all accumulated settings (host, port, user, key, args).
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").
	//	    SetPort("2222").
	//	    SetUser("deploy")
	//	cfg := node.GetConfig()
	//	fmt.Printf("Connecting to %s\n", cfg.SSHAddr())
	GetConfig() config.Config

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
	// Arguments are passed to playbooks for configuration.
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
	SetArgs(args map[string]string) NodeInterface

	// Connection management

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

	// Operations

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
	//	output1, _ := node.Run("uptime")
	//	output2, _ := node.Run("df -h")  // Reuses same connection
	//
	// Example without persistent connection:
	//
	//	node := ork.NewNode("server.example.com")
	//	output, err := node.Run("uptime")  // Creates one-time connection
	Run(cmd string) (string, error)

	// Playbook executes a named playbook on the remote server.
	// The playbook is retrieved from the global registry.
	// The current node configuration (including arguments set via SetArg/SetArgs)
	// is passed to the playbook.
	//
	// Returns an error if the playbook is not found in the registry or if
	// execution fails. The error message includes the playbook name and failure reason.
	//
	// Example:
	//
	//	node := ork.NewNode("server.example.com").
	//	    SetArg("username", "alice").
	//	    SetArg("shell", "/bin/bash")
	//
	//	if err := node.Playbook("user-create"); err != nil {
	//	    log.Fatalf("Playbook failed: %v", err)
	//	}
	Playbook(name string) error
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
//	//     cfg: config.Config{
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
func NewNode(host string) NodeInterface {
	return &nodeImplementation{
		cfg: config.Config{
			SSHHost:  host,
			SSHPort:  "22",
			RootUser: "root",
			SSHKey:   "id_rsa",
			Args:     make(map[string]string),
		},
		connected: false,
	}
}
