// Package skills provides built-in reusable skill implementations
// for common server automation tasks.
//
// Each subpackage contains skills for a specific domain:
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
// All skills implement the types.SkillInterface and can be used
// with the ork.Node.Skill() method or registered with a types.Registry.
//
// Example usage:
//
//	node := ork.NewNodeForHost("server.example.com")
//	result := node.Skill(skills.IDAptUpdate)
//
// Or with explicit skill creation:
//
//	skill := apt.NewAptUpdate()
//	skill.SetNodeConfig(node.GetNodeConfig())
//	result := skill.Run()
//
// Skills follow the idempotency principle: Check() determines if changes
// are needed, Run() executes the operation. The Result.Changed field indicates
// whether any modifications were made.
package skills
