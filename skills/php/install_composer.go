package php

// Package php documentation is in constants.go

import (
	"fmt"
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/ssh"
	"github.com/dracory/ork/types"
)

// InstallComposer installs Composer (PHP dependency manager) with signature verification.
//
// It downloads the Composer installer signature, downloads the installer itself,
// verifies the signature, installs Composer to /usr/local/bin/composer, and
// cleans up the temporary installer files.
//
// No args needed.
//
// Execution Flow:
//  1. Downloads installer signature from https://composer.github.io/installer.sig
//  2. Downloads installer from https://getcomposer.org/installer
//  3. Verifies signature
//  4. Installs Composer to /usr/local/bin/composer
//  5. Cleans up installer files
//
// Usage:
//
//	node.Run(php.NewInstallComposer())
//
// Idempotency:
//   - Check() verifies that /usr/local/bin/composer exists. When it does,
//     Run() is a no-op.
type InstallComposer struct {
	*types.BaseSkill
}

// Compile-time assertion that InstallComposer implements types.RunnableInterface.
var _ types.RunnableInterface = (*InstallComposer)(nil)

// Check determines if Composer needs to be installed.
// Returns true if /usr/local/bin/composer does not exist, false if it does.
func (s *InstallComposer) Check() (bool, error) {
	cfg := s.GetNodeConfig()

	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would check if Composer is installed")
		return true, nil
	}

	// Check if Composer binary exists.
	cmdCheck := types.Command{
		Command:     fmt.Sprintf("test -f %s", skills.ShellEscapeArg(ComposerBinaryPath)),
		Description: "Check if Composer is installed",
		Required:    false,
	}
	_, err := ssh.Run(cfg, cmdCheck)
	if err != nil {
		// Binary missing — needs install.
		return true, nil
	}
	return false, nil // already installed
}

// Run installs Composer with signature verification.
// Changed is true when Composer was installed, false when it was already
// present.
//
// Result.Details contains:
//   - output: Full output from the install commands
func (s *InstallComposer) Run() types.Result {
	cfg := s.GetNodeConfig()

	// Check if Composer is already installed (idempotency).
	needsChange, err := s.Check()
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check Composer installation",
			Error:   err,
		}
	}

	if !needsChange {
		return types.Result{
			Changed: false,
			Message: "Composer is already installed",
		}
	}

	// Check for dry-run mode.
	if cfg.IsDryRunMode {
		cfg.GetLoggerOrDefault().Info("dry-run: would install Composer")
		return types.Result{
			Changed: true,
			Message: "Would install Composer",
		}
	}

	// Download the installer signature.
	cmdDownloadSig := types.Command{
		Command:     fmt.Sprintf("wget -q -O %s %s", skills.ShellEscapeArg(ComposerInstallerSigPath), skills.ShellEscapeArg(ComposerSigUrl)),
		Description: "Download Composer installer signature",
	}
	if _, err := ssh.Run(cfg, cmdDownloadSig); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to download Composer signature",
			Error:   fmt.Errorf("failed to download signature from %s: %w", ComposerSigUrl, err),
		}
	}

	// Download the installer.
	cmdDownloadInstaller := types.Command{
		Command:     fmt.Sprintf("wget -q -O %s %s", skills.ShellEscapeArg(ComposerInstallerPath), skills.ShellEscapeArg(ComposerInstallerUrl)),
		Description: "Download Composer installer",
	}
	if _, err := ssh.Run(cfg, cmdDownloadInstaller); err != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to download Composer installer",
			Error:   fmt.Errorf("failed to download installer from %s: %w", ComposerInstallerUrl, err),
		}
	}

	// Verify the signature.
	cmdVerify := types.Command{
		Command: fmt.Sprintf(
			"php -r \"if (hash_file('sha384', '%s') === file_get_contents('%s')) { echo 'Installer verified'; } else { echo 'Installer corrupt'; unlink('%s'); exit(1); }\"",
			ComposerInstallerPath, ComposerInstallerSigPath, ComposerInstallerPath,
		),
		Description: "Verify Composer installer signature",
	}
	verifyOutput, err := ssh.Run(cfg, cmdVerify)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Composer signature verification failed",
			Error:   fmt.Errorf("signature verification failed: %w\nOutput: %s", err, verifyOutput),
		}
	}

	// Install Composer.
	cmdInstall := types.Command{
		Command:     fmt.Sprintf("php %s --install-dir=/usr/local/bin --filename=composer", skills.ShellEscapeArg(ComposerInstallerPath)),
		Description: "Install Composer to " + ComposerBinaryPath,
	}
	installOutput, err := ssh.Run(cfg, cmdInstall)
	if err != nil {
		return types.Result{
			Changed: false,
			Message: "Composer installation failed",
			Error:   fmt.Errorf("composer install failed: %w\nOutput: %s", err, installOutput),
		}
	}

	// Clean up installer files.
	cmdCleanup := types.Command{
		Command:     fmt.Sprintf("rm -f %s %s", skills.ShellEscapeArg(ComposerInstallerPath), skills.ShellEscapeArg(ComposerInstallerSigPath)),
		Description: "Clean up Composer installer files",
	}
	ssh.Run(cfg, cmdCleanup) // best effort

	cfg.GetLoggerOrDefault().Info("Composer installed")
	return types.Result{
		Changed: true,
		Message: "Composer installed",
		Details: map[string]string{
			"output":  installOutput,
			"verify":  verifyOutput,
			"binary":  ComposerBinaryPath,
		},
	}
}

// SetArgs sets the arguments for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *InstallComposer) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// WithNodeConfig sets the node config and returns InstallComposer for chaining.
// Shortcut alias to SetNodeConfig for fluent interface convenience.
func (s *InstallComposer) WithNodeConfig(cfg types.NodeConfig) *InstallComposer {
	s.BaseSkill.SetNodeConfig(cfg)
	return s
}

// SetArg sets a single argument for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *InstallComposer) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *InstallComposer) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *InstallComposer) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for this skill.
// Returns RunnableInterface for fluent method chaining.
func (s *InstallComposer) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewInstallComposer creates a new php-install-composer skill.
//
// Returns:
//
//	An InstallComposer skill configured with skills.IDPhpInstallComposer identifier
//	and description "Install Composer (PHP dependency manager)".
//
// Example:
//
//	node.Run(php.NewInstallComposer())
func NewInstallComposer() *InstallComposer {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDPhpInstallComposer)
	pb.SetDescription("Install Composer (PHP dependency manager)")
	return &InstallComposer{BaseSkill: pb}
}
