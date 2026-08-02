package ork_test

// Integration tests for Inventory concurrency control (SetMaxConcurrency).
//
// These tests verify that SetMaxConcurrency actually limits parallelism by
// comparing execution timing between sequential (MaxConcurrency=1) and
// parallel (MaxConcurrency=3+) modes. The absolute timings vary with Docker
// overhead, so we compare relative timing: parallel should be significantly
// faster than sequential.

import (
	"testing"
	"time"

	"github.com/dracory/ork"
)

// TestIntegration_Inventory_Concurrency_SerialVsParallel verifies that
// MaxConcurrency=1 (serial) is significantly slower than MaxConcurrency=3
// (parallel) for the same workload. This is a comparative test that is
// robust to Docker overhead variance.
func TestIntegration_Inventory_Concurrency_SerialVsParallel(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	// Create 3 nodes with persistent connections
	nodes := make([]ork.NodeInterface, 3)
	for i := range nodes {
		n := newTestNode(container)
		if err := n.Connect(); err != nil {
			t.Fatalf("Node %d Connect failed: %v", i, err)
		}
		defer n.Close()
		nodes[i] = n
	}

	// Measure serial execution (MaxConcurrency=1)
	invSerial := ork.NewInventory()
	for _, n := range nodes {
		invSerial.AddNode(n)
	}
	invSerial.SetMaxConcurrency(1)

	startSerial := time.Now()
	invSerial.RunCommand("sleep 1")
	serialElapsed := time.Since(startSerial)

	// Measure parallel execution (MaxConcurrency=3)
	invParallel := ork.NewInventory()
	for _, n := range nodes {
		invParallel.AddNode(n)
	}
	invParallel.SetMaxConcurrency(3)

	startParallel := time.Now()
	invParallel.RunCommand("sleep 1")
	parallelElapsed := time.Since(startParallel)

	t.Logf("Serial (MaxConcurrency=1):   %v", serialElapsed)
	t.Logf("Parallel (MaxConcurrency=3): %v", parallelElapsed)

	// Parallel should be at least 30% faster than serial.
	// With 3 * sleep 1: serial ~3s, parallel ~1s. Even with Docker overhead,
	// parallel should be noticeably faster.
	if parallelElapsed >= serialElapsed {
		t.Errorf("Expected parallel (%v) to be faster than serial (%v)", parallelElapsed, serialElapsed)
	}

	// Serial should take at least 2x the parallel time (3 nodes, sleep 1 each)
	// serial ~3s, parallel ~1s → ratio ~3x. Allow 1.5x as a conservative threshold.
	if serialElapsed < parallelElapsed*3/2 {
		t.Errorf("Expected serial to be at least 1.5x slower than parallel: serial=%v parallel=%v", serialElapsed, parallelElapsed)
	}
}

// TestIntegration_Inventory_Concurrency1 verifies that SetMaxConcurrency(1)
// causes commands to run sequentially. Three nodes each run `sleep 1`;
// with concurrency=1, total time should be at least 2 seconds (3 * 1s,
// allowing for some overlap in SSH session setup).
func TestIntegration_Inventory_Concurrency1(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node1 := newTestNode(container)
	node2 := newTestNode(container)
	node3 := newTestNode(container)

	for i, n := range []ork.NodeInterface{node1, node2, node3} {
		if err := n.Connect(); err != nil {
			t.Fatalf("Node %d Connect failed: %v", i, err)
		}
		defer n.Close()
	}

	inventory := ork.NewInventory()
	inventory.AddNode(node1)
	inventory.AddNode(node2)
	inventory.AddNode(node3)
	inventory.SetMaxConcurrency(1)

	start := time.Now()
	results := inventory.RunCommand("sleep 1")
	elapsed := time.Since(start)

	// With concurrency=1, 3 * sleep 1 should take at least 2.5 seconds
	// (3s of sleep, allowing small overlap in session setup)
	if elapsed < 2500*time.Millisecond {
		t.Errorf("Expected >= 2.5s with MaxConcurrency=1 (sequential), got %v", elapsed)
	}

	if len(results.Results) == 0 {
		t.Fatal("Expected at least 1 result, got 0")
	}

	for _, result := range results.Results {
		if result.Error != nil {
			t.Errorf("RunCommand failed: %v", result.Error)
		}
	}
}

// TestIntegration_Inventory_Concurrency3 verifies that SetMaxConcurrency(3)
// allows commands to run in parallel. Three nodes each run `sleep 1`;
// with concurrency=3, total time should be less than the sequential case.
func TestIntegration_Inventory_Concurrency3(t *testing.T) {
	container := setupSSHContainer(t)
	defer container.terminate(t)

	node1 := newTestNode(container)
	node2 := newTestNode(container)
	node3 := newTestNode(container)

	for i, n := range []ork.NodeInterface{node1, node2, node3} {
		if err := n.Connect(); err != nil {
			t.Fatalf("Node %d Connect failed: %v", i, err)
		}
		defer n.Close()
	}

	inventory := ork.NewInventory()
	inventory.AddNode(node1)
	inventory.AddNode(node2)
	inventory.AddNode(node3)
	inventory.SetMaxConcurrency(3)

	start := time.Now()
	results := inventory.RunCommand("sleep 1")
	elapsed := time.Since(start)

	// With concurrency=3 and persistent connections, 3 * sleep 1 should
	// take ~1s + overhead. Should be well under 3s (the sequential time).
	if elapsed >= 3*time.Second {
		t.Errorf("Expected < 3s with MaxConcurrency=3 (parallel), got %v", elapsed)
	}

	if len(results.Results) == 0 {
		t.Fatal("Expected at least 1 result, got 0")
	}

	for _, result := range results.Results {
		if result.Error != nil {
			t.Errorf("RunCommand failed: %v", result.Error)
		}
	}
}
