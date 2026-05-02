package examples

import (
	"testing"

	"github.com/dracory/ork"
)

func TestExampleSkillFluent(t *testing.T) {
	// This is a demonstration function, so we just verify it compiles
	// In a real scenario, this would require SSH mocking
	cfg := ork.NewNodeConfig().
		WithHost("server.example.com").
		WithPort("22").
		WithLogin("ubuntu").
		WithKey("/home/user/.ssh/id_rsa")

	skill := ork.NewSkill().
		WithID("check-connectivity").
		WithDescription("Check if server is reachable").
		WithDryRun(false).
		WithNodeConfig(*cfg)

	// Verify the skill was configured correctly
	if skill.GetID() != "check-connectivity" {
		t.Errorf("Expected ID to be 'check-connectivity', got %s", skill.GetID())
	}

	if skill.GetDescription() != "Check if server is reachable" {
		t.Errorf("Expected description to be 'Check if server is reachable', got %s", skill.GetDescription())
	}

	if skill.IsDryRun() {
		t.Error("Expected DryRun to be false")
	}
}

func TestExampleSkillWithArgs(t *testing.T) {
	cfg := ork.NewNodeConfig().
		WithHost("server.example.com").
		WithPort("22").
		WithLogin("ubuntu").
		WithKey("/home/user/.ssh/id_rsa")

	skill := ork.NewSkill().
		WithID("ping-with-args").
		WithDescription("Ping with custom args").
		WithArg("count", "5").
		WithArg("timeout", "10").
		WithNodeConfig(*cfg)

	// Verify the skill was configured correctly
	if skill.GetID() != "ping-with-args" {
		t.Errorf("Expected ID to be 'ping-with-args', got %s", skill.GetID())
	}

	if skill.GetArg("count") != "5" {
		t.Errorf("Expected arg 'count' to be '5', got %s", skill.GetArg("count"))
	}

	if skill.GetArg("timeout") != "10" {
		t.Errorf("Expected arg 'timeout' to be '10', got %s", skill.GetArg("timeout"))
	}
}

func TestExampleSkillMixedChaining(t *testing.T) {
	cfg := ork.NewNodeConfig().
		WithHost("server.example.com").
		WithPort("22").
		WithLogin("ubuntu").
		WithKey("/home/user/.ssh/id_rsa")

	skill := ork.NewSkill().
		WithID("mixed-example").
		WithDescription("Example using With* methods").
		WithNodeConfig(*cfg).
		WithDryRun(false).
		WithTimeout(30)

	// Verify the skill was configured correctly
	if skill.GetID() != "mixed-example" {
		t.Errorf("Expected ID to be 'mixed-example', got %s", skill.GetID())
	}

	if skill.GetDescription() != "Example using With* methods" {
		t.Errorf("Expected description to be 'Example using With* methods', got %s", skill.GetDescription())
	}

	if skill.IsDryRun() {
		t.Error("Expected DryRun to be false")
	}

	if skill.GetTimeout() != 30 {
		t.Errorf("Expected timeout to be 30, got %v", skill.GetTimeout())
	}
}
