# Ork vs Infrastructure-as-Code Tools Comparison

This document compares Ork with modern Infrastructure-as-Code (IaC) tools.

**Key Distinction:**
- **IaC Tools** (Pulumi, CloudFormation, etc.) = Create infrastructure (servers, networks, databases)
- **Ork** = Configure existing infrastructure (install software, manage users)

---

## Quick Comparison Table

| Tool | Language | Provider | Model | Scope |
|------|----------|----------|-------|-------|
| **Ork** | Go | Any (SSH) | Procedural | Configuration |
| **Pulumi** | TypeScript/Python/Go/C# | Multi-cloud | Declarative | Infrastructure |
| **CloudFormation** | JSON/YAML | AWS only | Declarative | Infrastructure |
| **Deployment Manager** | YAML/Python | GCP only | Declarative | Infrastructure |
| **Crossplane** | YAML (Kubernetes) | Multi-cloud | Declarative | Infrastructure |

---

## Pulumi

### Overview
Modern Infrastructure-as-Code using real programming languages instead of YAML/JSON.

### Pulumi Example (TypeScript)
```typescript
import * as aws from "@pulumi/aws";

// Create EC2 instance
const server = new aws.ec2.Instance("web-server", {
    ami: "ami-0c55b159cbfafe1f0",
    instanceType: "t2.micro",
    tags: {
        Name: "web-server",
    },
});

// Export the IP
export const publicIp = server.publicIp;
```

### Ork After Pulumi
```go
// After Pulumi creates the server, Ork configures it via Inventory
inv := ork.NewInventory()
webGroup := ork.NewGroup("webservers")
webGroup.AddNode(ork.NewNodeForHost(server.PublicIp).
    SetUser("ubuntu").
    SetKey("deploy.pem"))
inv.AddGroup(webGroup)

// Configure the server with skills
results := inv.Run(skills.NewUfwInstall())
summary := results.Summary()
```

### Pulumi vs Ork

| Aspect | Pulumi | Ork |
|--------|--------|-----|
| **Creates VMs** | ✅ Yes | ❌ No |
| **Configures Software** | ⚠️ Limited (user_data) | ✅ Yes |
| **Language** | ✅ TS/Python/Go/C# | ✅ Go |
| **State** | ✅ State file (backend) | ❌ None |
| **Type Safety** | ✅ Yes (your language) | ✅ Yes (Go) |
| **IDE Support** | ✅ Excellent | ✅ Excellent |

**Use together:** Pulumi provisions, Ork configures.

---

## AWS CloudFormation

### Overview
AWS-native Infrastructure-as-Code. JSON/YAML templates describing AWS resources.

### CloudFormation Example (YAML)
```yaml
AWSTemplateFormatVersion: '2010-09-09'
Resources:
  WebServer:
    Type: AWS::EC2::Instance
    Properties:
      ImageId: ami-0c55b159cbfafe1f0
      InstanceType: t2.micro
      KeyName: my-key
      SecurityGroups:
        - !Ref WebSecurityGroup
      UserData:
        Fn::Base64: |
          #!/bin/bash
          apt-get update
          apt-get install -y nginx

  WebSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Enable HTTP
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 80
          ToPort: 80
          CidrIp: 0.0.0.0/0
```

### CloudFormation vs Ork

| Aspect | CloudFormation | Ork |
|--------|----------------|-----|
| **Scope** | ⚠️ AWS only | ✅ Any SSH server |
| **Language** | ✅ JSON/YAML | ✅ Go |
| **UserData** | ⚠️ Limited scripting | ✅ Full playbook library |
| **Drift Detection** | ✅ Built-in | ⚠️ Manual (via Check) |
| **Rollback** | ✅ Automatic | ⚠️ Manual |

**CloudFormation UserData limitations:**
- Run once at boot
- No error handling
- Basic shell scripts only

**Ork advantages:**
- Rich playbook ecosystem
- Error handling
- Can re-run anytime
- Cross-platform (not AWS-specific)

---

## Google Cloud Deployment Manager

### Overview
GCP-native Infrastructure-as-Code. Similar to CloudFormation but for GCP.

### Deployment Manager Example (YAML)
```yaml
resources:
- name: vm-instance
  type: compute.v1.instance
  properties:
    zone: us-central1-a
    machineType: zones/us-central1-a/machineTypes/f1-micro
    disks:
    - deviceName: boot
      type: PERSISTENT
      boot: true
      autoDelete: true
      initializeParams:
        sourceImage: projects/debian-cloud/global/images/family/debian-9
    networkInterfaces:
    - network: global/networks/default
      accessConfigs:
      - name: External NAT
        type: ONE_TO_ONE_NAT
```

### Deployment Manager vs Ork

| Aspect | Deployment Manager | Ork |
|--------|-------------------|-----|
| **Scope** | ⚠️ GCP only | ✅ Any SSH server |
| **Templates** | ✅ Jinja2/Python | ✅ Go code |
| **Configuration** | ⚠️ Limited (startup scripts) | ✅ Full playbooks |
| **Multi-cloud** | ❌ No | ✅ Yes |

---

## Crossplane

### Overview
Kubernetes-native Infrastructure-as-Code. Uses Kubernetes controllers to provision cloud resources.

### Crossplane Example (YAML)
```yaml
apiVersion: compute.aws.crossplane.io/v1beta1
kind: Instance
metadata:
  name: my-ec2-instance
spec:
  forProvider:
    region: us-east-1
    imageId: ami-0c55b159cbfafe1f0
    instanceType: t2.micro
  providerConfigRef:
    name: aws-provider
```

### Crossplane Architecture
```
┌─────────────────┐
│   Kubernetes    │
│   Cluster       │
│                 │
│ ┌─────────────┐ │
│ │ Crossplane  │ │◄─── YAML resources
│ │ Controller  │ │      (like above)
│ └──────┬──────┘ │
└────────┼─────────┘
         │
    Cloud APIs
         │
   ┌─────┴──────┐
   │ AWS/GCP/   │
   │ Azure      │
   └────────────┘
```

### Crossplane vs Ork

| Aspect | Crossplane | Ork |
|--------|------------|-----|
| **Platform** | ✅ Kubernetes | ✅ Any Go runtime |
| **Interface** | ✅ YAML/Kubectl | ✅ Go API |
| **Creates Resources** | ✅ Yes | ❌ No |
| **Configures VMs** | ❌ No | ✅ Yes |
| **GitOps** | ✅ Native (ArgoCD/Flux) | ⚠️ User implements |

---

## Common Pattern: IaC + Ork

All these IaC tools work well with Ork in a two-phase workflow:

### Phase 1: Infrastructure (IaC Tool)
```yaml
# CloudFormation, Pulumi, etc. create infrastructure
Resources:
  WebServer:
    Type: AWS::EC2::Instance
    Properties:
      InstanceType: t2.micro
      # ... creates the server

Outputs:
  ServerIP:
    Value: !GetAtt WebServer.PublicIp
    Description: IP of created server
```

### Phase 2: Configuration (Ork)
```go
// Read IPs from IaC output
ips := getCloudFormationOutput("ServerIPs")

// Load secrets from vault (if needed)
vaultKeys, err := ork.VaultFileToKeysWithPrompt("vault.envenc")
if err != nil {
    log.Fatal(err)
}

// Configure with Ork via Inventory
inv := ork.NewInventory()
webGroup := ork.NewGroup("webservers")
for _, ip := range ips {
    node := ork.NewNodeForHost(ip).
        SetUser("ubuntu").
        SetKey("deploy.pem")
    // Set vault secrets as node args
    for key, value := range vaultKeys {
        node.SetArg(key, value)
    }
    webGroup.AddNode(node)
}
inv.AddGroup(webGroup)

// Security hardening across all servers
results := inv.Run(skills.NewUfwInstall())
results = inv.Run(skills.NewFail2banInstall())

// Install stack
results = inv.Run(skills.NewAptUpdate())

// Deploy application with per-node results
deploy := myapp.NewDeploy()
deploy.SetArg("version", "1.2.3")
results = inv.Run(deploy)
for host, result := range results.Results {
    if result.Error != nil {
        log.Printf("%s deploy failed: %v", host, result.Error)
    }
}
```

---

## When to Use Each

### Use IaC Tools (Pulumi/CloudFormation/etc.) when:
- Creating cloud infrastructure
- Managing VPCs, subnets, security groups
- Provisioning managed services (RDS, S3, etc.)
- Need infrastructure lifecycle (create/update/destroy)
- Multi-cloud or cloud-native approach

### Use Ork when:
- Configuring existing servers
- Installing and configuring software
- Managing users, SSH keys, firewalls
- Running maintenance tasks
- Application deployment
- Cross-platform (works on any SSH server)
- Secure secrets management (vault integration)
- Interactive configuration (prompt functions)

### Use Both when:
```
┌─────────────────────────────────────┐
│  1. Pulumi/CloudFormation           │
│     - Create EC2 instance           │
│     - Setup networking              │
│     - Output IP address             │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  2. Ork                             │
│     - SSH to new server             │
│     - Install security tools          │
│     - Configure users                 │
│     - Deploy application              │
└─────────────────────────────────────┘
```

---

## Summary

| Need | Tool Category | Examples |
|------|---------------|----------|
| **Create VMs** | IaC | Pulumi, CloudFormation, Crossplane |
| **Configure VMs** | Config Management | Ork, Ansible, Chef |
| **Full Stack** | Both | IaC + Ork |

**Remember:**
- IaC tools are for **infrastructure provisioning**
- Ork is for **configuration management**
- They are **complementary**, not competitive
