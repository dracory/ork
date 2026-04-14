# Testing Framework

This directory contains internal testing utilities for the Ork project.

## Test Coverage

All internal packages have comprehensive test coverage:

- `internal/sshtest/mock_client_test.go` - Tests for MockClient functionality
- `internal/playbooktest/helpers_test.go` - Tests for PlaybookTest helpers

Run tests with:
```bash
go test ./internal/...
```

## Packages

### sshtest

Mock SSH client for testing without real SSH connections.

**Usage:**

```go
import "github.com/dracory/ork/internal/sshtest"

mock := sshtest.NewMockClient()
mock.ExpectCommand("uptime", "up 5 days")
mock.ExpectError("id john", fmt.Errorf("no such user"))

mock.Connect()
output, err := mock.Run("uptime")
mock.AssertCommandRun("uptime")
```

**Methods:**

- `NewMockClient()` - Creates a new mock client
- `ExpectCommand(cmd, output)` - Sets expected output for a command
- `ExpectError(cmd, error)` - Sets expected error for a command
- `Connect()` - Simulates SSH connection
- `Run(cmd)` - Executes a command and returns predefined output/error
- `Close()` - Simulates closing the connection
- `AssertCommandRun(cmd)` - Verifies a command was executed
- `GetCommands()` - Returns all executed commands
- `Reset()` - Clears all recorded commands and expectations

### playbooktest

Test helpers for playbook testing with mock SSH.

**Usage:**

```go
import "github.com/dracory/ork/internal/playbooktest"

test := playbooktest.New(t)
defer test.Cleanup()

test.Setup()
test.ExpectCommand("apt-get update -qq", "")
test.ExpectCommand("apt list --upgradable", "nginx/stable 1.18.0")

pb := playbooks.NewAptStatus()
pb.SetNodeConfig(test.Config())
result := pb.Run()

test.AssertResultNoError(result)
test.AssertCommandRun("apt-get update -qq")
```

**Methods:**

- `New(t *testing.T)` - Creates a new PlaybookTest instance
- `Setup()` - Configures SSH override with mock client
- `Cleanup()` - Restores default SSH behavior
- `ExpectCommand(cmd, output)` - Sets expected command output
- `ExpectError(cmd, error)` - Sets expected command error
- `Config()` - Returns test configuration
- `SetConfig(cfg)` - Sets custom configuration
- `SetArg(key, value)` - Sets a single argument
- `SetArgs(args)` - Replaces the arguments map
- `MockClient()` - Returns the mock SSH client
- `AssertCommandRun(cmd)` - Verifies command was executed
- `AssertCommandNotRun(cmd)` - Verifies command was NOT executed
- `AssertNoError(err)` - Verifies error is nil
- `AssertError(err)` - Verifies error is non-nil
- `AssertErrorContains(err, text)` - Verifies error message contains text
- `AssertResultChanged(result)` - Verifies result indicates changes
- `AssertResultUnchanged(result)` - Verifies result indicates no changes
- `AssertResultNoError(result)` - Verifies result has no error
- `AssertResultError(result)` - Verifies result has an error
- `AssertResultMessageContains(result, text)` - Verifies result message contains text
- `GetCommands()` - Returns all executed commands
- `Reset()` - Clears all recorded commands and expectations

## Integration with SSH Package

The testing framework integrates with the SSH package via the `ssh.SetRunFunc()` function, which allows tests to override the SSH execution function with a mock implementation.

**Important:** Always call `test.Cleanup()` in a defer after `test.Setup()` to restore normal SSH behavior.

## Example Test Files

- `playbooks/ping/ping_mock_test.go` - Example tests for ping playbook
- `playbooks/apt/status_test.go` - Mixed dry-run and mock tests for apt-status

## Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./playbooks/ping/
go test ./playbooks/apt/

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

## Benefits

- **No SSH server required** - Tests run without needing actual SSH connections
- **Fast execution** - Mock tests are much faster than integration tests
- **Deterministic** - Mock outputs are consistent, no network variability
- **Isolated** - Tests don't depend on external systems
- **Comprehensive** - Can test error scenarios and edge cases easily
