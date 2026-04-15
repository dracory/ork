# Ork vs CloudFormation Comparison

## Quick Comparison

| Aspect | CloudFormation | Ork |
|--------|----------------|-----|
| **Language** | JSON/YAML | Go |
| **Purpose** | Infrastructure Provisioning | Configuration Management |
| **Layer** | Infrastructure (IaaS) | Software/Configuration |
| **Scope** | AWS only | Any SSH server |
| **Execution** | Declarative (template → stack) | Procedural (run now) |
| **State Management** | AWS CloudFormation (AWS-managed) | No state file |
| **Target** | AWS APIs | SSH to existing servers |
| **Idempotency** | Built-in (resource properties) | Playbook-level |
| **Lifecycle** | Create, update, delete stacks | Configure existing systems |
| **Type Safety** | Limited (JSON schema) | Yes (Go) |

## Core Difference

**CloudFormation = Building AWS infrastructure**  
**Ork = Configuring servers (AWS or elsewhere)**

```
┌─────────────────────────────────────────────────────────┐
│  CloudFormation Layer (AWS Infrastructure)               │
│  - Create EC2 instances                                 │
│  - Provision VPCs, subnets                              │
│  - Set up security groups, IAM roles                    │
│  - Create RDS databases                                │
│  - Configure load balancers (ALB/NLB)                   │
└─────────────────────────────────────────────────────────┘
                          ▼
              AWS resources now exist
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

## Architecture

### CloudFormation Architecture
```
┌──────────────┐
│ CloudFormation│
│ Template      │
│ (JSON/YAML)   │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ AWS Cloud     │
│ Formation     │
│ Service       │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ AWS APIs     │
│ (EC2, RDS,   │
│  S3, etc.)   │
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
│ (AWS or any) │
└──────────────┘
```

## CloudFormation Example (YAML)

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: Web server stack

Parameters:
  InstanceType:
    Type: String
    Default: t2.micro
    AllowedValues:
      - t2.micro
      - t2.small
      - t2.medium

Resources:
  # VPC
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      EnableDnsHostnames: true
      EnableDnsSupport: true
      Tags:
        - Key: Name
          Value: MyVPC

  # Public Subnet
  PublicSubnet:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.1.0/24
      MapPublicIpOnLaunch: true
      Tags:
        - Key: Name
          Value: PublicSubnet

  # Internet Gateway
  InternetGateway:
    Type: AWS::EC2::InternetGateway
    Properties:
      Tags:
        - Key: Name
          Value: MyIGW

  # Route Table
  RouteTable:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC

  # Route to Internet
  InternetRoute:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref RouteTable
      DestinationCidrBlock: 0.0.0.0/0
      GatewayId: !Ref InternetGateway

  # Security Group
  WebSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Enable HTTP and SSH access
      VpcId: !Ref VPC
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 22
          ToPort: 22
          CidrIp: 0.0.0.0/0
        - IpProtocol: tcp
          FromPort: 80
          ToPort: 80
          CidrIp: 0.0.0.0/0

  # EC2 Instance
  WebServer:
    Type: AWS::EC2::Instance
    Properties:
      ImageId: ami-0c55b159cbfafe1f0
      InstanceType: !Ref InstanceType
      SubnetId: !Ref PublicSubnet
      SecurityGroupIds:
        - !Ref WebSecurityGroup
      Tags:
        - Key: Name
          Value: WebServer
      UserData:
        Fn::Base64: |
          #!/bin/bash
          apt-get update
          apt-get install -y nginx

Outputs:
  InstanceId:
    Description: Instance ID
    Value: !Ref WebServer
  PublicIP:
    Description: Public IP address
    Value: !GetAtt WebServer.PublicIp
```

## Ork After CloudFormation

```go
package main

import (
    "fmt"
    "github.com/dracory/ork"
    "github.com/dracory/ork/playbooks"
)

func main() {
    // After CloudFormation creates the server, Ork configures it
    serverIP := "54.123.45.67" // From CloudFormation output
    
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

## Ork Inventory for Multiple AWS Servers

```go
// Configure multiple servers from CloudFormation stack
inv := ork.NewInventory()
inv.SetMaxConcurrency(5) // Configure parallel execution

// Add web servers
webGroup := ork.NewGroup("webservers")
for _, ip := range webServerIPs { // From CloudFormation outputs
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

### Phase 1: Infrastructure (CloudFormation)
```bash
# Create stack
aws cloudformation create-stack \
  --stack-name my-stack \
  --template-body file://template.yaml \
  --parameters ParameterKey=InstanceType,ParameterValue=t2.micro

# Wait for completion
aws cloudformation wait stack-create-complete --stack-name my-stack

# Get outputs
aws cloudformation describe-stacks \
  --stack-name my-stack \
  --query 'Stacks[0].Outputs'
```

### Phase 2: Configuration (Ork)
```go
// Use CloudFormation outputs to configure servers
serverIP := getCloudFormationOutput("PublicIP")

node := ork.NewNodeForHost(serverIP)
results := node.RunPlaybook(playbooks.NewUfwInstall())
```

### Phase 3: Integration
```bash
# Combined workflow
aws cloudformation create-stack ...  # Create infrastructure
./configure-servers                  # Run Ork to configure
```

## UserData vs Ork

### CloudFormation UserData
```yaml
UserData:
  Fn::Base64: |
    #!/bin/bash
    # Limitations:
    # - Runs only once at instance creation
    # - No error handling
    # - Basic shell scripts only
    # - Cannot re-run easily
    apt-get update
    apt-get install -y nginx
    systemctl start nginx
```

**Limitations:**
- Runs only at boot time
- No idempotency checks
- Difficult to debug failures
- Cannot update after creation
- Limited to basic scripts

### Ork Playbooks
```go
// Advantages:
// - Can run anytime
// - Built-in idempotency
// - Rich error handling
// - Check mode for preview
// - Reusable across servers

results := node.RunPlaybook(playbooks.NewNginxInstall())
if results.Results[host].Error != nil {
    log.Printf("Error: %v", results.Results[host].Error)
}

// Check what would change
checkResults := node.CheckPlaybook(playbooks.NewNginxInstall())
if checkResults.Results[host].Changed {
    fmt.Println("Nginx would be installed")
}
```

**Advantages:**
- Run anytime (not just at boot)
- Idempotent operations
- Rich error handling and logging
- Check mode for preview
- Reusable playbooks
- Can update configuration after creation

## State Management

### CloudFormation State
- Managed entirely by AWS
- Stored in AWS CloudFormation service
- Automatic drift detection
- Rollback capabilities
- Stack updates with change sets

```bash
# View stack status
aws cloudformation describe-stacks --stack-name my-stack

# Detect drift
aws cloudformation detect-stack-drift --stack-name my-stack

# Rollback to previous version
aws cloudformation rollback-stack --stack-name my-stack

# Delete stack (and all resources)
aws cloudformation delete-stack --stack-name my-stack
```

### Ork State
- No state file
- Idempotency via playbooks
- Check mode to preview changes
- Results returned after each run

```go
// Check what would change
results := node.CheckPlaybook(playbooks.NewUfwInstall())
if results.Results[host].Changed {
    fmt.Println("Changes would be made")
}

// Actually apply
results = node.RunPlaybook(playbooks.NewUfwInstall())
```

## Change Sets

### CloudFormation Change Sets
```bash
# Create change set to preview
aws cloudformation create-change-set \
  --stack-name my-stack \
  --change-set-name my-changes \
  --template-body file://updated-template.yaml

# Review changes
aws cloudformation describe-change-set \
  --stack-name my-stack \
  --change-set-name my-changes

# Execute change set
aws cloudformation execute-change-set \
  --stack-name my-stack \
  --change-set-name my-changes
```

### Ork Check Mode
```go
// Preview changes without applying
results := node.CheckPlaybook(playbooks.NewNginxInstall())
if results.Results[host].Changed {
    fmt.Printf("Would change: %s\n", results.Results[host].Message)
}

// Apply if desired
results = node.RunPlaybook(playbooks.NewNginxInstall())
```

## Cross-Stack References

### CloudFormation Cross-Stack References
```yaml
# Stack 1: VPC
Outputs:
  VpcId:
    Value: !Ref VPC
    Export:
      Name: !Sub "${AWS::StackName}-VpcId"

# Stack 2: Application (references Stack 1)
Resources:
  Subnet:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !ImportValue my-vpc-stack-VpcId
```

### Ork Inventory
```go
// Ork uses programmatic references
// No special cross-stack syntax needed

vpcStack := getCloudFormationStack("my-vpc-stack")
vpcID := vpcStack.Outputs["VpcId"]

// Use the value
node.SetArg("vpc_id", vpcID)
```

## When to Use Each

### Use CloudFormation when:
- Creating AWS infrastructure only
- Managing AWS-specific resources (VPC, RDS, S3, etc.)
- Need AWS-native integration
- Want AWS-managed state
- Require automatic rollback on failure
- Need change set previews
- Team uses AWS console heavily
- Want infrastructure as code for AWS
- Need AWS-specific features (IAM roles, policies)

### Use Ork when:
- Configuring existing servers (AWS or elsewhere)
- Installing and configuring software
- Managing users, SSH keys, security settings
- Deploying applications
- Running system updates
- Managing services
- Building Go applications
- Need type safety and compile-time checking
- Want programmatic control flow
- Cross-platform configuration (not just AWS)
- Embedding automation in Go applications

### Use Both Together
- CloudFormation provisions AWS infrastructure
- Ork configures the provisioned servers
- Common pattern: AWS IaC + Config Management

## Feature Comparison Table

| Feature | CloudFormation | Ork | Notes |
|---------|----------------|-----|-------|
| **Architecture** | ✅ IaC (AWS Infrastructure) | ✅ Config Management | CloudFormation creates AWS resources; Ork configures |
| **Scope** | ⚠️ AWS only | ✅ Any SSH server | CloudFormation is AWS-specific |
| **Execution Model** | ✅ Declarative (template → stack) | ✅ Procedural (run now) | CloudFormation declares state; Ork executes commands |
| **State Management** | ✅ AWS-managed (CloudFormation) | ❌ No state file | CloudFormation tracks state in AWS |
| **Creates VMs** | ✅ Yes (AWS EC2) | ❌ No | CloudFormation provisions AWS resources |
| **Configures Software** | ⚠️ Limited (UserData) | ✅ Yes | Ork has rich playbook library |
| **Manages Users** | ❌ No (via UserData only) | ✅ Yes | Ork manages system users |
| **Install Packages** | ⚠️ Limited (UserData) | ✅ Yes | Ork manages software packages |
| **Deploy Apps** | ❌ No | ✅ Yes | Ork deploys applications |
| **Run Commands** | ❌ No | ✅ Yes | Ork executes shell commands |
| **Parallel Execution** | ✅ Resource-level | ✅ Inventory-level (configurable) | Both support parallel operations |
| **Drift Detection** | ✅ Built-in (detect-stack-drift) | ⚠️ Manual (via Check) | CloudFormation auto-detects |
| **Rollback** | ✅ Automatic on failure | ⚠️ Manual | CloudFormation can auto-rollback |
| **Destroy Resources** | ✅ Yes (delete-stack) | ❌ No | CloudFormation manages lifecycle |
| **Change Sets** | ✅ Yes (preview changes) | ✅ Yes (Check mode) | Both can preview before applying |
| **Type Safety** | ⚠️ Limited (JSON schema) | ✅ Yes (Go) | Ork has compile-time checking |
| **IDE Support** | ⚠️ Basic (YAML/JSON) | ✅ Excellent (Go) | Ork has better tooling |
| **Preview Changes** | ✅ Yes (change sets) | ✅ Yes (Check mode) | Both can preview before applying |
| **Team Collaboration** | ✅ Yes (AWS-managed) | ⚠️ Manual | CloudFormation has built-in team features |
| **Import Existing** | ✅ Yes (import resources) | N/A | CloudFormation can import existing AWS resources |
| **Secrets Management** | ✅ AWS Secrets Manager/Parameter Store | ✅ envenc vault | Both handle secrets |
| **Testing** | ⚠️ Limited (StackSets) | ✅ Go testing + testcontainers | Ork has better testing support |
| **Policy as Code** | ✅ SCPs, GuardDuty | ❌ Manual | CloudFormation has AWS policy tools |
| **Learning Curve** | ⚠️ Medium (AWS concepts + YAML) | ✅ Low (Go knowledge) | CloudFormation requires AWS understanding |
| **Cross-Platform** | ❌ AWS only | ✅ Any SSH server | Ork works anywhere with SSH |
| **Language** | ✅ JSON/YAML | ✅ Go | CloudFormation uses declarative YAML/JSON |

## Concept Mapping

| CloudFormation Concept | Ork Equivalent |
|------------------------|----------------|
| Stack | Inventory (collection of nodes) |
| Resource | Node (server) |
| Template | Go code (programmatic) |
| Parameter | Node argument (SetArg) |
| Output | Results map |
| Change Set | CheckPlaybook() |
| Stack Update | RunPlaybook() |
| Stack Delete | No equivalent (Ork doesn't destroy) |
| Drift Detection | CheckPlaybook() |
| Rollback | Manual (re-run previous playbook) |
| Cross-Stack Reference | Programmatic variable passing |
| Resource Type | Playbook type |

## UserData vs Playbooks

| Aspect | CloudFormation UserData | Ork Playbooks |
|--------|------------------------|---------------|
| **Execution Time** | Once at boot | Anytime |
| **Idempotency** | Manual | Built-in |
| **Error Handling** | Limited | Rich |
| **Re-runnable** | No | Yes |
| **Debugging** | Difficult | Easy (logging) |
| **Complexity** | Basic scripts | Rich operations |
| **Updates** | Requires instance recreation | Can update running servers |
| **Preview** | No | Check mode |

## Summary

**CloudFormation Strengths:**
- AWS-native integration
- AWS-managed state
- Automatic rollback
- Change set previews
- IAM integration
- AWS-specific features
- Free (AWS service)
- AWS console integration

**Ork Strengths:**
- Go-native configuration management
- Rich playbook library
- Type safety and compile-time checking
- Simple, no state file
- Configures existing infrastructure
- Cross-platform (not just AWS)
- Embeddable in Go applications
- Rich error handling
- Check mode for preview

**Key Difference:**
- CloudFormation: AWS infrastructure provisioning with AWS-managed state
- Ork: Configuration management without state (works on any SSH server)

**Best Practice:**
- Use CloudFormation to create AWS infrastructure
- Use Ork to configure the infrastructure (even non-AWS servers)
- Combine for complete AWS infrastructure + configuration automation
