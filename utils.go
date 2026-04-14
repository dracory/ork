package ork

import (
	"fmt"
	"os"

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
