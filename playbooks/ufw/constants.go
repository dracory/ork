// Package ufw provides playbooks for managing the Uncomplicated Firewall (UFW).
// UFW is a simple interface for managing iptables firewall rules on Debian/Ubuntu systems.
package ufw

// Argument key constants for use with GetArg.
const (
	// ArgAllowSSH enables SSH access (port 22) - "true" or "false"
	ArgAllowSSH = "allow-ssh"

	// ArgAllowHTTP enables HTTP access (port 80) - "true" or "false"
	ArgAllowHTTP = "allow-http"

	// ArgAllowHTTPS enables HTTPS access (port 443) - "true" or "false"
	ArgAllowHTTPS = "allow-https"

	// ArgAllowPorts allows custom ports - comma-separated list (e.g., "8080,9000")
	ArgAllowPorts = "allow-ports"
)

// Default configuration constants.
const (
	// DefaultAllowSSH is the default for SSH access (true)
	DefaultAllowSSH = "true"

	// DefaultAllowHTTP is the default for HTTP access (false)
	DefaultAllowHTTP = "false"

	// DefaultAllowHTTPS is the default for HTTPS access (false)
	DefaultAllowHTTPS = "false"
)
