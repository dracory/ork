package user

import (
	"fmt"
	"log"
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
	output, _ := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, fmt.Sprintf("id %s", username))
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
func (u *UserCreate) Run() playbook.Result {
	cfg := u.GetConfig()
	username := u.GetArg(ArgUsername)
	sshKey := u.GetArg(ArgSSHKey)
	password := u.GetArg(ArgPassword)

	if username == "" {
		return playbook.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username is required (pass via --arg=username=value)"),
		}
	}

	log.Printf("Creating user '%s'...", username)

	// Create user with home directory and bash shell
	cmd := fmt.Sprintf("id %s &>/dev/null || useradd -m -s /bin/bash %s", username, username)
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to create user",
			Error:   fmt.Errorf("failed to create user: %w\nOutput: %s", err, output),
		}
	}

	// Add to sudo group
	_, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, fmt.Sprintf("usermod -aG sudo %s", username))

	// Set password if provided
	if password != "" {
		cmd = fmt.Sprintf("echo '%s:%s' | chpasswd", username, password)
		output, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
		if err != nil {
			log.Printf("Warning: Failed to set password for user '%s': %v", username, err)
		}
	}

	// Setup SSH key if provided
	if sshKey != "" {
		// Create .ssh directory with proper permissions
		cmd = fmt.Sprintf("mkdir -p /home/%s/.ssh && chmod 700 /home/%s/.ssh", username, username)
		output, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
		if err != nil {
			return playbook.Result{
				Changed: false,
				Message: "Failed to create .ssh directory",
				Error:   fmt.Errorf("failed to create .ssh directory: %w\nOutput: %s", err, output),
			}
		}

		// Add SSH public key to authorized_keys
		cmd = fmt.Sprintf("echo '%s' > /home/%s/.ssh/authorized_keys", sshKey, username)
		output, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
		if err != nil {
			return playbook.Result{
				Changed: false,
				Message: "Failed to add SSH key",
				Error:   fmt.Errorf("failed to add SSH key: %w\nOutput: %s", err, output),
			}
		}

		// Set permissions and ownership on .ssh directory
		cmd = fmt.Sprintf("chmod 600 /home/%s/.ssh/authorized_keys && chown -R %s:%s /home/%s/.ssh", username, username, username, username)
		output, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
		if err != nil {
			log.Printf("Warning: Failed to set permissions on .ssh directory for user '%s': %v", username, err)
		}
	}

	log.Printf("User '%s' created with sudo access", username)
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
