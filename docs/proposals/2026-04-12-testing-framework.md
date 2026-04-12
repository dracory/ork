# Proposal: Testing Framework

**Date:** 2026-04-12  
**Status:** Draft  
**Author:** System Review

## Problem Statement

Currently, there are no visible test files in the project. This creates several risks:

- No automated verification of playbook behavior
- Difficult to refactor with confidence
- No regression testing
- Hard to validate SSH connection logic
- Manual testing is time-consuming and error-prone

## Proposed Solution

Implement a comprehensive testing framework with:

1. **Unit tests** for individual components
2. **Integration tests** with mock SSH servers
3. **End-to-end tests** with Docker containers
4. **Test helpers** for common patterns

## Testing Architecture

### 1. Mock SSH Client

```go
package sshtest

type MockClient struct {
    Commands []string
    Outputs  map[string]string
    Errors   map[string]error
    Connected bool
}

func NewMockClient() *MockClient {
    return &MockClient{
        Commands: []string{},
        Outputs:  make(map[string]string),
        Errors:   make(map[string]error),
    }
}

func (m *MockClient) Connect() error {
    m.Connected = true
    return nil
}

func (m *MockClient) Run(cmd string) (string, error) {
    m.Commands = append(m.Commands, cmd)
    
    if err, ok := m.Errors[cmd]; ok {
        return "", err
    }
    
    if output, ok := m.Outputs[cmd]; ok {
        return output, nil
    }
    
    return "", nil
}

func (m *MockClient) Close() error {
    m.Connected = false
    return nil
}

func (m *MockClient) ExpectCommand(cmd, output string) {
    m.Outputs[cmd] = output
}

func (m *MockClient) ExpectError(cmd string, err error) {
    m.Errors[cmd] = err
}

func (m *MockClient) AssertCommandRun(t *testing.T, cmd string) {
    for _, c := range m.Commands {
        if c == cmd {
            return
        }
    }
    t.Errorf("Expected command '%s' was not run", cmd)
}
```

### 2. Playbook Test Helpers

```go
package playbooktest

import (
    "testing"
    "github.com/dracory/ork/config"
    "github.com/dracory/ork/playbook"
)

type PlaybookTest struct {
    t          *testing.T
    mockClient *sshtest.MockClient
    config     config.Config
}

func New(t *testing.T) *PlaybookTest {
    return &PlaybookTest{
        t:          t,
        mockClient: sshtest.NewMockClient(),
        config: config.Config{
            SSHHost:  "test.example.com",
            SSHPort:  "22",
            RootUser: "root",
            SSHKey:   "test_key",
        },
    }
}

func (pt *PlaybookTest) ExpectCommand(cmd, output string) *PlaybookTest {
    pt.mockClient.ExpectCommand(cmd, output)
    return pt
}

func (pt *PlaybookTest) ExpectError(cmd string, err error) *PlaybookTest {
    pt.mockClient.ExpectError(cmd, err)
    return pt
}

func (pt *PlaybookTest) Run(pb playbook.Playbook) error {
    // Inject mock client
    return pb.Run(pt.config)
}

func (pt *PlaybookTest) AssertCommandRun(cmd string) {
    pt.mockClient.AssertCommandRun(pt.t, cmd)
}

func (pt *PlaybookTest) AssertNoError(err error) {
    if err != nil {
        pt.t.Fatalf("Expected no error, got: %v", err)
    }
}

func (pt *PlaybookTest) AssertError(err error) {
    if err == nil {
        pt.t.Fatal("Expected error, got nil")
    }
}
```

## Test Examples

### Unit Test: Ping Playbook

```go
package playbooks_test

import (
    "testing"
    "github.com/dracory/ork/playbooks"
    "github.com/dracory/ork/internal/playbooktest"
)

func TestPing_Success(t *testing.T) {
    test := playbooktest.New(t)
    
    // Setup expectations
    test.ExpectCommand("uptime", " 10:30:01 up 5 days, 2:15, 1 user, load average: 0.5, 0.3, 0.2")
    
    // Run playbook
    pb := playbooks.NewPing()
    err := test.Run(pb)
    
    // Assertions
    test.AssertNoError(err)
    test.AssertCommandRun("uptime")
}

func TestPing_ConnectionFailure(t *testing.T) {
    test := playbooktest.New(t)
    
    // Setup expectations
    test.ExpectError("uptime", fmt.Errorf("connection refused"))
    
    // Run playbook
    pb := playbooks.NewPing()
    err := test.Run(pb)
    
    // Assertions
    test.AssertError(err)
}
```

### Unit Test: AptUpgrade Playbook

```go
func TestAptUpgrade_WithUpdates(t *testing.T) {
    test := playbooktest.New(t)
    
    // Setup expectations
    test.ExpectCommand("apt-get upgrade -y", "Reading package lists...\nUpgraded 5 packages")
    
    // Run playbook
    pb := playbooks.NewAptUpgrade()
    err := test.Run(pb)
    
    // Assertions
    test.AssertNoError(err)
    test.AssertCommandRun("apt-get upgrade -y")
}

func TestAptUpgrade_NoUpdates(t *testing.T) {
    test := playbooktest.New(t)
    
    // Setup expectations
    test.ExpectCommand("apt-get upgrade -y", "0 upgraded, 0 newly installed")
    
    // Run playbook
    pb := playbooks.NewAptUpgrade()
    err := test.Run(pb)
    
    // Assertions
    test.AssertNoError(err)
}
```

### Unit Test: UserCreate Playbook

```go
func TestUserCreate_NewUser(t *testing.T) {
    test := playbooktest.New(t)
    test.config.Args = map[string]string{"username": "john"}
    
    // User doesn't exist
    test.ExpectError("id john", fmt.Errorf("no such user"))
    test.ExpectCommand("adduser --disabled-password --gecos '' john", "Adding user john...")
    test.ExpectCommand("usermod -aG sudo john", "")
    
    // Run playbook
    pb := playbooks.NewUserCreate()
    err := test.Run(pb)
    
    // Assertions
    test.AssertNoError(err)
    test.AssertCommandRun("adduser --disabled-password --gecos '' john")
    test.AssertCommandRun("usermod -aG sudo john")
}

func TestUserCreate_UserExists(t *testing.T) {
    test := playbooktest.New(t)
    test.config.Args = map[string]string{"username": "john"}
    
    // User already exists
    test.ExpectCommand("id john", "uid=1000(john) gid=1000(john) groups=1000(john)")
    
    // Run playbook
    pb := playbooks.NewUserCreate()
    err := test.Run(pb)
    
    // Should handle gracefully or skip
    test.AssertNoError(err)
}

func TestUserCreate_MissingUsername(t *testing.T) {
    test := playbooktest.New(t)
    // No username in args
    
    // Run playbook
    pb := playbooks.NewUserCreate()
    err := test.Run(pb)
    
    // Should error
    test.AssertError(err)
    if !strings.Contains(err.Error(), "username") {
        t.Errorf("Expected error about missing username, got: %v", err)
    }
}
```

### Unit Test: SwapCreate Playbook

```go
func TestSwapCreate_Success(t *testing.T) {
    test := playbooktest.New(t)
    test.config.Args = map[string]string{"size": "2"}
    
    // No existing swap
    test.ExpectCommand("swapon --show=NAME --noheadings", "")
    
    // Create swap commands
    test.ExpectCommand("fallocate -l 2G /swapfile && chmod 600 /swapfile && mkswap /swapfile && swapon /swapfile", "")
    test.ExpectCommand("grep -q '/swapfile' /etc/fstab && echo 'exists' || echo 'missing'", "missing")
    test.ExpectCommand("echo '/swapfile none swap sw 0 0' >> /etc/fstab", "")
    
    // Run playbook
    pb := playbooks.NewSwapCreate()
    err := test.Run(pb)
    
    // Assertions
    test.AssertNoError(err)
}

func TestSwapCreate_AlreadyExists(t *testing.T) {
    test := playbooktest.New(t)
    test.config.Args = map[string]string{"size": "2"}
    
    // Swap already exists
    test.ExpectCommand("swapon --show=NAME --noheadings", "/swapfile")
    
    // Run playbook
    pb := playbooks.NewSwapCreate()
    err := test.Run(pb)
    
    // Should error or skip
    test.AssertError(err)
}
```

### Integration Test: SSH Connection

```go
package ssh_test

import (
    "testing"
    "github.com/dracory/ork/ssh"
    "github.com/dracory/ork/internal/sshtest"
)

func TestSSHClient_ConnectAndRun(t *testing.T) {
    // Start mock SSH server
    server := sshtest.NewMockServer(t)
    defer server.Close()
    
    server.ExpectCommand("echo hello", "hello")
    
    // Create client
    client := ssh.NewClient(server.Host(), server.Port(), "testuser", server.KeyPath())
    
    // Connect
    err := client.Connect()
    if err != nil {
        t.Fatalf("Failed to connect: %v", err)
    }
    defer client.Close()
    
    // Run command
    output, err := client.Run("echo hello")
    if err != nil {
        t.Fatalf("Failed to run command: %v", err)
    }
    
    if output != "hello" {
        t.Errorf("Expected 'hello', got '%s'", output)
    }
}
```

### End-to-End Test with Docker

```go
package e2e_test

import (
    "testing"
    "github.com/dracory/ork/config"
    "github.com/dracory/ork/playbooks"
    "github.com/dracory/ork/internal/dockertest"
)

func TestAptUpgrade_E2E(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }
    
    // Start Ubuntu container with SSH
    container := dockertest.NewUbuntuContainer(t)
    defer container.Stop()
    
    cfg := config.Config{
        SSHHost:  container.Host(),
        SSHPort:  container.Port(),
        RootUser: "root",
        SSHKey:   container.KeyPath(),
    }
    
    // Run apt-update first
    aptUpdate := playbooks.NewAptUpdate()
    err := aptUpdate.Run(cfg)
    if err != nil {
        t.Fatalf("apt-update failed: %v", err)
    }
    
    // Run apt-upgrade
    aptUpgrade := playbooks.NewAptUpgrade()
    err = aptUpgrade.Run(cfg)
    if err != nil {
        t.Fatalf("apt-upgrade failed: %v", err)
    }
}
```

## Test Organization

```
ork/
├── ssh/
│   ├── ssh.go
│   └── ssh_test.go
├── playbook/
│   ├── playbook.go
│   └── playbook_test.go
├── playbooks/
│   ├── ping.go
│   ├── ping_test.go
│   ├── apt.go
│   ├── apt_test.go
│   ├── user.go
│   └── user_test.go
├── internal/
│   ├── sshtest/
│   │   ├── mock_client.go
│   │   └── mock_server.go
│   ├── playbooktest/
│   │   └── helpers.go
│   └── dockertest/
│       └── containers.go
└── e2e/
    ├── apt_test.go
    ├── user_test.go
    └── swap_test.go
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run unit tests only
go test -short ./...

# Run with coverage
go test -cover ./...

# Run with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestPing_Success ./playbooks

# Run with verbose output
go test -v ./...

# Run E2E tests
go test ./e2e/...
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run unit tests
        run: go test -short -cover ./...
      
      - name: Run integration tests
        run: go test -cover ./...
      
      - name: Run E2E tests
        run: go test ./e2e/...
```

## Implementation Plan

### Phase 1: Test Infrastructure
- Create mock SSH client
- Create playbook test helpers
- Set up test organization

### Phase 2: Unit Tests
- Write tests for all existing playbooks
- Achieve >80% code coverage
- Add table-driven tests

### Phase 3: Integration & E2E
- Create mock SSH server
- Set up Docker test containers
- Write end-to-end tests

### Phase 4: CI/CD
- Set up GitHub Actions
- Add coverage reporting
- Add test badges to README

## Benefits

- **Confidence**: Refactor without fear
- **Quality**: Catch bugs before production
- **Documentation**: Tests show how to use the code
- **Regression Prevention**: Ensure bugs don't come back
- **Faster Development**: Quick feedback loop

## Success Metrics

- >80% code coverage
- All playbooks have unit tests
- CI/CD runs tests automatically
- Zero flaky tests
- Tests run in <30 seconds

## Open Questions

1. Should we use testify or standard library?
2. How to handle tests that require real SSH servers?
3. Should we test against multiple OS versions?
4. How to test timeout and retry logic?
