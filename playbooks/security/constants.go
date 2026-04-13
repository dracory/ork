// Package security provides playbooks for system security hardening and configuration.
// These playbooks help secure servers by applying industry-standard security settings
// to SSH, kernel parameters, and installing security monitoring tools.
package security

// Argument key constants for SSH hardening playbook.
const (
	// ArgNonRootUser specifies the non-root user to verify before disabling root login
	ArgNonRootUser = "non-root-user"

	// ArgSSHConfigPath specifies the SSH configuration file path
	ArgSSHConfigPath = "ssh-config-path"

	// ArgMaxAuthTries specifies the maximum authentication attempts
	ArgMaxAuthTries = "max-auth-tries"

	// ArgClientAliveInterval specifies the client alive interval in seconds
	ArgClientAliveInterval = "client-alive-interval"

	// ArgClientAliveCountMax specifies the client alive count max
	ArgClientAliveCountMax = "client-alive-count-max"

	// ArgSysctlConfigPath specifies the sysctl configuration file path
	ArgSysctlConfigPath = "sysctl-config-path"

	// ArgSysctlDropInPath specifies the sysctl drop-in file path
	ArgSysctlDropInPath = "sysctl-dropin-path"
)

// Default configuration constants for security playbooks.
const (
	// DefaultNonRootUser is the default non-root username to verify
	DefaultNonRootUser = "deploy"

	// DefaultSSHConfigPath is the default SSH configuration file path
	DefaultSSHConfigPath = "/etc/ssh/sshd_config"

	// DefaultMaxAuthTries is the default maximum authentication attempts
	DefaultMaxAuthTries = "3"

	// DefaultClientAliveInterval is the default client alive interval (seconds)
	DefaultClientAliveInterval = "300"

	// DefaultClientAliveCountMax is the default client alive count max
	DefaultClientAliveCountMax = "2"

	// DefaultSysctlConfigPath is the default sysctl configuration file path
	DefaultSysctlConfigPath = "/etc/sysctl.conf"

	// DefaultSysctlDropInPath is the default sysctl drop-in directory path
	DefaultSysctlDropInPath = "/etc/sysctl.d/99-security-hardening.conf"
)
