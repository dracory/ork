package ssh

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// becomePromptTemplate is the prompt string ork passes to sudo -p.
// It is intentionally unique so it can be reliably detected in the output
// stream and stripped out. The %s placeholder is replaced with a unique ID
// per execution (matching Ansible's approach).
const becomePromptTemplate = "[sudo via ork, key=%s] password: "

// becomeSuccessTemplate is the success marker echoed before the wrapped command.
// The connection layer watches for it to confirm escalation succeeded before
// sending command stdin. The %s placeholder is the same unique ID as the prompt.
const becomeSuccessTemplate = "BECOME-SUCCESS-%s"

// sudo error strings (from lib/ansible/plugins/become/sudo.py).
// Used to detect wrong/missing passwords and fail with a clear error instead
// of hanging or producing confusing output.
var (
	sudoFailErrs    = []string{"Sorry, try again."}
	sudoMissingErrs = []string{"Sorry, a password is required to run sudo", "sudo: a password is required"}
)

// becomeState tracks the 3-state machine for sequencing password delivery
// and command stdin, matching Ansible's awaiting_prompt → awaiting_escalation
// → ready_to_send pipeline.
type becomeState int

const (
	stateAwaitingPrompt     becomeState = iota
	stateAwaitingEscalation             // prompt seen, waiting for success marker or error
	stateReadyToSend                    // escalation confirmed, command stdin can be sent
	stateDone                           // terminal
)

// becomeWriter captures combined stdout/stderr output, detects the sudo prompt,
// success marker, and error strings, and drives the state machine. It writes
// the password to the stdin pipe when the prompt appears, and writes command
// stdin after the success marker appears. Prompt, marker, and error lines are
// suppressed from the captured output so callers get clean output.
type becomeWriter struct {
	mu        sync.Mutex
	buf       bytes.Buffer
	remainder []byte // incomplete trailing line carried to next Write
	prompt    string
	success   string
	password  string
	cmdStdin  string
	stdinPipe io.WriteCloser
	state     becomeState
	sentCh    chan struct{} // closed when password has been sent
	readyCh   chan struct{} // closed when escalation succeeded (ready to send cmd stdin)
	errCh     chan error    // receives a non-nil error on wrong/missing password
}

func (w *becomeWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// After escalation is confirmed, just accumulate output without inspection.
	if w.state >= stateReadyToSend {
		w.buf.Write(p)
		return len(p), nil
	}

	// Process line-by-line, carrying any incomplete trailing line to the next
	// Write (matches Ansible's _examine_output remainder handling).
	data := append(w.remainder, p...)
	w.remainder = nil

	for {
		nl := bytes.IndexByte(data, '\n')
		if nl < 0 {
			// No complete line. Check for the prompt in the incomplete data —
			// sudo -S -p writes the prompt WITHOUT a trailing newline, so we
			// must detect it here, not just in handleLine. This matches
			// Ansible's check_password_prompt which uses splitlines() (returns
			// incomplete trailing data as a line too).
			if w.state == stateAwaitingPrompt && w.prompt != "" {
				lineStr := string(bytes.TrimSpace(data))
				if strings.HasPrefix(lineStr, strings.TrimSpace(w.prompt)) {
					io.WriteString(w.stdinPipe, w.password+"\n")
					w.state = stateAwaitingEscalation
					close(w.sentCh)
					// Discard the prompt data — it should not appear in output.
					w.remainder = nil
					return len(p), nil
				}
			}
			// Not a prompt — keep as remainder for next Write.
			w.remainder = data
			break
		}
		line := data[:nl+1] // include the newline
		data = data[nl+1:]

		// Only inspect lines while negotiating escalation.
		if w.state < stateReadyToSend {
			if handled := w.handleLine(line); handled {
				// line was a prompt/success/error line — do not append to buf
				continue
			}
		}
		w.buf.Write(line)
	}

	return len(p), nil
}

// handleLine inspects one complete line during the escalation negotiation.
// Returns true if the line is a prompt/success/error line and should be
// suppressed from the captured output.
func (w *becomeWriter) handleLine(line []byte) bool {
	trimmed := bytes.TrimRight(line, "\r\n")
	lineStr := string(bytes.TrimSpace(trimmed))

	// 1. Prompt detection (only while awaiting the prompt).
	// Match Ansible: strip both the prompt and the line, then check startswith.
	if w.state == stateAwaitingPrompt && w.prompt != "" && strings.HasPrefix(lineStr, strings.TrimSpace(w.prompt)) {
		io.WriteString(w.stdinPipe, w.password+"\n")
		w.state = stateAwaitingEscalation
		close(w.sentCh)
		return true
	}

	// 2. Success marker detection.
	if w.success != "" && bytes.Contains(trimmed, []byte(w.success)) {
		// Escalation confirmed — send command stdin (if any), then close stdin.
		if w.cmdStdin != "" {
			io.WriteString(w.stdinPipe, w.cmdStdin)
		}
		w.stdinPipe.Close()
		w.state = stateReadyToSend
		close(w.readyCh)
		return true
	}

	// 3. Wrong password detection.
	for _, msg := range sudoFailErrs {
		if bytes.Contains(trimmed, []byte(msg)) {
			select {
			case w.errCh <- fmt.Errorf("incorrect sudo password"):
			default:
				// error already sent
			}
			w.state = stateDone
			return true
		}
	}

	// 4. Missing password detection (sudo needs a password but none was provided).
	for _, msg := range sudoMissingErrs {
		if bytes.Contains(trimmed, []byte(msg)) {
			select {
			case w.errCh <- fmt.Errorf("sudo requires a password but none was provided"):
			default:
				// error already sent
			}
			w.state = stateDone
			return true
		}
	}

	return false
}

// RunWithBecome executes a command with prompt-triggered sudo password delivery.
// becomePrompt is the exact prompt string passed to sudo -p (ssh.Run generates it
// and passes it here so the writer knows what to detect in the output stream).
// becomeSuccess is the success marker echoed before the wrapped command.
// If becomePassword is empty, it behaves like Run (with optional stdin via strings.NewReader).
// If becomePassword is non-empty, it uses StdinPipe + the state machine to deliver
// the password on demand, then the command stdin after escalation succeeds.
func (c *Client) RunWithBecome(cmd string, becomePassword string, becomePrompt string, becomeSuccess string, stdin string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("not connected, call Connect() first")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// No become password — simple path (existing behavior + optional stdin).
	// sudo -n is used by the caller in this path so sudo fails fast if a
	// password is actually required, instead of hanging. This is the
	// NOPASSWD / cached-credentials path.
	if becomePassword == "" {
		if stdin != "" {
			session.Stdin = strings.NewReader(stdin)
		}
		output, err := session.CombinedOutput(cmd)
		return string(output), err
	}

	// Become path — state machine + prompt detection.
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	writer := &becomeWriter{
		prompt:    becomePrompt,
		success:   becomeSuccess,
		password:  becomePassword,
		cmdStdin:  stdin,
		stdinPipe: stdinPipe,
		state:     stateAwaitingPrompt,
		sentCh:    make(chan struct{}),
		readyCh:   make(chan struct{}),
		errCh:     make(chan error, 1),
	}

	session.Stdout = writer
	session.Stderr = writer

	if err := session.Start(cmd); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Safety net: if the session ends without the success marker ever
	// appearing (e.g. network drop, command killed), close stdin so the
	// pipe is not left open. handleLine normally closes stdin when it
	// sees the success marker; this covers the abnormal-termination case.
	// Closing an already-closed pipe is harmless (we guard with the state).
	doneCh := make(chan struct{})
	go func() {
		select {
		case <-writer.readyCh:
			// stdin already closed by handleLine — nothing to do.
		case <-doneCh:
			// session finished without success marker — close stdin.
			_ = closeOnce(stdinPipe, writer)
		}
	}()

	waitErr := session.Wait()
	close(doneCh)

	// A wrong/missing password is reported via errCh; surface it preferentially.
	select {
	case err := <-writer.errCh:
		return writer.buf.String(), err
	default:
	}

	return writer.buf.String(), waitErr
}

// closeOnce closes the stdin pipe only if escalation has not already closed it.
// This prevents goroutine leaks without double-closing. It also transitions
// the state to stateDone so the state machine is not left in an inconsistent
// state after abnormal termination.
func closeOnce(stdinPipe io.WriteCloser, w *becomeWriter) error {
	w.mu.Lock()
	ready := w.state >= stateReadyToSend
	if !ready {
		w.state = stateDone
	}
	w.mu.Unlock()
	if ready {
		return nil // already closed by handleLine
	}
	return stdinPipe.Close()
}

// BuildBecomePrompt builds the unique sudo prompt string for the given ID.
// Exported so both ssh.Run (in functions.go) and tests can use it.
func BuildBecomePrompt(id string) string {
	return fmt.Sprintf(becomePromptTemplate, id)
}

// BuildBecomeSuccess builds the success marker string for the given ID.
// Exported so both ssh.Run (in functions.go) and tests can use it.
func BuildBecomeSuccess(id string) string {
	return fmt.Sprintf(becomeSuccessTemplate, id)
}

// generateBecomeID produces a short unique string via crypto/rand (16 hex chars).
// Do not use a timestamp — parallel executions can collide. This matches
// Ansible's self._id approach (32 random lowercase letters via secrets.choice).
func generateBecomeID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to a timestamp-based ID if crypto/rand fails (extremely unlikely).
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// sentChannel is a small helper for tests to check if the password was sent.
func (w *becomeWriter) passwordSent() bool {
	select {
	case <-w.sentCh:
		return true
	default:
		return false
	}
}

// readyChannel is a small helper for tests to check if escalation succeeded.
func (w *becomeWriter) escalationReady() bool {
	select {
	case <-w.readyCh:
		return true
	default:
		return false
	}
}

// buffer returns the captured output (for tests).
func (w *becomeWriter) buffer() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

// stateValue returns the current state (for tests).
func (w *becomeWriter) stateValue() becomeState {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.state
}

// newBecomeWriterForTest constructs a becomeWriter wired to the given pipe.
// Exposed for unit tests that drive the writer directly without an SSH session.
func newBecomeWriterForTest(prompt, success, password, cmdStdin string, stdinPipe io.WriteCloser) *becomeWriter {
	return &becomeWriter{
		prompt:    prompt,
		success:   success,
		password:  password,
		cmdStdin:  cmdStdin,
		stdinPipe: stdinPipe,
		state:     stateAwaitingPrompt,
		sentCh:    make(chan struct{}),
		readyCh:   make(chan struct{}),
		errCh:     make(chan error, 1),
	}
}
