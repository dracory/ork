package ping

import (
	"fmt"
	"testing"

	"github.com/dracory/ork/internal/playbooktest"
)

// TestPing_Run_WithMock demonstrates using the mock SSH client for testing.
func TestPing_Run_WithMock(t *testing.T) {
	test := playbooktest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand("uptime", " 10:30:01 up 5 days,  2 users,  load average: 0.01, 0.05, 0.00")

	pb := NewPing()
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultNoError(result)
	test.AssertResultUnchanged(result)
	test.AssertCommandRun("uptime")
	test.AssertResultMessageContains(result, "is alive")
}

// TestPing_Run_WithMockError demonstrates testing error scenarios.
func TestPing_Run_WithMockError(t *testing.T) {
	test := playbooktest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectError("uptime", fmt.Errorf("connection refused"))

	pb := NewPing()
	pb.SetNodeConfig(test.Config())
	result := pb.Run()

	test.AssertResultError(result)
	test.AssertErrorContains(result.Error, "failed to ping")
}

// TestPing_Check_WithMock demonstrates testing the Check method.
func TestPing_Check_WithMock(t *testing.T) {
	test := playbooktest.New(t)
	defer test.Cleanup()

	test.Setup()
	test.ExpectCommand("uptime", " 10:30:01 up 5 days")

	pb := NewPing()
	pb.SetNodeConfig(test.Config())
	needsChange, err := pb.Check()

	test.AssertNoError(err)
	if needsChange {
		t.Error("Expected Check to return false for read-only operation")
	}
	test.AssertCommandRun("uptime")
}
