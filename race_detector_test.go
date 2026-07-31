package ork

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/dracory/ork/types"
)

// TestNodeRunConcurrentRace demonstrates the data race that occurs when
// multiple goroutines call node.Run() with the SAME shared skill instance
// but DIFFERENT node configs (different hosts).
//
// The root cause: node.Run() calls skill.SetNodeConfig(), skill.SetDryRun(),
// and skill.SetBecomeUser() on the shared skill BEFORE calling skill.Run().
// When N goroutines do this concurrently with different configs, they race
// on the skill's internal fields — goroutine A's config leaks into
// goroutine B's Run().
//
// This test detects the race by checking for cross-contamination:
// Each goroutine uses a node with a UNIQUE host, and the skill's Run()
// reports which host it sees. Without cloning, goroutines will see
// each other's hosts (or crash with `fatal error: concurrent map writes`).
//
// For environments with cgo/gcc available, also run with -race for the
// definitive detector:
//
//	CGO_ENABLED=1 go test -race -run TestNodeRunConcurrentRace -v
//
// Expected behavior BEFORE the fix:  test fails — goroutines see wrong hosts
// (cross-contamination) or crash with `fatal error: concurrent map read and
// map write`.
//
// Expected behavior AFTER the fix:  test passes — each goroutine sees only
// its own host because each operates on its own clone.
func TestNodeRunConcurrentRace(t *testing.T) {
	const goroutines = 100
	const iterations = 10

	t.Logf("=== TestNodeRunConcurrentRace: %d goroutines x %d iterations, each with unique host ===", goroutines, iterations)
	t.Logf("Expected AFTER fix: each goroutine sees its own host (no cross-contamination)")

	// Create ONE shared skill instance — this simulates the registry pattern
	// where a skill is registered once and executed by multiple nodes.
	// Pre-set one arg so the args map is non-empty (triggers map race too).
	sharedSkill := &raceSkill{
		BaseSkill: types.NewBaseSkill(),
	}
	sharedSkill.SetID("race-skill")
	sharedSkill.SetDescription("Skill for race detection test")
	sharedSkill.SetArg("preset", "shared-value")
	t.Logf("Shared skill created: id=%q, preset arg set", sharedSkill.GetID())

	// Create N nodes, each with a DIFFERENT host.
	// This is the key: without cloning, goroutine A's host leaks into goroutine B.
	nodes := make([]NodeInterface, goroutines)
	for i := 0; i < goroutines; i++ {
		nodes[i] = NewNodeForHost(fmt.Sprintf("host-%d.example.com", i))
	}

	totalCrossContaminated := 0
	totalErrors := 0
	totalCorrect := 0

	for iter := 0; iter < iterations; iter++ {
		// Use a channel as a start barrier — all goroutines start at the same
		// time to maximize contention and increase the chance of detecting
		// the race.
		startBarrier := make(chan struct{})

		var wg sync.WaitGroup
		wg.Add(goroutines)

		// Collect results from each goroutine
		type goroutineResult struct {
			goroutineID  int
			expectedHost string
			seenHost     string
			seenArgCount int
			seenPreset   string
			err          error
		}
		results := make([]goroutineResult, goroutines)

		for i := 0; i < goroutines; i++ {
			i := i
			go func() {
				defer wg.Done()
				expectedHost := fmt.Sprintf("host-%d.example.com", i)

				// Wait for the start signal — maximizes contention
				<-startBarrier

				// Run via the node — internally calls SetNodeConfig/SetDryRun
				// on the shared skill. THIS is the race: concurrent writes.
				r := nodes[i].Run(sharedSkill)
				res, ok := r.Results[expectedHost]
				if !ok {
					// The result key is the node's host. If the skill saw a
					// different host (due to race), the result might be under
					// a different key. Check all keys.
					for _, v := range r.Results {
						res = v
						break
					}
					results[i] = goroutineResult{
						goroutineID:  i,
						expectedHost: expectedHost,
						err:          fmt.Errorf("no result for expected host %q (results map has %d entries)", expectedHost, len(r.Results)),
					}
					return
				}

				// Parse the message from raceSkill.Run():
				// "ran with N args, host=H, preset=P"
				count, host, preset, perr := parseRaceMessage(res.Message)
				if perr != nil {
					results[i] = goroutineResult{
						goroutineID:  i,
						expectedHost: expectedHost,
						err:          fmt.Errorf("failed to parse message %q: %w", res.Message, perr),
					}
					return
				}

				results[i] = goroutineResult{
					goroutineID:  i,
					expectedHost: expectedHost,
					seenHost:     host,
					seenArgCount: count,
					seenPreset:   preset,
					err:          res.Error,
				}
			}()
		}

		// Release all goroutines at the same time
		close(startBarrier)
		wg.Wait()

		// === Analyze results for this iteration ===
		iterCrossContaminated := 0
		iterErrors := 0
		iterCorrect := 0

		for _, r := range results {
			if r.err != nil {
				iterErrors++
			} else if r.seenHost != r.expectedHost {
				iterCrossContaminated++
			} else {
				iterCorrect++
			}
		}

		totalCrossContaminated += iterCrossContaminated
		totalErrors += iterErrors
		totalCorrect += iterCorrect

		// Log per-iteration summary
		status := "PASS"
		if iterCrossContaminated > 0 || iterErrors > 0 {
			status = "FAIL"
		}
		t.Logf("Iteration %2d/%d: correct=%d, cross-contaminated=%d, errors=%d [%s]",
			iter+1, iterations, iterCorrect, iterCrossContaminated, iterErrors, status)

		// Log details for any cross-contaminated goroutines
		for _, r := range results {
			if r.err != nil {
				t.Logf("  goroutine %d: ERROR: %v", r.goroutineID, r.err)
			} else if r.seenHost != r.expectedHost {
				t.Logf("  goroutine %d: CROSS-CONTAMINATED (expected %s, saw %s)",
					r.goroutineID, r.expectedHost, r.seenHost)
			}
		}
	}

	// === Final summary ===
	t.Logf("")
	t.Logf("=== Final Summary (across %d iterations) ===", iterations)
	t.Logf("  Total goroutine-runs:  %d", goroutines*iterations)
	t.Logf("  Correct:                %d", totalCorrect)
	t.Logf("  Cross-contaminated:      %d", totalCrossContaminated)
	t.Logf("  Errors:                 %d", totalErrors)

	if totalCrossContaminated > 0 {
		t.Errorf("DATA RACE detected: %d goroutine-run(s) saw the WRONG host (cross-contamination from concurrent SetNodeConfig on shared skill)", totalCrossContaminated)
	}

	if totalErrors > 0 {
		t.Errorf("DATA RACE detected: %d goroutine-run(s) returned errors (possible fatal crash or corrupted state)", totalErrors)
	}

	if totalCorrect == goroutines*iterations {
		t.Logf("PASS: All %d goroutine-runs saw their own host — no cross-contamination detected", goroutines*iterations)
	}
}

// parseRaceMessage parses a raceSkill.Run() message:
// "ran with N args, host=H, preset=P"
// Returns (argCount, host, preset, error).
func parseRaceMessage(msg string) (int, string, string, error) {
	// Expected format: "ran with N args, host=H, preset=P"
	parts := strings.Split(msg, ", ")
	if len(parts) != 3 {
		return 0, "", "", fmt.Errorf("expected 3 parts, got %d in %q", len(parts), msg)
	}
	// parts[0] = "ran with N args"
	countStr := strings.TrimPrefix(strings.TrimSuffix(parts[0], " args"), "ran with ")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to parse count from %q: %w", parts[0], err)
	}
	// parts[1] = "host=H"
	host := strings.TrimPrefix(parts[1], "host=")
	// parts[2] = "preset=P"
	preset := strings.TrimPrefix(parts[2], "preset=")
	return count, host, preset, nil
}

// raceSkill is a minimal skill that records its config without doing SSH.
// Its Run() method reads the args map and node config — this triggers the
// concurrent map read/write race when multiple goroutines share the same
// instance.
type raceSkill struct {
	*types.BaseSkill
}

func (r *raceSkill) Check() (bool, error) {
	// Read args — races with concurrent SetArgs from other goroutines
	_ = r.GetArgs()
	return true, nil
}

func (r *raceSkill) Run() types.Result {
	// Read args and node config — races with concurrent setters from other goroutines.
	// We report host, arg count, and the preset arg value so the test can detect:
	//   1. Cross-contamination (wrong host)
	//   2. Map corruption (wrong arg count or missing preset)
	//   3. Fatal crash (concurrent map writes — kills the process)
	args := r.GetArgs()
	cfg := r.GetNodeConfig()
	preset := r.GetArg("preset")
	return types.Result{
		Changed: true,
		Message: fmt.Sprintf("ran with %d args, host=%s, preset=%s", len(args), cfg.SSHHost, preset),
	}
}
