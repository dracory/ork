package caddy

import "github.com/dracory/ork/types"

// runSub propagates the node config to a sub-skill and runs it, returning the
// single result. This mirrors what node.Run does internally (SetNodeConfig +
// Run) without requiring a Node, allowing caddy skills to compose sibling
// skills (apt, fs, systemctl) directly.
//
// The sub-skill must be freshly constructed (not shared across goroutines) so
// that SetNodeConfig does not race with concurrent executions.
func runSub(skill types.RunnableInterface, cfg types.NodeConfig) types.Result {
	skill.SetNodeConfig(cfg)
	skill.SetDryRun(cfg.IsDryRunMode)
	return skill.Run()
}
