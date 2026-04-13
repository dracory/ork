package user

import (
	"fmt"
	"log"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// UserStatus shows user information.
type UserStatus struct {
	*playbook.BasePlaybook
}

// Check always returns false since UserStatus is read-only.
func (u *UserStatus) Check() (bool, error) {
	return false, nil
}

// Run displays user status and returns detailed result.
func (u *UserStatus) Run() playbook.Result {
	cfg := u.GetConfig()
	username := u.GetArg(ArgUsername)

	if username != "" {
		// Check specific user
		log.Printf("Checking user: %s", username)

		cmd := fmt.Sprintf("id %s", username)
		output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
		if err != nil {
			return playbook.Result{
				Changed: false,
				Message: fmt.Sprintf("User '%s' does not exist", username),
				Error:   fmt.Errorf("user '%s' not found", username),
			}
		}
		log.Println(output)

		// Check if user has sudo
		cmd = fmt.Sprintf("groups %s", username)
		groupsOutput, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
		if err == nil {
			log.Printf("Groups: %s", groupsOutput)
		}

		return playbook.Result{
			Changed: false,
			Message: fmt.Sprintf("User info for '%s'", username),
			Details: map[string]string{"info": output, "groups": groupsOutput},
		}
	}

	// List all non-system users
	log.Println("Listing all system users...")

	cmd := "awk -F: '$3 >= 1000 && $3 < 65534 {print $1}' /etc/passwd"
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to list users",
			Error:   fmt.Errorf("failed to list users: %w", err),
		}
	}

	if output == "" {
		log.Println("No non-system users found")
		return playbook.Result{
			Changed: false,
			Message: "No non-system users found",
		}
	}

	log.Println("Users:")
	log.Println(output)
	return playbook.Result{
		Changed: false,
		Message: "Non-system users listed",
		Details: map[string]string{"users": output},
	}
}

// NewUserStatus creates a new user-status playbook.
func NewUserStatus() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDUserStatus)
	pb.SetDescription("Show user information")
	return &UserStatus{BasePlaybook: pb}
}
