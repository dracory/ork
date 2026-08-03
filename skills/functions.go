package skills

import (
	"strings"

	"github.com/dracory/ork/ssh"
)

// ShellEscapeArg escapes a string for safe use as an unquoted shell argument.
// It wraps the value in single quotes and escapes any embedded single quotes
// using the POSIX sequence '\". This prevents shell injection when
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

// ShellEscapeContent escapes file content for safe use in a printf '%s'
// argument. It wraps the content in single quotes and escapes embedded
// single quotes using the POSIX sequence '\”.
// This is the shared version used by fs.FileCreate, security.AideInstall,
// security.AuditdInstall, and other skills that write file content via printf.
func ShellEscapeContent(content string) string {
	return "'" + strings.ReplaceAll(content, "'", "'\\''") + "'"
}

// ShellEscapeGrep escapes a string for safe use as a literal grep BRE pattern,
// then shell-escapes the result for safe interpolation into a command.
// This prevents regex metacharacters in the input (e.g. ".", "*") from being
// interpreted as grep pattern syntax, and prevents shell injection.
//
// Usage:
//
//	safe := skills.ShellEscapeGrep(packageName)
//	cmd := fmt.Sprintf("grep -i -- ^%s/", safe)
func ShellEscapeGrep(s string) string {
	replacer := strings.NewReplacer(
		`\\`, `\\\\`,
		`.`, `\.`,
		`*`, `\*`,
		`[`, `\[`,
		`]`, `\]`,
		`^`, `\^`,
		`$`, `\$`,
	)
	return ShellEscapeArg(replacer.Replace(s))
}
