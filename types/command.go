package types

// Command represents a shell command with its description.
//
// It is used to display and execute shell commands in a structured way.
//
// Also useful when running in dry-run mode to see what commands would be executed.
//
// Example:
//   - Command: "ls -la"
//   - Description: "List all files in long format"
//
// Usage:
//   - command := types.Command{Command: "ls -la", Description: "List all files in long format"}
type Command struct {
	Command     string
	Description string
}
