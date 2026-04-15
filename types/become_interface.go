package types

// BecomeInterface defines privilege escalation capabilities.
// It allows running commands as a different user via sudo.
//
// Example:
//
//	node.SetBecomeUser("root")
//	node.RunCommand("apt-get update")  // Runs: sudo -u root apt-get update
type BecomeInterface interface {
	// SetBecomeUser sets the user to become when executing commands.
	// If empty, no privilege escalation is performed.
	// Returns BecomeInterface for fluent method chaining.
	SetBecomeUser(user string) BecomeInterface

	// GetBecomeUser returns the configured become user.
	// Returns empty string if not set.
	GetBecomeUser() string
}

// BaseBecome provides a default implementation of BecomeInterface.
// Embed this in structs that need privilege escalation support.
type BaseBecome struct {
	becomeUser string
}

// SetBecomeUser sets the user to become when executing commands.
// Returns BecomeInterface for fluent method chaining.
func (b *BaseBecome) SetBecomeUser(user string) BecomeInterface {
	b.becomeUser = user
	return b
}

// GetBecomeUser returns the configured become user.
// Returns empty string if not set.
func (b *BaseBecome) GetBecomeUser() string {
	return b.becomeUser
}
