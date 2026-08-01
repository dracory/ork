package ssh

import (
	"errors"

	"golang.org/x/crypto/ssh"
)

// IsExitError returns true if the error is a command exit error (i.e., the
// command was executed on the remote server but returned a non-zero exit
// code), as opposed to a connection or session creation failure.
//
// This allows callers to distinguish between "the command ran and exited
// non-zero" (which may be a valid result for commands like
// `systemctl is-enabled` that exit non-zero for disabled units) and "the SSH
// connection itself failed" (which is always a real error that should be
// propagated).
//
// Usage:
//
//	output, err := ssh.Run(cfg, cmd)
//	if err != nil {
//	    if ssh.IsExitError(err) {
//	        // Command ran but exited non-zero — may be a valid result
//	    } else {
//	        // SSH connection or session failure — propagate as error
//	    }
//	}
func IsExitError(err error) bool {
	var exitErr *ssh.ExitError
	return errors.As(err, &exitErr)
}

// NewExitError creates a *ssh.ExitError for testing purposes.
// This allows tests to simulate command exit failures without importing
// golang.org/x/crypto/ssh directly.
func NewExitError() error {
	return &ssh.ExitError{}
}
