package fs

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// FileCreate creates a file with content, optional ownership, and permissions.
// This is a readable, idempotent abstraction over writing files via SSH.
//
// Usage:
//
//	node.Run(fs.NewFileCreate().SetArgs(map[string]string{
//	    fs.ArgPath:     "/var/www/myapp/config.json",
//	    fs.ArgContent:  `{"key": "value"}`,
//	    fs.ArgOwner:    "www-data:www-data",
//	    fs.ArgMode:     "644",
//	    fs.ArgOverwrite: "true",
//	}))
//
// Args:
//   - path: File path to create (required, must be absolute)
//   - content: Content to write (optional, empty creates empty file)
//   - owner: Owner in user:group format (optional)
//   - mode: Permissions in octal (optional, default "644")
//   - overwrite: Overwrite if file exists (optional, default "false")
//
// Idempotency:
//   - Check() returns false if file exists with matching content and permissions
//   - Check() returns true if file doesn't exist or content/mode mismatch
type FileCreate struct {
	*types.BaseSkill
}

// Check determines if the file needs to be created or updated.
// Returns true if file doesn't exist or content/mode mismatch.
func (f *FileCreate) Check() (bool, error) {
	cfg := f.GetNodeConfig()
	path := f.GetArg(ArgPath)

	if err := validatePath(path); err != nil {
		return false, err
	}

	// In dry-run mode, assume changes are needed without running SSH commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if file exists with correct content and mode")
		return true, nil
	}

	// Check if file exists
	if !fileExists(cfg, path) {
		// File doesn't exist — needs creation
		return true, nil
	}

	// File exists — check overwrite flag
	overwrite := f.GetArg(ArgOverwrite)
	if !isTrue(overwrite) {
		// Don't overwrite — no change needed
		return false, nil
	}

	// Overwrite is true — check if content differs
	content := f.GetArg(ArgContent)
	currentContent := fileContent(cfg, path)
	if currentContent != content {
		return true, nil
	}

	// Content matches — check mode
	mode := f.GetArg(ArgMode)
	if mode == "" {
		mode = DefaultFileMode
	}
	if err := validateMode(mode); err != nil {
		return false, err
	}
	currentMode := getMode(cfg, path)
	if currentMode != mode {
		return true, nil
	}

	// Check owner
	owner := f.GetArg(ArgOwner)
	if owner != "" {
		if err := validateOwner(owner); err != nil {
			return false, err
		}
		currentOwner := getOwner(cfg, path)
		if currentOwner != owner {
			return true, nil
		}
	}

	return false, nil
}

// Run creates the file with content, ownership, and permissions.
// Changed is true when the file was created or overwritten.
func (f *FileCreate) Run() types.Result {
	path := f.GetArg(ArgPath)
	content := f.GetArg(ArgContent)
	owner := f.GetArg(ArgOwner)
	mode := f.GetArg(ArgMode)
	overwrite := f.GetArg(ArgOverwrite)
	if mode == "" {
		mode = DefaultFileMode
	}
	if overwrite == "" {
		overwrite = DefaultOverwrite
	}

	if err := validatePath(path); err != nil {
		return types.Result{Changed: false, Message: "Invalid path", Error: err}
	}
	if owner != "" {
		if err := validateOwner(owner); err != nil {
			return types.Result{Changed: false, Message: "Invalid owner", Error: err}
		}
	}
	if err := validateMode(mode); err != nil {
		return types.Result{Changed: false, Message: "Invalid mode", Error: err}
	}

	cfg := f.GetNodeConfig()
	escPath := skills.ShellEscapeArg(path)

	// Check if already correct
	needsChange, err := f.Check()
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to check file state", Error: err}
	}
	if !needsChange {
		return types.Result{
			Changed: false,
			Message: "File already exists with correct content and mode: " + path,
		}
	}

	// Check if file exists and overwrite is false (skip in dry-run mode)
	if !cfg.IsDryRunMode {
		if fileExists(cfg, path) && !isTrue(overwrite) {
			return types.Result{
				Changed: false,
				Message: "File already exists and overwrite is false: " + path,
			}
		}
	}

	// Write content using a heredoc-free approach: printf is safer for content
	// with special characters. We escape single quotes in content.
	escContent := shellEscapeContent(content)
	cmdWrite := types.Command{
		Command:     fmt.Sprintf("printf '%%s' %s > %s", escContent, escPath),
		Description: "Write file: " + path,
	}

	var cmdChown, cmdChmod types.Command
	if owner != "" {
		cmdChown = types.Command{
			Command:     fmt.Sprintf("chown %s %s", skills.ShellEscapeArg(owner), escPath),
			Description: "Set owner: " + owner + " on " + path,
		}
	}
	cmdChmod = types.Command{
		Command:     fmt.Sprintf("chmod %s %s", skills.ShellEscapeArg(mode), escPath),
		Description: "Set mode: " + mode + " on " + path,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdWrite.Command)
		if cmdChown.Command != "" {
			cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdChown.Command)
		}
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdChmod.Command)
		return types.Result{
			Changed: true,
			Message: "Would create file: " + path,
		}
	}

	cfg.GetLoggerOrDefault().Info("creating file", "path", path)
	output, err := ssh.Run(cfg, cmdWrite)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to write file",
			Error:   fmt.Errorf("failed to write file: %w\nOutput: %s", err, output),
		}
	}

	if cmdChown.Command != "" {
		chownOutput, chownErr := ssh.Run(cfg, cmdChown)
		if chownErr != nil {
			return types.Result{
				Changed: true,
				Message: "File created but failed to set owner: " + path,
				Error:   fmt.Errorf("failed to set owner: %w\nOutput: %s", chownErr, chownOutput),
				Details: map[string]string{"path": path, "owner": owner},
			}
		}
	}

	chmodOutput, chmodErr := ssh.Run(cfg, cmdChmod)
	if chmodErr != nil {
		return types.Result{
			Changed: true,
			Message: "File created but failed to set mode: " + path,
			Error:   fmt.Errorf("failed to set mode: %w\nOutput: %s", chmodErr, chmodOutput),
			Details: map[string]string{"path": path, "mode": mode},
		}
	}

	cfg.GetLoggerOrDefault().Info("file created", "path", path)
	return types.Result{
		Changed: true,
		Message: "File created: " + path,
		Details: map[string]string{
			"path":  path,
			"mode":  mode,
			"owner": owner,
		},
	}
}

// shellEscapeContent escapes file content for safe use in a printf '%s' argument.
// It wraps the content in single quotes and escapes embedded single quotes
// using the POSIX sequence '\”.
func shellEscapeContent(content string) string {
	return "'" + strings.ReplaceAll(content, "'", "'\\''") + "'"
}

// SetPath sets the file path and returns FileCreate for chaining.
func (f *FileCreate) SetPath(path string) *FileCreate {
	f.BaseSkill.SetArg(ArgPath, path)
	return f
}

// SetContent sets the file content and returns FileCreate for chaining.
func (f *FileCreate) SetContent(content string) *FileCreate {
	f.BaseSkill.SetArg(ArgContent, content)
	return f
}

// SetOwner sets the owner (user:group) and returns FileCreate for chaining.
func (f *FileCreate) SetOwner(owner string) *FileCreate {
	f.BaseSkill.SetArg(ArgOwner, owner)
	return f
}

// SetMode sets the permissions (octal, e.g. "644") and returns FileCreate for chaining.
func (f *FileCreate) SetMode(mode string) *FileCreate {
	f.BaseSkill.SetArg(ArgMode, mode)
	return f
}

// SetOverwrite sets whether to overwrite if file exists and returns FileCreate for chaining.
func (f *FileCreate) SetOverwrite(overwrite bool) *FileCreate {
	f.BaseSkill.SetArg(ArgOverwrite, fmt.Sprintf("%v", overwrite))
	return f
}

// WithNodeConfig sets the node config and returns FileCreate for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (f *FileCreate) WithNodeConfig(cfg types.NodeConfig) *FileCreate {
	f.BaseSkill.SetNodeConfig(cfg)
	return f
}

// SetArgs sets the arguments and returns FileCreate for fluent chaining.
func (f *FileCreate) SetArgs(args map[string]string) types.RunnableInterface {
	f.BaseSkill.SetArgs(args)
	return f
}

// SetArg sets a single argument and returns FileCreate for fluent chaining.
func (f *FileCreate) SetArg(key, value string) types.RunnableInterface {
	f.BaseSkill.SetArg(key, value)
	return f
}

// SetID sets the ID and returns FileCreate for fluent chaining.
func (f *FileCreate) SetID(id string) types.RunnableInterface {
	f.BaseSkill.SetID(id)
	return f
}

// SetDescription sets the description and returns FileCreate for fluent chaining.
func (f *FileCreate) SetDescription(description string) types.RunnableInterface {
	f.BaseSkill.SetDescription(description)
	return f
}

// SetTimeout sets the timeout and returns FileCreate for fluent chaining.
func (f *FileCreate) SetTimeout(timeout time.Duration) types.RunnableInterface {
	f.BaseSkill.SetTimeout(timeout)
	return f
}

// NewFileCreate creates a new fs-file-create skill.
func NewFileCreate() *FileCreate {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSFileCreate)
	pb.SetDescription("Create file with content, ownership, and permissions")
	return &FileCreate{BaseSkill: pb}
}
