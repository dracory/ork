# Ork vs Terraform Comparison

## Quick Comparison

| Aspect | Terraform | Ork |
|--------|-----------|-----|
| **Language** | HCL (HashiCorp Configuration Language) | Go |
| **Purpose** | Infrastructure Provisioning | Configuration Management |
| **Layer** | Infrastructure (IaaS) | Software/Configuration |
| **State Management** | State file (local or remote) | No state file |
| **Target** | Cloud APIs, hypervisors | SSH to existing servers |
| **Execution** | Declarative (plan → apply) | Procedural (run now) |
| **Idempotency** | Built into providers | Playbook-level |
| **Lifecycle** | Create, update, destroy resources | Configure existing systems |

## Core Difference

**Terraform = Building the house**  
**Ork = Furnishing and maintaining the house**

```
┌─────────────────────────────────────────────────────────┐
│  Terraform Layer (Infrastructure)                       │
│  - Create EC2 instances                                 │
│  - Provision VPCs, subnets                              │
│  - Set up load balancers                                │
│  - Create databases                                     │
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
└─────────────────────────────────────────────────────────┘
```

## What Each Tool Does

### Terraform
**Infrastructure as Code** - Creates and manages infrastructure resources.

```hcl
# main.tf - Create AWS infrastructure
provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "web" {
  count         = 3
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  tags = {
    Name = "web-server-${count.index}"
  }
}

resource "aws_security_group" "web" {
  name = "web-sg"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

**Terraform workflow:**
```bash
terraform init      # Download providers
terraform plan      # Preview changes
terraform apply     # Create/update infrastructure
terraform destroy   # Remove infrastructure
```

### Ork
**Configuration Management** - Configures existing servers.

```go
// main.go - Configure the servers Terraform created
node := ork.NewNodeForHost("web-server-1.example.com").
    SetUser("ubuntu").
    SetKey("mykey.pem")

// Install and configure software
node.RunPlaybook(playbooks.NewAptUpdate())
node.RunPlaybook(playbooks.NewAptUpgrade())

// Custom playbook for application deployment
result := node.RunPlaybook(myapp.NewDeploy())
```

## Execution Model

### Terraform (Declarative with State)
```
Desired State (HCL)
       │
       ▼
┌──────────────┐
│  Terraform   │──► Compare with state file
│   Plan       │
└──────────────┘
       │
       ▼
Actual Changes Needed
       │
       ▼
┌──────────────┐
│  Terraform   │──► Execute via Cloud APIs
│   Apply      │
└──────────────┘
       │
       ▼
Update State File
```

**Key characteristics:**
- Maintains state file tracking all resources
- Compares desired vs actual state
- Only makes necessary changes
- Can destroy what it created

### Ork (Procedural)
```
Your Go Program
       │
       ▼
┌──────────────┐
│   SSH to     │
│   Server     │
└──────────────┘
       │
       ▼
Execute Commands/Playbooks
       │
       ▼
Get Results
```

**Key characteristics:**
- No state file
- Executes what you tell it, when you tell it
- Cannot "unprovision" (no destroy)
- One-way configuration

## Use Case Examples

### Scenario: Web Application Stack

**Terraform handles:**
```hcl
# Create the infrastructure
resource "aws_vpc" "main" { cidr_block = "10.0.0.0/16" }

resource "aws_instance" "web" {
  count = 2
  ami   = "ami-12345678"
  # ... creates EC2 instances
}

resource "aws_rds_instance" "db" {
  # ... creates managed database
}

resource "aws_elb" "web" {
  # ... creates load balancer
}
```

**Ork handles:**
```go
// After Terraform creates servers, Ork configures them
webNodes := []string{"web1.internal", "web2.internal"}

for _, host := range webNodes {
    node := ork.NewNodeForHost(host).
        SetUser("ubuntu").
        SetKey("deploy.pem")

    // Configure each web server
    node.RunPlaybook(playbooks.NewUserCreate()).
        SetArg("username", "deployer")

    node.RunPlaybook(playbooks.NewUfwInstall())
    node.RunPlaybook(playbooks.NewFail2banInstall())

    // Install and start nginx
    node.RunCommand("sudo apt-get install -y nginx")
    node.RunCommand("sudo systemctl enable nginx")
}
```

## Idempotency Comparison

### Terraform Idempotency
```hcl
# Running 'terraform apply' twice:
# First run: Creates 2 EC2 instances
# Second run: No changes (instances already exist)

resource "aws_instance" "web" {
  count = 2
  # ...
}
```

Terraform tracks resources in state file. Second apply = no-op.

### Ork Idempotency
```go
// Running twice:
// First run: Installs nginx, creates user
// Second run: Same commands, nginx already installed

// Playbooks handle idempotency internally
result := node.RunPlaybook(playbooks.NewAptInstall())
if result.Changed {
    log.Println("Nginx was installed")
} else {
    log.Println("Nginx already installed")
}
```

Ork playbooks check state before acting (e.g., "is nginx installed?").

## Complementary Usage

Terraform and Ork work well together:

```hcl
# terraform/main.tf - Create infrastructure
resource "aws_instance" "web" {
  count = 3
  # ...

  # Output the IPs for Ork to use
}

output "web_ips" {
  value = aws_instance.web.*.private_ip
}
```

```go
// main.go - Configure the created servers
// Read IPs from Terraform output
ips := getTerraformOutput("web_ips")

for _, ip := range ips {
    node := ork.NewNodeForHost(ip)
    node.RunPlaybook(playbooks.NewPing())
}
```

**Typical workflow:**
1. `terraform apply` - Create infrastructure
2. Wait for instances to be ready
3. Ork configures the new servers
4. Application is deployed and running

## Feature Comparison

| Feature | Terraform | Ork |
|---------|-----------|-----|
| **Create VMs** | Yes (AWS, Azure, GCP, etc.) | No |
| **Configure Software** | Limited (user_data) | Yes |
| **Manage Users** | No | Yes |
| **Install Packages** | No | Yes |
| **Deploy Apps** | No | Yes |
| **Run Commands** | No | Yes |
| **Parallel Execution** | Resource-level | Node-level (planned) |
| **Rolling Updates** | Via resource lifecycle | Manual |
| **State Tracking** | Yes (state file) | No |
| **Drift Detection** | Yes | Manual (via Check) |
| **Destroy Resources** | Yes | No |

## When to Use Each

### Use Terraform when:
- Creating cloud infrastructure (AWS, Azure, GCP)
- Managing network topology (VPCs, subnets, firewalls)
- Provisioning VMs, databases, load balancers
- You need infrastructure lifecycle (create/update/destroy)
- Multi-cloud or hybrid infrastructure

### Use Ork when:
- Configuring existing servers
- Installing and configuring software
- Managing users, SSH keys, security settings
- Running maintenance tasks (updates, backups)
- Application deployment
- You have Go expertise and want type safety

### Use Both when:
- Full stack automation needed
- Terraform creates infrastructure
- Ork configures the created servers
- Terraform manages resource lifecycle
- Ork manages configuration drift

## Summary

**Terraform:**
- "I need 3 servers in AWS with a load balancer"
- Creates/manages infrastructure
- Has state, can destroy

**Ork:**
- "Install nginx and deploy my app on these 3 servers"
- Configures existing infrastructure
- No state, runs on demand

**Together:**
```bash
# 1. Create infrastructure
terraform apply

# 2. Configure it
go run configure.go

# Full stack deployed!
```
