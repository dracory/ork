package playbooks

import (
	"fmt"
	"log"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/ssh"
)

// UserCreate creates a new non-root user with sudo access.
type UserCreate struct{}

// Name returns the playbook identifier.
func (u *UserCreate) Name() string {
	return "user-create"
}

// Description returns what this playbook does.
func (u *UserCreate) Description() string {
	return "Create a new user with sudo access (username via args['username'])"
}

// Run creates the user.
func (u *UserCreate) Run(cfg config.Config) error {
	username := cfg.GetArg("username")
	if username == "" {
		return fmt.Errorf("username is required (pass via --arg=username=value)")
	}

	log.Printf("Creating user '%s'...", username)

	// Create user
	cmd := fmt.Sprintf("id %s &>/dev/null || adduser --disabled-password --gecos '' %s", username, username)
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
	if err != nil {
		return fmt.Errorf("failed to create user: %w\nOutput: %s", err, output)
	}

	// Add to sudo group
	_, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, fmt.Sprintf("usermod -aG sudo %s", username))

	log.Printf("User '%s' created with sudo access", username)
	return nil
}

// NewUserCreate creates a new user-create playbook.
func NewUserCreate() *UserCreate {
	return &UserCreate{}
}

// UserDelete removes a user.
type UserDelete struct{}

// Name returns the playbook identifier.
func (u *UserDelete) Name() string {
	return "user-delete"
}

// Description returns what this playbook does.
func (u *UserDelete) Description() string {
	return "Delete a user (username via args['username'])"
}

// Run removes the user.
func (u *UserDelete) Run(cfg config.Config) error {
	username := cfg.GetArg("username")
	if username == "" {
		return fmt.Errorf("username is required (pass via --arg=username=value)")
	}

	log.Printf("Deleting user '%s'...", username)

	cmd := fmt.Sprintf("deluser %s 2>/dev/null || userdel %s 2>/dev/null || true", username, username)
	_, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	log.Printf("User '%s' deleted", username)
	return nil
}

// NewUserDelete creates a new user-delete playbook.
func NewUserDelete() *UserDelete {
	return &UserDelete{}
}

// UserStatus shows user information.
type UserStatus struct{}

// Name returns the playbook identifier.
func (u *UserStatus) Name() string {
	return "user-status"
}

// Description returns what this playbook does.
func (u *UserStatus) Description() string {
	return "Show user information"
}

// Run displays user status.
func (u *UserStatus) Run(cfg config.Config) error {
	username := cfg.GetArg("username")
	if username == "" {
		// Show all users
		output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "cat /etc/passwd | grep -E 'bash|zsh' | cut -d: -f1")
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}
		log.Printf("Shell users:\n%s", output)
		return nil
	}

	// Show specific user
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, fmt.Sprintf("id %s", username))
	if err != nil {
		return fmt.Errorf("user '%s' not found", username)
	}

	log.Printf("User info for '%s':\n%s", username, output)
	return nil
}

// NewUserStatus creates a new user-status playbook.
func NewUserStatus() *UserStatus {
	return &UserStatus{}
}
