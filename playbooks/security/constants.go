// Package security provides playbooks for system security hardening and configuration.
// These playbooks help secure servers by applying industry-standard security settings
// to SSH, kernel parameters, and installing security monitoring tools.
package security

// Argument key constants for SSH hardening playbook.
const (
	// ArgNonRootUser specifies the non-root user to verify before disabling root login
	ArgNonRootUser = "non-root-user"
)
