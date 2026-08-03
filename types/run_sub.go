package types

// RunSub propagates the node config to a sub-skill and runs it, returning the
// single result. This mirrors what node.Run does internally (SetNodeConfig +
// SetDryRun + Run) without requiring a Node, allowing skills to compose
// sibling skills (apt, fs, systemctl) directly.
//
// The sub-skill must be freshly constructed (not shared across goroutines) so
// that SetNodeConfig does not race with concurrent executions.
func RunSub(skill RunnableInterface, cfg NodeConfig) Result {
	skill.SetNodeConfig(cfg)
	skill.SetDryRun(cfg.IsDryRunMode)
	return skill.Run()
}
