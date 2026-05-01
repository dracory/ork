package ork

import "github.com/dracory/ork/types"

// NewSkill creates a new base skill for custom skill implementations.
// Use this for creating custom skills with fluent configuration.
//
// Example:
//
//	skill := ork.NewSkill().
//	    WithID("my-skill").
//	    WithDescription("What this skill does").
//	    WithDryRun(false)
//
//	node := ork.NewNodeForHost("server.example.com")
//	node.SetNodeConfig(cfg)
//	result := skill.Run()
func NewSkill() *types.BaseSkill {
	return types.NewBaseSkill()
}
