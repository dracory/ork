package caddy

import (
	"time"

	"github.com/dracory/ork/skills"
	"github.com/dracory/ork/skills/systemctl"
	"github.com/dracory/ork/types"
)

// Status shows the status of the Caddy systemd service.
// This is a read-only skill: it never modifies system state, so Changed is
// always false and Check always returns false.
//
// Usage:
//
//	node.Run(caddy.NewStatus())
//
// Execution Flow:
//  1. Runs `systemctl status caddy --no-pager -l` on the remote server
//  2. Returns the full status output in Result.Details["output"]
//
// Result Details:
//   - output: Full output from systemctl status
//   - service: The unit name that was queried ("caddy")
type Status struct {
	*types.BaseSkill
}

// Compile-time assertion that Status implements types.RunnableInterface.
var _ types.RunnableInterface = (*Status)(nil)

// Check always returns false since Status is read-only.
func (s *Status) Check() (bool, error) {
	return false, nil
}

// Run shows the Caddy service status.
// Changed is always false since this is a read-only operation.
func (s *Status) Run() types.Result {
	cfg := s.GetNodeConfig()
	statusResult := runSub(systemctl.NewStatus().SetService(DefaultCaddyService), cfg)

	return types.Result{
		Changed: false,
		Message: "Caddy status:\n" + statusResult.Details["output"],
		Details: statusResult.Details,
	}
}

// SetArgs sets the arguments for the Caddy status query.
// Returns Status for fluent method chaining.
func (s *Status) SetArgs(args map[string]string) types.RunnableInterface {
	s.BaseSkill.SetArgs(args)
	return s
}

// SetArg sets a single argument for the Caddy status query.
// Returns Status for fluent method chaining.
func (s *Status) SetArg(key, value string) types.RunnableInterface {
	s.BaseSkill.SetArg(key, value)
	return s
}

// SetID sets the ID for the Caddy status query.
// Returns Status for fluent method chaining.
func (s *Status) SetID(id string) types.RunnableInterface {
	s.BaseSkill.SetID(id)
	return s
}

// SetDescription sets the description for the Caddy status query.
// Returns Status for fluent method chaining.
func (s *Status) SetDescription(description string) types.RunnableInterface {
	s.BaseSkill.SetDescription(description)
	return s
}

// SetTimeout sets the timeout for the Caddy status query.
// Returns Status for fluent method chaining.
func (s *Status) SetTimeout(timeout time.Duration) types.RunnableInterface {
	s.BaseSkill.SetTimeout(timeout)
	return s
}

// NewStatus creates a new caddy-status skill.
//
// Returns a Status skill configured with skills.IDCaddyStatus identifier and
// description "Show Caddy systemd service status".
//
// Example:
//
//	node.Run(caddy.NewStatus())
func NewStatus() *Status {
	pb := types.NewBaseSkill()
	pb.SetID(skills.IDCaddyStatus)
	pb.SetDescription("Show Caddy systemd service status")
	return &Status{BaseSkill: pb}
}
