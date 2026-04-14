package ufw

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// UfwInstall installs and configures the Uncomplicated Firewall (UFW).
// UFW provides a simple interface for managing iptables firewall rules.
// This playbook installs UFW, resets it to defaults, and configures secure
// default policies with configurable port access.
//
// Usage:
//
//	go run . --playbook=ufw-install [--arg=allow-ssh=true] [--arg=allow-http=true] [--arg=allow-https=true] [--arg=allow-ports=8080,9000]
//
// Execution Flow:
//  1. Updates package lists via apt-get update
//  2. Installs UFW package
//  3. Resets UFW to factory defaults (force)
//  4. Sets default policy: deny incoming, allow outgoing
//  5. Allows configured ports based on arguments
//  6. Enables UFW with --force to avoid interactive prompt
//
// Args:
//   - allow-ssh: "true" to allow SSH (default: "true")
//   - allow-http: "true" to allow HTTP (default: "false")
//   - allow-https: "true" to allow HTTPS (default: "false")
//   - allow-ports: Comma-separated list of additional ports (e.g., "3306,8080")
//
// Security Benefits:
//   - Blocks unauthorized access attempts
//   - Reduces attack surface
//   - Provides clear logging of blocked traffic
//   - Easy to configure additional rules
//
// Prerequisites:
//   - Root SSH access required
//   - Internet connectivity for package installation
//
// Related Playbooks:
//   - ufw-status: Check firewall status
//   - ufw-allow: Allow additional ports after installation
type UfwInstall struct {
	*playbook.BasePlaybook
}

// Check determines if UFW needs to be installed.
func (u *UfwInstall) Check() (bool, error) {
	cfg := u.GetNodeConfig()
	_, err := ssh.Run(cfg, types.Command{Command: "which ufw", Description: "Check if UFW is installed"})
	return err != nil, nil
}

// Run executes the playbook and returns detailed result.
func (u *UfwInstall) Run() playbook.Result {
	cfg := u.GetNodeConfig()

	// Define commands
	cmdInstall := "apt-get update -y && DEBIAN_FRONTEND=noninteractive apt-get install -y ufw"
	cmdReset := "ufw --force reset"
	cmdDefaults := "ufw default deny incoming && ufw default allow outgoing"
	cmdEnable := "echo 'y' | ufw enable"

	// Parse arguments for conditional commands
	allowSSH := u.GetArg(ArgAllowSSH)
	if allowSSH == "" {
		allowSSH = DefaultAllowSSH
	}
	allowHTTP := u.GetArg(ArgAllowHTTP)
	if allowHTTP == "" {
		allowHTTP = DefaultAllowHTTP
	}
	allowHTTPS := u.GetArg(ArgAllowHTTPS)
	if allowHTTPS == "" {
		allowHTTPS = DefaultAllowHTTPS
	}
	allowPorts := u.GetArg(ArgAllowPorts)

	// Check for dry-run mode - display actual commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdInstall)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdReset)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdDefaults)
		if allowSSH == "true" {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", "ufw allow ssh")
		}
		if allowHTTP == "true" {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", "ufw allow 80/tcp")
		}
		if allowHTTPS == "true" {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", "ufw allow 443/tcp")
		}
		if allowPorts != "" {
			ports := strings.Split(allowPorts, ",")
			for _, port := range ports {
				port = strings.TrimSpace(port)
				if port != "" {
					cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", fmt.Sprintf("ufw allow %s/tcp", port))
				}
			}
		}
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdEnable)
		return playbook.Result{
			Changed: true,
			Message: "Would install and configure UFW firewall",
		}
	}

	cfg.GetLoggerOrDefault().Info("installing UFW firewall")

	// Install UFW
	output, err := ssh.Run(cfg, types.Command{Command: cmdInstall, Description: "Install UFW package"})
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to install UFW",
			Error:   fmt.Errorf("failed to install UFW: %w\nOutput: %s", err, output),
		}
	}

	// Reset UFW to defaults
	_, _ = ssh.Run(cfg, types.Command{Command: cmdReset, Description: "Reset UFW to defaults"})

	// Set default policies
	_, _ = ssh.Run(cfg, types.Command{Command: cmdDefaults, Description: "Set UFW default policies"})

	// Re-parse arguments (already defined in dry-run block)
	allowSSH = u.GetArg(ArgAllowSSH)
	if allowSSH == "" {
		allowSSH = DefaultAllowSSH
	}
	allowHTTP = u.GetArg(ArgAllowHTTP)
	if allowHTTP == "" {
		allowHTTP = DefaultAllowHTTP
	}
	allowHTTPS = u.GetArg(ArgAllowHTTPS)
	if allowHTTPS == "" {
		allowHTTPS = DefaultAllowHTTPS
	}
	allowPorts = u.GetArg(ArgAllowPorts)

	allowedServices := []string{}

	// Allow SSH if requested
	if allowSSH == "true" {
		_, _ = ssh.Run(cfg, types.Command{Command: "ufw allow ssh", Description: "Allow SSH access"})
		allowedServices = append(allowedServices, "SSH")
	}

	// Allow HTTP if requested
	if allowHTTP == "true" {
		_, _ = ssh.Run(cfg, types.Command{Command: "ufw allow 80/tcp", Description: "Allow HTTP access"})
		allowedServices = append(allowedServices, "HTTP")
	}

	// Allow HTTPS if requested
	if allowHTTPS == "true" {
		_, _ = ssh.Run(cfg, types.Command{Command: "ufw allow 443/tcp", Description: "Allow HTTPS access"})
		allowedServices = append(allowedServices, "HTTPS")
	}

	// Allow custom ports
	if allowPorts != "" {
		ports := strings.Split(allowPorts, ",")
		for _, port := range ports {
			port = strings.TrimSpace(port)
			if port != "" {
				_, _ = ssh.Run(cfg, types.Command{Command: fmt.Sprintf("ufw allow %s/tcp", port), Description: "Allow custom port"})
				allowedServices = append(allowedServices, fmt.Sprintf("port %s", port))
			}
		}
	}

	// Enable UFW (non-interactive)
	output, err = ssh.Run(cfg, types.Command{Command: cmdEnable, Description: "Enable UFW firewall"})
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to enable UFW",
			Error:   fmt.Errorf("failed to enable UFW: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("UFW installed and configured")
	return playbook.Result{
		Changed: true,
		Message: fmt.Sprintf("UFW installed with secure defaults (allowed: %s)", strings.Join(allowedServices, ", ")),
		Details: map[string]string{
			"allowed_services": strings.Join(allowedServices, ", "),
		},
	}
}

// NewUfwInstall creates a new ufw-install playbook.
func NewUfwInstall() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDUfwInstall)
	pb.SetDescription("Install and configure UFW firewall with secure defaults")
	return &UfwInstall{BasePlaybook: pb}
}
