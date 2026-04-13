// Package swap provides playbooks for managing swap files on Linux systems.
// It supports creating, deleting, and checking swap file status with configurable
// size, unit, and swappiness settings.
package swap

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// SwapCreate creates a swap file of the specified size.
// This playbook creates a swap file, formats it as swap space, enables it,
// adds it to /etc/fstab for persistence across reboots, and configures
// the kernel swappiness parameter.
//
// Usage:
//
//	go run . --playbook=swap-create [--arg=size=2] [--arg=unit=gb] [--arg=swappiness=10]
//
// Arguments:
//   - size: Swap file size as integer (default: "1")
//   - unit: Size unit, either "gb" or "mb" (default: "gb")
//   - swappiness: Kernel swappiness value 0-100 (default: "10")
//
// Execution Flow:
//  1. Validates size argument is a positive integer
//  2. Validates unit is "gb" or "mb"
//  3. Checks if swap already exists (uses Check method)
//  4. Creates swap file with fallocate at /swapfile
//  5. Sets secure permissions (chmod 600)
//  6. Formats as swap with mkswap
//  7. Enables swap with swapon
//  8. Adds entry to /etc/fstab for persistence
//  9. Configures vm.swappiness via sysctl and /etc/sysctl.conf
//  10. Reports success with swap status
//
// Expected Output:
//   - Success: "Created <size> swap file" with details
//   - Failure: Error with specific failure reason
//
// Result Details:
//   - size: Human-readable size (e.g., "2GB", "512MB")
//   - file: Path to swap file ("/swapfile")
//   - swappiness: Configured swappiness value
//   - status: Output from swapon --show command
//
// Use Cases:
//   - Add swap to memory-constrained systems
//   - Configure swap for database workloads (low swappiness)
//   - Initial server setup
//
// Idempotency:
//   - Reports Changed=false if swap already exists
//   - Reports Changed=true when new swap is created
//
// Safety:
//   - Validates all arguments before making system changes
//   - Uses secure permissions on swap file (600)
//   - Falls back gracefully if fstab or sysctl updates fail (logs warning)
type SwapCreate struct {
	*playbook.BasePlaybook
}

// Check determines if swap needs to be created.
// Per the playbook interface convention, returns true if swap needs to be
// created (no swap exists), false if swap already exists.
//
// This method runs swapon --show to detect active swap devices.
// An empty output indicates no swap is active.
func (s *SwapCreate) Check() (bool, error) {
	cfg := s.GetConfig()
	output, err := ssh.RunOnce(cfg.SSHHost, cfg.SSHPort, cfg.RootUser, cfg.SSHKey, "swapon --show=NAME --noheadings")
	if err != nil {
		return false, err
	}
	// If output is empty, no swap exists - need to create
	return strings.TrimSpace(output) == "", nil
}

// Run creates the swap file and returns detailed result.
// Changed is true when a new swap file is created, false if swap already exists.
//
// This method reads arguments using ArgSize, ArgUnit, and ArgSwappiness constants,
// applying defaults from DefaultSize, DefaultUnit, and DefaultSwappiness when
// arguments are not provided.
//
// Result.Details contains:
//   - size: Human-readable swap size (e.g., "2GB")
//   - file: Swap file path ("/swapfile")
//   - swappiness: Configured kernel swappiness value
//   - status: Output from swapon --show showing active swap
func (s *SwapCreate) Run() playbook.Result {
	cfg := s.GetConfig()

	// Arguments
	sizeStr := s.GetArg(ArgSize)
	unit := s.GetArg(ArgUnit)
	swappiness := s.GetArg(ArgSwappiness)

	// Defaults
	if sizeStr == "" {
		sizeStr = DefaultSize
	}
	if unit == "" {
		unit = DefaultUnit
	}
	if swappiness == "" {
		swappiness = DefaultSwappiness
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

	log.Printf("Creating %s swap file...", sizeDesc)

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
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDSwapCreate identifier
//	and description indicating required arguments and defaults.
//
// Default Configuration:
//
//	The returned playbook uses these defaults if arguments are not provided:
//	- size: "1" (1 unit)
//	- unit: "gb" (gigabytes)
//	- swappiness: "10" (low swappiness, prefers RAM)
func NewSwapCreate() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDSwapCreate)
	pb.SetDescription("Create a swap file (size via args['size'], unit via args['unit']='gb'|'mb', swappiness via args['swappiness']=10, defaults: 1GB, swappiness=10)")
	return &SwapCreate{BasePlaybook: pb}
}
