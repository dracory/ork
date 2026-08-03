package security

import (
	"fmt"
	"strings"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// SshKeyGen generates a server-side SSH keypair for a user so it can
// authenticate to a git host (e.g. GitLab/GitHub) for git clone / git pull.
// This is the server's own key — NOT the operator's local key.
//
// The public key is returned in the result message so it can be added to the
// git host as a Deploy Key (e.g. GitLab: Settings → Repository → Deploy Keys).
//
// Idempotent: if the private key already exists, generation is skipped and
// the existing public key is returned.
//
// Usage:
//
//	node.Run(security.NewSshKeyGen().SetUsername("deploy"))
//	// or with full configuration:
//	node.Run(security.NewSshKeyGen().
//	    SetUsername("deploy").
//	    SetKeyType("ed25519").
//	    SetComment("deploy@my-server"))
//
// Args:
//   - username: User account to generate the keypair for (required)
//   - key-type: SSH key type (default: ed25519)
//   - comment:  Comment embedded in the public key (-C). If empty, ssh-keygen
//     uses its default (<user>@<host>).
//   - key-path: Private key file path (default: /home/<username>/.ssh/id_<key-type>)
//
// Execution Flow:
//  1. Ensures the user's ~/.ssh directory exists with correct ownership/permissions
//  2. Generates the keypair only if the private key doesn't already exist
//  3. Reads and returns the public key
//
// Prerequisites:
//   - Root SSH access (the connected user must be able to chown to <username>)
//   - The target user must already exist (use user-create first)
//
// Related Playbooks:
//   - user-create: Create the target user before generating its keypair
type SshKeyGen struct {
	*types.BaseSkill
}

// Compile-time assertion that SshKeyGen implements types.RunnableInterface.
var _ types.RunnableInterface = (*SshKeyGen)(nil)

// Check always returns true; idempotency is handled inside Run via test -f.
func (s *SshKeyGen) Check() (bool, error) {
	return true, nil
}

// Run generates the server SSH keypair if it doesn't exist and returns the
// public key in the result message.
func (s *SshKeyGen) Run() types.Result {
	cfg := s.GetNodeConfig()

	username := s.GetArg(ArgUsername)
	if username == "" {
		return types.Result{
			Changed: false,
			Message: "Username is required",
			Error:   fmt.Errorf("username is required (pass via --arg=username=value)"),
		}
	}

	keyType := s.GetArg(ArgKeyType)
	if keyType == "" {
		keyType = DefaultKeyType
	}

	comment := s.GetArg(ArgComment)

	keyPath := s.GetArg(ArgKeyPath)
	if keyPath == "" {
		keyPath = fmt.Sprintf("/home/%s/.ssh/id_%s", username, keyType)
	}
	sshDir := keyPath[:strings.LastIndex(keyPath, "/")]

	// Ensure the .ssh directory exists and is owned by the target user.
	// Wrapped in sh -c so && stays inside sudo's scope when BecomeUser is set.
	innerDir := fmt.Sprintf("mkdir -p %s && chmod 700 %s && chown %s:%s %s",
		skills.ShellEscapeArg(sshDir), skills.ShellEscapeArg(sshDir),
		skills.ShellEscapeArg(username), skills.ShellEscapeArg(username),
		skills.ShellEscapeArg(sshDir))
	cmdEnsureDir := types.Command{
		Command:     fmt.Sprintf("sh -c %s", skills.ShellEscapeArg(innerDir)),
		Description: "Ensure .ssh directory exists for " + username,
		Required:    true,
	}

	// Generate the keypair only if it doesn't already exist (idempotent).
	// Using a single shell command with `test -f ... || ssh-keygen ...` avoids
	// the ssh.Run error-suppression behavior for Required:false commands —
	// ssh.Run returns nil error for non-required failures, so a separate
	// check call would not reliably signal "key missing".
	keygenCmd := fmt.Sprintf("sudo -u %s ssh-keygen -t %s -f %s -N ''",
		skills.ShellEscapeArg(username), skills.ShellEscapeArg(keyType), skills.ShellEscapeArg(keyPath))
	if comment != "" {
		keygenCmd += fmt.Sprintf(" -C %s", skills.ShellEscapeArg(comment))
	}
	innerGen := fmt.Sprintf("test -f %s || %s", skills.ShellEscapeArg(keyPath), keygenCmd)
	cmdGenIfMissing := types.Command{
		Command:     fmt.Sprintf("sh -c %s", skills.ShellEscapeArg(innerGen)),
		Description: "Generate SSH keypair for " + username + " (if missing)",
		Required:    true,
	}

	// Read the public key so it can be surfaced to the operator.
	cmdReadPub := types.Command{
		Command:     fmt.Sprintf("cat %s", skills.ShellEscapeArg(keyPath+".pub")),
		Description: "Read SSH public key",
		Required:    true,
	}

	// Dry-run mode: log commands and return without executing.
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdEnsureDir.Command, "description", cmdEnsureDir.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdGenIfMissing.Command, "description", cmdGenIfMissing.Description)
		cfg.GetLoggerOrDefault().Info("dry-run: would run command", "cmd", cmdReadPub.Command, "description", cmdReadPub.Description)
		return types.Result{
			Changed: true,
			Message: fmt.Sprintf("Would generate SSH keypair for %s", username),
		}
	}

	cfg.GetLoggerOrDefault().Info("ensuring .ssh directory", "user", username)
	if output, err := ssh.Run(cfg, cmdEnsureDir); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to create .ssh directory for " + username,
			Error:   fmt.Errorf("%w: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("generating SSH keypair if missing", "user", username, "key_path", keyPath)
	if output, err := ssh.Run(cfg, cmdGenIfMissing); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to generate SSH keypair",
			Error:   fmt.Errorf("%w: %s", err, output),
		}
	}

	cfg.GetLoggerOrDefault().Info("reading SSH public key", "key_path", keyPath+".pub")
	pubKeyOutput, err := ssh.Run(cfg, cmdReadPub)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to read SSH public key",
			Error:   fmt.Errorf("%w: %s", err, pubKeyOutput),
		}
	}
	pubKey := strings.TrimSpace(pubKeyOutput)

	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("SSH keypair ready for %s\n\nAdd this public key to your git host as a Deploy Key (e.g. GitLab: Settings → Repository → Deploy Keys):\n%s", username, pubKey),
		Details: map[string]string{
			"username":  username,
			"key_path":  keyPath,
			"key_type":  keyType,
			"pub_key":   pubKey,
		},
	}
}

// SetArgs sets the arguments for SSH key generation.
// Returns SshKeyGen for fluent method chaining.
func (s *SshKeyGen) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetUsername sets the user account to generate the keypair for (required).
// Returns SshKeyGen for chaining.
func (s *SshKeyGen) SetUsername(username string) *SshKeyGen {
	s.BaseSkill.SetArg(ArgUsername, username)
	return s
}

// SetKeyType sets the SSH key type (ed25519, rsa, ecdsa). Defaults to ed25519.
// Returns SshKeyGen for chaining.
func (s *SshKeyGen) SetKeyType(keyType string) *SshKeyGen {
	s.BaseSkill.SetArg(ArgKeyType, keyType)
	return s
}

// SetComment sets the comment embedded in the public key (-C). If empty,
// ssh-keygen uses its default (<user>@<host>).
// Returns SshKeyGen for chaining.
func (s *SshKeyGen) SetComment(comment string) *SshKeyGen {
	s.BaseSkill.SetArg(ArgComment, comment)
	return s
}

// SetKeyPath sets the private key file path. Defaults to
// /home/<username>/.ssh/id_<key-type>.
// Returns SshKeyGen for chaining.
func (s *SshKeyGen) SetKeyPath(keyPath string) *SshKeyGen {
	s.BaseSkill.SetArg(ArgKeyPath, keyPath)
	return s
}

// SetArg sets a single argument for SSH key generation.
// Returns SshKeyGen for fluent method chaining.
func (s *SshKeyGen) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for SSH key generation.
// Returns SshKeyGen for fluent method chaining.
func (s *SshKeyGen) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for SSH key generation.
// Returns SshKeyGen for fluent method chaining.
func (s *SshKeyGen) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for SSH key generation.
// Returns SshKeyGen for fluent method chaining.
func (s *SshKeyGen) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewSshKeyGen creates a new ssh-key-gen skill.
func NewSshKeyGen() *SshKeyGen {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDSshKeyGen)
	pb.SetDescription("Generate a server-side SSH keypair for a user (prints public key for git host Deploy Key)")
	return &SshKeyGen{BaseSkill: pb}
}
