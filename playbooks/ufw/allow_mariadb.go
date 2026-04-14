package ufw

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/playbook"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// AllowMariaDB configures UFW firewall rules for MariaDB database access.
// This playbook can allow access from specific IP addresses or from any IP (less secure).
// It is recommended to restrict access to known application server IPs only.
//
// Usage:
//
//	go run . --playbook=ufw-allow-mariadb [--arg=ip=<ip_address>] [--arg=port=<port>]
//
// Args:
//   - ip: IP address(es) to allow (default: "any" - allows all IPs)
//     Supports comma-separated list for multiple IPs
//     Example: --arg=ip=192.168.1.10,192.168.1.11
//   - port: MariaDB port (default: "3306")
//
// Execution Flow:
//
//	If ip not provided or is "any":
//	  1. Opens port to all IP addresses
//	  2. Logs security warning
//	If specific IP(s) provided:
//	  1. Splits comma-separated IP list
//	  2. For each IP, creates UFW rule allowing access from that IP only
//	  3. Logs each allowed IP
//	Finally:
//	  4. Displays current MariaDB-related firewall rules
//
// Security Recommendations:
//   - PRODUCTION: Always specify IP addresses, never use "any"
//   - DEVELOPMENT: "any" may be acceptable for testing
//   - Multiple IPs: Use comma-separated list for application servers
//   - Example: --arg=ip=10.0.0.5,10.0.0.6,10.0.0.7
//
// Prerequisites:
//   - UFW must be installed and enabled
//   - Root SSH access required
//
// Related Playbooks:
//   - ufw-install: Install UFW firewall
//   - ufw-status: Verify MariaDB rules are active
type AllowMariaDB struct {
	*playbook.BasePlaybook
}

// Check determines if UFW rules need to be configured.
func (u *AllowMariaDB) Check() (bool, error) {
	return true, nil
}

// Run executes the playbook and returns detailed result.
func (u *AllowMariaDB) Run() playbook.Result {
	cfg := u.GetNodeConfig()
	ip := cfg.GetArgOr(ArgIP, "")
	mariaDBPort := cfg.GetArgOr(ArgPort, "3306")

	if ip == "" || ip == "any" {
		cfg.GetLoggerOrDefault().Warn("allowing MariaDB access from ANY IP")
		cmdAllowAny := types.Command{Command: fmt.Sprintf("ufw allow %s/tcp", mariaDBPort), Description: "Allow MariaDB access from any IP"}
		output, err := ssh.Run(cfg, cmdAllowAny)
		if err != nil {
			return playbook.Result{
				Changed: false,
				Message: "Failed to allow MariaDB access",
				Error:   fmt.Errorf("failed to allow MariaDB: %w\nOutput: %s", err, output),
			}
		}
		return playbook.Result{
			Changed: true,
			Message: fmt.Sprintf("MariaDB port %s is now open to all IPs", mariaDBPort),
		}
	}

	// Allow from specific IP(s)
	ips := strings.Split(ip, ",")
	allowedIPs := []string{}
	for _, singleIP := range ips {
		singleIP = strings.TrimSpace(singleIP)
		cfg.GetLoggerOrDefault().Info("allowing MariaDB access", "ip", singleIP)
		cmdAllowIP := types.Command{Command: fmt.Sprintf("ufw allow from %s to any port %s", singleIP, mariaDBPort), Description: "Allow MariaDB access from specific IP"}
		output, err := ssh.Run(cfg, cmdAllowIP)
		if err != nil {
			cfg.GetLoggerOrDefault().Warn("could not allow IP", "ip", singleIP, "error", err)
		} else {
			allowedIPs = append(allowedIPs, singleIP)
			cfg.GetLoggerOrDefault().Info("UFW output", "output", output)
		}
	}

	return playbook.Result{
		Changed: len(allowedIPs) > 0,
		Message: fmt.Sprintf("Allowed %d IPs to access MariaDB port %s", len(allowedIPs), mariaDBPort),
		Details: map[string]string{
			"allowed_ips": strings.Join(allowedIPs, ","),
		},
	}
}

// NewAllowMariaDB creates a new ufw-allow-mariadb playbook.
func NewAllowMariaDB() playbook.PlaybookInterface {
	pb := playbook.NewBasePlaybook()
	pb.SetID(playbook.IDUfwAllowMariaDB)
	pb.SetDescription("Configure UFW firewall rules for MariaDB access")
	return &AllowMariaDB{BasePlaybook: pb}
}
