// Package playbooks provides reusable playbook implementations for common
// server automation tasks.
//
// Each subpackage contains playbooks for a specific domain:
//
//   - apt: Package management (update, upgrade, status)
//   - ping: SSH connectivity checks
//   - reboot: Server reboot with optional reconnection wait
//   - swap: Swap file management (create, delete, status)
//   - user: User management (create, delete, status)
//
// All playbooks implement the playbook.PlaybookInterface and can be used
// with the ork.Node.Playbook() method or registered with a playbook.Registry.
//
// Example usage:
//
//	node := ork.NewNodeForHost("server.example.com")
//	result := node.Playbook(playbook.IDAptUpdate)
//
// Or with explicit playbook creation:
//
//	pb := apt.NewAptUpdate()
//	pb.SetConfig(node.GetConfig())
//	result := pb.Run()
//
// Playbooks follow the idempotency principle: Check() determines if changes
// are needed, Run() executes the operation. The Result.Changed field indicates
// whether any modifications were made.
package playbooks
