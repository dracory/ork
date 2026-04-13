package user

// Package user documentation is in status.go

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// UserCreate creates a new non-root user with sudo access.
type UserCreate struct {
	*playbook.BasePlaybook
}

// Check determines if user needs to be created.
// Returns true if user doesn't exist, false if user already exists.
func (u *UserCreate) Check() (bool, error) {
	cfg := u.GetConfig()
	username := u.GetArg(ArgUsername)
	if username == "" {
		return false, fmt.Errorf("username is required (pass via --arg=username=value)")
	}
	output, _ := ssh.Run(cfg, fmt.Sprintf("id %s", username))
	return !strings.Contains(output, username), nil
}

// Run creates a non-root system user with sudo privileges and SSH access.
// This playbook sets up a standard user account suitable for day-to-day server management,
// including proper SSH key authentication and sudo access for administrative tasks.
//
// Usage:
//
//	go run . --playbook=user-create [--arg=username=<name>] [--arg=ssh-key=<public_key>] [--arg=password=<password>]
//
// Parameters (passed via args):
//   - username: Name of the user to create (required, via --arg=username=<name>)
//   - ssh-key: Public SSH key to add to authorized_keys (for key-based authentication)
//   - password: Initial password for the user (optional, less secure than SSH keys)
//
// Execution Flow:
//  1. Validates username parameter (required)
//  2. Creates user with home directory and bash shell using useradd
//  3. Adds user to sudo group for administrative privileges
//  4. Sets password if provided (uses chpasswd for secure handling)
//  5. Creates .ssh directory with proper permissions (700)
//  6. Adds SSH public key to authorized_keys
//  7. Sets secure permissions on authorized_keys (600)
//  8. Changes ownership of .ssh directory to new user
//
// Security Considerations:
//   - SSH key authentication is preferred over password authentication
//   - Passwords are set via chpasswd which handles hashing automatically
//   - .ssh directory has restrictive permissions (700)
//   - authorized_keys file has restrictive permissions (600)
//
// Prerequisites:
//   - Root SSH access required for user creation
//   - User must not already exist (checked via id command)
//
// Args:
//   - username (string, required): Username to create
//   - ssh-key (string, optional): SSH public key for authorized_keys
//   - password (string, optional): Initial password for the user
//   - shell (string, optional): Login shell (default: /bin/bash)
//   - group (string, optional): Primary group (default: same as username)
//   - sudo-group (string, optional): Sudo/admin group name (default: sudo)
//   - home-dir (string, optional): Home directory path (default: /home/<username>)
func (u *UserCreate) Run() playbook.Result {
	cfg := u.GetConfig()
	username := u.GetArg(ArgUsername)
	sshKey := u.GetArg(ArgSSHKey)
	password := u.GetArg(ArgPassword)
	shell := u.GetArg(ArgShell)
	group := u.GetArg(ArgGroup)
	sudoGroup := u.GetArg(ArgSudoGroup)

	// Apply defaults
	if shell == "" {
		shell = DefaultShell
	}
	if sudoGroup == "" {
		sudoGroup = DefaultSudoGroup
	}

	if username == "" {
		return playbook.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username is required (pass via --arg=username=value)"),
		}
	}

	cfg.GetLoggerOrDefault().Info("creating user", "username", username)

	// Build useradd command with options
	useraddOpts := fmt.Sprintf("-m -s %s", shell)
	if group != "" {
		useraddOpts = fmt.Sprintf("%s -g %s", useraddOpts, group)
	}
	cmd := fmt.Sprintf("id %s &>/dev/null || useradd %s %s", username, useraddOpts, username)
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to create user",
			Error:   fmt.Errorf("failed to create user: %w\nOutput: %s", err, output),
		}
	}

	// Add to sudo group
	_, _ = ssh.Run(cfg, fmt.Sprintf("usermod -aG %s %s", sudoGroup, username))

	// Set password if provided
	if password != "" {
		cmd = fmt.Sprintf("echo '%s:%s' | chpasswd", username, password)
		output, err = ssh.Run(cfg, cmd)
		if err != nil {
			cfg.GetLoggerOrDefault().Warn("failed to set password", "username", username, "error", err)
		}
	}

	// Setup SSH key if provided
	if sshKey != "" {
		// Determine home directory
		homeDir := u.GetArg(ArgHomeDir)
		if homeDir == "" {
			homeDir = fmt.Sprintf("/home/%s", username)
		}

		// Create .ssh directory with proper permissions
		cmd = fmt.Sprintf("mkdir -p %s/.ssh && chmod 700 %s/.ssh", homeDir, homeDir)
		output, err = ssh.Run(cfg, cmd)
		if err != nil {
			return playbook.Result{
				Changed: false,
				Message: "Failed to create .ssh directory",
				Error:   fmt.Errorf("failed to create .ssh directory: %w\nOutput: %s", err, output),
			}
		}

		// Add SSH public key to authorized_keys
		cmd = fmt.Sprintf("echo '%s' > %s/.ssh/authorized_keys", sshKey, homeDir)
		output, err = ssh.Run(cfg, cmd)
		if err != nil {
			return playbook.Result{
				Changed: false,
				Message: "Failed to add SSH key",
				Error:   fmt.Errorf("failed to add SSH key: %w\nOutput: %s", err, output),
			}
		}

		// Set permissions and ownership on .ssh directory
		cmd = fmt.Sprintf("chmod 600 %s/.ssh/authorized_keys && chown -R %s:%s %s/.ssh", homeDir, username, username, homeDir)
		output, err = ssh.Run(cfg, cmd)
		if err != nil {
			cfg.GetLoggerOrDefault().Warn("failed to set permissions on .ssh directory", "username", username, "error", err)
		}
	}

	cfg.GetLoggerOrDefault().Info("user created with sudo access", "username", username)
	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("User '%s' created with sudo access", username),
	}
}

// NewUserCreate creates a new user-create playbook.
func NewUserCreate() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDUserCreate)
	pb.SetDescription("Create a new user with sudo access (username via args['username'])")
	return &UserCreate{BasePlaybook: pb}
}
