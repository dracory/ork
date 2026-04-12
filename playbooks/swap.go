package playbooks

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dracory/ork/config"
	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// SwapCreate creates a swap file of the specified size.
// Size is in GB, default is 1GB if not specified via Args["size"].
type SwapCreate struct {
	cfg  config.Config
	opts *playbook.PlaybookOptions
}

// GetID returns the playbook identifier.
func (s *SwapCreate) GetID() string {
	return playbook.IDSwapCreate
}

// SetID sets the playbook identifier.
func (s *SwapCreate) SetID(id string) playbook.PlaybookInterface {
	return s
}

// GetDescription returns what this playbook does.
func (s *SwapCreate) GetDescription() string {
	return "Create a swap file (size in GB via args['size'], default 1GB)"
}

// SetDescription sets the playbook description.
func (s *SwapCreate) SetDescription(description string) playbook.PlaybookInterface {
	return s
}

// GetConfig returns the current node configuration.
func (s *SwapCreate) GetConfig() config.Config {
	return s.cfg
}

// GetOptions returns the current playbook options.
func (s *SwapCreate) GetOptions() *playbook.PlaybookOptions {
	return s.opts
}

// SetConfig sets the node configuration for this playbook.
func (s *SwapCreate) SetConfig(cfg config.Config) playbook.PlaybookInterface {
	s.cfg = cfg
	return s
}

// SetOptions sets the playbook-specific options.
func (s *SwapCreate) SetOptions(opts *playbook.PlaybookOptions) playbook.PlaybookInterface {
	s.opts = opts
	return s
}

// Check determines if swap needs to be created.
// Returns true if no swap exists, false if swap already exists.
func (s *SwapCreate) Check() (bool, error) {
	output, err := ssh.RunOnce(s.cfg.SSHHost, s.cfg.SSHPort, s.cfg.RootUser, s.cfg.SSHKey, "swapon --show=NAME --noheadings")
	if err != nil {
		return false, err
	}
	// If output is empty, no swap exists - need to create
	return strings.TrimSpace(output) == "", nil
}

// Run creates the swap file and returns detailed result.
func (s *SwapCreate) Run() playbook.Result {
	sizeStr := s.cfg.GetArgOr("size", "1")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		return playbook.Result{
			Changed: false,
			Message: "Invalid swap size",
			Error:   fmt.Errorf("invalid swap size: %s (must be positive integer)", sizeStr),
		}
	}

	// Check if swap already exists
	needsCreate, err := s.Check()
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to check swap status",
			Error:   err,
		}
	}

	if !needsCreate {
		return playbook.Result{
			Changed: false,
			Message: "Swap already exists",
		}
	}

	log.Printf("Creating %dGB swap file...", size)

	// Create swap file
	cmd := fmt.Sprintf("fallocate -l %dG /swapfile && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile", size)
	output, err := ssh.RunOnce(s.cfg.SSHHost, s.cfg.SSHPort, s.cfg.RootUser, s.cfg.SSHKey, cmd)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to create swap",
			Error:   fmt.Errorf("failed to create swap: %w\nOutput: %s", err, output),
		}
	}

	// Add to fstab if not already there
	output, _ = ssh.RunOnce(s.cfg.SSHHost, s.cfg.SSHPort, s.cfg.RootUser, s.cfg.SSHKey, "grep -q '/swapfile' /etc/fstab && echo 'exists' || echo 'missing'")
	if strings.TrimSpace(output) == "missing" {
		_, err = ssh.RunOnce(s.cfg.SSHHost, s.cfg.SSHPort, s.cfg.RootUser, s.cfg.SSHKey, "echo '/swapfile none swap sw 0 0' >> /etc/fstab")
		if err != nil {
			log.Printf("Warning: failed to add swap to fstab: %v", err)
		}
	}

	log.Printf("Swap file created successfully (%dGB)", size)
	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("Created %dGB swap file", size),
		Details: map[string]string{
			"size_gb": fmt.Sprintf("%d", size),
			"file":    "/swapfile",
		},
	}
}

// NewSwapCreate creates a new swap-create playbook.
func NewSwapCreate() *SwapCreate {
	return &SwapCreate{}
}

// SwapDelete removes the swap file.
type SwapDelete struct {
	cfg  config.Config
	opts *playbook.PlaybookOptions
}

// GetID returns the playbook identifier.
func (s *SwapDelete) GetID() string {
	return playbook.IDSwapDelete
}

// SetID sets the playbook identifier.
func (s *SwapDelete) SetID(id string) playbook.PlaybookInterface {
	return s
}

// GetDescription returns what this playbook does.
func (s *SwapDelete) GetDescription() string {
	return "Remove the swap file"
}

// SetDescription sets the playbook description.
func (s *SwapDelete) SetDescription(description string) playbook.PlaybookInterface {
	return s
}

// GetConfig returns the current node configuration.
func (s *SwapDelete) GetConfig() config.Config {
	return s.cfg
}

// GetOptions returns the current playbook options.
func (s *SwapDelete) GetOptions() *playbook.PlaybookOptions {
	return s.opts
}

// SetConfig sets the node configuration for this playbook.
func (s *SwapDelete) SetConfig(cfg config.Config) playbook.PlaybookInterface {
	s.cfg = cfg
	return s
}

// SetOptions sets the playbook-specific options.
func (s *SwapDelete) SetOptions(opts *playbook.PlaybookOptions) playbook.PlaybookInterface {
	s.opts = opts
	return s
}

// Check determines if swap needs to be removed.
// Returns true if swap exists, false if no swap exists.
func (s *SwapDelete) Check() (bool, error) {
	output, err := ssh.RunOnce(s.cfg.SSHHost, s.cfg.SSHPort, s.cfg.RootUser, s.cfg.SSHKey, "swapon --show=NAME --noheadings")
	if err != nil {
		return false, err
	}
	// If output is not empty, swap exists - need to remove
	return strings.TrimSpace(output) != "", nil
}

// Run removes the swap file and returns detailed result.
func (s *SwapDelete) Run() playbook.Result {
	// Check if swap exists
	needsDelete, err := s.Check()
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to check swap status",
			Error:   err,
		}
	}

	if !needsDelete {
		return playbook.Result{
			Changed: false,
			Message: "No swap file to remove",
		}
	}

	log.Println("Removing swap file...")

	// Turn off swap
	_, _ = ssh.RunOnce(s.cfg.SSHHost, s.cfg.SSHPort, s.cfg.RootUser, s.cfg.SSHKey, "swapoff /swapfile 2>/dev/null || true")

	// Remove from fstab
	_, _ = ssh.RunOnce(s.cfg.SSHHost, s.cfg.SSHPort, s.cfg.RootUser, s.cfg.SSHKey, `sed -i '/\/swapfile/d' /etc/fstab`)

	// Delete file
	_, err = ssh.RunOnce(s.cfg.SSHHost, s.cfg.SSHPort, s.cfg.RootUser, s.cfg.SSHKey, "rm -f /swapfile")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to remove swap file",
			Error:   fmt.Errorf("failed to remove swap file: %w", err),
		}
	}

	log.Println("Swap file removed successfully")
	return playbook.Result{
		Changed: true,
		Message: "Swap file removed",
		Details: map[string]string{
			"file": "/swapfile",
		},
	}
}

// NewSwapDelete creates a new swap-delete playbook.
func NewSwapDelete() *SwapDelete {
	return &SwapDelete{}
}

// SwapStatus shows current swap usage.
type SwapStatus struct {
	cfg  config.Config
	opts *playbook.PlaybookOptions
}

// GetID returns the playbook identifier.
func (s *SwapStatus) GetID() string {
	return playbook.IDSwapStatus
}

// SetID sets the playbook identifier.
func (s *SwapStatus) SetID(id string) playbook.PlaybookInterface {
	return s
}

// GetDescription returns what this playbook does.
func (s *SwapStatus) GetDescription() string {
	return "Show swap status and usage"
}

// SetDescription sets the playbook description.
func (s *SwapStatus) SetDescription(description string) playbook.PlaybookInterface {
	return s
}

// GetConfig returns the current node configuration.
func (s *SwapStatus) GetConfig() config.Config {
	return s.cfg
}

// GetOptions returns the current playbook options.
func (s *SwapStatus) GetOptions() *playbook.PlaybookOptions {
	return s.opts
}

// SetConfig sets the node configuration for this playbook.
func (s *SwapStatus) SetConfig(cfg config.Config) playbook.PlaybookInterface {
	s.cfg = cfg
	return s
}

// SetOptions sets the playbook-specific options.
func (s *SwapStatus) SetOptions(opts *playbook.PlaybookOptions) playbook.PlaybookInterface {
	s.opts = opts
	return s
}

// Check always returns false since SwapStatus is read-only.
func (s *SwapStatus) Check() (bool, error) {
	return false, nil
}

// Run displays swap status and returns detailed result.
func (s *SwapStatus) Run() playbook.Result {
	output, err := ssh.RunOnce(s.cfg.SSHHost, s.cfg.SSHPort, s.cfg.RootUser, s.cfg.SSHKey, "swapon --show")
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
func NewSwapStatus() *SwapStatus {
	return &SwapStatus{}
}
