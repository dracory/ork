// Package types provides shared types for ork packages.
package types

// Result represents the outcome of a playbook execution.
// It indicates whether any changes were made and provides details about the execution.
type Result struct {
	// Changed indicates whether the playbook made any changes to the system.
	// false means the system was already in the desired state.
	// true means the playbook modified the system.
	Changed bool

	// Message is a human-readable description of what happened.
	Message string

	// Details contains additional information about the execution.
	// Keys are field names, values are string representations.
	Details map[string]string

	// Error is non-nil if the playbook failed to execute.
	// When Error is non-nil, Changed may be true if some changes occurred before the failure.
	Error error
}

// Results contains per-node results from any operation.
type Results struct {
	Results map[string]Result
}

// Summary returns aggregated statistics.
func (r Results) Summary() Summary {
	var s Summary
	for _, res := range r.Results {
		s.Total++
		if res.Error != nil {
			s.Failed++
		} else if res.Changed {
			s.Changed++
		} else {
			s.Unchanged++
		}
	}
	return s
}

// Summary holds aggregated result statistics.
type Summary struct {
	Total     int
	Changed   int
	Unchanged int
	Failed    int
}
