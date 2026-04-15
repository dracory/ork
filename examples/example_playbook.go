package examples

import (
	"github.com/dracory/ork/skills/apt"
	"github.com/dracory/ork/skills/ping"
	"github.com/dracory/ork/types"
)

// ExamplePlaybook demonstrates a basic playbook that runs multiple skills in sequence.
// This playbook checks system status and updates if needed.
type ExamplePlaybook struct {
	*types.BasePlaybook
}

// NewExamplePlaybook creates a new example playbook.
func NewExamplePlaybook() types.RunnableInterface {
	playbook := types.NewBasePlaybook()
	playbook.SetID("example-playbook")
	playbook.SetDescription("Example playbook demonstrating sequential skill execution")
	return &ExamplePlaybook{BasePlaybook: playbook}
}

// Run executes the playbook with custom orchestration logic.
func (e *ExamplePlaybook) Run() types.Result {
	cfg := e.GetNodeConfig()

	// Step 1: Check connectivity
	pingSkill := ping.NewPing()
	pingSkill.SetNodeConfig(cfg)
	pingResult := pingSkill.Run()

	if pingResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to check connectivity",
			Error:   pingResult.Error,
		}
	}

	// Step 2: Check for package updates
	updateSkill := apt.NewAptUpdate()
	updateSkill.SetNodeConfig(cfg)
	updateResult := updateSkill.Run()

	if updateResult.Error != nil {
		return types.Result{
			Changed: false,
			Message: "Failed to update package list",
			Error:   updateResult.Error,
		}
	}

	return types.Result{
		Changed: updateResult.Changed,
		Message: "Example playbook completed successfully",
		Details: map[string]string{
			"ping":   pingResult.Message,
			"update": updateResult.Message,
		},
	}
}
