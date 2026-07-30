package user

// Package user documentation is in status.go

import "strings"

// shellEscapeArg escapes a string for safe use as an unquoted shell argument.
// It wraps the value in single quotes and escapes any embedded single quotes
// using the POSIX sequence '\''. This prevents shell injection when
// interpolating user-supplied values (usernames, group names, etc.) into
// shell commands.
//
// Usage:
//
//	safe := shellEscapeArg(username)
//	cmd := fmt.Sprintf("id %s", safe)
func shellEscapeArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
