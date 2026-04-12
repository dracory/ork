package ork

import (
	"fmt"

	"github.com/dracory/ork/playbook"
)

// RunSSH executes a single SSH command on a remote server.
// It connects, runs the command, and disconnects automatically.
//
// The host parameter specifies the remote server (hostname or IP).
// The cmd parameter is the shell command to execute.
// Optional configuration can be provided via functional options.
//
// Default values: port 22, user "root", key "id_rsa", timeout 30s.
//
// Returns the command output as a string and any error that occurred.
// If the command execution fails, the error message includes the command
// and failure reason.
//
// Example:
//
//	output, err := ork.RunSSH("server.example.com", "uptime")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(output)
//
// Example with options:
//
//	output, err := ork.RunSSH("server.example.com", "uptime",
//	    ork.WithPort("2222"),
//	    ork.WithUser("deploy"),
//	    ork.WithKey("production.prv"),
//	)
func RunSSH(host, cmd string, opts ...Option) (string, error) {
	// Build configuration from options
	cfg := applyOptions(host, opts...)

	// Execute command using ssh.RunOnce
	output, err := sshRunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
	if err != nil {
		return "", fmt.Errorf("failed to execute command '%s' on %s:%s: %w", cmd, cfg.SSHHost, cfg.SSHPort, err)
	}

	return output, nil
}

// RunPlaybook executes a named playbook on a remote server.
// The playbook is retrieved from the global registry and executed with the provided configuration.
//
// The name parameter specifies the playbook to execute (e.g., "apt-update", "user-create").
// The host parameter specifies the remote server (hostname or IP).
// Optional configuration can be provided via functional options.
//
// Default values: port 22, user "root", key "id_rsa", timeout 30s.
//
// Returns an error if the playbook is not found in the registry or if execution fails.
// The error message includes the playbook name and failure reason.
//
// Example:
//
//	err := ork.RunPlaybook("apt-update", "server.example.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Example with options and arguments:
//
//	err := ork.RunPlaybook("user-create", "server.example.com",
//	    ork.WithPort("2222"),
//	    ork.WithUser("deploy"),
//	    ork.WithArg("username", "alice"),
//	    ork.WithArg("shell", "/bin/bash"),
//	)
func RunPlaybook(name, host string, opts ...Option) error {
	// Build configuration from options
	cfg := applyOptions(host, opts...)

	// Retrieve playbook from global registry
	pb, ok := defaultRegistry.Get(name)
	if !ok {
		return fmt.Errorf("playbook '%s' not found in registry", name)
	}

	// Execute playbook with configuration
	if err := pb.Run(cfg); err != nil {
		return fmt.Errorf("playbook '%s' failed on %s:%s: %w", name, cfg.SSHHost, cfg.SSHPort, err)
	}

	return nil
}

// ListPlaybooks returns a list of all registered playbook names.
// This includes all built-in playbooks and any custom playbooks registered via RegisterPlaybook.
//
// The returned slice contains playbook names that can be used with RunPlaybook or GetPlaybook.
//
// Example:
//
//	names := ork.ListPlaybooks()
//	for _, name := range names {
//	    fmt.Println(name)
//	}
func ListPlaybooks() []string {
	return defaultRegistry.Names()
}

// GetPlaybook retrieves a playbook by name from the global registry.
// Returns the playbook and true if found, or nil and false if not found.
//
// This is useful for inspecting playbook details (name, description) before execution.
//
// Example:
//
//	pb, ok := ork.GetPlaybook("apt-update")
//	if ok {
//	    fmt.Printf("%s: %s\n", pb.Name(), pb.Description())
//	}
func GetPlaybook(name string) (playbook.Playbook, bool) {
	return defaultRegistry.Get(name)
}

// RegisterPlaybook adds a custom playbook to the global registry.
// Once registered, the playbook can be executed via RunPlaybook or Node.Playbook.
//
// If a playbook with the same name already exists, it will be replaced.
//
// Example:
//
//	customPlaybook := playbook.NewSimplePlaybook(
//	    "custom-task",
//	    "Performs a custom automation task",
//	    func(cfg config.Config) error {
//	        // Custom playbook logic
//	        return nil
//	    },
//	)
//	ork.RegisterPlaybook(customPlaybook)
//
//	// Now it can be used
//	err := ork.RunPlaybook("custom-task", "server.example.com")
func RegisterPlaybook(pb playbook.Playbook) {
	defaultRegistry.Register(pb)
}
