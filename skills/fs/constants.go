// Package fs provides readable, idempotent abstractions over common filesystem
// operations (mkdir, chown, chmod, ln, cp, mv, rm, test).
//
// These are general-purpose primitives reusable across any project. They wrap
// raw shell commands with validation, idempotency checks, and structured
// results. Each skill reports Changed=true when it modifies the system and
// Changed=false when the system is already in the desired state.
//
// Skills in this package:
//   - DirCreate: Create directory with ownership and permissions
//   - DirExists: Check if directory exists (read-only)
//   - DirDelete: Delete a directory
//   - FileCreate: Create file with content, ownership, and permissions
//   - FileExists: Check if file exists (read-only)
//   - FileDelete: Delete a single file
//   - FileCopy: Copy a file on the remote server
//   - ChangeOwner: Change file/directory ownership (chown)
//   - ChangeMode: Change file/directory permissions (chmod)
//   - SymlinkCreate: Create or update symbolic link (ln -sf)
//   - Rename: Rename/move file or directory (mv)
//   - Remove: Generic remove file or directory (rm)
package fs

import (
	"fmt"
	"strings"
)

// Argument key constants for use with GetArg.
const (
	// ArgPath is the file or directory path argument key.
	ArgPath = "path"

	// ArgSrc is the source path argument key (for copy/rename).
	ArgSrc = "src"

	// ArgDst is the destination path argument key (for copy/rename).
	ArgDst = "dst"

	// ArgTarget is the symlink target path argument key.
	ArgTarget = "target"

	// ArgLink is the symlink path itself argument key.
	ArgLink = "link"

	// ArgContent is the file content argument key (for FileCreate).
	ArgContent = "content"

	// ArgOwner is the owner (user:group) argument key.
	ArgOwner = "owner"

	// ArgMode is the permissions (octal, e.g. "755") argument key.
	ArgMode = "mode"

	// ArgOverwrite is the overwrite-if-exists flag argument key ("true"/"false").
	ArgOverwrite = "overwrite"

	// ArgForce is the force flag argument key ("true"/"false").
	ArgForce = "force"

	// ArgRecursive is the recursive flag argument key ("true"/"false").
	ArgRecursive = "recursive"

	// ArgParents is the create-parent-dirs flag argument key ("true"/"false").
	ArgParents = "parents"
)

// Default configuration constants for fs playbooks.
const (
	// DefaultDirMode is the default directory permissions.
	DefaultDirMode = "755"

	// DefaultFileMode is the default file permissions.
	DefaultFileMode = "644"

	// DefaultParents is the default for creating parent directories.
	DefaultParents = "true"

	// DefaultOverwrite is the default for overwriting existing files.
	DefaultOverwrite = "false"

	// DefaultForce is the default for force operations.
	DefaultForce = "false"

	// DefaultRecursive is the default for recursive operations.
	DefaultRecursive = "false"
)

// validatePath validates that a path is a sane absolute path with no shell
// metacharacters. This prevents shell injection when the path is interpolated
// into commands like mkdir, chmod, chown, rm, etc.
// Returns an error if the path is empty, not absolute, or contains any
// character outside [A-Za-z0-9._\-/].
func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must be absolute (got %q)", path)
	}
	for _, r := range path {
		if !isSafePathRune(r) {
			return fmt.Errorf("path contains disallowed character %q (allowed: letters, digits, '.', '_', '-', '/')", r)
		}
	}
	return nil
}

// isSafePathRune reports whether r is allowed in a path.
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

// validateOwner validates that an owner string is in user:group format with
// safe characters. Allowed: letters, digits, underscore, hyphen, and colon
// separator.
func validateOwner(owner string) error {
	if owner == "" {
		return fmt.Errorf("owner cannot be empty")
	}
	parts := strings.Split(owner, ":")
	if len(parts) > 2 {
		return fmt.Errorf("owner must be in user:group format, got %q", owner)
	}
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("owner components cannot be empty, got %q", owner)
		}
		for _, r := range part {
			if !isSafeOwnerRune(r) {
				return fmt.Errorf("owner contains disallowed character %q (allowed: letters, digits, '_', '-')", r)
			}
		}
	}
	return nil
}

// isSafeOwnerRune reports whether r is allowed in an owner/user/group name.
func isSafeOwnerRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= 'A' && r <= 'Z':
		return true
	case r >= '0' && r <= '9':
		return true
	case r == '_', r == '-':
		return true
	}
	return false
}

// validateMode validates that a mode string is a valid octal string of 3 or 4
// digits (e.g. "755", "644", "0600").
func validateMode(mode string) error {
	if mode == "" {
		return fmt.Errorf("mode cannot be empty")
	}
	if len(mode) < 3 || len(mode) > 4 {
		return fmt.Errorf("mode must be 3 or 4 octal digits, got %q", mode)
	}
	for _, r := range mode {
		if r < '0' || r > '7' {
			return fmt.Errorf("mode must be octal (0-7), got %q", mode)
		}
	}
	return nil
}

// isTrue returns true if the string value represents a truthy value.
func isTrue(s string) bool {
	return s == "true" || s == "1" || s == "yes"
}
