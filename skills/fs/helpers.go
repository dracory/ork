package fs

import (
	"fmt"
	"strings"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// This file contains shared helper functions used by all fs skills.
// They exist to eliminate duplication: each filesystem check (does dir exist?
// what's the current owner? what's the current mode?) has a single source of
// truth here, so a bug fix or behavior change only needs to happen once.
//
// These are internal helpers, not skills. Skills wrap user-facing operations
// with validation, logging, dry-run handling, and structured results. These
// helpers wrap raw ssh.Run calls with no frills.

// dirExists reports whether a directory exists at path on the remote server.
// Returns false if the path doesn't exist or if the SSH command fails.
func dirExists(cfg types.NodeConfig, path string) bool {
	cmd := types.Command{
		Command:     fmt.Sprintf("test -d %s", skills.ShellEscapeArg(path)),
		Description: "Check if directory exists: " + path,
		Required:    true, // propagate non-zero exit so we can distinguish exists/not-exists
	}
	_, err := ssh.Run(cfg, cmd)
	return err == nil
}

// fileExists reports whether a regular file exists at path on the remote server.
// Returns false if the path doesn't exist or if the SSH command fails.
func fileExists(cfg types.NodeConfig, path string) bool {
	cmd := types.Command{
		Command:     fmt.Sprintf("test -f %s", skills.ShellEscapeArg(path)),
		Description: "Check if file exists: " + path,
		Required:    true, // propagate non-zero exit so we can distinguish exists/not-exists
	}
	_, err := ssh.Run(cfg, cmd)
	return err == nil
}

// pathExists reports whether any file, directory, or symlink exists at path.
// Uses test -e. Returns false if the path doesn't exist or if the SSH command fails.
func pathExists(cfg types.NodeConfig, path string) bool {
	cmd := types.Command{
		Command:     fmt.Sprintf("test -e %s", skills.ShellEscapeArg(path)),
		Description: "Check if path exists: " + path,
		Required:    true, // propagate non-zero exit so we can distinguish exists/not-exists
	}
	_, err := ssh.Run(cfg, cmd)
	return err == nil
}

// getOwner returns the current owner (user:group) of path, or empty string on error.
func getOwner(cfg types.NodeConfig, path string) string {
	cmd := types.Command{
		Command:     fmt.Sprintf("stat -c '%%U:%%G' %s", skills.ShellEscapeArg(path)),
		Description: "Get current owner of: " + path,
	}
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}

// getMode returns the current octal mode (e.g. "755") of path, or empty string on error.
func getMode(cfg types.NodeConfig, path string) string {
	cmd := types.Command{
		Command:     fmt.Sprintf("stat -c '%%a' %s", skills.ShellEscapeArg(path)),
		Description: "Get current mode of: " + path,
	}
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}

// getSymlinkTarget returns the raw target of a symlink (without resolving
// intermediate symlinks), or empty string on error. Uses `readlink` without
// `-f` to preserve the exact target stored in the symlink, ensuring correct
// idempotency comparisons.
func getSymlinkTarget(cfg types.NodeConfig, link string) string {
	cmd := types.Command{
		Command:     fmt.Sprintf("readlink %s", skills.ShellEscapeArg(link)),
		Description: "Get symlink target of: " + link,
	}
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}

// fileContent returns the content of a file, or empty string on error.
func fileContent(cfg types.NodeConfig, path string) string {
	cmd := types.Command{
		Command:     fmt.Sprintf("cat %s", skills.ShellEscapeArg(path)),
		Description: "Read content of: " + path,
	}
	output, err := ssh.Run(cfg, cmd)
	if err != nil {
		return ""
	}
	return output
}

// validateDestructivePath validates that a path is safe for destructive
// operations (rm -rf, etc.). It calls validatePath first, then additionally
// rejects paths with fewer than 2 path components after the root slash.
// This prevents catastrophic operations on root-level directories like
// "/", "/var", "/usr", "/etc", "/home", etc.
//
// Examples:
//   - "/"        → rejected (0 components)
//   - "/var"     → rejected (1 component)
//   - "/var/www" → allowed   (2 components)
func validateDestructivePath(path string) error {
	if err := validatePath(path); err != nil {
		return err
	}
	trimmed := strings.Trim(path, "/")
	components := strings.Split(trimmed, "/")
	if len(components) < 2 {
		return fmt.Errorf("path %q is too short for destructive operations: need at least 2 path components (e.g. /var/www), got %d", path, len(components))
	}
	return nil
}

// filesIdentical reports whether two files have identical content using cmp -s.
// Returns false if either file doesn't exist or if the SSH command fails.
func filesIdentical(cfg types.NodeConfig, src, dst string) bool {
	cmd := types.Command{
		Command:     fmt.Sprintf("cmp -s %s %s", skills.ShellEscapeArg(src), skills.ShellEscapeArg(dst)),
		Description: "Compare content of " + src + " and " + dst,
		Required:    true, // propagate non-zero exit so we can distinguish identical/different
	}
	_, err := ssh.Run(cfg, cmd)
	return err == nil
}
