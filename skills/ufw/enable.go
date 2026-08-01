package ufw

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// Enable enables the UFW firewall.
// This skill activates UFW to start enforcing firewall rules, and optionally
// opens SSH, HTTP, HTTPS, and custom ports before enabling.
//
// Usage:
//
//	node.Run(ufw.NewEnable().
//	    SetAllowSSH(true).
//	    SetAllowHTTP(true).
//	    SetAllowHTTPS(true).
//	    SetAllowPorts("8080", "9000"))
//
// Execution Flow:
//  1. Allows configured ports (SSH/HTTP/HTTPS/custom) if requested
//  2. Executes `ufw --force enable`
//  3. Returns success/failure result
//
// Args:
//   - allow-ssh: "true" to allow SSH (default: "true")
//   - allow-http: "true" to allow HTTP (default: "false")
//   - allow-https: "true" to allow HTTPS (default: "false")
//   - allow-ports: Comma-separated list of additional ports (e.g., "3306,8080")
//
// Prerequisites:
//   - UFW must be installed
//   - Root SSH access required
//   - Ensure SSH port is allowed before enabling
//
// Related Playbooks:
//   - ufw-disable: Disable UFW firewall
//   - ufw-status: Verify UFW status
//   - ufw-install: Install and enable UFW
type Enable struct {
	*types.BaseSkill
}

// Compile-time assertion that Enable implements types.RunnableInterface.
var _ types.RunnableInterface = (*Enable)(nil)

// Check determines if UFW needs to be enabled.
func (e *Enable) Check() (bool, error) {
	return true, nil
}

// Run executes the skill and enables UFW.
func (e *Enable) Run() types.Result {
	cfg := e.GetNodeConfig()

	// Parse arguments
	allowSSH := e.GetArg(ArgAllowSSH)
	if allowSSH == "" {
		allowSSH = DefaultAllowSSH
	}
	allowHTTP := e.GetArg(ArgAllowHTTP)
	if allowHTTP == "" {
		allowHTTP = DefaultAllowHTTP
	}
	allowHTTPS := e.GetArg(ArgAllowHTTPS)
	if allowHTTPS == "" {
		allowHTTPS = DefaultAllowHTTPS
	}
	allowPorts := e.GetArg(ArgAllowPorts)

	// Define commands
	cmdEnable := types.Command{
		Command:     "ufw --force enable",
		Description: "Enable UFW firewall",
		Required:    true,
	}
	cmdAllowSSH := types.Command{
		Command:     "ufw allow ssh",
		Description: "Allow SSH access",
	}
	cmdAllowHTTP := types.Command{
		Command:     "ufw allow 80/tcp",
		Description: "Allow HTTP access",
	}
	cmdAllowHTTPS := types.Command{
		Command:     "ufw allow 443/tcp",
		Description: "Allow HTTPS access",
	}

	// Check for dry-run mode
	if cfg.IsDryRunMode {
		if allowSSH == "true" {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdAllowSSH.Command)
		}
		if allowHTTP == "true" {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdAllowHTTP.Command)
		}
		if allowHTTPS == "true" {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdAllowHTTPS.Command)
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
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdEnable.Command)
		return types.Result{
			Changed: true,
			Message: "Would enable UFW firewall",
		}
	}

	allowedServices := []string{}

	// Allow SSH if requested
	if allowSSH == "true" {
		_, _ = ssh.Run(cfg, cmdAllowSSH)
		allowedServices = append(allowedServices, "SSH")
	}

	// Allow HTTP if requested
	if allowHTTP == "true" {
		_, _ = ssh.Run(cfg, cmdAllowHTTP)
		allowedServices = append(allowedServices, "HTTP")
	}

	// Allow HTTPS if requested
	if allowHTTPS == "true" {
		_, _ = ssh.Run(cfg, cmdAllowHTTPS)
		allowedServices = append(allowedServices, "HTTPS")
	}

	// Allow custom ports
	if allowPorts != "" {
		ports := strings.Split(allowPorts, ",")
		for _, port := range ports {
			port = strings.TrimSpace(port)
			if port != "" {
				cmdAllowPort := types.Command{Command: fmt.Sprintf("ufw allow %s/tcp", port), Description: "Allow custom port"}
				_, _ = ssh.Run(cfg, cmdAllowPort)
				allowedServices = append(allowedServices, fmt.Sprintf("port %s", port))
			}
		}
	}

	// Enable UFW
	output, err := ssh.Run(cfg, cmdEnable)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to enable UFW",
			Error:   err,
		}
	}

	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("Enabled UFW firewall (allowed: %s)", strings.Join(allowedServices, ", ")),
		Details: map[string]string{
			"output":           output,
			"allowed_services": strings.Join(allowedServices, ", "),
		},
	}
}

// SetArgs sets the arguments for enable.
// Returns Enable for fluent method chaining.
func (e *Enable) SetArgs(args map[string]string) types.RunnableInterface {
	e.BaseSkill.SetArgs(args)
	return e
}

// SetArg sets a single argument for enable.
// Returns Enable for fluent method chaining.
func (e *Enable) SetArg(key, value string) types.RunnableInterface {
	e.BaseSkill.SetArg(key, value)
	return e
}

// SetAllowSSH sets whether to allow SSH (port 22) and returns Enable for chaining.
func (e *Enable) SetAllowSSH(allow bool) *Enable {
	e.BaseSkill.SetArg(ArgAllowSSH, fmt.Sprintf("%v", allow))
	return e
}

// SetAllowHTTP sets whether to allow HTTP (port 80) and returns Enable for chaining.
func (e *Enable) SetAllowHTTP(allow bool) *Enable {
	e.BaseSkill.SetArg(ArgAllowHTTP, fmt.Sprintf("%v", allow))
	return e
}

// SetAllowHTTPS sets whether to allow HTTPS (port 443) and returns Enable for chaining.
func (e *Enable) SetAllowHTTPS(allow bool) *Enable {
	e.BaseSkill.SetArg(ArgAllowHTTPS, fmt.Sprintf("%v", allow))
	return e
}

// SetAllowPorts adds custom ports to allow and returns Enable for chaining.
// Pass each port as a separate argument. Multiple calls accumulate — ports
// from each call are appended to the existing list (duplicates are removed).
// To reset the list, call SetArg(ArgAllowPorts, "") directly.
func (e *Enable) SetAllowPorts(ports ...string) *Enable {
	existing := e.BaseSkill.GetArg(ArgAllowPorts)
	merged := mergePorts(existing, ports)
	e.BaseSkill.SetArg(ArgAllowPorts, merged)
	return e
}

// SetID sets the ID for enable.
// Returns Enable for fluent method chaining.
func (e *Enable) SetID(id string) types.RunnableInterface {
	e.BaseSkill.SetID(id)
	return e
}

// SetDescription sets the description for enable.
// Returns Enable for fluent method chaining.
func (e *Enable) SetDescription(description string) types.RunnableInterface {
	e.BaseSkill.SetDescription(description)
	return e
}

// SetTimeout sets the timeout for enable.
// Returns Enable for fluent method chaining.
func (e *Enable) SetTimeout(timeout time.Duration) types.RunnableInterface {
	e.BaseSkill.SetTimeout(timeout)
	return e
}

// NewEnable creates a new ufw-enable skill.
func NewEnable() *Enable {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDUfwEnable)
	pb.SetDescription("Enable UFW firewall")
	return &Enable{BaseSkill: pb}
}
