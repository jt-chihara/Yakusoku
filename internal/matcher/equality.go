package matcher

import (
	"fmt"
	"reflect"
)

// EqualityMatcher performs exact equality matching.
type EqualityMatcher struct{}

// NewEqualityMatcher creates a new EqualityMatcher.
func NewEqualityMatcher() *EqualityMatcher {
	return &EqualityMatcher{}
}

// Name returns "equality".
func (m *EqualityMatcher) Name() string {
	return "equality"
}

// Match compares expected and actual values for exact equality.
func (m *EqualityMatcher) Match(expected, actual interface{}) (*MatchResult, error) {
	if deepEqual(expected, actual) {
		return &MatchResult{Matched: true}, nil
	}
	return &MatchResult{
		Matched: false,
		Diff:    fmt.Sprintf("expected %v (%T), got %v (%T)", expected, expected, actual, actual),
	}, nil
}

func deepEqual(expected, actual interface{}) bool {
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil || actual == nil {
		return false
	}

	expVal := reflect.ValueOf(expected)
	actVal := reflect.ValueOf(actual)

	if expVal.Kind() != actVal.Kind() {
		return false
	}

	switch expVal.Kind() {
	case reflect.Slice:
		return sliceEqual(expVal, actVal)
	case reflect.Map:
		return mapEqual(expVal, actVal)
	default:
		return reflect.DeepEqual(expected, actual)
	}
}

func sliceEqual(expected, actual reflect.Value) bool {
	if expected.Len() != actual.Len() {
		return false
	}
	for i := 0; i < expected.Len(); i++ {
		if !deepEqual(expected.Index(i).Interface(), actual.Index(i).Interface()) {
			return false
		}
	}
	return true
}

func mapEqual(expected, actual reflect.Value) bool {
	if expected.Len() != actual.Len() {
		return false
	}
	for _, key := range expected.MapKeys() {
		expElem := expected.MapIndex(key)
		actElem := actual.MapIndex(key)
		if !actElem.IsValid() {
			return false
		}
		if !deepEqual(expElem.Interface(), actElem.Interface()) {
			return false
		}
	}
	return true
}
