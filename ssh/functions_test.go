package ssh

import "testing"

// TestShellEscapeArg_PlainString verifies that a plain string is wrapped in single quotes.
func TestShellEscapeArg_PlainString(t *testing.T) {
	got := ShellEscapeArg("testuser")
	want := "'testuser'"
	if got != want {
		t.Errorf("ShellEscapeArg(%q) = %q, want %q", "testuser", got, want)
	}
}

// TestShellEscapeArg_EmptyString verifies that an empty string is wrapped in single quotes.
func TestShellEscapeArg_EmptyString(t *testing.T) {
	got := ShellEscapeArg("")
	want := "''"
	if got != want {
		t.Errorf("ShellEscapeArg(%q) = %q, want %q", "", got, want)
	}
}

// TestShellEscapeArg_SingleQuote verifies that embedded single quotes are escaped.
func TestShellEscapeArg_SingleQuote(t *testing.T) {
	got := ShellEscapeArg("foo'bar")
	want := "'foo'\\''bar'"
	if got != want {
		t.Errorf("ShellEscapeArg(%q) = %q, want %q", "foo'bar", got, want)
	}
}

// TestShellEscapeArg_MultipleSingleQuotes verifies multiple embedded single quotes are escaped.
func TestShellEscapeArg_MultipleSingleQuotes(t *testing.T) {
	got := ShellEscapeArg("'foo'bar'")
	want := "''\\''foo'\\''bar'\\'''"
	if got != want {
		t.Errorf("ShellEscapeArg(%q) = %q, want %q", "'foo'bar'", got, want)
	}
}

// TestShellEscapeArg_SemicolonInjection verifies that shell command injection via semicolon is prevented.
func TestShellEscapeArg_SemicolonInjection(t *testing.T) {
	got := ShellEscapeArg("foo; rm -rf /")
	want := "'foo; rm -rf /'"
	if got != want {
		t.Errorf("ShellEscapeArg(%q) = %q, want %q", "foo; rm -rf /", got, want)
	}
}

// TestShellEscapeArg_BacktickInjection verifies that backtick command substitution is prevented.
func TestShellEscapeArg_BacktickInjection(t *testing.T) {
	got := ShellEscapeArg("foo`whoami`")
	want := "'foo`whoami`'"
	if got != want {
		t.Errorf("ShellEscapeArg(%q) = %q, want %q", "foo`whoami`", got, want)
	}
}

// TestShellEscapeArg_DollarInjection verifies that $ variable expansion is prevented.
func TestShellEscapeArg_DollarInjection(t *testing.T) {
	got := ShellEscapeArg("foo$HOME")
	want := "'foo$HOME'"
	if got != want {
		t.Errorf("ShellEscapeArg(%q) = %q, want %q", "foo$HOME", got, want)
	}
}

// TestShellEscapeArg_PipeInjection verifies that pipe-based injection is prevented.
func TestShellEscapeArg_PipeInjection(t *testing.T) {
	got := ShellEscapeArg("foo | cat /etc/passwd")
	want := "'foo | cat /etc/passwd'"
	if got != want {
		t.Errorf("ShellEscapeArg(%q) = %q, want %q", "foo | cat /etc/passwd", got, want)
	}
}

// TestShellEscapeArg_AndInjection verifies that && injection is prevented.
func TestShellEscapeArg_AndInjection(t *testing.T) {
	got := ShellEscapeArg("foo && whoami")
	want := "'foo && whoami'"
	if got != want {
		t.Errorf("ShellEscapeArg(%q) = %q, want %q", "foo && whoami", got, want)
	}
}

// TestShellEscapeArg_NewlineInjection verifies that newline-based injection is prevented.
func TestShellEscapeArg_NewlineInjection(t *testing.T) {
	got := ShellEscapeArg("foo\nwhoami")
	want := "'foo\nwhoami'"
	if got != want {
		t.Errorf("ShellEscapeArg(%q) = %q, want %q", "foo\nwhoami", got, want)
	}
}
