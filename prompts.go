package ork

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"strconv"

	"github.com/dracory/ork/types"
	"golang.org/x/term"
)

// PromptPassword securely prompts for a password from stdin.
// The password is not echoed to the terminal.
func PromptPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(bytePassword), nil
}

// readInput reads a single value with optional visibility.
// If private is true, input is hidden (like a password).
// If private is false, input is visible.
func readInput(prompt string, private bool) (string, error) {
	if private {
		return PromptPassword(prompt)
	}

	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return "", nil
}

// PromptForString prompts for a string value.
func PromptForString(prompt string) (string, error) {
	return PromptForStringWithDefault(prompt, "")
}

// PromptForStringWithDefault prompts for a string value with a default.
func PromptForStringWithDefault(prompt, defaultValue string) (string, error) {
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s [%s]: ", prompt, defaultValue)
	} else {
		prompt = fmt.Sprintf("%s: ", prompt)
	}

	value, err := readInput(prompt, false)
	if err != nil {
		return "", err
	}

	if value == "" && defaultValue != "" {
		return defaultValue, nil
	}
	return value, nil
}

// PromptForPassword prompts for a password (hidden input).
func PromptForPassword(prompt string) (string, error) {
	return readInput(prompt+": ", true)
}

// PromptForPasswordWithConfirmation prompts for a password with confirmation.
func PromptForPasswordWithConfirmation(prompt string) (string, error) {
	value, err := PromptForPassword(prompt)
	if err != nil {
		return "", err
	}

	confirmed, err := PromptForPassword("Confirm " + prompt)
	if err != nil {
		return "", err
	}

	if value != confirmed {
		return "", fmt.Errorf("confirmation mismatch")
	}
	return value, nil
}

// PromptForInt prompts for an integer value.
func PromptForInt(prompt string) (int, error) {
	return PromptForIntWithDefault(prompt, 0)
}

// PromptForIntWithDefault prompts for an integer value with a default.
func PromptForIntWithDefault(prompt string, defaultValue int) (int, error) {
	defaultStr := ""
	if defaultValue != 0 {
		defaultStr = fmt.Sprintf("%d", defaultValue)
	}

	valueStr, err := PromptForStringWithDefault(prompt, defaultStr)
	if err != nil {
		return 0, err
	}

	if valueStr == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("must be a valid integer")
	}
	return value, nil
}

// PromptForBool prompts for a boolean value (yes/no).
func PromptForBool(prompt string) (bool, error) {
	return PromptForBoolWithDefault(prompt, false)
}

// PromptForBoolWithDefault prompts for a boolean value with a default.
func PromptForBoolWithDefault(prompt string, defaultValue bool) (bool, error) {
	defaultStr := ""
	if defaultValue {
		defaultStr = "yes"
	} else {
		defaultStr = "no"
	}

	valueStr, err := PromptForStringWithDefault(prompt+" (yes/no)", defaultStr)
	if err != nil {
		return false, err
	}

	if valueStr == "" {
		return defaultValue, nil
	}

	switch valueStr {
	case "yes", "y", "Y", "true", "1":
		return true, nil
	case "no", "n", "N", "false", "0":
		return false, nil
	default:
		return false, fmt.Errorf("must be yes/no")
	}
}

// PromptWithOptions prompts the user to select from a list of options.
// Returns the index of the selected option (0-based).
func PromptWithOptions(prompt string, options []string) (int, error) {
	if len(options) == 0 {
		return 0, fmt.Errorf("no options provided")
	}

	fmt.Println(prompt)
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i+1, option)
	}

	selection, err := PromptForInt("Select option")
	if err != nil {
		return 0, err
	}

	if selection < 1 || selection > len(options) {
		return 0, fmt.Errorf("invalid selection")
	}

	return selection - 1, nil
}

// PromptMultiple prompts for multiple variables using a configuration.
// Returns a map of variable names to user-provided values.
// Skips prompts for variables that already exist in the provided values map.
// Optionally pass existing values as a second argument.
//
// Example:
//
//	prompts := []types.PromptConfig{
//	    {Name: "username", Prompt: "Username", Default: "admin", Required: true},
//	    {Name: "password", Prompt: "Password", Private: true, Confirm: true, Required: true},
//	}
//	results, err := ork.PromptMultiple(prompts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	username := results["username"]
//	password := results["password"]
func PromptMultiple(configs []types.PromptConfig, existingValues ...map[string]string) (types.PromptResult, error) {
	results := make(types.PromptResult)

	// Copy existing values if provided
	if len(existingValues) > 0 {
		maps.Copy(results, existingValues[0])
	}

	for _, cfg := range configs {
		// Skip if already provided
		if _, exists := results[cfg.Name]; exists {
			continue
		}

		var value string
		var err error

		// Handle confirmation
		if cfg.Confirm {
			value, err = PromptForPasswordWithConfirmation(cfg.Prompt)
		} else if cfg.Private {
			value, err = PromptForPassword(cfg.Prompt)
		} else {
			value, err = PromptForStringWithDefault(cfg.Prompt, cfg.Default)
		}

		if err != nil {
			return nil, fmt.Errorf("prompt for %s failed: %w", cfg.Name, err)
		}

		// Handle required fields
		if cfg.Required && value == "" {
			return nil, fmt.Errorf("%s is required", cfg.Name)
		}

		// Handle validation
		if cfg.Validate != nil {
			if err := cfg.Validate(value); err != nil {
				return nil, fmt.Errorf("validation failed for %s: %w", cfg.Name, err)
			}
		}

		results[cfg.Name] = value
	}

	return results, nil
}
