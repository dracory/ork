// Package playbooks provides built-in reusable playbook implementations
// for common server automation tasks.
//
// Each subpackage contains playbooks for a specific domain:
//
//   - apt: Package management (update, upgrade, status)
//   - fail2ban: Fail2ban service management (status, start, stop)
//   - mariadb: MariaDB database management (status, start, stop)
//   - ping: SSH connectivity checks
//   - reboot: Server reboot with optional reconnection wait
//   - swap: Swap file management (create, delete, status)
//   - ufw: Uncomplicated Firewall management (status, enable, disable)
//   - user: User management (create, delete, status)
//
// All playbooks implement the types.PlaybookInterface and can be used
// with the ork.Node.Playbook() method or registered with a types.Registry.
//
// Example usage:
//
//	node := ork.NewNodeForHost("server.example.com")
//	result := node.Playbook(playbooks.IDAptUpdate)
//
// Or with explicit playbook creation:
//
//	pb := apt.NewAptUpdate()
//	pb.SetNodeConfig(node.GetNodeConfig())
//	result := pb.Run()
//
// Playbooks follow the idempotency principle: Check() determines if changes
// are needed, Run() executes the operation. The Result.Changed field indicates
// whether any modifications were made.
package playbooks
