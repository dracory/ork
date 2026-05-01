package examples

import (
	"fmt"

	"github.com/dracory/ork"
)

// ExampleSkillFluent demonstrates using skills with fluent With* chaining.
// This shows how to configure skills using the fluent interface pattern.
func ExampleSkillFluent() {
	// Create node config with fluent chaining
	cfg := ork.NewNodeConfig().
		WithHost("server.example.com").
		WithPort("22").
		WithLogin("ubuntu").
		WithKey("/home/user/.ssh/id_rsa")

	// Create a skill and configure it using fluent chaining - just like ork.NewCommand()
	skill := ork.NewSkill().
		WithID("check-connectivity").
		WithDescription("Check if server is reachable").
		WithDryRun(false).
		WithNodeConfig(*cfg)

	// Check if changes are needed
	needsRun, err := skill.Check()
	if err != nil {
		fmt.Printf("Check failed: %v\n", err)
		return
	}

	if needsRun {
		// Run the skill
		result := skill.Run()
		if result.Error != nil {
			fmt.Printf("Skill failed: %v\n", result.Error)
		} else {
			fmt.Printf("Skill succeeded: %s\n", result.Message)
		}
	} else {
		fmt.Println("No changes needed")
	}
}

// ExampleSkillWithArgs demonstrates configuring skills with arguments.
func ExampleSkillWithArgs() {
	// Create node config with fluent chaining
	cfg := ork.NewNodeConfig().
		WithHost("server.example.com").
		WithPort("22").
		WithLogin("ubuntu").
		WithKey("/home/user/.ssh/id_rsa")

	// Create skill with fluent chaining including arguments
	skill := ork.NewSkill().
		WithID("ping-with-args").
		WithDescription("Ping with custom args").
		WithArg("count", "5").
		WithArg("timeout", "10").
		WithNodeConfig(*cfg)

	result := skill.Run()

	if result.Error != nil {
		fmt.Printf("Failed: %v\n", result.Error)
	} else {
		fmt.Printf("Success: %s\n", result.Message)
	}
}

// ExampleSkillMixedChaining demonstrates using fluent With* methods.
func ExampleSkillMixedChaining() {
	// Create node config with fluent chaining
	cfg := ork.NewNodeConfig().
		WithHost("server.example.com").
		WithPort("22").
		WithLogin("ubuntu").
		WithKey("/home/user/.ssh/id_rsa")

	// Create skill with fluent chaining using With* methods
	skill := ork.NewSkill().
		WithID("mixed-example").
		WithDescription("Example using With* methods").
		WithNodeConfig(*cfg).
		WithDryRun(false).
		WithTimeout(30)

	result := skill.Run()

	if result.Error != nil {
		fmt.Printf("Failed: %v\n", result.Error)
	} else {
		fmt.Printf("Success: %s\n", result.Message)
	}
}
