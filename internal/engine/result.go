package engine

// CheckResult holds the outcome of a single check execution.
// Status is one of "pass", "fail", or "skipped".
type CheckResult struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Status   string `json:"status"`
	Reason   string `json:"reason,omitempty"`
	Message  string `json:"message,omitempty"`
	Fix      string `json:"fix,omitempty"`
}

// RunReport is the top-level output of a Run() call, suitable for both
// human rendering and JSON serialization via --format=json.
type RunReport struct {
	Passed        int           `json:"passed"`
	Failed        int           `json:"failed"`
	BlockerFailed bool          `json:"blocker_failed"`
	Checks        []CheckResult `json:"checks"`
}
