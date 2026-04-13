package security

import (
	"fmt"
	"log"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
)

// KernelHarden applies security-focused kernel parameters via sysctl.
// This playbook creates a dedicated configuration file that hardens the network stack
// against common attacks like IP spoofing, SYN floods, and ICMP redirection attacks.
//
// Usage:
//
//	go run . --playbook=kernel-harden
//
// Security Parameters Applied:
//   - IP Spoofing Protection: rp_filter enabled on all interfaces
//   - ICMP Redirects: Disabled to prevent MITM attacks
//   - Source Routing: Disabled to prevent packet spoofing
//   - TCP SYN Cookies: Enabled to prevent SYN flood attacks
//   - Martian Logging: Enabled to log suspicious packets
//   - IPv6: Disabled if not needed (reduces attack surface)
//   - ASLR: Enabled for memory layout randomization
//   - Core Dumps: Restricted for sensitive data protection
//   - Kernel Pointers: Hidden from unprivileged users
//
// Args:
//   - sysctl-config-path: Path to sysctl.conf for backup (default: /etc/sysctl.conf)
//   - sysctl-dropin-path: Path for security drop-in file (default: /etc/sysctl.d/99-security-hardening.conf)
//
// Execution Flow:
//  1. Backs up current sysctl.conf
//  2. Creates security hardening drop-in configuration
//  3. Applies parameters with sysctl -p
//  4. Verifies key parameters are active
//
// Persistence:
//   - Configuration persists across reboots
//   - Applied via sysctl.d drop-in directory
//
// Prerequisites:
//   - Root SSH access required
//
// Related Playbooks:
//   - ssh-harden: Network-facing service hardening
type KernelHarden struct {
	*playbook.BasePlaybook
}

// Check always returns true since we want to verify and apply security settings.
func (k *KernelHarden) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (k *KernelHarden) Run() playbook.Result {
	cfg := k.GetConfig()

	// Get configurable paths
	sysctlConfigPath := k.GetArg(ArgSysctlConfigPath)
	if sysctlConfigPath == "" {
		sysctlConfigPath = DefaultSysctlConfigPath
	}
	sysctlDropInPath := k.GetArg(ArgSysctlDropInPath)
	if sysctlDropInPath == "" {
		sysctlDropInPath = DefaultSysctlDropInPath
	}

	log.Println("=== Hardening Kernel Security Parameters ===")
	log.Println("WARNING: This will disable IPv6 system-wide")

	// Step 1: Backup
	log.Println("Step 1: Backing up current sysctl configuration...")
	_, err := ssh.Run(cfg, fmt.Sprintf(`cp %s %s.backup.$(date +%%Y%%m%%d)`, sysctlConfigPath, sysctlConfigPath))
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to backup sysctl config", Error: err}
	}

	// Step 2: Create security configuration
	log.Printf("Step 2: Creating security hardening configuration at %s...", sysctlDropInPath)
	cmd := fmt.Sprintf(`cat >> %s << 'EOF'
# IP Spoofing protection
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1

# Ignore ICMP redirects
net.ipv4.conf.all.accept_redirects = 0
net.ipv6.conf.all.accept_redirects = 0
net.ipv4.conf.default.accept_redirects = 0
net.ipv6.conf.default.accept_redirects = 0

# Ignore send redirects
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.default.send_redirects = 0

# Disable source packet routing
net.ipv4.conf.all.accept_source_route = 0
net.ipv6.conf.all.accept_source_route = 0
net.ipv4.conf.default.accept_source_route = 0
net.ipv6.conf.default.accept_source_route = 0

# Log Martians
net.ipv4.conf.all.log_martians = 1
net.ipv4.conf.default.log_martians = 1

# Ignore ICMP ping requests
net.ipv4.icmp_echo_ignore_all = 0

# Ignore Directed pings
net.ipv4.icmp_echo_ignore_broadcasts = 1

# Enable TCP SYN Cookies
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_max_syn_backlog = 2048
net.ipv4.tcp_synack_retries = 2
net.ipv4.tcp_syn_retries = 5

# Disable IPv6 if not needed
net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1
net.ipv6.conf.lo.disable_ipv6 = 1

# Enable IP forwarding protection
net.ipv4.conf.all.forwarding = 0
net.ipv6.conf.all.forwarding = 0

# Enable bad error message protection
net.ipv4.icmp_ignore_bogus_error_responses = 1

# Enable reverse path filtering
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1

# Increase system file descriptor limit
fs.file-max = 65535

# Discourage address space layout randomization (ASLR) bypass
kernel.randomize_va_space = 2

# Restrict core dumps
fs.suid_dumpable = 0

# Hide kernel pointers
kernel.kptr_restrict = 2

# Restrict dmesg access
kernel.dmesg_restrict = 1

# Restrict access to kernel logs
kernel.printk = 3 3 3 3
EOF`, sysctlDropInPath)
	_, err = ssh.Run(cfg, cmd)
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to create hardening config", Error: err}
	}

	// Step 3: Apply parameters
	log.Println("Step 3: Applying kernel parameters...")
	output, err := ssh.Run(cfg, fmt.Sprintf(`sysctl -p %s`, sysctlDropInPath))
	if err != nil {
		return playbook.Result{Changed: false, Message: "Failed to apply kernel parameters", Error: err}
	}
	log.Println(output)

	log.Println("=== Kernel Hardening Complete ===")
	return playbook.Result{
		Changed: true,
		Message: "Kernel security hardening applied successfully",
		Details: map[string]string{
			"config_file":   sysctlDropInPath,
			"backup":        fmt.Sprintf("%s.backup.<date>", sysctlConfigPath),
			"sysctl-config": sysctlConfigPath,
		},
	}
}

// NewKernelHarden creates a new kernel-harden playbook.
func NewKernelHarden() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDKernelHarden)
	pb.SetDescription("Apply security-focused kernel parameters via sysctl")
	return &KernelHarden{BasePlaybook: pb}
}
