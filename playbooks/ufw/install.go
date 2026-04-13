package ufw

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
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
	cfg := u.GetConfig()
	_, err := ssh.Run(cfg, "which ufw")
	return err != nil, nil
}

// Run executes the playbook and returns detailed result.
func (u *UfwInstall) Run() playbook.Result {
	cfg := u.GetConfig()

	cfg.GetLoggerOrDefault().Info("installing UFW firewall")

	// Install UFW
	output, err := ssh.Run(cfg, "apt-get update -y && DEBIAN_FRONTEND=noninteractive apt-get install -y ufw")
	if err != nil {
		return playbook.Result{
			Changed: false,
			Message: "Failed to install UFW",
			Error:   fmt.Errorf("failed to install UFW: %w\nOutput: %s", err, output),
		}
	}

	// Reset UFW to defaults
	_, _ = ssh.Run(cfg, "ufw --force reset")

	// Set default policies
	_, _ = ssh.Run(cfg, "ufw default deny incoming && ufw default allow outgoing")

	// Parse arguments
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

	allowedServices := []string{}

	// Allow SSH if requested
	if allowSSH == "true" {
		_, _ = ssh.Run(cfg, "ufw allow ssh")
		allowedServices = append(allowedServices, "SSH")
	}

	// Allow HTTP if requested
	if allowHTTP == "true" {
		_, _ = ssh.Run(cfg, "ufw allow 80/tcp")
		allowedServices = append(allowedServices, "HTTP")
	}

	// Allow HTTPS if requested
	if allowHTTPS == "true" {
		_, _ = ssh.Run(cfg, "ufw allow 443/tcp")
		allowedServices = append(allowedServices, "HTTPS")
	}

	// Allow custom ports
	if allowPorts != "" {
		ports := strings.Split(allowPorts, ",")
		for _, port := range ports {
			port = strings.TrimSpace(port)
			if port != "" {
				_, _ = ssh.Run(cfg, fmt.Sprintf("ufw allow %s/tcp", port))
				allowedServices = append(allowedServices, fmt.Sprintf("port %s", port))
			}
		}
	}

	// Enable UFW (non-interactive)
	output, err = ssh.Run(cfg, "echo 'y' | ufw enable")
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
