package playbooks

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/ssh"
)

// SwapCreate creates a swap file of the specified size.
// Size is in GB, default is 1GB if not specified via Args["size"].
type SwapCreate struct{}

// Name returns the playbook identifier.
func (s *SwapCreate) Name() string {
	return "swap-create"
}

// Description returns what this playbook does.
func (s *SwapCreate) Description() string {
	return "Create a swap file (size in GB via args['size'], default 1GB)"
}

// Run creates the swap file.
func (s *SwapCreate) Run(cfg config.Config) error {
	sizeStr := cfg.GetArgOr("size", "1")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		return fmt.Errorf("invalid swap size: %s (must be positive integer)", sizeStr)
	}

	log.Printf("Creating %dGB swap file...", size)

	// Check if swap already exists
	output, _ := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "swapon --show=NAME --noheadings")
	if strings.TrimSpace(output) != "" {
		return fmt.Errorf("swap already exists: %s", strings.TrimSpace(output))
	}

	// Create swap file
	cmd := fmt.Sprintf("fallocate -l %dG /swapfile && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile", size)
	output, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
	if err != nil {
		return fmt.Errorf("failed to create swap: %w\nOutput: %s", err, output)
	}

	// Add to fstab if not already there
	output, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "grep -q '/swapfile' /etc/fstab && echo 'exists' || echo 'missing'")
	if strings.TrimSpace(output) == "missing" {
		_, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "echo '/swapfile none swap sw 0 0' >> /etc/fstab")
		if err != nil {
			log.Printf("Warning: failed to add swap to fstab: %v", err)
		}
	}

	log.Printf("Swap file created successfully (%dGB)", size)
	return nil
}

// NewSwapCreate creates a new swap-create playbook.
func NewSwapCreate() *SwapCreate {
	return &SwapCreate{}
}

// SwapDelete removes the swap file.
type SwapDelete struct{}

// Name returns the playbook identifier.
func (s *SwapDelete) Name() string {
	return "swap-delete"
}

// Description returns what this playbook does.
func (s *SwapDelete) Description() string {
	return "Remove the swap file"
}

// Run removes the swap file.
func (s *SwapDelete) Run(cfg config.Config) error {
	log.Println("Removing swap file...")

	// Turn off swap
	_, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "swapoff /swapfile 2>/dev/null || true")

	// Remove from fstab
	_, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, `sed -i '/\/swapfile/d' /etc/fstab`)

	// Delete file
	_, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "rm -f /swapfile")
	if err != nil {
		return fmt.Errorf("failed to remove swap file: %w", err)
	}

	log.Println("Swap file removed successfully")
	return nil
}

// NewSwapDelete creates a new swap-delete playbook.
func NewSwapDelete() *SwapDelete {
	return &SwapDelete{}
}

// SwapStatus shows current swap usage.
type SwapStatus struct{}

// Name returns the playbook identifier.
func (s *SwapStatus) Name() string {
	return "swap-status"
}

// Description returns what this playbook does.
func (s *SwapStatus) Description() string {
	return "Show swap status and usage"
}

// Run displays swap status.
func (s *SwapStatus) Run(cfg config.Config) error {
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "swapon --show")
	if err != nil {
		return fmt.Errorf("failed to get swap status: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		log.Println("No swap is currently active")
	} else {
		log.Printf("Swap status:\n%s", output)
	}

	return nil
}

// NewSwapStatus creates a new swap-status playbook.
func NewSwapStatus() *SwapStatus {
	return &SwapStatus{}
}
