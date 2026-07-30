package skills

import "strings"

// ShellEscapeArg escapes a string for safe use as an unquoted shell argument.
// It wraps the value in single quotes and escapes any embedded single quotes
// using the POSIX sequence '\''. This prevents shell injection when
// interpolating user-supplied values (usernames, package names, etc.) into
// shell commands.
//
// Usage:
//
//	safe := skills.ShellEscapeArg(username)
//	cmd := fmt.Sprintf("id %s", safe)
func ShellEscapeArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
