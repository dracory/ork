package playbook

import (
	"strings"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/ssh"
)

// CheckExists runs a check command and returns true if the command succeeds.
// This is useful for checking if a file exists, a service is running, etc.
// Returns false if the command fails or produces no output.
func CheckExists(client *ssh.Client, checkCmd string) bool {
	output, err := client.Run(checkCmd)
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) != ""
}

// EnsureState ensures a desired state by running a check command first.
// If the check fails, it runs the apply command to achieve the desired state.
// Returns true if changes were made (apply was run), false if no changes needed.
// Returns an error if either command fails.
func EnsureState(client *ssh.Client, checkCmd, applyCmd string) (bool, error) {
	// Check if already in desired state
	output, err := client.Run(checkCmd)
	if err == nil && strings.TrimSpace(output) != "" {
		// Already in desired state
		return false, nil
	}

	// Apply the change
	_, err = client.Run(applyCmd)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Execute runs a playbook and returns a Result.
// This is a convenience wrapper that calls pb.Run(cfg).
func Execute(pb Playbook, cfg config.Config) Result {
	return pb.Run(cfg)
}
