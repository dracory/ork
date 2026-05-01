package ork

import "github.com/dracory/ork/types"

// NewPlaybook creates a new base playbook for custom playbook implementations.
// Use this for creating custom playbooks with fluent configuration.
//
// Example:
//
//	pb := ork.NewPlaybook().
//	    WithID("my-playbook").
//	    WithDescription("What this playbook does").
//	    WithDryRun(false)
//
//	node := ork.NewNodeForHost("server.example.com")
//	node.SetNodeConfig(cfg)
//	result := pb.Run()
func NewPlaybook() *types.BasePlaybook {
	return types.NewBasePlaybook()
}
