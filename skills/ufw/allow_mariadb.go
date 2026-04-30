package ufw

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
	"github.com/samber/lo"
)

// AllowMariaDB configures UFW firewall rules for MariaDB database access.
// This skill can allow access from specific IP addresses or from any IP (less secure).
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
	*types.BaseSkill
}

// Check determines if UFW rules need to be configured.
func (u *AllowMariaDB) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and returns detailed result.
func (u *AllowMariaDB) Run() types.Result {
	cfg := u.GetNodeConfig()
	ip := cfg.GetArgOr(ArgIP, "")
	mariaDBPort := cfg.GetArgOr(ArgPort, "3306")

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		allowedIPs, _ := u.allowIPs(cfg, ip, mariaDBPort, true)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would configure UFW for MariaDB port %s", mariaDBPort),
			Details: map[string]string{
				"allowed_ips": strings.Join(allowedIPs, ","),
			},
		}
	}

	// Execute for real
	allowedIPs, err := u.allowIPs(cfg, ip, mariaDBPort, false)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to allow MariaDB access", Error: err}
	}

	if ip == "" || ip == "any" {
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("MariaDB port %s is now open to all IPs", mariaDBPort),
		}
	}

	return types.Result{
		Changed: len(allowedIPs) > 0,
		Message: fmt.Sprintf("Allowed %d IPs to access MariaDB port %s", len(allowedIPs), mariaDBPort),
		Details: map[string]string{
			"allowed_ips": strings.Join(allowedIPs, ","),
		},
	}
}

// SetArgs sets the arguments for MariaDB UFW rules.
// Returns AllowMariaDB for fluent method chaining.
func (u *AllowMariaDB) SetArgs(args map[string]string) types.RunnableInterface {
	u.BaseSkill.SetArgs(args)
	return u
}

// SetArg sets a single argument for MariaDB UFW rules.
// Returns AllowMariaDB for fluent method chaining.
func (u *AllowMariaDB) SetArg(key, value string) types.RunnableInterface {
	u.BaseSkill.SetArg(key, value)
	return u
}

// SetID sets the ID for MariaDB UFW rules.
// Returns AllowMariaDB for fluent method chaining.
func (u *AllowMariaDB) SetID(id string) types.RunnableInterface {
	u.BaseSkill.SetID(id)
	return u
}

// SetDescription sets the description for MariaDB UFW rules.
// Returns AllowMariaDB for fluent method chaining.
func (u *AllowMariaDB) SetDescription(description string) types.RunnableInterface {
	u.BaseSkill.SetDescription(description)
	return u
}

// SetTimeout sets the timeout for MariaDB UFW rules.
// Returns AllowMariaDB for fluent method chaining.
func (u *AllowMariaDB) SetTimeout(timeout time.Duration) types.RunnableInterface {
	u.BaseSkill.SetTimeout(timeout)
	return u
}

// allowIPs executes IP processing with the appropriate method based on IP value
func (u *AllowMariaDB) allowIPs(cfg types.NodeConfig, ip string, mariaDBPort string, isDryRun bool) ([]string, error) {
	if ip == "" || ip == "any" {
		return u.allowAnyIP(cfg, mariaDBPort, isDryRun)
	}

	ips := lo.Filter(
		lo.Map(strings.Split(ip, ","), func(item string, index int) string {
			return strings.TrimSpace(item)
		}),
		func(item string, index int) bool {
			return item != ""
		},
	)

	return u.allowSpecificIPs(cfg, ips, mariaDBPort, isDryRun)
}

// allowAnyIP handles allowing MariaDB access from any IP
func (u *AllowMariaDB) allowAnyIP(cfg types.NodeConfig, mariaDBPort string, isDryRun bool) ([]string, error) {
	cmd := types.Command{
		Command:     fmt.Sprintf("ufw allow %s/tcp", mariaDBPort),
		Description: "Allow MariaDB access from any IP",
	}

	if isDryRun {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmd.Command, "description", cmd.Description)
		return []string{"any"}, nil
	}
	cfg.GetLoggerOrDefault().Warn("allowing MariaDB access from ANY IP")
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to allow MariaDB: %w\nOutput: %s", err, output)
	}
	return []string{"any"}, nil
}

// allowSpecificIPs handles allowing MariaDB access from specific IP addresses
func (u *AllowMariaDB) allowSpecificIPs(cfg types.NodeConfig, ips []string, mariaDBPort string, isDryRun bool) ([]string, error) {
	allowedIPs := []string{}
	for _, singleIP := range ips {
		cmd := types.Command{
			Command:     fmt.Sprintf("ufw allow from %s to any port %s", singleIP, mariaDBPort),
			Description: "Allow MariaDB access from IP: " + singleIP,
		}

		if isDryRun {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmd.Command, "description", cmd.Description)
			allowedIPs = append(allowedIPs, singleIP)
			continue
		}

		cfg.GetLoggerOrDefault().Info("allowing MariaDB access", "ip", singleIP)

		output, err := ssh.Run(cfg, cmd)
		if err != nil {
			cfg.GetLoggerOrDefault().Warn("could not allow IP", "ip", singleIP, "error", err)
		} else {
			allowedIPs = append(allowedIPs, singleIP)
			cfg.GetLoggerOrDefault().Info("UFW output", "output", output)
		}
	}
	return allowedIPs, nil
}

// NewAllowMariaDB creates a new ufw-allow-mariadb skill.
func NewAllowMariaDB() types.RunnableInterface {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUfwAllowMariaDB)
	pb.SetDescription("Configure UFW firewall rules for MariaDB access")
	return &AllowMariaDB{BaseSkill: pb}
}
