# Ork vs Pulumi Comparison

## Quick Comparison

| Aspect | Pulumi | Ork |
|--------|--------|-----|
| **Language** | TypeScript, Python, Go, C#, Java | Go |
| **Purpose** | Infrastructure Provisioning | Configuration Management |
| **Layer** | Infrastructure (IaaS) | Software/Configuration |
| **State Management** | State file (local or remote backends) | No state file |
| **Target** | Cloud APIs (AWS, Azure, GCP, etc.) | SSH to existing servers |
| **Execution** | Declarative (plan → apply → destroy) | Procedural (run now) |
| **Idempotency** | Built into providers | Playbook-level |
| **Lifecycle** | Create, update, destroy resources | Configure existing systems |
| **Type Safety** | Yes (your chosen language) | Yes (Go) |
| **IDE Support** | Excellent (language servers) | Excellent (Go) |

## Core Difference

**Pulumi = Building the house**  
**Ork = Furnishing and maintaining the house**

```
┌─────────────────────────────────────────────────────────┐
│  Pulumi Layer (Infrastructure)                           │
│  - Create EC2 instances                                 │
│  - Provision VPCs, subnets                              │
│  - Set up load balancers                                │
│  - Create databases                                     │
│  - Configure cloud resources                            │
└─────────────────────────────────────────────────────────┘
                          ▼
              Servers now exist
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Ork Layer (Configuration)                              │
│  - Install nginx, mysql                                 │
│  - Configure users, SSH keys                            │
│  - Deploy applications                                  │
│  - Run security updates                                 │
│  - Manage system services                               │
└─────────────────────────────────────────────────────────┘
```

## What Each Tool Does

### Pulumi (Infrastructure Provisioning)
- Creates cloud resources (VMs, databases, networks)
- Manages infrastructure lifecycle
- Uses real programming languages (not YAML)
- Tracks infrastructure state
- Supports multi-cloud deployments
- Provides drift detection
- Enables preview before changes

### Ork (Configuration Management)
- Configures existing servers
- Installs and manages software
- Manages users and permissions
- Deploys applications
- Runs security updates
- Manages system services
- Executes commands across fleets

## Architecture

### Pulumi Architecture
```
┌──────────────┐
│ Pulumi CLI    │
│ or SDK        │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ State File   │◄─── Local file or remote backend
│ (JSON)       │     (S3, Azure Blob, GCS, etc.)
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Cloud APIs   │
│ (AWS, Azure, │
│  GCP, etc.)  │
└──────────────┘
```

### Ork Architecture
```
┌──────────────┐
│ Your Go      │
│ Application  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Ork Library  │
│ (Nodes,      │
│  Playbooks,  │
│  Inventory)  │
└──────┬───────┘
       │ SSH
       ▼
┌──────────────┐
│ Target       │
│ Servers      │
└──────────────┘
```

## Pulumi Example (TypeScript)

```typescript
import * as aws from "@pulumi/aws";

// Create VPC
const vpc = new aws.ec2.Vpc("my-vpc", {
    cidrBlock: "10.0.0.0/16",
    enableDnsHostnames: true,
    enableDnsSupport: true,
});

// Create subnet
const subnet = new aws.ec2.Subnet("my-subnet", {
    vpcId: vpc.id,
    cidrBlock: "10.0.1.0/24",
    mapPublicIpOnLaunch: true,
});

// Create security group
const sg = new aws.ec2.SecurityGroup("web-sg", {
    vpcId: vpc.id,
    ingress: [
        { protocol: "tcp", fromPort: 22, toPort: 22, cidrBlocks: ["0.0.0.0/0"] },
        { protocol: "tcp", fromPort: 80, toPort: 80, cidrBlocks: ["0.0.0.0/0"] },
    ],
});

// Create EC2 instance
const server = new aws.ec2.Instance("web-server", {
    ami: "ami-0c55b159cbfafe1f0",
    instanceType: "t2.micro",
    subnetId: subnet.id,
    vpcSecurityGroupIds: [sg.id],
    tags: {
        Name: "web-server",
        Environment: "production",
    },
});

// Export outputs
export const serverIp = server.publicIp;
export const serverId = server.id;
```

## Ork After Pulumi

```go
package main

import (
    "fmt"
    "github.com/dracory/ork"
    "github.com/dracory/ork/playbooks"
)

func main() {
    // After Pulumi creates the server, Ork configures it
    serverIP := "54.123.45.67" // From Pulumi output
    
    // Create node
    node := ork.NewNodeForHost(serverIP).
        SetUser("ubuntu").
        SetKey("/path/to/deploy.pem")
    
    // Configure the server with playbooks
    results := node.
        RunPlaybook(playbooks.NewAptUpdate()).
        RunPlaybook(playbooks.NewAptUpgrade()).
        RunPlaybook(playbooks.NewUfwInstall()).
        RunPlaybook(playbooks.NewNginxInstall())
    
    // Check results
    result := results.Results[serverIP]
    if result.Error != nil {
        fmt.Printf("Error: %v\n", result.Error)
    } else {
        fmt.Printf("Success: %s\n", result.Message)
    }
}
```

## Ork Inventory for Multiple Servers

```go
// Configure multiple servers from Pulumi
inv := ork.NewInventory()
inv.SetMaxConcurrency(5) // Configure parallel execution

// Add web servers
webGroup := ork.NewGroup("webservers")
for _, ip := range webServerIPs { // From Pulumi outputs
    webGroup.AddNode(ork.NewNodeForHost(ip).
        SetUser("ubuntu").
        SetKey("/path/to/deploy.pem"))
}
inv.AddGroup(webGroup)

// Configure all web servers in parallel
results := inv.RunPlaybook(playbooks.NewNginxInstall())
summary := results.Summary()
fmt.Printf("Changed: %d, Failed: %d\n", summary.Changed, summary.Failed)
```

## Typical Workflow

### Phase 1: Infrastructure (Pulumi)
```bash
# Preview changes
pulumi preview

# Apply infrastructure changes
pulumi up

# Export server IPs
pulumi stack output serverIp
```

### Phase 2: Configuration (Ork)
```go
// Use Pulumi outputs to configure servers
serverIP := pulumiStack.Output("serverIp")

node := ork.NewNodeForHost(serverIP)
results := node.RunPlaybook(playbooks.NewUfwInstall())
```

### Phase 3: Integration
```bash
# Combined workflow
pulumi up          # Create infrastructure
./configure-servers # Run Ork to configure
```

## State Management

### Pulumi State
- Tracks all managed resources
- Stores in local file or remote backend
- Enables drift detection
- Supports import of existing resources
- Provides state locking for teams

```bash
# View state
pulumi stack export

# Import existing resource
pulumi import aws:ec2/instance:Instance web-server i-1234567890abcdef0

# Remove from state (keep resource)
pulumi state rm aws:ec2/instance:Instance web-server
```

### Ork State
- No state file
- Idempotency via playbooks
- Check mode to preview changes
- Results returned after each run
- Manual tracking if needed

```go
// Check what would change
results := node.CheckPlaybook(playbooks.NewUfwInstall())
if results.Results[host].Changed {
    fmt.Println("Changes would be made")
}

// Actually apply
results = node.RunPlaybook(playbooks.NewUfwInstall())
```

## Drift Detection

### Pulumi Drift Detection
```bash
# Check for drift (resources changed outside Pulumi)
pulumi refresh

# View what changed
pulumi diff
```

### Ork Drift Detection
```go
// Check if configuration matches desired state
results := node.CheckPlaybook(playbooks.NewUfwInstall())
if results.Results[host].Changed {
    fmt.Println("Configuration has drifted")
    // Re-apply playbook
    node.RunPlaybook(playbooks.NewUfwInstall())
}
```

## Multi-Language Support

### Pulumi Languages
- TypeScript/JavaScript
- Python
- Go
- C# (.NET)
- Java
- YAML (for simple cases)

### Ork Languages
- Go (only)

**Note:** Pulumi's multi-language support is for infrastructure definition. Ork is Go-only for configuration automation.

## When to Use Each

### Use Pulumi when:
- Creating cloud infrastructure (AWS, Azure, GCP)
- Managing network topology (VPCs, subnets, firewalls)
- Provisioning VMs, databases, load balancers
- You need infrastructure lifecycle (create/update/destroy)
- Multi-cloud or hybrid infrastructure
- You prefer real programming languages over YAML
- Team has TypeScript/Python/Go/C# expertise
- Need state tracking and drift detection

### Use Ork when:
- Configuring existing servers
- Installing and configuring software
- Managing users, SSH keys, security settings
- Deploying applications
- Running system updates
- Managing services
- Building Go applications
- Need type safety and compile-time checking
- Want programmatic control flow
- Embedding automation in Go applications

### Use Both Together
- Pulumi provisions infrastructure
- Ork configures the provisioned servers
- Common pattern: IaC + Config Management

## Feature Comparison Table

| Feature | Pulumi | Ork | Notes |
|---------|--------|-----|-------|
| **Architecture** | ✅ IaC (Infrastructure) | ✅ Config Management | Pulumi creates; Ork configures |
| **Execution Model** | ✅ Declarative (plan → apply) | ✅ Procedural (run now) | Pulumi declares state; Ork executes commands |
| **State Management** | ✅ State file (remote backends) | ❌ No state file | Pulumi tracks infrastructure state |
| **Creates VMs** | ✅ Yes (cloud APIs) | ❌ No | Pulumi provisions resources |
| **Configures Software** | ⚠️ Limited (user_data/cloud-init) | ✅ Yes | Ork has rich playbook library |
| **Manages Users** | ❌ No | ✅ Yes | Ork manages system users |
| **Install Packages** | ❌ No | ✅ Yes | Ork manages software packages |
| **Deploy Apps** | ❌ No | ✅ Yes | Ork deploys applications |
| **Run Commands** | ❌ No | ✅ Yes | Ork executes shell commands |
| **Parallel Execution** | ✅ Resource-level | ✅ Inventory-level (configurable) | Both support parallel operations |
| **Drift Detection** | ✅ Built-in (refresh) | ⚠️ Manual (via Check) | Pulumi auto-detects; Ork manual |
| **Rollback** | ✅ Automatic (state rollback) | ⚠️ Manual | Pulumi can undo changes |
| **Destroy Resources** | ✅ Yes (destroy command) | ❌ No | Pulumi manages lifecycle |
| **Multi-Language** | ✅ TS, Python, Go, C#, Java | ✅ Go (only) | Pulumi supports many languages |
| **Type Safety** | ✅ Yes (your language) | ✅ Yes (Go) | Both have compile-time checking |
| **IDE Support** | ✅ Excellent (language servers) | ✅ Excellent (Go) | Both have good tooling |
| **Preview Changes** | ✅ Yes (preview/diff) | ✅ Yes (Check mode) | Both can preview before applying |
| **Team Collaboration** | ✅ Yes (state backends, locking) | ⚠️ Manual | Pulumi has built-in team features |
| **Import Existing** | ✅ Yes (import command) | N/A | Pulumi can import existing resources |
| **Secrets Management** | ✅ Pulumi Secrets | ✅ envenc vault | Both handle secrets |
| **Testing** | ✅ Pulumi Test (unit/integration) | ✅ Go testing + testcontainers | Both support testing |
| **Policy as Code** | ✅ CrossGuard | ❌ Manual | Pulumi has policy framework |
| **Learning Curve** | ⚠️ Medium (IaC concepts) | ✅ Low (Go knowledge) | Pulumi requires IaC understanding |

## Concept Mapping

| Pulumi Concept | Ork Equivalent |
|----------------|----------------|
| Resource | Node (server) |
| Stack | Inventory (collection of nodes) |
| Provider | SSH connection |
| State file | No equivalent (idempotent playbooks) |
| `pulumi up` | `RunPlaybook()` |
| `pulumi preview` | `CheckPlaybook()` |
| `pulumi destroy` | No equivalent (Ork doesn't destroy) |
| `pulumi refresh` | `CheckPlaybook()` (drift detection) |
| `pulumi import` | No equivalent |
| Component (multiple resources) | Group (multiple nodes) |
| Output | Results map |

## Summary

**Pulumi Strengths:**
- Real programming languages for infrastructure
- Multi-cloud support
- State management and drift detection
- Team collaboration features
- Rich ecosystem of providers
- Preview before applying changes

**Ork Strengths:**
- Go-native configuration management
- Rich playbook library
- Type safety and compile-time checking
- Simple, no state file
- Configures existing infrastructure
- Embeddable in Go applications

**Key Difference:**
- Pulumi: Infrastructure provisioning with state tracking
- Ork: Configuration management without state

**Best Practice:**
- Use Pulumi to create infrastructure
- Use Ork to configure the infrastructure
- Combine for complete infrastructure + configuration automation
