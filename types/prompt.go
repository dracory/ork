package types

// PromptConfig defines a single prompt configuration
type PromptConfig struct {
	Name     string             // Variable name
	Prompt   string             // Prompt message to display
	Private  bool               // Hide input (true) or show it (false)
	Default  string             // Default value if user provides no input
	Confirm  bool               // Require confirmation (for passwords)
	Validate func(string) error // Validation function
	Required bool               // Whether the field is required
}

// PromptResult contains the results of a prompt session
type PromptResult map[string]string
