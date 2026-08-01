package systemctl

// helpers.go contains small shared utilities used across the systemctl
// skill files. Keeping them here avoids implicit cross-file dependencies
// (e.g. a helper defined in enable.go but used by disable.go).

// ternary returns ifTrue when cond is true, ifFalse otherwise.
// Used for inline conditionals in description/message strings where an
// if-statement would break fluent chaining.
func ternary(cond bool, ifTrue, ifFalse string) string {
	if cond {
		return ifTrue
	}
	return ifFalse
}
