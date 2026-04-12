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
type UserCreate struct {
	cfg  config.Config
	opts *playbook.PlaybookOptions
}

// GetID returns the playbook identifier.
func (u *UserCreate) GetID() string {
	return playbook.IDUserCreate
}

// SetID sets the playbook identifier.
func (u *UserCreate) SetID(id string) playbook.Playbook {
	return u
}

// GetDescription returns what this playbook does.
func (u *UserCreate) GetDescription() string {
	return "Create a new user with sudo access (username via args['username'])"
}

// SetDescription sets the playbook description.
func (u *UserCreate) SetDescription(description string) playbook.Playbook {
	return u
}

// GetConfig returns the current node configuration.
func (u *UserCreate) GetConfig() config.Config {
	return u.cfg
}

// SetConfig sets the node configuration for this playbook.
func (u *UserCreate) SetConfig(cfg config.Config) playbook.Playbook {
	u.cfg = cfg
	return u
}

// GetOptions returns the current playbook options.
func (u *UserCreate) GetOptions() *playbook.PlaybookOptions {
	return u.opts
}

// SetOptions sets the playbook-specific options.
func (u *UserCreate) SetOptions(opts *playbook.PlaybookOptions) playbook.Playbook {
	u.opts = opts
	return u
}

// Check determines if user needs to be created.
// Returns true if user doesn't exist, false if user already exists.
func (u *UserCreate) Check() (bool, error) {
	username := u.cfg.GetArg("username")
	if username == "" {
		return false, fmt.Errorf("username is required (pass via --arg=username=value)")
	}
	output, _ := ssh.RunOnce(u.cfg.SSHHost, u.cfg.SSHPort, u.cfg.RootUser, u.cfg.SSHKey, fmt.Sprintf("id %s", username))
	return !strings.Contains(output, username), nil
}

// Run creates the user and returns detailed result.
func (u *UserCreate) Run() playbook.Result {
	username := u.cfg.GetArg("username")
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
	output, err := ssh.RunOnce(u.cfg.SSHHost, u.cfg.SSHPort, u.cfg.RootUser, u.cfg.SSHKey, cmd)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to create user",
			Error:   fmt.Errorf("failed to create user: %w\nOutput: %s", err, output),
		}
	}

	// Add to sudo group
	_, _ = ssh.RunOnce(u.cfg.SSHHost, u.cfg.SSHPort, u.cfg.RootUser, u.cfg.SSHKey, fmt.Sprintf("usermod -aG sudo %s", username))

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
type UserDelete struct {
	cfg  config.Config
	opts *playbook.PlaybookOptions
}

// GetID returns the playbook identifier.
func (u *UserDelete) GetID() string {
	return playbook.IDUserDelete
}

// SetID sets the playbook identifier.
func (u *UserDelete) SetID(id string) playbook.Playbook {
	return u
}

// GetDescription returns what this playbook does.
func (u *UserDelete) GetDescription() string {
	return "Delete a user (username via args['username'])"
}

// SetDescription sets the playbook description.
func (u *UserDelete) SetDescription(description string) playbook.Playbook {
	return u
}

// GetConfig returns the current node configuration.
func (u *UserDelete) GetConfig() config.Config {
	return u.cfg
}

// SetConfig sets the node configuration for this playbook.
func (u *UserDelete) SetConfig(cfg config.Config) playbook.Playbook {
	u.cfg = cfg
	return u
}

// GetOptions returns the current playbook options.
func (u *UserDelete) GetOptions() *playbook.PlaybookOptions {
	return u.opts
}

// SetOptions sets the playbook-specific options.
func (u *UserDelete) SetOptions(opts *playbook.PlaybookOptions) playbook.Playbook {
	u.opts = opts
	return u
}

// Check determines if user exists and can be deleted.
// Returns true if user exists, false if user doesn't exist.
func (u *UserDelete) Check() (bool, error) {
	username := u.cfg.GetArg("username")
	if username == "" {
		return false, fmt.Errorf("username is required (pass via --arg=username=value)")
	}
	output, _ := ssh.RunOnce(u.cfg.SSHHost, u.cfg.SSHPort, u.cfg.RootUser, u.cfg.SSHKey, fmt.Sprintf("id %s", username))
	return strings.Contains(output, username), nil
}

// Run removes the user and returns detailed result.
func (u *UserDelete) Run() playbook.Result {
	username := u.cfg.GetArg("username")
	if username == "" {
		return playbook.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username is required (pass via --arg=username=value)"),
		}
	}

	log.Printf("Deleting user '%s'...", username)

	cmd := fmt.Sprintf("deluser %s 2>/dev/null || userdel %s 2>/dev/null || true", username, username)
	_, err := ssh.RunOnce(u.cfg.SSHHost, u.cfg.SSHPort, u.cfg.RootUser, u.cfg.SSHKey, cmd)
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
type UserStatus struct {
	cfg  config.Config
	opts *playbook.PlaybookOptions
}

// GetID returns the playbook identifier.
func (u *UserStatus) GetID() string {
	return playbook.IDUserStatus
}

// SetID sets the playbook identifier.
func (u *UserStatus) SetID(id string) playbook.Playbook {
	return u
}

// GetDescription returns what this playbook does.
func (u *UserStatus) GetDescription() string {
	return "Show user information"
}

// SetDescription sets the playbook description.
func (u *UserStatus) SetDescription(description string) playbook.Playbook {
	return u
}

// GetConfig returns the current node configuration.
func (u *UserStatus) GetConfig() config.Config {
	return u.cfg
}

// SetConfig sets the node configuration for this playbook.
func (u *UserStatus) SetConfig(cfg config.Config) playbook.Playbook {
	u.cfg = cfg
	return u
}

// GetOptions returns the current playbook options.
func (u *UserStatus) GetOptions() *playbook.PlaybookOptions {
	return u.opts
}

// SetOptions sets the playbook-specific options.
func (u *UserStatus) SetOptions(opts *playbook.PlaybookOptions) playbook.Playbook {
	u.opts = opts
	return u
}

// Check always returns false since UserStatus is read-only.
func (u *UserStatus) Check() (bool, error) {
	return false, nil
}

// Run displays user status and returns detailed result.
func (u *UserStatus) Run() playbook.Result {
	username := u.cfg.GetArg("username")
	if username == "" {
		// Show all users
		output, err := ssh.RunOnce(u.cfg.SSHHost, u.cfg.SSHPort, u.cfg.RootUser, u.cfg.SSHKey, "cat /etc/passwd | grep -E 'bash|zsh' | cut -d: -f1")
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
	output, err := ssh.RunOnce(u.cfg.SSHHost, u.cfg.SSHPort, u.cfg.RootUser, u.cfg.SSHKey, fmt.Sprintf("id %s", username))
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
