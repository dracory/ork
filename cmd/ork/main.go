package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dracory/ork/vault"
	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "vault":
		handleVaultCommand()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleVaultCommand() {
	if len(os.Args) < 3 {
		printVaultUsage()
		os.Exit(1)
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "init":
		handleVaultInit()
	case "set":
		handleVaultSet()
	case "get":
		handleVaultGet()
	case "delete":
		handleVaultDelete()
	case "list":
		handleVaultList()
	case "changepassword":
		handleVaultChangePassword()
	case "ui":
		handleVaultUI()
	default:
		fmt.Printf("Unknown vault command: %s\n\n", subcommand)
		printVaultUsage()
		os.Exit(1)
	}
}

func handleVaultInit() {
	if len(os.Args) < 4 {
		fmt.Println("Error: vault path required")
		fmt.Println("\nUsage: ork vault init <path>")
		os.Exit(1)
	}

	path := os.Args[3]

	// Check if stdin is piped (for automation)
	stat, _ := os.Stdin.Stat()
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	var password string
	var err error

	if isPiped {
		// Read password from stdin once for automation
		password, err = readPasswordFromStdin()
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Using password from stdin")
	} else {
		// Interactive mode - prompt for password
		password, err = promptPassword("Enter password: ")
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			os.Exit(1)
		}

		// Confirm password
		confirm, err := promptPassword("Confirm password: ")
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			os.Exit(1)
		}

		if password != confirm {
			fmt.Println("Error: passwords do not match")
			os.Exit(1)
		}
	}

	// Create vault
	v, err := vault.Create(path, password)
	if err != nil {
		fmt.Printf("Error creating vault: %v\n", err)
		os.Exit(1)
	}

	if err := v.Close(); err != nil {
		fmt.Printf("Error closing vault: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Vault created: %s\n", path)
}

func handleVaultSet() {
	if len(os.Args) < 5 {
		fmt.Println("Error: vault path and key required")
		fmt.Println("\nUsage: ork vault set <path> <key> [value]")
		os.Exit(1)
	}

	path := os.Args[3]
	key := os.Args[4]

	var value string
	var password string
	var err error

	if len(os.Args) >= 6 {
		value = strings.Join(os.Args[5:], " ")
		password, err = readPasswordFromStdin()
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Read both value and password from stdin
		stat, _ := os.Stdin.Stat()
		isPiped := (stat.Mode() & os.ModeCharDevice) == 0

		if isPiped {
			// Piped mode: read all lines upfront to avoid scanner buffer conflicts
			scanner := bufio.NewScanner(os.Stdin)
			lineNum := 0
			for scanner.Scan() {
				line := scanner.Text()
				if lineNum == 0 {
					value = line
				} else if lineNum == 1 {
					password = strings.TrimSpace(line)
				}
				lineNum++
			}
			if err := scanner.Err(); err != nil {
				fmt.Printf("Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
			if value == "" {
				fmt.Println("Error: value required via stdin")
				os.Exit(1)
			}
			if password == "" {
				fmt.Println("Error: password required via stdin")
				os.Exit(1)
			}
		} else {
			// Interactive mode: read value from stdin, then prompt for password
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				value = scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				fmt.Printf("Error reading value: %v\n", err)
				os.Exit(1)
			}
			password, err = promptPassword("Enter vault password: ")
			if err != nil {
				fmt.Printf("Error reading password: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// Open vault
	v, err := vault.Open(path, password)
	if err != nil {
		fmt.Printf("Error opening vault: %v\n", err)
		os.Exit(1)
	}
	defer v.Close()

	// Set value
	if err := v.KeySet(key, value); err != nil {
		fmt.Printf("Error setting key: %v\n", err)
		os.Exit(1)
	}

	if err := v.Save(); err != nil {
		fmt.Printf("Error saving vault: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Key set: %s\n", key)
}

func handleVaultGet() {
	if len(os.Args) < 5 {
		fmt.Println("Error: vault path and key required")
		fmt.Println("\nUsage: ork vault get <path> <key>")
		os.Exit(1)
	}

	path := os.Args[3]
	key := os.Args[4]

	// Prompt for password
	password, err := readPasswordFromStdin()
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		os.Exit(1)
	}

	// Open vault
	v, err := vault.Open(path, password)
	if err != nil {
		fmt.Printf("Error opening vault: %v\n", err)
		os.Exit(1)
	}
	defer v.Close()

	// Get value
	value, err := v.KeyGet(key)
	if err != nil {
		fmt.Printf("Error getting key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(value)
}

func handleVaultDelete() {
	if len(os.Args) < 5 {
		fmt.Println("Error: vault path and key required")
		fmt.Println("\nUsage: ork vault delete <path> <key>")
		os.Exit(1)
	}

	path := os.Args[3]
	key := os.Args[4]

	// Prompt for password
	password, err := readPasswordFromStdin()
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		os.Exit(1)
	}

	// Open vault
	v, err := vault.Open(path, password)
	if err != nil {
		fmt.Printf("Error opening vault: %v\n", err)
		os.Exit(1)
	}
	defer v.Close()

	// Delete key
	if err := v.KeyDelete(key); err != nil {
		fmt.Printf("Error deleting key: %v\n", err)
		os.Exit(1)
	}

	if err := v.Save(); err != nil {
		fmt.Printf("Error saving vault: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Key deleted: %s\n", key)
}

func handleVaultList() {
	if len(os.Args) < 4 {
		fmt.Println("Error: vault path required")
		fmt.Println("\nUsage: ork vault list <path>")
		os.Exit(1)
	}

	path := os.Args[3]

	// Prompt for password
	password, err := readPasswordFromStdin()
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		os.Exit(1)
	}

	// Open vault
	v, err := vault.Open(path, password)
	if err != nil {
		fmt.Printf("Error opening vault: %v\n", err)
		os.Exit(1)
	}
	defer v.Close()

	// List keys
	keys := v.KeyList()
	for _, key := range keys {
		fmt.Println(key)
	}
}

func handleVaultChangePassword() {
	if len(os.Args) < 4 {
		fmt.Println("Error: vault path required")
		fmt.Println("\nUsage: ork vault changepassword <path>")
		os.Exit(1)
	}

	path := os.Args[3]

	// Prompt for current password
	currentPassword, err := promptPassword("Enter current password: ")
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		os.Exit(1)
	}

	// Prompt for new password
	newPassword, err := promptPassword("Enter new password: ")
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		os.Exit(1)
	}

	// Confirm new password
	confirm, err := promptPassword("Confirm new password: ")
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		os.Exit(1)
	}

	if newPassword != confirm {
		fmt.Println("Error: passwords do not match")
		os.Exit(1)
	}

	// Open vault with current password
	v, err := vault.Open(path, currentPassword)
	if err != nil {
		fmt.Printf("Error opening vault: %v\n", err)
		os.Exit(1)
	}
	defer v.Close()

	// Change password
	if err := v.ChangePassword(newPassword); err != nil {
		fmt.Printf("Error changing password: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Password changed successfully")
}

// readPasswordFromStdin reads password from stdin (for automation via pipes)
func readPasswordFromStdin() (string, error) {
	// Check if stdin is a pipe or has data available
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// stdin is a pipe or redirected, read from it
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			password := strings.TrimSpace(scanner.Text())
			if password == "" {
				return "", fmt.Errorf("password cannot be empty")
			}
			return password, nil
		}
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("no password provided")
	}

	// stdin is a terminal, prompt for password
	return promptPassword("Enter vault password: ")
}

// promptPassword prompts for password without echo
func promptPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	defer fmt.Println()

	var password []byte
	var err error

	if term.IsTerminal(int(os.Stdin.Fd())) {
		password, err = term.ReadPassword(int(os.Stdin.Fd()))
	} else {
		// Not a terminal, read normally
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			password = []byte(strings.TrimSpace(scanner.Text()))
		} else {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", fmt.Errorf("no password provided")
		}
	}

	if err != nil {
		return "", err
	}

	if len(password) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}

	return string(password), nil
}

func handleVaultUI() {
	if len(os.Args) < 4 {
		fmt.Println("Error: vault path required")
		fmt.Println("\nUsage: ork vault ui <path> [address]")
		fmt.Println("  path    Vault file path")
		fmt.Println("  address Listen address (default: 127.0.0.1:38080)")
		os.Exit(1)
	}

	path := os.Args[3]
	address := "127.0.0.1:38080"
	if len(os.Args) > 4 {
		address = os.Args[4]
	}

	// Check if vault exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Error: vault file does not exist: %s\n", path)
		os.Exit(1)
	}

	fmt.Printf("Starting Ork Vault UI on http://%s\n", address)
	fmt.Printf("Vault file: %s\n", path)
	fmt.Println("Press Ctrl+C to stop")

	// Start HTTP server
	if err := startUIServer(path, address); err != nil {
		fmt.Printf("Error starting UI server: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Ork - SSH-based server automation")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ork <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  vault    Vault operations for secure secrets management")
	fmt.Println("  help     Show this help message")
	fmt.Println()
	fmt.Println("Use 'ork vault' without arguments for vault subcommand help")
}

func printVaultUsage() {
	fmt.Println("Ork Vault - Secure secrets management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ork vault <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init            Create a new vault")
	fmt.Println("  set             Set a key in the vault")
	fmt.Println("  get             Get a value from the vault")
	fmt.Println("  delete          Delete a key from the vault")
	fmt.Println("  list            List all keys in the vault")
	fmt.Println("  changepassword  Change vault password")
	fmt.Println("  ui              Start web UI for vault management")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ork vault init secrets.vault")
	fmt.Println("  ork vault set secrets.vault DB_HOST localhost")
	fmt.Println("  ork vault get secrets.vault DB_HOST")
	fmt.Println("  ork vault delete secrets.vault DB_HOST")
	fmt.Println("  ork vault list secrets.vault")
	fmt.Println("  ork vault changepassword secrets.vault")
	fmt.Println("  ork vault ui secrets.vault")
}
