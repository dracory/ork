package playbooks

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// SwapCreate creates a swap file of the specified size.
// Size is in GB, default is 1GB if not specified via Args["size"].
type swapCreate struct {
	*playbook.BasePlaybook
}

// Check determines if swap needs to be created.
// Returns true if no swap exists, false if swap already exists.
func (s *swapCreate) Check() (bool, error) {
	cfg := s.GetConfig()
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "swapon --show=NAME --noheadings")
	if err != nil {
		return false, err
	}
	// If output is empty, no swap exists - need to create
	return strings.TrimSpace(output) == "", nil
}

// Run creates the swap file and returns detailed result.
//
// Args:
//   - size: Swap size (default: "1")
//   - unit: "gb" or "mb" (default: "gb")
//   - swappiness: Kernel swappiness 0-100 (default: "10")
func (s *swapCreate) Run() playbook.Result {
	cfg := s.GetConfig()

	// Arguments
	sizeStr := s.GetArg("size")
	unit := s.GetArg("unit")
	swappiness := s.GetArg("swappiness")

	// Defaults
	if sizeStr == "" {
		sizeStr = "1" // Default to 1GB
	}
	if unit == "" {
		unit = "gb" // Default to GB
	}
	if swappiness == "" {
		swappiness = "10"
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		return playbook.Result{
			Changed: false,
			Message: "Invalid swap size",
			Error:   fmt.Errorf("invalid swap size: %s (must be positive integer)", sizeStr),
		}
	}

	// Normalize unit
	unit = strings.ToLower(unit)
	var sizeFlag string
	var sizeDesc string
	switch unit {
	case "mb", "m":
		sizeFlag = "M"
		sizeDesc = fmt.Sprintf("%dMB", size)
	case "gb", "g":
		sizeFlag = "G"
		sizeDesc = fmt.Sprintf("%dGB", size)
	default:
		return playbook.Result{
			Changed: false,
			Message: "Invalid unit",
			Error:   fmt.Errorf("invalid unit: %s (use 'gb' or 'mb')", unit),
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
	cmd := fmt.Sprintf("fallocate -l %d%s /swapfile && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile", size, sizeFlag)
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, cmd)
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to create swap",
			Error:   fmt.Errorf("failed to create swap: %w\nOutput: %s", err, output),
		}
	}

	// Add to fstab if not already there using tee for visibility
	output, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "grep -q '/swapfile' /etc/fstab && echo 'exists' || echo 'missing'")
	if strings.TrimSpace(output) == "missing" {
		_, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "echo '/swapfile none swap sw 0 0' | tee -a /etc/fstab")
		if err != nil {
			log.Printf("Warning: failed to add swap to fstab: %v", err)
		}
	}

	// Configure swappiness
	_, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey,
		fmt.Sprintf("sysctl vm.swappiness=%s && grep -q 'vm.swappiness' /etc/sysctl.conf && sed -i 's/vm.swappiness=.*/vm.swappiness=%s/' /etc/sysctl.conf || echo 'vm.swappiness=%s' | tee -a /etc/sysctl.conf",
			swappiness, swappiness, swappiness))
	if err != nil {
		log.Printf("Warning: failed to set swappiness: %v", err)
	}

	// Get final swap status
	status, _ := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "swapon --show")

	log.Printf("Swap file created successfully (%s)", sizeDesc)
	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("Created %s swap file", sizeDesc),
		Details: map[string]string{
			"size":       sizeDesc,
			"file":       "/swapfile",
			"swappiness": swappiness,
			"status":     status,
		},
	}
}

// NewSwapCreate creates a new swap-create playbook.
func NewSwapCreate() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDSwapCreate)
	pb.SetDescription("Create a swap file (size via args['size'], unit via args['unit']='gb'|'mb', swappiness via args['swappiness']=10, defaults: 1GB, swappiness=10)")
	return &swapCreate{BasePlaybook: pb}
}

// SwapDelete removes the swap file.
type swapDelete struct {
	*playbook.BasePlaybook
}

// Check determines if swap needs to be removed.
// Returns true if swap exists, false if no swap exists.
func (s *swapDelete) Check() (bool, error) {
	cfg := s.GetConfig()
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "swapon --show=NAME --noheadings")
	if err != nil {
		return false, err
	}
	// If output is not empty, swap exists - need to remove
	return strings.TrimSpace(output) != "", nil
}

// Run removes the swap file and returns detailed result.
func (s *swapDelete) Run() playbook.Result {
	cfg := s.GetConfig()
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
	_, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "swapoff /swapfile 2>/dev/null || true")

	// Remove from fstab
	_, _ = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, `sed -i '/\/swapfile/d' /etc/fstab`)

	// Delete file
	_, err = ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "rm -f /swapfile")
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
func NewSwapDelete() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDSwapDelete)
	pb.SetDescription("Remove the swap file")
	return &swapDelete{BasePlaybook: pb}
}

// SwapStatus shows current swap usage.
type swapStatus struct {
	*playbook.BasePlaybook
}

// Check always returns false since SwapStatus is read-only.
func (s *swapStatus) Check() (bool, error) {
	return false, nil
}

// Run displays swap status and returns detailed result.
func (s *swapStatus) Run() playbook.Result {
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
	return &swapStatus{BasePlaybook: pb}
}
