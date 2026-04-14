---
path: troubleshooting.md
page-type: reference
summary: Common issues and solutions when working with Ork.
tags: [troubleshooting, errors, faq]
created: 2025-04-14
updated: 2025-04-14
version: 1.0.0
---

# Troubleshooting

This document covers common issues and their solutions when using Ork.

## Connection Issues

### "failed to connect to server.example.com:22"

**Problem**: SSH connection cannot be established.

**Possible Causes & Solutions**:

1. **Wrong hostname or IP**
   ```go
   // Verify the host is correct
   node := ork.NewNodeForHost("server.example.com")
   ```
   Test with: `ping server.example.com`

2. **Wrong port**
   ```go
   // Default is 22, but your server might use a different port
   node.SetPort("2222")  // or your custom port
   ```

3. **SSH key not found**
   ```go
   // Default key is "id_rsa" in ~/.ssh/
   // Make sure the key exists:
   node.SetKey("your_key_file")
   ```
   Verify: `ls -la ~/.ssh/`

4. **Key permissions too open**
   ```bash
   # Fix permissions
   chmod 600 ~/.ssh/your_key_file
   chmod 700 ~/.ssh/
   ```

5. **User doesn't exist**
   ```go
   // Default user is "root"
   node.SetUser("your_username")
   ```

### "host cannot be empty"

**Problem**: Trying to connect without setting a host.

**Solution**: Use `NewNodeForHost()` or set SSHHost explicitly:

```go
// Correct
node := ork.NewNodeForHost("server.example.com")

// Incorrect - this creates empty host
node := ork.NewNode()
```

## Playbook Issues

### "playbook 'xyz' not found in registry"

**Problem**: Trying to run a playbook by ID that isn't registered.

**Solution**: Use the correct ID or run the playbook directly:

```go
// By ID (must be registered)
results := node.RunPlaybookByID(playbook.IDAptUpdate)

// Direct instance (preferred)
results := node.RunPlaybook(playbooks.NewAptUpdate())
```

Check available IDs in `playbook/constants.go`.

### "username is required"

**Problem**: Running a user playbook without required arguments.

**Solution**: Set the required argument:

```go
node.SetArg("username", "alice")
results := node.RunPlaybook(playbooks.NewUserCreate())
```

### Playbook reports "Changed: false" but should have changed

**Problem**: Idempotency check incorrectly reports no changes needed.

**Possible Causes**:
1. Check command is wrong
2. System state detection is incorrect

**Debug**:
```go
// Manually run the check command
results := node.RunCommand("your-check-command")
log.Println(results.Results["host"].Message)
```

## SSH Authentication Issues

### "ssh: handshake failed: ssh: unable to authenticate"

**Problem**: SSH key authentication failed.

**Solutions**:

1. **Key not authorized on server**
   ```bash
   # On server, add your public key
   cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
   ```

2. **Wrong key file**
   ```go
   // Verify you're using the private key
   node.SetKey("id_rsa")  // not id_rsa.pub
   ```

3. **SSH agent not running**
   ```bash
   # Start ssh-agent
   eval $(ssh-agent -s)
   ssh-add ~/.ssh/your_key
   ```

### Permission Denied (publickey)

**Problem**: Server rejects key authentication.

**Solutions**:

1. **Enable key authentication on server**:
   ```bash
   # In /etc/ssh/sshd_config
   PubkeyAuthentication yes
   # Then restart SSH
   systemctl restart sshd
   ```

2. **Check authorized_keys permissions**:
   ```bash
   chmod 700 ~/.ssh
   chmod 600 ~/.ssh/authorized_keys
   ```

## Dry-Run Mode Issues

### Changes happening in dry-run mode

**Problem**: Commands are executing even with dry-run enabled.

**Solution**: This should not happen. Ensure you're using Ork's SSH functions:

```go
// Use ssh.Run which checks IsDryRunMode
output, err := ssh.Run(cfg, "command")

// NOT os/exec or other methods
```

### Dry-run mode not propagating

**Problem**: Dry-run set on inventory but nodes still executing.

**Solution**: Verify propagation is working:

```go
inv := ork.NewInventory()
inv.SetDryRunMode(true)

// Check individual nodes
for _, node := range inv.GetNodes() {
    if !node.GetDryRunMode() {
        log.Println("WARNING: Node doesn't have dry-run mode!")
    }
}
```

## Results and Error Handling

### Results map is empty

**Problem**: Running operations but results.Results is empty.

**Possible Causes**:

1. **Wrong host key in results**
   ```go
   // Results are keyed by host
   result := results.Results["server.example.com"]
   // NOT by IP or other identifier
   ```

2. **Node not added to group/inventory**
   ```go
   group := ork.NewGroup("webservers")
   group.AddNode(node)  // Don't forget this!
   ```

### Cannot access Result fields

**Problem**: Trying to access result but getting nil pointer.

**Solution**: Check if host exists in results:

```go
result, ok := results.Results["server.example.com"]
if !ok {
    log.Fatal("No result for host")
}
```

## Group and Inventory Issues

### Group operations not running on all nodes

**Problem**: Running playbook on group but only some nodes execute.

**Debug**:
```go
// Check nodes in group
nodes := group.GetNodes()
log.Printf("Group has %d nodes", len(nodes))

// Verify node configuration
for _, node := range nodes {
    log.Printf("Node: %s, Port: %s", node.GetHost(), node.GetPort())
}
```

### Inventory results missing some hosts

**Problem**: Inventory RunPlaybook missing results for some nodes.

**Possible Causes**:

1. **Max concurrency limiting**
   ```go
   // Increase if needed (default is 10)
   inv.SetMaxConcurrency(20)
   ```

2. **Duplicate hosts in different groups**
   - Results from the last processed group will overwrite earlier ones

## Performance Issues

### Operations are slow

**Problem**: Each command takes a long time to execute.

**Solutions**:

1. **Use persistent connections**
   ```go
   node.Connect()
   defer node.Close()
   // All subsequent commands reuse the connection
   ```

2. **Check network latency**
   ```bash
   ping -c 10 your-server.com
   ```

3. **Use inventory concurrency for multiple nodes**
   ```go
   inv.SetMaxConcurrency(50)  // Run up to 50 in parallel
   ```

### Memory usage high with many nodes

**Problem**: High memory usage when managing many servers.

**Solution**: Limit concurrent connections:

```go
inv.SetMaxConcurrency(10)  // Don't overwhelm your system
```

## Testing Issues

### Tests fail with "connection refused"

**Problem**: Integration tests can't connect to test containers.

**Solution**: Ensure Docker is running:

```bash
docker ps
# If not running, start Docker Desktop or docker service
```

### Mock not working in tests

**Problem**: SSH calls still hitting real servers in tests.

**Solution**: Ensure you're using the mockable variable:

```go
// In test
original := sshRunOnce
sshRunOnce = mockFn
defer func() { sshRunOnce = original }()
```

## Debugging Tips

### Enable Verbose Logging

```go
import (
    "log/slog"
    "os"
)

// Create verbose logger
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

node.SetLogger(logger)
```

### Inspect Configuration

```go
// Print full configuration
log.Printf("Node config: %+v", node.GetNodeConfig())

// Print args
log.Printf("Args: %v", node.GetArgs())
```

### Trace Execution

```go
// Step-by-step execution
log.Println("1. Creating node...")
node := ork.NewNodeForHost("server.com")

log.Println("2. Setting args...")
node.SetArg("key", "value")

log.Println("3. Running playbook...")
results := node.RunPlaybook(playbooks.NewPing())
log.Printf("4. Results: %+v", results)
```

## Common Error Messages

| Error | Meaning | Solution |
|-------|---------|----------|
| "failed to connect" | SSH connection failed | Check host, port, credentials |
| "playbook not found" | Unknown playbook ID | Use correct ID from constants |
| "username is required" | Missing required arg | Set arg with SetArg() |
| "host cannot be empty" | Node created without host | Use NewNodeForHost() |
| "[dry-run]" marker | Expected in dry-run mode | Normal behavior |

## Getting Help

If issues persist:

1. **Check documentation**
   - [Getting Started](getting_started.md)
   - [API Reference](api_reference.md)
   - [Architecture](architecture.md)

2. **Review examples**
   - See `*_test.go` files for usage examples
   - Check README.md in repository

3. **Enable debug logging**
   - Use `slog.LevelDebug` for verbose output

4. **Test in dry-run mode**
   - Verify operations before running for real

## See Also

- [Getting Started](getting_started.md) - Basic usage
- [Configuration](configuration.md) - Configuration options
- [Development](development.md) - Development guide
