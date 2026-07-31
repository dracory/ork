package skills

import "testing"

func TestShellEscapeContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "''"},
		{"no special chars", "hello", "'hello'"},
		{"single quote", "it's", `'it'\''s'`},
		{"multiple single quotes", "a'b'c", `'a'\''b'\''c'`},
		{"consecutive single quotes", "''", `''\'''\'''`},
		{"double quote (not escaped)", `say "hi"`, `'say "hi"'`},
		{"dollar sign (not escaped)", "$HOME", "'$HOME'"},
		{"backtick (not escaped)", "test`cmd", "'test`cmd'"},
		{"backslash (not escaped)", `C:\path`, `'C:\path'`},
		{"newline (not escaped)", "line1\nline2", "'line1\nline2'"},
		{"mixed", `it's a "test" with $VAR`, `'it'\''s a "test" with $VAR'`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShellEscapeContent(tt.input)
			if got != tt.want {
				t.Errorf("ShellEscapeContent(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestShellEscapeArg(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "''"},
		{"no special chars", "hello", "'hello'"},
		{"single quote", "it's", `'it'\''s'`},
		{"multiple single quotes", "a'b'c", `'a'\''b'\''c'`},
		{"double quote (not escaped)", `say "hi"`, `'say "hi"'`},
		{"dollar sign (not escaped)", "$HOME", "'$HOME'"},
		{"backtick (not escaped)", "test`cmd", "'test`cmd'"},
		{"backslash (not escaped)", `C:\path`, `'C:\path'`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShellEscapeArg(tt.input)
			if got != tt.want {
				t.Errorf("ShellEscapeArg(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
