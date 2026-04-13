package swap

import (
	"fmt"
	"log"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// SwapStatus shows current swap usage.
type SwapStatus struct {
	*playbook.BasePlaybook
}

// Check always returns false since SwapStatus is read-only.
func (s *SwapStatus) Check() (bool, error) {
	return false, nil
}

// Run displays swap status and returns detailed result.
func (s *SwapStatus) Run() playbook.Result {
	cfg := s.GetConfig()
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "swapon --show")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to get swap status",
			Error:   fmt.Errorf("failed to get swap status: %w", err),
		}
	}

	if strings.TrimSpace(output) == "" {
		log.Println("No swap is currently active")
		return playbook.Result{
			Changed: false,
			Message: "No swap is currently active",
			Details: map[string]string{
				"active": "false",
			},
		}
	}

	log.Printf("Swap status:\n%s", output)
	return playbook.Result{
		Changed: false, // Read-only operation
		Message: "Swap is active",
		Details: map[string]string{
			"active": "true",
			"status": output,
		},
	}
}

// NewSwapStatus creates a new swap-status playbook.
func NewSwapStatus() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDSwapStatus)
	pb.SetDescription("Show swap status and usage")
	return &SwapStatus{BasePlaybook: pb}
}
