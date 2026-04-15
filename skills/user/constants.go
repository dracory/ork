package user

// Argument key constants for use with GetArg.
const (
	// ArgUsername is the username argument key
	ArgUsername = "username"
	// ArgSSHKey is the SSH public key argument key
	ArgSSHKey = "ssh-key"
	// ArgPassword is the password argument key
	ArgPassword = "password"

	// ArgShell is the user's login shell
	ArgShell = "shell"

	// ArgGroup is the primary group for the user
	ArgGroup = "group"

	// ArgHomeDir is the home directory path
	ArgHomeDir = "home-dir"

	// ArgSudoGroup is the sudo/admin group name
	ArgSudoGroup = "sudo-group"
)

// Default configuration constants for user playbooks.
const (
	// DefaultUsername is the default username (empty - must be provided)
	DefaultUsername = ""

	// DefaultShell is the default login shell
	DefaultShell = "/bin/bash"

	// DefaultGroup is the default primary group (empty = same as username)
	DefaultGroup = ""

	// DefaultSudoGroup is the default sudo/admin group
	DefaultSudoGroup = "sudo"
)
