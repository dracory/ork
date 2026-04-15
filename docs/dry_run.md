# Dry-Run Mode

Preview what changes would be made without actually executing commands on the server. Safety is enforced at the SSH execution layer - **no commands execute on the server in dry-run mode**.

## Enable Dry-Run

### Node Level

```go
node := ork.NewNodeForHost("server.example.com").
    SetDryRunMode(true)
results := node.RunSkill(skills.NewAptUpgrade())
// Commands are logged but not executed
```

### Group Level

```go
webGroup := ork.NewGroup("webservers")
webGroup.SetDryRunMode(true)
webGroup.AddNode(node1)
webGroup.AddNode(node2)
// All nodes inherit dry-run mode
```

### Inventory Level

```go
inv := ork.NewInventory()
inv.SetDryRunMode(true)
inv.AddGroup(webGroup)
results := inv.RunCommand("uptime")
// All groups and nodes inherit dry-run mode
```

## How It Works

1. **Safety at execution layer**: `ssh.Run()` checks `cfg.IsDryRunMode` and returns `"[dry-run]"` without executing commands
2. **Automatic propagation**: Dry-run mode propagates from Inventory → Groups → Nodes at execution time
3. **Thread-safe**: Uses mutex protection for concurrent access to dry-run state

## Detecting Dry-Run in Skills

```go
func (s *MySkill) Run() skill.Result {
    output, _ := ssh.Run(s.cfg, "apt-get upgrade -y")

    if output == "[dry-run]" {
        return skill.Result{
            Changed: true,
            Message: "Would run: apt-get upgrade -y",
        }
    }
    // Normal execution handling...
}
```

**Note:** Even if a skill doesn't check for the `[dry-run]` marker, **safety is guaranteed** - no commands execute on the server when dry-run mode is enabled.

## Best Practices

1. **Always test with dry-run first**: Before running potentially destructive operations
2. **Use with CheckSkill**: Combine dry-run with idempotency checks for maximum safety
3. **Review output**: Check the dry-run output to ensure it matches expectations
4. **Document dry-run behavior**: Custom skills should document their dry-run behavior

## Example Workflow

```go
// Step 1: Check if changes needed
checkResults := node.CheckSkill(skills.NewAptUpgrade())
if !checkResults.Results["server.example.com"].Changed {
    log.Println("No updates needed")
    return
}

// Step 2: Dry-run to preview
node.SetDryRunMode(true)
dryRunResults := node.RunSkill(skills.NewAptUpgrade())
log.Printf("Dry-run output: %s", dryRunResults.Results["server.example.com"].Message)

// Step 3: Disable dry-run and execute
node.SetDryRunMode(false)
results := node.RunSkill(skills.NewAptUpgrade())
log.Printf("Execution result: %s", results.Results["server.example.com"].Message)
```
