// Package swap provides playbooks for managing swap files on Linux systems.
// It supports creating, deleting, and checking swap file status with configurable
// size, unit, and swappiness settings.
package swap

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// SwapCreate creates a swap file of the specified size.
// This skill creates a swap file, formats it as swap space, enables it,
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
	*types.BaseSkill
}

// Check determines if swap needs to be created.
// Per the skill interface convention, returns true if swap needs to be
// created (no swap exists), false if swap already exists.
//
// This method runs swapon --show to detect active swap devices.
// An empty output indicates no swap is active.
func (s *SwapCreate) Check() (bool, error) {
	cfg := s.GetNodeConfig()
	cmdCheckExists := types.Command{Command: "swapon --show=NAME --noheadings", Description: "Check if swap exists"}
	output, err := ssh.Run(cfg, cmdCheckExists)
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
func (s *SwapCreate) Run() types.Result {
	cfg := s.GetNodeConfig()

	// Arguments
	sizeStr := s.GetArg(ArgSize)
	unit := s.GetArg(ArgUnit)
	swappiness := s.GetArg(ArgSwappiness)
	swapFilePath := s.GetArg(ArgSwapFilePath)

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
	if swapFilePath == "" {
		swapFilePath = DefaultSwapFilePath
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		return types.Result{
			Changed: false,
			Message: "Invalid swap size",
			Error:   fmt.Errorf("invalid swap size: %s (must be positive integer)", sizeStr),
		}
	}

	// Normalize unit and convert to megabytes
	unit = strings.ToLower(unit)
	var sizeMB int
	var sizeDesc string
	switch unit {
	case "mb", "m":
		sizeMB = size
		sizeDesc = fmt.Sprintf("%dMB", size)
	case "gb", "g":
		sizeMB = size * 1024
		sizeDesc = fmt.Sprintf("%dGB", size)
	default:
		return types.Result{
			Changed: false,
			Message: "Invalid unit",
			Error:   fmt.Errorf("invalid unit: %s (use 'gb' or 'mb')", unit),
		}
	}

	// Check if swap already exists
	needsCreate, err := s.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check swap status",
			Error:   err,
		}
	}

	if !needsCreate {
		return types.Result{
			Changed: false,
			Message: "Swap already exists",
		}
	}

	cfg.GetLoggerOrDefault().Info("creating swap file", "size", sizeDesc, "path", swapFilePath)

	cmdCreate := types.Command{Command: fmt.Sprintf("dd if=/dev/zero of=%s bs=1M count=%d && chmod 600 %s", swapFilePath, sizeMB, swapFilePath), Description: "Create swap file"}
	cmdMakeSwap := types.Command{Command: fmt.Sprintf("mkswap %s", swapFilePath), Description: "Make swap file"}
	cmdCheckFstab := types.Command{Command: fmt.Sprintf("grep -q '%s none swap sw 0 0' /etc/fstab && echo 'exists' || echo 'missing'", swapFilePath), Description: "Check if swap in fstab"}
	cmdAddFstab := types.Command{Command: fmt.Sprintf("echo '%s none swap sw 0 0' | tee -a /etc/fstab", swapFilePath), Description: "Add swap to fstab"}
	cmdSwappiness := types.Command{Command: fmt.Sprintf("sysctl vm.swappiness=%s && grep -q 'vm.swappiness' /etc/sysctl.conf && sed -i 's/vm.swappiness=.*/vm.swappiness=%s/' /etc/sysctl.conf || echo 'vm.swappiness=%s' | tee -a /etc/sysctl.conf", swappiness, swappiness, swappiness), Description: "Configure swappiness"}
	cmdStatus := types.Command{Command: "swapon --show", Description: "Get swap status"}

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		dryRunCreateCmd := fmt.Sprintf("dd if=/dev/zero of=%s bs=1M count=%d && chmod 600 %s", swapFilePath, sizeMB, swapFilePath)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", dryRunCreateCmd)
		cfg.GetLoggerOrDefault().Info("dry-run: would make swap", "cmd", cmdMakeSwap.Command)
		cfg.GetLoggerOrDefault().Info("dry-run: would enable swap", "cmd", fmt.Sprintf("swapon %s", swapFilePath))
		cfg.GetLoggerOrDefault().Info("dry-run: would check fstab", "cmd", cmdCheckFstab.Command)
		cfg.GetLoggerOrDefault().Info("dry-run: would add to fstab", "cmd", cmdAddFstab.Command)
		cfg.GetLoggerOrDefault().Info("dry-run: would configure swappiness", "cmd", cmdSwappiness.Command)
		cfg.GetLoggerOrDefault().Info("dry-run: would get swap status", "cmd", cmdStatus.Command)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would create %s swap file at %s", sizeDesc, swapFilePath),
		}
	}

	output, err := ssh.Run(cfg, cmdCreate)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to create swap",
			Error:   fmt.Errorf("failed to create swap: %w\nOutput: %s", err, output),
		}
	}

	// Make swap file
	_, err = ssh.Run(cfg, cmdMakeSwap)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to make swap",
			Error:   fmt.Errorf("failed to make swap: %w", err),
		}
	}

	// Enable swap
	cmdEnableSwap := types.Command{Command: fmt.Sprintf("swapon %s", swapFilePath), Description: "Enable swap"}
	_, err = ssh.Run(cfg, cmdEnableSwap)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to enable swap",
			Error:   fmt.Errorf("failed to enable swap: %w", err),
		}
	}

	// Add to fstab if not already there using tee for visibility
	output, _ = ssh.Run(cfg, cmdCheckFstab)
	if strings.TrimSpace(output) == "missing" {
		_, err = ssh.Run(cfg, cmdAddFstab)
		if err != nil {
			cfg.GetLoggerOrDefault().Warn("failed to add swap to fstab", "error", err)
		}
	}

	// Configure swappiness
	_, err = ssh.Run(cfg, cmdSwappiness)
	if err != nil {
		cfg.GetLoggerOrDefault().Warn("failed to set swappiness", "error", err)
	}

	// Get final swap status
	status, _ := ssh.Run(cfg, cmdStatus)

	cfg.GetLoggerOrDefault().Info("swap file created", "size", sizeDesc, "path", swapFilePath)
	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("Created %s swap file", sizeDesc),
		Details: map[string]string{
			"size":       sizeDesc,
			"file":       swapFilePath,
			"swappiness": swappiness,
			"status":     status,
		},
	}
}

// SetArgs sets the arguments for swap creation.
// Returns SwapCreate for fluent method chaining.
func (s *SwapCreate) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetArg sets a single argument for swap creation.
// Returns SwapCreate for fluent method chaining.
func (s *SwapCreate) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for swap creation.
// Returns SwapCreate for fluent method chaining.
func (s *SwapCreate) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for swap creation.
// Returns SwapCreate for fluent method chaining.
func (s *SwapCreate) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for swap creation.
// Returns SwapCreate for fluent method chaining.
func (s *SwapCreate) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewSwapCreate creates a new swap-create skill.
//
// Returns:
//
//	A PlaybookInterface implementation configured with IDSwapCreate identifier
//	and description indicating required arguments and defaults.
//
// Default Configuration:
//
//	The returned skill uses these defaults if arguments are not provided:
//	- size: "1" (1 unit)
//	- unit: "gb" (gigabytes)
//	- swappiness: "10" (low swappiness, prefers RAM)
func NewSwapCreate() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDSwapCreate)
	pb.SetDescription("Create a swap file (size via args['size'], unit via args['unit']='gb'|'mb', swappiness via args['swappiness']=10, defaults: 1GB, swappiness=10)")
	return &SwapCreate{BaseSkill: pb}
}
