package swap

import (
	"fmt"
	"strings"
)

// Argument key constants for use with GetArg.
const (
	// ArgSize is the swap file size argument key
	ArgSize = "size"

	// ArgUnit is the size unit argument key (gb|mb)
	ArgUnit = "unit"

	// ArgSwappiness is the kernel swappiness argument key
	ArgSwappiness = "swappiness"

	// ArgSwapFilePath is the path for the swap file
	ArgSwapFilePath = "swapfile-path"
)

// Default configuration constants for swap playbooks.
const (
	// DefaultSize is the default swap file size (1)
	DefaultSize = "1"

	// DefaultUnit is the default size unit (gb)
	DefaultUnit = "gb"

	// DefaultSwappiness is the default kernel swappiness value (10)
	// Lower values prefer RAM over swap, better for database workloads
	DefaultSwappiness = "10"

	// DefaultSwapFilePath is the default path for the swap file
	DefaultSwapFilePath = "/swapfile"
)

// validateSwapFilePath validates that a swap file path is a sane absolute path
// with no shell metacharacters. This prevents shell injection when the path is
// interpolated into commands like `dd`, `mkswap`, `swapon`, and `sed`.
// Returns an error if the path is empty, not absolute, or contains any
// character outside [A-Za-z0-9._\-/].
func validateSwapFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("swap file path cannot be empty")
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("swap file path must be absolute (got %q)", path)
	}
	for _, r := range path {
		if !isSafePathRune(r) {
			return fmt.Errorf("swap file path contains disallowed character %q (allowed: letters, digits, '.', '_', '-', '/')", r)
		}
	}
	return nil
}

// isSafePathRune reports whether r is allowed in a swap file path.
func isSafePathRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= 'A' && r <= 'Z':
		return true
	case r >= '0' && r <= '9':
		return true
	case r == '/', r == '.', r == '_', r == '-':
		return true
	}
	return false
}
