package skills

import "github.com/dracory/ork/ssh"

// ShellEscapeArg escapes a string for safe use as an unquoted shell argument.
// It wraps the value in single quotes and escapes any embedded single quotes
// using the POSIX sequence '\”. This prevents shell injection when
// interpolating user-supplied values (usernames, package names, etc.) into
// shell commands.
//
// This is a re-export of ssh.ShellEscapeArg for convenience so skill
// implementations don't need to import the ssh package directly.
//
// Usage:
//
//	safe := skills.ShellEscapeArg(username)
//	cmd := fmt.Sprintf("id %s", safe)
func ShellEscapeArg(s string) string {
	return ssh.ShellEscapeArg(s)
}
