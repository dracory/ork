package ork

import (
	"fmt"
	"maps"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// nodeImplementation is the default implementation of NodeInterface.
// It wraps config.Config and optionally maintains a persistent SSH connection.
//
// The nodeImplementation struct stores all configuration in a config.Config and tracks
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
	cfg       config.Config
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

// GetConfig returns a copy of the underlying config.Config.
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
//	cfg := node.GetConfig()
//	fmt.Printf("Connecting to %s\n", cfg.SSHAddr())
func (n *nodeImplementation) GetConfig() config.Config {
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
func (n *nodeImplementation) RunCommand(cmd string) (string, error) {
	if n.sshClient != nil && n.connected {
		output, err := n.sshClient.Run(cmd)
		if err != nil {
			return "", fmt.Errorf("failed to execute command '%s': %w", cmd, err)
		}
		return output, nil
	}

	output, err := sshRunOnce(n.cfg.SSHHost, n.cfg.SSHPort, n.cfg.RootUser, n.cfg.SSHKey, cmd)
	if err != nil {
		return "", fmt.Errorf("failed to execute command '%s': %w", cmd, err)
	}
	return output, nil
}

// RunPlaybook executes a named playbook on the remote server.
// The playbook is retrieved from the global registry.
// The current node configuration (including arguments set via SetArg/SetArgs)
// is passed to the playbook.
//
// Optional PlaybookOptions can be provided to override node-level arguments for this
// specific execution. Playbook-level args take precedence over node-level args.
//
// Example:
//
//	node := ork.NewNode("server.example.com")
//
//	// Without options - uses node-level arguments
//	result := node.RunPlaybook("ping")
//
//	// With options - per-playbook arguments override node-level
//	result := node.RunPlaybook("swap-create", playbook.PlaybookOptions{
//	    Args: map[string]string{"size": "2"},
//	})
//
// RunPlaybook executes a named playbook and returns detailed result information.
// This is the preferred method for executing playbooks as it provides idempotency support
// through the Result.Changed field.
func (n *nodeImplementation) RunPlaybook(name string, opts ...playbook.PlaybookOptions) playbook.Result {
	pb, ok := defaultRegistry.Get(name)
	if !ok {
		return playbook.Result{
			Changed: false,
			Message: fmt.Sprintf("playbook '%s' not found in registry", name),
			Error:   fmt.Errorf("playbook '%s' not found in registry", name),
		}
	}

	// Start with node-level config
	cfg := n.cfg

	// If PlaybookOptions provided, merge playbook-level args with node-level args
	// Playbook-level args take precedence
	if len(opts) > 0 {
		mergedArgs := make(map[string]string)

		// Copy node-level args first
		maps.Copy(mergedArgs, n.cfg.Args)

		// Override with playbook-level args
		maps.Copy(mergedArgs, opts[0].Args)

		cfg.Args = mergedArgs
	}

	// Configure playbook using fluent setters
	pb.SetConfig(cfg)
	if len(opts) > 0 {
		pb.SetOptions(&opts[0])
	}
	return pb.Run()
}
