package ssh

import (
	"fmt"
	"os/user"
	"sync"

	"github.com/dracory/ork/types"
)

// runFunc is the function used to execute SSH commands.
// Can be overridden for testing.
var runFunc func(types.NodeConfig, types.Command) (string, error)
var runFuncMu sync.RWMutex

// SetRunFunc sets a custom function for executing SSH commands.
// This is intended for testing purposes only.
// Call with nil to restore the default behavior.
func SetRunFunc(fn func(types.NodeConfig, types.Command) (string, error)) {
	runFuncMu.Lock()
	defer runFuncMu.Unlock()
	runFunc = fn
}

// runSingleCommand is a convenience function that connects, runs a command, and closes.
// Use this for single commands where you don't need to maintain the connection.
// The host parameter should be just the hostname, port is the SSH port (empty defaults to 22).
// This is a lower-level function; prefer using Run() for playbook development.
func runSingleCommand(host, port, user, key string, cmd types.Command) (string, error) {
	client := NewClient(host, port, user, key)
	if err := client.Connect(); err != nil {
		return "", err
	}
	defer client.Close()
	return client.Run(cmd.Command)
}

// PrivateKeyPath constructs the absolute path to an SSH private key file.
// It combines the current user's home directory with the .ssh directory
// and the provided key filename.
func PrivateKeyPath(sshKey string) string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	return usr.HomeDir + "/.ssh/" + sshKey
}

// Run connects to a node using NodeConfig and executes a command.
// It extracts SSH connection settings (SSHHost, SSHPort, SSHLogin, SSHKey)
// from the config and runs the command, returning the output.
//
// If cfg.BecomeUser is set, the command is wrapped with sudo -u <user>.
//
// SAFETY: When cfg.IsDryRunMode is true, this function will NOT execute
// any commands on the server. Instead, it logs the command and returns
// "[dry-run]" as the output. This ensures no accidental changes in dry-run mode.
//
// IMPORTANT: The callers must not rely on the safety net of this function.
// They should handle the dry-run mode themselves before calling this function.
// This is just a final safety net, as a last resort for implementation mistakes.
func Run(cfg types.NodeConfig, cmd types.Command) (string, error) {
	// Check if a test override is set
	runFuncMu.RLock()
	fn := runFunc
	runFuncMu.RUnlock()

	if fn != nil {
		return fn(cfg, cmd)
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run", "command:", cmd.Command, "description:", cmd.Description)
		// Return marker that playbook can detect
		return "[dry-run]", nil
	}

	// Wrap command with sudo if become user is set
	commandToRun := cmd.Command
	if cfg.BecomeUser != "" {
		commandToRun = fmt.Sprintf("sudo -u %s %s", cfg.BecomeUser, cmd.Command)
	}

	return runSingleCommand(cfg.SSHHost, cfg.SSHPort, cfg.SSHLogin, cfg.SSHKey, types.Command{Command: commandToRun, Description: cmd.Description})
}
