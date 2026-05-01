package types

import (
	"errors"
	"testing"
)

func TestResults_FirstResult(t *testing.T) {
	tests := []struct {
		name     string
		results  Results
		wantZero bool
	}{
		{
			name:     "empty results",
			results:  Results{Results: map[string]Result{}},
			wantZero: true,
		},
		{
			name: "single result",
			results: Results{Results: map[string]Result{
				"host1": {Changed: true, Message: "success"},
			}},
			wantZero: false,
		},
		{
			name: "multiple results",
			results: Results{Results: map[string]Result{
				"host1": {Changed: true, Message: "success1"},
				"host2": {Changed: false, Message: "success2"},
			}},
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.results.FirstResult()
			isZero := result.Message == "" && result.Details == nil && result.Error == nil

			if isZero != tt.wantZero {
				t.Errorf("FirstResult() zero = %v, want %v", isZero, tt.wantZero)
			}

			if !tt.wantZero && !isZero {
				// Verify we got a valid result
				if result.Message == "" {
					t.Error("Expected non-empty message")
				}
			}
		})
	}
}

func TestResults_Summary(t *testing.T) {
	tests := []struct {
		name    string
		results Results
		want    Summary
	}{
		{
			name:    "empty results",
			results: Results{Results: map[string]Result{}},
			want:    Summary{Total: 0, Changed: 0, Unchanged: 0, Failed: 0},
		},
		{
			name: "all changed",
			results: Results{Results: map[string]Result{
				"host1": {Changed: true},
				"host2": {Changed: true},
			}},
			want: Summary{Total: 2, Changed: 2, Unchanged: 0, Failed: 0},
		},
		{
			name: "all unchanged",
			results: Results{Results: map[string]Result{
				"host1": {Changed: false},
				"host2": {Changed: false},
			}},
			want: Summary{Total: 2, Changed: 0, Unchanged: 2, Failed: 0},
		},
		{
			name: "mixed",
			results: Results{Results: map[string]Result{
				"host1": {Changed: true},
				"host2": {Changed: false},
				"host3": {Error: errors.New("failed")},
			}},
			want: Summary{Total: 3, Changed: 1, Unchanged: 1, Failed: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.results.Summary()
			if got != tt.want {
				t.Errorf("Summary() = %v, want %v", got, tt.want)
			}
		})
	}
}
