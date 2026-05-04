package types

// Command represents a shell command with its description and optional execution settings.
//
// It is used to display and execute shell commands in a structured way.
//
// Also useful when running in dry-run mode to see what commands would be executed.
//
// Example:
//   - Command: "ls -la"
//   - Description: "List all files in long format"
//   - Chdir: "/var/www"
//   - BecomeUser: "www-data"
//   - Required: true
//
// Usage:
//   - command := types.Command{Command: "ls -la", Description: "List all files in long format", Chdir: "/var/www"}
type Command struct {
	Command     string
	Description string
	Chdir       string // Working directory for command execution
	BecomeUser  string // User to become when executing command (e.g., "postgres", "www-data")
	Required    bool   // Whether the command must succeed for execution to continue
}
