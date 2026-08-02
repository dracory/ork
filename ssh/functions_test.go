package ssh

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/dracory/ork/types"
)

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

// TestRun_DryRun_SensitiveRedacted verifies that a sensitive command is redacted in dry-run logs.
func TestRun_DryRun_SensitiveRedacted(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       logger,
	}

	cmd := types.Command{
		Command:     "echo 'secret_password' | chpasswd",
		Description: "Set user password",
		Sensitive:   true,
	}

	output, err := Run(cfg, cmd)
	if err != nil {
		t.Fatalf("Expected no error in dry-run mode, got: %v", err)
	}
	if output != "[dry-run]" {
		t.Errorf("Expected '[dry-run]' output, got '%s'", output)
	}

	logOutput := buf.String()
	if strings.Contains(logOutput, "secret_password") {
		t.Errorf("Sensitive command should be redacted in dry-run log, but log contains: %s", logOutput)
	}
	if strings.Contains(logOutput, "[redacted]") {
		// Expected: the redacted placeholder appears
	} else {
		t.Errorf("Expected '[redacted]' in dry-run log, but log is: %s", logOutput)
	}
}

// TestRun_DryRun_NonSensitiveNotRedacted verifies that a non-sensitive command is logged as-is in dry-run.
func TestRun_DryRun_NonSensitiveNotRedacted(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       logger,
	}

	cmd := types.Command{
		Command:     "id testuser",
		Description: "Check if user exists",
		Sensitive:   false,
	}

	_, err := Run(cfg, cmd)
	if err != nil {
		t.Fatalf("Expected no error in dry-run mode, got: %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "id testuser") {
		t.Errorf("Non-sensitive command should appear in dry-run log, but log is: %s", logOutput)
	}
}

// TestRun_NotRequired_SensitiveRedacted verifies that a sensitive command is redacted
// in the warning log when a non-required command exits non-zero.
func TestRun_NotRequired_SensitiveRedacted(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	// Use SetRunSingleCommandFunc to simulate an exit error (non-zero exit code)
	SetRunSingleCommandFunc(func(host, port, user, key string, cmd types.Command, kexAlgorithms []string, hostKeyAlgorithms []string) (string, error) {
		return "some output", NewExitError()
	})
	defer SetRunSingleCommandFunc(nil)

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       logger,
	}

	cmd := types.Command{
		Command:     "echo 'secret_password' | chpasswd",
		Description: "Set user password",
		Required:    false,
		Sensitive:   true,
	}

	output, err := Run(cfg, cmd)
	if err != nil {
		t.Fatalf("Expected no error for non-required exit failure, got: %v", err)
	}
	if output != "some output" {
		t.Errorf("Expected output 'some output', got '%s'", output)
	}

	logOutput := buf.String()
	if strings.Contains(logOutput, "secret_password") {
		t.Errorf("Sensitive command should be redacted in warning log, but log contains: %s", logOutput)
	}
	if !strings.Contains(logOutput, "[redacted]") {
		t.Errorf("Expected '[redacted]' in warning log, but log is: %s", logOutput)
	}
}
