package ork

import (
	"fmt"
	"time"

	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// commandImplementation is the default implementation of CommandInterface.
type commandImplementation struct {
	*types.BaseSkill
	command  string
	required bool
	chdir    string
}

// NewCommand creates a new Command with default values.
// The command is not required to succeed by default.
//
// Example:
//
//	command := ork.NewCommand().
//	    SetDescription("Restart application").
//	    SetCommand("pm2 restart app").
//	    SetRequired(true)
//
//	node := ork.NewNodeForHost("server.example.com")
//	result := node.Run(command)
func NewCommand() CommandInterface {
	skill := types.NewBaseSkill()
	skill.SetID("command")
	skill.SetDescription("Execute shell command")

	return &commandImplementation{
		BaseSkill: skill,
		required:  false,
	}
}

// GetBecomeUser returns the configured become user.
func (c *commandImplementation) GetBecomeUser() string {
	return c.BaseSkill.GetBecomeUser()
}

// SetBecomeUser sets the user to become when executing commands.
// Returns CommandInterface for fluent method chaining.
func (c *commandImplementation) SetBecomeUser(user string) types.BecomeInterface {
	c.BaseSkill.SetBecomeUser(user)
	return c
}

// GetID returns the unique identifier for this command.
func (c *commandImplementation) GetID() string {
	return c.BaseSkill.GetID()
}

// SetID sets the unique identifier for this command.
// Returns CommandInterface for fluent method chaining.
func (c *commandImplementation) SetID(id string) types.RunnableInterface {
	c.BaseSkill.SetID(id)
	return c
}

// GetDescription returns a short description of what the command does.
func (c *commandImplementation) GetDescription() string {
	return c.BaseSkill.GetDescription()
}

// SetDescription sets a description of what the command does.
// Returns CommandInterface for fluent method chaining.
func (c *commandImplementation) SetDescription(description string) types.RunnableInterface {
	c.BaseSkill.SetDescription(description)
	return c
}

// GetNodeConfig returns the current node configuration for this command.
func (c *commandImplementation) GetNodeConfig() types.NodeConfig {
	return c.BaseSkill.GetNodeConfig()
}

// SetNodeConfig sets the node configuration for this command execution.
// Returns CommandInterface for fluent method chaining.
func (c *commandImplementation) SetNodeConfig(cfg types.NodeConfig) types.RunnableInterface {
	c.BaseSkill.SetNodeConfig(cfg)
	return c
}

// GetArg retrieves a single argument value by key.
func (c *commandImplementation) GetArg(key string) string {
	return c.BaseSkill.GetArg(key)
}

// SetArg sets a single argument value.
// Returns CommandInterface for fluent method chaining.
func (c *commandImplementation) SetArg(key, value string) types.RunnableInterface {
	c.BaseSkill.SetArg(key, value)
	return c
}

// GetArgs returns the entire arguments map.
func (c *commandImplementation) GetArgs() map[string]string {
	return c.BaseSkill.GetArgs()
}

// SetArgs replaces the entire arguments map.
// Returns CommandInterface for fluent method chaining.
func (c *commandImplementation) SetArgs(args map[string]string) types.RunnableInterface {
	c.BaseSkill.SetArgs(args)
	return c
}

// IsDryRun returns true if this is a dry-run execution.
func (c *commandImplementation) IsDryRun() bool {
	return c.BaseSkill.IsDryRun()
}

// SetDryRun sets whether to simulate execution without making changes.
// Returns CommandInterface for fluent method chaining.
func (c *commandImplementation) SetDryRun(dryRun bool) types.RunnableInterface {
	c.BaseSkill.SetDryRun(dryRun)
	return c
}

// GetTimeout returns the maximum duration for command execution.
func (c *commandImplementation) GetTimeout() time.Duration {
	return c.BaseSkill.GetTimeout()
}

// SetTimeout sets the maximum duration for command execution.
// Returns CommandInterface for fluent method chaining.
func (c *commandImplementation) SetTimeout(timeout time.Duration) types.RunnableInterface {
	c.BaseSkill.SetTimeout(timeout)
	return c
}

// SetCommand sets the shell command to execute.
// Returns CommandInterface for fluent method chaining.
func (c *commandImplementation) SetCommand(cmd string) CommandInterface {
	c.command = cmd
	return c
}

// WithCommand sets the shell command and returns CommandInterface for chaining.
// Alternative to SetCommand for consistent named-method fluent chaining.
func (c *commandImplementation) WithCommand(cmd string) CommandInterface {
	c.command = cmd
	return c
}

// SetRequired sets whether the command must succeed.
// Returns CommandInterface for fluent method chaining.
func (c *commandImplementation) SetRequired(required bool) CommandInterface {
	c.required = required
	return c
}

// WithRequired sets whether the command must succeed and returns CommandInterface for chaining.
// Alternative to SetRequired for consistent named-method fluent chaining.
func (c *commandImplementation) WithRequired(required bool) CommandInterface {
	c.required = required
	return c
}

// SetChdir sets the working directory for command execution.
// The command will be executed as `cd <dir> && <command>`.
// Returns CommandInterface for fluent method chaining.
func (c *commandImplementation) SetChdir(dir string) CommandInterface {
	c.chdir = dir
	return c
}

// WithChdir sets the working directory and returns CommandInterface for chaining.
// Alternative to SetChdir for consistent named-method fluent chaining.
func (c *commandImplementation) WithChdir(dir string) CommandInterface {
	c.chdir = dir
	return c
}

// WithDescription sets a description and returns CommandInterface for chaining.
func (c *commandImplementation) WithDescription(description string) CommandInterface {
	c.BaseSkill.SetDescription(description)
	return c
}

// WithID sets the ID and returns CommandInterface for chaining.
func (c *commandImplementation) WithID(id string) CommandInterface {
	c.BaseSkill.SetID(id)
	return c
}

// WithArg sets a single argument and returns CommandInterface for chaining.
func (c *commandImplementation) WithArg(key, value string) CommandInterface {
	c.BaseSkill.SetArg(key, value)
	return c
}

// WithArgs replaces the arguments map and returns CommandInterface for chaining.
func (c *commandImplementation) WithArgs(args map[string]string) CommandInterface {
	c.BaseSkill.SetArgs(args)
	return c
}

// WithNodeConfig sets the node config and returns CommandInterface for chaining.
func (c *commandImplementation) WithNodeConfig(cfg types.NodeConfig) CommandInterface {
	c.BaseSkill.SetNodeConfig(cfg)
	return c
}

// WithDryRun sets dry-run mode and returns CommandInterface for chaining.
func (c *commandImplementation) WithDryRun(dryRun bool) CommandInterface {
	c.BaseSkill.SetDryRun(dryRun)
	return c
}

// WithTimeout sets the timeout and returns CommandInterface for chaining.
func (c *commandImplementation) WithTimeout(timeout interface{}) CommandInterface {
	if td, ok := timeout.(time.Duration); ok {
		c.BaseSkill.SetTimeout(td)
	}
	return c
}

// WithBecomeUser sets the become user and returns CommandInterface for chaining.
func (c *commandImplementation) WithBecomeUser(user string) CommandInterface {
	c.BaseSkill.SetBecomeUser(user)
	return c
}

// Check always returns false since commands are not idempotent.
// Commands are one-off shell commands that don't have a check phase.
func (c *commandImplementation) Check() (bool, error) {
	return false, nil
}

// Run executes the command using the node configuration.
// Returns a Result with the execution outcome.
// Respects cfg.IsDryRunMode - logs and returns dry-run marker when enabled.
func (c *commandImplementation) Run() types.Result {
	cfg := c.GetNodeConfig()

	// Validate command is set
	if c.command == "" {
		return types.Result{
			Changed: false,
			Message: "Command cannot be empty",
			Error:   fmt.Errorf("command cannot be empty"),
		}
	}

	// Set chdir in node config for ssh.Run to handle
	// Note: We modify the config directly since GetNodeConfig returns a copy
	if c.chdir != "" {
		cfg.Chdir = c.chdir
	}

	// Set become user in node config for ssh.Run to handle
	if c.GetBecomeUser() != "" {
		cfg.BecomeUser = c.GetBecomeUser()
	}

	// Set the modified config back on the BaseSkill for consistency
	if c.chdir != "" || c.GetBecomeUser() != "" {
		c.BaseSkill.SetNodeConfig(cfg)
	}

	// Use the modified cfg for the rest of the method
	// (the local cfg variable already has our modifications)

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run", "command", c.command, "description", c.GetDescription())
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Would execute: %s", c.command),
		}
	}

	// Log the command execution
	cfg.GetLoggerOrDefault().Info("running command", "command", c.command, "description", c.GetDescription())

	// Execute the command
	cmd := types.Command{Command: c.command, Description: c.GetDescription()}
	output, err := ssh.Run(cfg, cmd)

	if err != nil {
		// If command is required, return error
		if c.required {
			return types.Result{
				Changed: false,
				Message: fmt.Sprintf("Command failed (required): %s", c.GetDescription()),
				Error:   fmt.Errorf("command failed: %w", err),
			}
		}

		// If not required, log warning but return success
		cfg.GetLoggerOrDefault().Warn("command failed but not required", "command", c.command, "error", err)
		return types.Result{
			Changed: false,
			Message: fmt.Sprintf("Command failed but not required: %s", c.GetDescription()),
			Details: map[string]string{
				"output": output,
				"error":  err.Error(),
			},
		}
	}

	// Success
	return types.Result{
		Changed: true,
		Message: c.GetDescription(),
		Details: map[string]string{
			"output": output,
		},
	}
}
