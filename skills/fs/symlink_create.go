package fs

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// SymlinkCreate creates or updates a symbolic link (ln -sf).
// If the symlink already points to the correct target, no change is made.
//
// Usage:
//
//	node.Run(fs.NewSymlinkCreate().SetArgs(map[string]string{
//	    fs.ArgTarget: "/opt/node/bin/pm2",
//	    fs.ArgLink:   "/usr/local/bin/pm2",
//	}))
//
// Args:
//   - target: Path the symlink points to (required, must be absolute)
//   - link: Path of the symlink itself (required, must be absolute)
//
// Idempotency:
//   - Check() returns false if symlink already points to correct target
//   - Check() returns true if symlink missing or points elsewhere
type SymlinkCreate struct {
	*types.BaseSkill
}

// Check determines if the symlink needs to be created or updated.
// Returns true if symlink is missing or points to wrong target.
func (s *SymlinkCreate) Check() (bool, error) {
	cfg := s.GetNodeConfig()
	target := s.GetArg(ArgTarget)
	link := s.GetArg(ArgLink)

	if err := validatePath(target); err != nil {
		return false, err
	}
	if err := validatePath(link); err != nil {
		return false, err
	}

	// In dry-run mode, assume changes are needed without running SSH commands
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if symlink creation is needed")
		return true, nil
	}

	// Check if symlink exists and what it points to
	currentTarget := getSymlinkTarget(cfg, link)
	return currentTarget != target, nil
}

// Run creates or updates the symlink.
// Changed is true when symlink was created or updated, false if already correct.
func (s *SymlinkCreate) Run() types.Result {
	target := s.GetArg(ArgTarget)
	link := s.GetArg(ArgLink)

	if err := validatePath(target); err != nil {
		return types.Result{Changed: false, Message: "Invalid target", Error: err}
	}
	if err := validatePath(link); err != nil {
		return types.Result{Changed: false, Message: "Invalid link path", Error: err}
	}

	cfg := s.GetNodeConfig()

	cmdLn := types.Command{
		Command:     fmt.Sprintf("ln -sf %s %s", skills.ShellEscapeArg(target), skills.ShellEscapeArg(link)),
		Description: "Create symlink: " + link + " -> " + target,
	}

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdLn.Command)
		return types.Result{
			Changed: true,
			Message: "Would create symlink: " + link + " -> " + target,
		}
	}

	needsChange, err := s.Check()
	if err != nil {
		return types.Result{Changed: false, Message: "Failed to check symlink", Error: err}
	}
	if !needsChange {
		return types.Result{
			Changed: false,
			Message: "Symlink already correct: " + link + " -> " + target,
		}
	}

	cfg.GetLoggerOrDefault().Info("creating symlink", "link", link, "target", target)
	output, err := ssh.Run(cfg, cmdLn)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to create symlink",
			Error:   fmt.Errorf("failed to create symlink: %w\nOutput: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("symlink created", "link", link, "target", target)
	return types.Result{
		Changed: true,
		Message: "Symlink created: " + link + " -> " + target,
		Details: map[string]string{"link": link, "target": target},
	}
}

// SetTarget sets the symlink target path and returns SymlinkCreate for chaining.
func (s *SymlinkCreate) SetTarget(target string) *SymlinkCreate {
	s.BaseSkill.SetArg(ArgTarget, target)
	return s
}

// SetLink sets the symlink path itself and returns SymlinkCreate for chaining.
func (s *SymlinkCreate) SetLink(link string) *SymlinkCreate {
	s.BaseSkill.SetArg(ArgLink, link)
	return s
}

// WithNodeConfig sets the node config and returns SymlinkCreate for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (s *SymlinkCreate) WithNodeConfig(cfg types.NodeConfig) *SymlinkCreate {
	s.BaseSkill.SetNodeConfig(cfg)
	return s
}

// SetArgs sets the arguments and returns SymlinkCreate for fluent chaining.
func (s *SymlinkCreate) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetArg sets a single argument and returns SymlinkCreate for fluent chaining.
func (s *SymlinkCreate) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID and returns SymlinkCreate for fluent chaining.
func (s *SymlinkCreate) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description and returns SymlinkCreate for fluent chaining.
func (s *SymlinkCreate) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout and returns SymlinkCreate for fluent chaining.
func (s *SymlinkCreate) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewSymlinkCreate creates a new fs-symlink-create skill.
func NewSymlinkCreate() *SymlinkCreate {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDFSSymlinkCreate)
	pb.SetDescription("Create or update symbolic link (ln -sf)")
	return &SymlinkCreate{BaseSkill: pb}
}
