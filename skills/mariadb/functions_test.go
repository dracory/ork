package mariadb

import "testing"

func TestMariadbEscapeSQLQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no quotes", "hello", "hello"},
		{"single quote", "it's", "it''s"},
		{"multiple single quotes", "a'b'c", "a''b''c"},
		{"consecutive single quotes", "''", "''''"},
		{"double quote (not escaped)", `say "hi"`, `say "hi"`},
		{"backtick (not escaped)", "test`db", "test`db"},
		{"backslash (not escaped)", `C:\path`, `C:\path`},
		{"mixed", `it's a "test"`, `it''s a "test"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mariadbEscapeSQLQuote(tt.input)
			if got != tt.want {
				t.Errorf("mariadbEscapeSQLQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMariadbEscapeShellQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no quotes", "hello", "hello"},
		{"single quote", "it's", `it'\''s`},
		{"multiple single quotes", "a'b'c", `a'\''b'\''c`},
		{"consecutive single quotes", "''", `'\'''\''`},
		{"double quote (not escaped)", `say "hi"`, `say "hi"`},
		{"dollar sign (not escaped)", "$HOME", "$HOME"},
		{"backtick (not escaped)", "test`cmd", "test`cmd"},
		{"backslash (not escaped)", `C:\path`, `C:\path`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mariadbEscapeShellQuote(tt.input)
			if got != tt.want {
				t.Errorf("mariadbEscapeShellQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMariadbEscapeSQLIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no backticks", "mydb", "mydb"},
		{"single backtick", "my`db", "my``db"},
		{"multiple backticks", "a`b`c", "a``b``c"},
		{"consecutive backticks", "``", "````"},
		{"single quote (not escaped)", "it's", "it's"},
		{"double quote (not escaped)", `say "hi"`, `say "hi"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mariadbEscapeSQLIdentifier(tt.input)
			if got != tt.want {
				t.Errorf("mariadbEscapeSQLIdentifier(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMariadbEscapeShellDoubleQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no special chars", "hello", "hello"},
		{"double quote", `say "hi"`, `say \"hi\"`},
		{"dollar sign", "$HOME", `\$HOME`},
		{"backtick", "test`cmd", "test\\`cmd"},
		{"backslash", `C:\path`, `C:\\path`},
		{"single quote (not escaped)", "it's", "it's"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mariadbEscapeShellDoubleQuote(tt.input)
			if got != tt.want {
				t.Errorf("mariadbEscapeShellDoubleQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
