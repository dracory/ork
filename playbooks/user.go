package playbooks

import (
	"fmt"
	"log"
	"strings"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// UserCreate creates a new non-root user with sudo access.
type UserCreate struct{}

// Name returns the playbook identifier.
func (u *UserCreate) Name() string {
	return playbook.NameUserCreate
}

// Description returns what this playbook does.
func (u *UserCreate) Description() string {
	return "Create a new user with sudo access (username via args['username'])"
}

// Check determines if user needs to be created.
// Returns true if user doesn't exist, false if user already exists.
func (u *UserCreate) Check(cfg config.Config) (bool, error) {
	username := cfg.GetArg("username")
	if username == "" {
		return false, fmt.Errorf("username is required (pass via --arg=username=value)")
	}
	output, _ := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, fmt.Sprintf("id %s", username))
	return !strings.Contains(output, username), nil
}

// Run creates the user and returns detailed result.
func (u *UserCreate) Run(cfg config.Config) playbook.Result {
	username := cfg.GetArg("username")
	if username == "" {
		return playbook.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username is required (pass via --arg=username=value)"),
		}
	}

	log.Printf("Creating user '%s'...", username)

	// Create user
	cmd := fmt.Sprintf("id %s &>/dev/null || adduser --disabled-password --gecos '' %s", username, username)
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

	log.Printf("User '%s' created with sudo access", username)
	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("User '%s' created with sudo access", username),
	}
}

// NewUserCreate creates a new user-create playbook.
func NewUserCreate() *UserCreate {
	return &UserCreate{}
}

// UserDelete removes a user.
type UserDelete struct{}

// Name returns the playbook identifier.
func (u *UserDelete) Name() string {
	return playbook.NameUserDelete
}

// Description returns what this playbook does.
func (u *UserDelete) Description() string {
	return "Delete a user (username via args['username'])"
}

// Check determines if user exists and can be deleted.
// Returns true if user exists, false if user doesn't exist.
func (u *UserDelete) Check(cfg config.Config) (bool, error) {
	username := cfg.GetArg("username")
	if username == "" {
		return false, fmt.Errorf("username is required (pass via --arg=username=value)")
	}
	output, _ := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, fmt.Sprintf("id %s", username))
	return strings.Contains(output, username), nil
}

// Run removes the user and returns detailed result.
func (u *UserDelete) Run(cfg config.Config) playbook.Result {
	username := cfg.GetArg("username")
	if username == "" {
		return playbook.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username is required (pass via --arg=username=value)"),
		}
	}

	log.Printf("Deleting user '%s'...", username)

	cmd := fmt.Sprintf("deluser %s 2>/dev/null || userdel %s 2>/dev/null || true", username, username)
	_, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to delete user",
			Error:   fmt.Errorf("failed to delete user: %w", err),
		}
	}

	log.Printf("User '%s' deleted", username)
	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("User '%s' deleted", username),
	}
}

// NewUserDelete creates a new user-delete playbook.
func NewUserDelete() *UserDelete {
	return &UserDelete{}
}

// UserStatus shows user information.
type UserStatus struct{}

// Name returns the playbook identifier.
func (u *UserStatus) Name() string {
	return playbook.NameUserStatus
}

// Description returns what this playbook does.
func (u *UserStatus) Description() string {
	return "Show user information"
}

// Check always returns false since UserStatus is read-only.
func (u *UserStatus) Check(cfg config.Config) (bool, error) {
	return false, nil
}

// Run displays user status and returns detailed result.
func (u *UserStatus) Run(cfg config.Config) playbook.Result {
	username := cfg.GetArg("username")
	if username == "" {
		// Show all users
		output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "cat /etc/passwd | grep -E 'bash|zsh' | cut -d: -f1")
		if err != nil {
			return playbook.Result{
				Changed: false,
				Message: "Failed to list users",
				Error:   fmt.Errorf("failed to list users: %w", err),
			}
		}
		log.Printf("Shell users:\n%s", output)
		return playbook.Result{
			Changed: false,
			Message: "Shell users listed",
			Details: map[string]string{"users": output},
		}
	}

	// Show specific user
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, fmt.Sprintf("id %s", username))
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: fmt.Sprintf("User '%s' not found", username),
			Error:   fmt.Errorf("user '%s' not found", username),
		}
	}

	log.Printf("User info for '%s':\n%s", username, output)
	return playbook.Result{
		Changed: false,
		Message: fmt.Sprintf("User info for '%s'", username),
		Details: map[string]string{"info": output},
	}
}

// NewUserStatus creates a new user-status playbook.
func NewUserStatus() *UserStatus {
	return &UserStatus{}
}
