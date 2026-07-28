package mariadb

import "strings"

// shellEscapedBacktick is a backtick prefixed with a backslash, for use inside
// shell double-quoted strings where a literal backtick is needed (e.g. SQL
// identifiers). Used in fmt.Sprintf format strings via concatenation.
const shellEscapedBacktick = "\\" + "`"

// mariadbEscapeSQLQuote escapes single quotes for safe use inside SQL string
// literals. The standard SQL escaping is to double the single quote: ' → ”.
//
// Usage:
//
//	escaped := mariadbEscapeSQLQuote(password)
//	cmd := fmt.Sprintf(`ALTER USER 'root' IDENTIFIED BY '%s'`, escaped)
func mariadbEscapeSQLQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// mariadbEscapeShellQuote escapes a string for safe use inside single-quoted
// shell arguments. The standard POSIX shell escaping is to close the quote,
// insert an escaped literal quote, and reopen: ' → '\”.
//
// Usage:
//
//	escaped := mariadbEscapeShellQuote(password)
//	cmd := fmt.Sprintf(`MYSQL_PWD='%s' mysqldump ...`, escaped)
func mariadbEscapeShellQuote(s string) string {
	return strings.ReplaceAll(s, "'", `'\''`)
}

// mariadbEscapeShellDoubleQuote escapes a string for safe use inside
// double-quoted shell arguments. In POSIX shell double quotes, the characters
// \, ", $, and ` are special and must be escaped with a backslash.
//
// Usage:
//
//	escaped := mariadbEscapeShellDoubleQuote(mariadbEscapeSQLQuote(password))
//	cmd := fmt.Sprintf(`mysql -e "ALTER USER 'root' IDENTIFIED BY '%s';"`, escaped)
func mariadbEscapeShellDoubleQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, `$`, `\$`)
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}

// mariadbEscapeSQLIdentifier escapes backticks for safe use inside
// backtick-quoted SQL identifiers (database names, table names). The standard
// escaping is to double the backtick: ` → “.
//
// Usage:
//
//	escaped := mariadbEscapeSQLIdentifier(dbName)
//	cmd := fmt.Sprintf("CREATE DATABASE `%s`", escaped)
func mariadbEscapeSQLIdentifier(s string) string {
	return strings.ReplaceAll(s, "`", "``")
}
