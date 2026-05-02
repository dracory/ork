package examples

import (
	"testing"

	"github.com/dracory/ork"
)

func TestExampleCommand(t *testing.T) {
	// Test that we can create a command with fluent chaining
	command := ork.NewCommand().
		WithDescription("Check server uptime").
		WithCommand("uptime").
		WithRequired(true)

	// Verify the command was configured correctly
	if command.GetDescription() != "Check server uptime" {
		t.Errorf("Expected description to be 'Check server uptime', got %s", command.GetDescription())
	}

	// Note: We can't test GetCommand() since CommandInterface doesn't expose it directly
	// This is a demonstration test to verify the fluent API compiles correctly
}

func TestExampleCommandInventory(t *testing.T) {
	// Test that we can create a command for inventory usage
	command := ork.NewCommand().
		WithDescription("Restart application").
		WithCommand("pm2 restart app").
		WithRequired(true)

	if command.GetDescription() != "Restart application" {
		t.Errorf("Expected description to be 'Restart application', got %s", command.GetDescription())
	}
}

func TestExampleCommandNotRequired(t *testing.T) {
	// Test that we can create a command that's not required
	command := ork.NewCommand().
		WithDescription("Non-critical operation").
		WithCommand("some-non-critical-command").
		WithRequired(false)

	if command.GetDescription() != "Non-critical operation" {
		t.Errorf("Expected description to be 'Non-critical operation', got %s", command.GetDescription())
	}
}

func TestExampleCommandWithBecome(t *testing.T) {
	// Test that we can create a command with become user
	command := ork.NewCommand().
		WithDescription("Backup database as postgres user").
		WithCommand("pg_dump mydb").
		WithRequired(true).
		WithBecomeUser("postgres")

	if command.GetDescription() != "Backup database as postgres user" {
		t.Errorf("Expected description to be 'Backup database as postgres user', got %s", command.GetDescription())
	}

	if command.GetBecomeUser() != "postgres" {
		t.Errorf("Expected become user to be 'postgres', got %s", command.GetBecomeUser())
	}
}

func TestExampleCommandWithChdir(t *testing.T) {
	// Test that we can create a command with chdir
	command := ork.NewCommand().
		WithDescription("List files in web directory").
		WithCommand("ls -la").
		WithRequired(true).
		WithChdir("/var/www")

	if command.GetDescription() != "List files in web directory" {
		t.Errorf("Expected description to be 'List files in web directory', got %s", command.GetDescription())
	}

	// Note: We can't test GetChdir() since CommandInterface doesn't expose it directly
	// This is a demonstration test to verify the fluent API compiles correctly
}
