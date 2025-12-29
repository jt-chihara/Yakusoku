// Package matcher provides matching rule implementations for contract verification.
package matcher

// MatchResult represents the result of a match operation.
type MatchResult struct {
	Matched bool
	Diff    string
}

// Matcher is the interface that all matchers must implement.
type Matcher interface {
	// Name returns the matcher type name (e.g., "equality", "type", "regex").
	Name() string

	// Match compares expected and actual values according to the matcher's rules.
	Match(expected, actual interface{}) (*MatchResult, error)
}
