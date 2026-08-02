package ssh

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// pipeStub simulates an SSH stdin pipe for becomeWriter tests.
// It captures everything written to it so tests can assert on the writes.
type pipeStub struct {
	mu   sync.Mutex
	buf  bytes.Buffer
	open bool
}

func newPipeStub() *pipeStub {
	return &pipeStub{open: true}
}

func (p *pipeStub) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.open {
		return 0, io.ErrClosedPipe
	}
	return p.buf.Write(b)
}

func (p *pipeStub) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.open = false
	return nil
}

func (p *pipeStub) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.buf.String()
}

func (p *pipeStub) isOpen() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.open
}

// --- BuildBecomePrompt / BuildBecomeSuccess / generateBecomeID ---

func TestBuildBecomePrompt(t *testing.T) {
	got := BuildBecomePrompt("abc123")
	want := "[sudo via ork, key=abc123] password: "
	if got != want {
		t.Errorf("BuildBecomePrompt(%q) = %q, want %q", "abc123", got, want)
	}
}

func TestBuildBecomeSuccess(t *testing.T) {
	got := BuildBecomeSuccess("abc123")
	want := "BECOME-SUCCESS-abc123"
	if got != want {
		t.Errorf("BuildBecomeSuccess(%q) = %q, want %q", "abc123", got, want)
	}
}

func TestGenerateBecomeID_UniquenessAndLength(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := generateBecomeID()
		if len(id) != 16 { // 8 bytes → 16 hex chars
			t.Errorf("generateBecomeID() returned length %d, want 16 (id=%q)", len(id), id)
		}
		if ids[id] {
			t.Errorf("generateBecomeID() returned duplicate id %q on iteration %d", id, i)
		}
		ids[id] = true
	}
}

// --- becomeWriter state machine tests ---

// TestBecomeWriter_PromptInSingleWrite verifies that the writer detects the
// prompt in a single write (WITHOUT a trailing newline, matching real sudo -S
// behavior), sends the password, advances to awaiting_escalation, and
// suppresses the prompt from the buffer.
func TestBecomeWriter_PromptInSingleWrite(t *testing.T) {
	pipe := newPipeStub()
	w := newBecomeWriterForTest(
		"[sudo via ork, key=test] password: ",
		"BECOME-SUCCESS-test",
		"secret123",
		"",
		pipe,
	)

	// Real sudo -S -p writes the prompt WITHOUT a trailing newline.
	promptData := "[sudo via ork, key=test] password: "
	n, err := w.Write([]byte(promptData))
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if n != len(promptData) {
		t.Errorf("Write returned %d bytes, want %d", n, len(promptData))
	}

	// Password should have been sent to the pipe
	if pipe.String() != "secret123\n" {
		t.Errorf("Expected password 'secret123\\n' sent to pipe, got %q", pipe.String())
	}

	// State should be awaiting_escalation
	if w.stateValue() != stateAwaitingEscalation {
		t.Errorf("Expected state stateAwaitingEscalation, got %d", w.stateValue())
	}

	// passwordSent should be true
	if !w.passwordSent() {
		t.Error("Expected passwordSent() to be true")
	}

	// escalationReady should be false (no success marker yet)
	if w.escalationReady() {
		t.Error("Expected escalationReady() to be false before success marker")
	}

	// The prompt should be suppressed from the buffer
	if w.buffer() != "" {
		t.Errorf("Expected empty buffer (prompt suppressed), got %q", w.buffer())
	}
}

// TestBecomeWriter_PromptSplitAcrossWrites verifies that the writer detects
// the prompt when it arrives in fragments across multiple Write calls (without
// a trailing newline, matching real sudo -S behavior).
func TestBecomeWriter_PromptSplitAcrossWrites(t *testing.T) {
	pipe := newPipeStub()
	w := newBecomeWriterForTest(
		"[sudo via ork, key=test] password: ",
		"BECOME-SUCCESS-test",
		"secret123",
		"",
		pipe,
	)

	// Write the prompt in fragments — no newline (real sudo doesn't send one)
	w.Write([]byte("[sudo via o"))
	w.Write([]byte("rk, key=test] password: "))

	// Password should have been sent
	if pipe.String() != "secret123\n" {
		t.Errorf("Expected password 'secret123\\n' sent to pipe, got %q", pipe.String())
	}

	if w.stateValue() != stateAwaitingEscalation {
		t.Errorf("Expected state stateAwaitingEscalation, got %d", w.stateValue())
	}

	// Buffer should be empty (prompt suppressed)
	if w.buffer() != "" {
		t.Errorf("Expected empty buffer (prompt suppressed), got %q", w.buffer())
	}
}

// TestBecomeWriter_SuccessMarkerAfterPrompt verifies that after the prompt is
// detected and the password sent, the success marker is detected, command stdin
// is sent, stdin is closed, and the state advances to ready_to_send.
func TestBecomeWriter_SuccessMarkerAfterPrompt(t *testing.T) {
	pipe := newPipeStub()
	w := newBecomeWriterForTest(
		"[sudo via ork, key=test] password: ",
		"BECOME-SUCCESS-test",
		"secret123",
		"hello-world",
		pipe,
	)

	// Phase 1: send the prompt (no newline — real sudo behavior)
	w.Write([]byte("[sudo via ork, key=test] password: "))
	if !w.passwordSent() {
		t.Fatal("Expected password to be sent after prompt")
	}

	// Phase 2: send the success marker
	w.Write([]byte("BECOME-SUCCESS-test\n"))

	if w.stateValue() != stateReadyToSend {
		t.Errorf("Expected state stateReadyToSend, got %d", w.stateValue())
	}

	if !w.escalationReady() {
		t.Error("Expected escalationReady() to be true after success marker")
	}

	// Pipe should contain password + command stdin
	expected := "secret123\nhello-world"
	if pipe.String() != expected {
		t.Errorf("Expected pipe to contain %q, got %q", expected, pipe.String())
	}

	// Pipe should be closed
	if pipe.isOpen() {
		t.Error("Expected pipe to be closed after success marker")
	}

	// Buffer should be empty (both prompt and marker suppressed)
	if w.buffer() != "" {
		t.Errorf("Expected empty buffer, got %q", w.buffer())
	}
}

// TestBecomeWriter_SuccessMarkerWithoutPrompt verifies the NOPASSWD case:
// sudo doesn't prompt (NOPASSWD or cached credentials), the success marker
// appears directly. The password is never sent, command stdin is sent on
// the success marker.
func TestBecomeWriter_SuccessMarkerWithoutPrompt(t *testing.T) {
	pipe := newPipeStub()
	w := newBecomeWriterForTest(
		"[sudo via ork, key=test] password: ",
		"BECOME-SUCCESS-test",
		"secret123",
		"hello-world",
		pipe,
	)

	// Send the success marker directly (no prompt — NOPASSWD case)
	w.Write([]byte("BECOME-SUCCESS-test\n"))

	// Password should NOT have been sent
	if w.passwordSent() {
		t.Error("Expected password NOT to be sent in NOPASSWD case")
	}

	// State should be ready_to_send (success marker advances directly)
	if w.stateValue() != stateReadyToSend {
		t.Errorf("Expected state stateReadyToSend, got %d", w.stateValue())
	}

	// Pipe should contain only command stdin (no password)
	if pipe.String() != "hello-world" {
		t.Errorf("Expected pipe to contain only 'hello-world', got %q", pipe.String())
	}

	// Pipe should be closed
	if pipe.isOpen() {
		t.Error("Expected pipe to be closed after success marker")
	}
}

// TestBecomeWriter_WrongPassword verifies that "Sorry, try again." is detected
// and an error is sent to errCh.
func TestBecomeWriter_WrongPassword(t *testing.T) {
	pipe := newPipeStub()
	w := newBecomeWriterForTest(
		"[sudo via ork, key=test] password: ",
		"BECOME-SUCCESS-test",
		"wrongpass",
		"",
		pipe,
	)

	// Send the prompt (no newline — real sudo behavior), then the wrong-password error
	w.Write([]byte("[sudo via ork, key=test] password: "))
	w.Write([]byte("Sorry, try again.\n"))

	// errCh should have the incorrect password error
	select {
	case err := <-w.errCh:
		if err == nil {
			t.Fatal("Expected non-nil error from errCh")
		}
		if !strings.Contains(err.Error(), "incorrect sudo password") {
			t.Errorf("Expected error to contain 'incorrect sudo password', got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for error from errCh")
	}

	// State should be done
	if w.stateValue() != stateDone {
		t.Errorf("Expected state stateDone, got %d", w.stateValue())
	}
}

// TestBecomeWriter_MissingPassword verifies that "Sorry, a password is required"
// is detected and an error is sent to errCh.
func TestBecomeWriter_MissingPassword(t *testing.T) {
	pipe := newPipeStub()
	w := newBecomeWriterForTest(
		"[sudo via ork, key=test] password: ",
		"BECOME-SUCCESS-test",
		"",
		"",
		pipe,
	)

	// Send the missing-password error directly (no prompt)
	w.Write([]byte("Sorry, a password is required to run sudo\n"))

	select {
	case err := <-w.errCh:
		if err == nil {
			t.Fatal("Expected non-nil error from errCh")
		}
		if !strings.Contains(err.Error(), "sudo requires a password") {
			t.Errorf("Expected error to contain 'sudo requires a password', got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for error from errCh")
	}

	if w.stateValue() != stateDone {
		t.Errorf("Expected state stateDone, got %d", w.stateValue())
	}
}

// TestBecomeWriter_OutputAfterReadyToSend verifies that after escalation is
// confirmed, output is accumulated without inspection (no false matches on
// prompt-like command output).
func TestBecomeWriter_OutputAfterReadyToSend(t *testing.T) {
	pipe := newPipeStub()
	w := newBecomeWriterForTest(
		"[sudo via ork, key=test] password: ",
		"BECOME-SUCCESS-test",
		"secret123",
		"",
		pipe,
	)

	// Complete the escalation (prompt without newline — real sudo behavior)
	w.Write([]byte("[sudo via ork, key=test] password: "))
	w.Write([]byte("BECOME-SUCCESS-test\n"))

	// Now write output that contains the prompt string — it should be
	// accumulated as-is, not interpreted as a prompt.
	commandOutput := "[sudo via ork, key=test] password: this is just command output\n"
	w.Write([]byte(commandOutput))

	if w.buffer() != commandOutput {
		t.Errorf("Expected command output to be accumulated as-is after ready_to_send, got %q", w.buffer())
	}
}

// TestBecomeWriter_ConcurrentWrites verifies that the writer is safe for
// concurrent calls from stdout and stderr goroutines.
func TestBecomeWriter_ConcurrentWrites(t *testing.T) {
	pipe := newPipeStub()
	w := newBecomeWriterForTest(
		"[sudo via ork, key=test] password: ",
		"BECOME-SUCCESS-test",
		"secret123",
		"",
		pipe,
	)

	// Complete the escalation first so concurrent writes just accumulate
	// (prompt without newline — real sudo behavior)
	w.Write([]byte("[sudo via ork, key=test] password: "))
	w.Write([]byte("BECOME-SUCCESS-test\n"))

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			w.Write([]byte("line from goroutine\n"))
		}(i)
	}
	wg.Wait()

	// All 10 lines should be in the buffer
	output := w.buffer()
	lines := strings.Count(output, "line from goroutine\n")
	if lines != 10 {
		t.Errorf("Expected 10 lines in buffer, got %d", lines)
	}
}

// TestBecomeWriter_NoPromptNoMarker verifies that if neither the prompt nor
// the success marker ever appears, the password is never sent and the buffer
// accumulates output.
func TestBecomeWriter_NoPromptNoMarker(t *testing.T) {
	pipe := newPipeStub()
	w := newBecomeWriterForTest(
		"[sudo via ork, key=test] password: ",
		"BECOME-SUCCESS-test",
		"secret123",
		"",
		pipe,
	)

	// Write some output that is neither a prompt nor a marker
	w.Write([]byte("some command output\n"))
	w.Write([]byte("more output\n"))

	// Password should not have been sent
	if w.passwordSent() {
		t.Error("Expected password NOT to be sent when no prompt appears")
	}

	// State should still be awaiting_prompt
	if w.stateValue() != stateAwaitingPrompt {
		t.Errorf("Expected state stateAwaitingPrompt, got %d", w.stateValue())
	}

	// Buffer should contain the output
	expected := "some command output\nmore output\n"
	if w.buffer() != expected {
		t.Errorf("Expected buffer %q, got %q", expected, w.buffer())
	}
}

// TestBecomeWriter_PromptWithNewline verifies that the writer also detects the
// prompt when it DOES include a trailing newline (some sudo implementations
// may emit one). This is the fallback path through handleLine.
func TestBecomeWriter_PromptWithNewline(t *testing.T) {
	pipe := newPipeStub()
	w := newBecomeWriterForTest(
		"[sudo via ork, key=test] password: ",
		"BECOME-SUCCESS-test",
		"secret123",
		"",
		pipe,
	)

	// Prompt WITH a trailing newline — should be detected by handleLine
	w.Write([]byte("[sudo via ork, key=test] password: \n"))

	if pipe.String() != "secret123\n" {
		t.Errorf("Expected password 'secret123\\n' sent to pipe, got %q", pipe.String())
	}
	if w.stateValue() != stateAwaitingEscalation {
		t.Errorf("Expected state stateAwaitingEscalation, got %d", w.stateValue())
	}
	if w.buffer() != "" {
		t.Errorf("Expected empty buffer (prompt suppressed), got %q", w.buffer())
	}
}

// TestBecomeWriter_PromptWithPartialLineThenNewline verifies that the writer
// detects the prompt when the first write contains a partial prompt (no
// newline) and the second write completes it with a newline.
func TestBecomeWriter_PromptWithPartialLineThenNewline(t *testing.T) {
	pipe := newPipeStub()
	w := newBecomeWriterForTest(
		"[sudo via ork, key=test] password: ",
		"BECOME-SUCCESS-test",
		"secret123",
		"",
		pipe,
	)

	// First write: partial prompt, no newline — goes to remainder, not yet detected
	w.Write([]byte("[sudo via ork, key=te"))
	// Password should NOT be sent yet (partial prompt)
	if pipe.String() != "" {
		t.Errorf("Expected no password sent for partial prompt, got %q", pipe.String())
	}

	// Second write: completes the prompt with a newline
	w.Write([]byte("st] password: \n"))

	if pipe.String() != "secret123\n" {
		t.Errorf("Expected password 'secret123\\n' sent to pipe, got %q", pipe.String())
	}
	if w.stateValue() != stateAwaitingEscalation {
		t.Errorf("Expected state stateAwaitingEscalation, got %d", w.stateValue())
	}
	if w.buffer() != "" {
		t.Errorf("Expected empty buffer (prompt suppressed), got %q", w.buffer())
	}
}

// --- ssh.Run command wrapping tests (via SetRunSingleCommandFunc) ---

// TestRun_BecomeUserNoPassword_UsesFailFast verifies that when BecomeUser is set
// but BecomePassword is empty, the command uses sudo -H -n (fail-fast).
func TestRun_BecomeUserNoPassword_UsesFailFast(t *testing.T) {
	var capturedCmd types.Command
	SetRunSingleCommandFunc(func(host, port, user, key string, cmd types.Command, kexAlgorithms []string, hostKeyAlgorithms []string, becomePassword string, becomePrompt string, becomeSuccess string) (string, error) {
		capturedCmd = cmd
		if becomePassword != "" {
			t.Errorf("Expected empty becomePassword, got %q", becomePassword)
		}
		return "output", nil
	})
	defer SetRunSingleCommandFunc(nil)

	cfg := types.NodeConfig{
		SSHHost:      "localhost",
		SSHPort:      "22",
		SSHLogin:     "root",
		SSHKey:       "test",
		BecomeUser:   "postgres",
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	cmd := types.Command{
		Command:    "psql -l",
		BecomeUser: "postgres",
		Required:   true,
	}

	Run(cfg, cmd)

	if !strings.HasPrefix(capturedCmd.Command, "sudo -H -n -u 'postgres' ") {
		t.Errorf("Expected command to start with 'sudo -H -n -u \\'postgres\\' ', got: %s", capturedCmd.Command)
	}
}

// TestRun_BecomeUserWithPassword_UsesPromptDetection verifies that when both
// BecomeUser and BecomePassword are set, the command uses sudo -H -S -p with
// the success marker, and the password/prompt/success are passed through.
func TestRun_BecomeUserWithPassword_UsesPromptDetection(t *testing.T) {
	var capturedCmd types.Command
	var capturedPassword, capturedPrompt, capturedSuccess string
	SetRunSingleCommandFunc(func(host, port, user, key string, cmd types.Command, kexAlgorithms []string, hostKeyAlgorithms []string, becomePassword string, becomePrompt string, becomeSuccess string) (string, error) {
		capturedCmd = cmd
		capturedPassword = becomePassword
		capturedPrompt = becomePrompt
		capturedSuccess = becomeSuccess
		return "output", nil
	})
	defer SetRunSingleCommandFunc(nil)

	cfg := types.NodeConfig{
		SSHHost:        "localhost",
		SSHPort:        "22",
		SSHLogin:       "root",
		SSHKey:         "test",
		BecomeUser:     "postgres",
		BecomePassword: "secret-pass",
		IsDryRunMode:   false,
		Logger:         slog.Default(),
	}

	cmd := types.Command{
		Command:    "psql -l",
		BecomeUser: "postgres",
		Required:   true,
	}

	Run(cfg, cmd)

	// Password should be passed through
	if capturedPassword != "secret-pass" {
		t.Errorf("Expected becomePassword 'secret-pass', got %q", capturedPassword)
	}

	// Prompt should match the template
	if !strings.HasPrefix(capturedPrompt, "[sudo via ork, key=") {
		t.Errorf("Expected becomePrompt to start with '[sudo via ork, key=', got %q", capturedPrompt)
	}

	// Success marker should match the template
	if !strings.HasPrefix(capturedSuccess, "BECOME-SUCCESS-") {
		t.Errorf("Expected becomeSuccess to start with 'BECOME-SUCCESS-', got %q", capturedSuccess)
	}

	// Command should use sudo -H -S -p
	if !strings.HasPrefix(capturedCmd.Command, "sudo -H -S -p ") {
		t.Errorf("Expected command to start with 'sudo -H -S -p ', got: %s", capturedCmd.Command)
	}

	// Command should contain -u 'postgres'
	if !strings.Contains(capturedCmd.Command, "-u 'postgres'") {
		t.Errorf("Expected command to contain -u 'postgres', got: %s", capturedCmd.Command)
	}

	// Command should contain the success marker echo (shell-escaped as '\'' inside bash -c)
	if !strings.Contains(capturedCmd.Command, "echo '\\''BECOME-SUCCESS-") {
		t.Errorf("Expected command to contain echo '\\''BECOME-SUCCESS-...' (shell-escaped), got: %s", capturedCmd.Command)
	}
}

// TestRun_BecomeUserWithPassword_FullCommandStructure verifies the EXACT
// command structure, including that the semicolon is inside bash -c (inside
// sudo's scope). This is the regression test for the bug where the semicolon
// was at the top level, causing the actual command to run as the wrong user.
func TestRun_BecomeUserWithPassword_FullCommandStructure(t *testing.T) {
	var capturedCmd types.Command
	var capturedSuccess string
	SetRunSingleCommandFunc(func(host, port, user, key string, cmd types.Command, kexAlgorithms []string, hostKeyAlgorithms []string, becomePassword string, becomePrompt string, becomeSuccess string) (string, error) {
		capturedCmd = cmd
		capturedSuccess = becomeSuccess
		return "output", nil
	})
	defer SetRunSingleCommandFunc(nil)

	cfg := types.NodeConfig{
		SSHHost:        "localhost",
		SSHPort:        "22",
		SSHLogin:       "root",
		SSHKey:         "test",
		BecomeUser:     "postgres",
		BecomePassword: "secret-pass",
		IsDryRunMode:   false,
		Logger:         slog.Default(),
	}

	cmd := types.Command{
		Command:    "psql -l",
		BecomeUser: "postgres",
		Required:   true,
	}

	Run(cfg, cmd)

	// The exact expected structure:
	// sudo -H -S -p '[sudo via ork, key=<id>] password: ' -u 'postgres' bash -c 'echo '\''BECOME-SUCCESS-<id>'\''; psql -l'
	//
	// The critical part is "bash -c '...'" wrapping the echo;cmd so the
	// semicolon is inside sudo's scope. Without bash -c, the shell would
	// split at the semicolon and psql would run as the original user.
	successMarker := capturedSuccess
	cmdStr := capturedCmd.Command

	expectedPrefix := "sudo -H -S -p '[sudo via ork, key="
	if !strings.HasPrefix(cmdStr, expectedPrefix) {
		t.Errorf("Command should start with %q\nGot: %s", expectedPrefix, cmdStr)
	}

	// Must contain "bash -c" (the critical wrapper)
	if !strings.Contains(cmdStr, " bash -c '") {
		t.Fatalf("Command must contain ' bash -c '\\''...' — the semicolon must be inside bash -c (inside sudo scope)\nGot: %s", cmdStr)
	}

	// Verify the semicolon is INSIDE the bash -c quoting, not at the top level.
	bashIdx := strings.Index(cmdStr, " bash -c '")
	if bashIdx < 0 {
		t.Fatalf("Could not find ' bash -c '\\''' in command: %s", cmdStr)
	}
	afterBash := cmdStr[bashIdx+len(" bash -c '"):]
	// The closing quote should be at the very end
	if !strings.HasSuffix(afterBash, "'") {
		t.Errorf("bash -c argument should end with '\\''\nGot: %s", afterBash)
	}
	inner := afterBash[:len(afterBash)-1] // strip closing quote
	// The semicolon should be inside this inner string
	if !strings.Contains(inner, "; ") {
		t.Errorf("Semicolon should be inside bash -c argument\nInner: %s", inner)
	}
	// Verify there is NO semicolon at the top level (outside bash -c)
	beforeBash := cmdStr[:bashIdx]
	if strings.Contains(beforeBash, ";") {
		t.Errorf("There should be NO semicolon before 'bash -c' — it would split at the top level\nBefore bash -c: %s", beforeBash)
	}

	// Verify the success marker appears inside the bash -c argument
	if !strings.Contains(inner, successMarker) {
		t.Errorf("Success marker %q should be inside bash -c argument\nInner: %s", successMarker, inner)
	}

	// Verify the actual command (psql -l) appears inside the bash -c argument
	if !strings.Contains(inner, "psql -l") {
		t.Errorf("Command 'psql -l' should be inside bash -c argument\nInner: %s", inner)
	}
}

// TestRun_BecomeUserWithPasswordAndStdin verifies that BecomeUser + Stdin is
// supported (no error) when BecomePassword is set.
func TestRun_BecomeUserWithPasswordAndStdin(t *testing.T) {
	var capturedStdin string
	SetRunSingleCommandFunc(func(host, port, user, key string, cmd types.Command, kexAlgorithms []string, hostKeyAlgorithms []string, becomePassword string, becomePrompt string, becomeSuccess string) (string, error) {
		capturedStdin = cmd.Stdin
		return "output", nil
	})
	defer SetRunSingleCommandFunc(nil)

	cfg := types.NodeConfig{
		SSHHost:        "localhost",
		SSHPort:        "22",
		SSHLogin:       "root",
		SSHKey:         "test",
		BecomeUser:     "root",
		BecomePassword: "secret-pass",
		IsDryRunMode:   false,
		Logger:         slog.Default(),
	}

	cmd := types.Command{
		Command:    "cat",
		BecomeUser: "root",
		Stdin:      "hello-from-stdin",
		Required:   true,
	}

	_, err := Run(cfg, cmd)
	if err != nil {
		t.Fatalf("Expected no error for BecomeUser + Stdin with password, got: %v", err)
	}

	if capturedStdin != "hello-from-stdin" {
		t.Errorf("Expected stdin 'hello-from-stdin' to be passed through, got %q", capturedStdin)
	}
}

// TestRun_DryRun_DoesNotLogBecomePassword verifies that BecomePassword is never
// logged in dry-run mode.
func TestRun_DryRun_DoesNotLogBecomePassword(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	cfg := types.NodeConfig{
		IsDryRunMode:   true,
		BecomeUser:     "root",
		BecomePassword: "super-secret-password",
		Logger:         logger,
	}

	cmd := types.Command{
		Command:     "id",
		Description: "Check identity",
	}

	_, err := Run(cfg, cmd)
	if err != nil {
		t.Fatalf("Expected no error in dry-run mode, got: %v", err)
	}

	logOutput := buf.String()
	if strings.Contains(logOutput, "super-secret-password") {
		t.Errorf("BecomePassword should NOT appear in dry-run log, but log contains: %s", logOutput)
	}
}

// TestRun_DryRun_DoesNotLogStdin verifies that cmd.Stdin is never logged in
// dry-run mode.
func TestRun_DryRun_DoesNotLogStdin(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       logger,
	}

	cmd := types.Command{
		Command:     "cat",
		Description: "Read stdin",
		Stdin:       "secret-stdin-data",
	}

	_, err := Run(cfg, cmd)
	if err != nil {
		t.Fatalf("Expected no error in dry-run mode, got: %v", err)
	}

	logOutput := buf.String()
	if strings.Contains(logOutput, "secret-stdin-data") {
		t.Errorf("Stdin should NOT appear in dry-run log, but log contains: %s", logOutput)
	}
}
