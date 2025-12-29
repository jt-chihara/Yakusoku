package matcher

import (
	"fmt"

	"github.com/jt-chihara/yakusoku/internal/contract"
)

// Comparator orchestrates matching using the appropriate matchers.
type Comparator struct {
	matchers map[string]Matcher
}

// NewComparator creates a new Comparator with default matchers.
func NewComparator() *Comparator {
	c := &Comparator{
		matchers: make(map[string]Matcher),
	}
	// Register default matchers
	c.RegisterMatcher(NewEqualityMatcher())
	return c
}

// RegisterMatcher registers a matcher.
func (c *Comparator) RegisterMatcher(m Matcher) {
	c.matchers[m.Name()] = m
}

// GetMatcher returns a matcher by name.
func (c *Comparator) GetMatcher(name string) (Matcher, bool) {
	m, ok := c.matchers[name]
	return m, ok
}

// Compare compares expected and actual values using the specified matching rules.
func (c *Comparator) Compare(expected, actual interface{}, rules contract.MatchingRules) (*MatchResult, error) {
	// If no rules, use equality matching
	if len(rules.Body) == 0 && len(rules.Headers) == 0 && len(rules.Query) == 0 {
		return c.matchers["equality"].Match(expected, actual)
	}

	// For now, just do equality matching
	// TODO: implement path-based matching with rules
	return c.matchers["equality"].Match(expected, actual)
}

// CompareBody compares body values using the specified body matching rules.
func (c *Comparator) CompareBody(expected, actual interface{}, rules map[string]contract.MatcherSet) (*MatchResult, error) {
	if len(rules) == 0 {
		return c.matchers["equality"].Match(expected, actual)
	}

	// For now, use equality matching
	// TODO: implement path-based matching
	return c.matchers["equality"].Match(expected, actual)
}

// CompareHeaders compares headers using the specified header matching rules.
func (c *Comparator) CompareHeaders(expected, actual map[string]interface{}, rules map[string]contract.MatcherSet) (*MatchResult, error) {
	if expected == nil && actual == nil {
		return &MatchResult{Matched: true}, nil
	}

	// Check all expected headers are present
	for key, expVal := range expected {
		actVal, ok := actual[key]
		if !ok {
			return &MatchResult{
				Matched: false,
				Diff:    fmt.Sprintf("missing header: %s", key),
			}, nil
		}

		result, err := c.matchers["equality"].Match(expVal, actVal)
		if err != nil {
			return nil, err
		}
		if !result.Matched {
			return &MatchResult{
				Matched: false,
				Diff:    fmt.Sprintf("header %s: %s", key, result.Diff),
			}, nil
		}
	}

	return &MatchResult{Matched: true}, nil
}
