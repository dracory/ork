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

	// ArgIP specifies IP address(es) to allow - single IP or comma-separated list
	ArgIP = "ip"

	// ArgPort specifies the port number to configure
	ArgPort = "port"

	// ArgProtocol specifies the protocol - "tcp" or "udp"
	ArgProtocol = "protocol"

	// ArgComment specifies an optional comment for the rule
	ArgComment = "comment"

	// ArgNumber specifies the rule number for delete operations
	ArgNumber = "number"

	// ArgIncoming specifies the incoming policy for default rules
	ArgIncoming = "incoming"

	// ArgOutgoing specifies the outgoing policy for default rules
	ArgOutgoing = "outgoing"
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
