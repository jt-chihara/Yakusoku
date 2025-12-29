package verifier

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jt-chihara/yakusoku/internal/contract"
)

// CompareResult holds the result of a comparison.
type CompareResult struct {
	Match bool
	Diff  string
}

// Comparer compares expected and actual values.
type Comparer struct{}

// NewComparer creates a new Comparer.
func NewComparer() *Comparer {
	return &Comparer{}
}

// CompareStatus compares status codes.
func (c *Comparer) CompareStatus(expected, actual int) CompareResult {
	if expected == actual {
		return CompareResult{Match: true}
	}
	return CompareResult{
		Match: false,
		Diff:  fmt.Sprintf("expected status %d, got %d", expected, actual),
	}
}

// CompareHeaders compares headers.
func (c *Comparer) CompareHeaders(expected map[string]interface{}, actual map[string]string) CompareResult {
	if expected == nil {
		return CompareResult{Match: true}
	}

	var diffs []string
	for key, expVal := range expected {
		actVal, ok := actual[key]
		if !ok {
			diffs = append(diffs, fmt.Sprintf("missing header: %s", key))
			continue
		}
		if fmt.Sprintf("%v", expVal) != actVal {
			diffs = append(diffs, fmt.Sprintf("header %s: expected %v, got %s", key, expVal, actVal))
		}
	}

	if len(diffs) > 0 {
		return CompareResult{Match: false, Diff: strings.Join(diffs, "; ")}
	}
	return CompareResult{Match: true}
}

// CompareBody compares body content.
func (c *Comparer) CompareBody(expected, actual interface{}, rules map[string]contract.MatcherSet) (*CompareResult, error) {
	if expected == nil {
		return &CompareResult{Match: true}, nil
	}

	diffs := c.compareValues("$", expected, actual)
	if len(diffs) > 0 {
		return &CompareResult{Match: false, Diff: strings.Join(diffs, "; ")}, nil
	}
	return &CompareResult{Match: true}, nil
}

func (c *Comparer) compareValues(path string, expected, actual interface{}) []string {
	if expected == nil {
		return nil
	}

	expVal := reflect.ValueOf(expected)
	actVal := reflect.ValueOf(actual)

	switch expVal.Kind() {
	case reflect.Map:
		return c.compareMaps(path, expVal, actVal)
	case reflect.Slice:
		return c.compareSlices(path, expVal, actVal)
	default:
		if !reflect.DeepEqual(expected, actual) {
			return []string{fmt.Sprintf("%s: expected %v, got %v", path, expected, actual)}
		}
		return nil
	}
}

func (c *Comparer) compareMaps(path string, expected, actual reflect.Value) []string {
	if actual.Kind() != reflect.Map {
		return []string{fmt.Sprintf("%s: expected object, got %v", path, actual.Kind())}
	}

	var diffs []string
	for _, key := range expected.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		expElem := expected.MapIndex(key)
		actElem := actual.MapIndex(key)

		if !actElem.IsValid() {
			diffs = append(diffs, fmt.Sprintf("%s.%s: missing field", path, keyStr))
			continue
		}

		childDiffs := c.compareValues(path+"."+keyStr, expElem.Interface(), actElem.Interface())
		diffs = append(diffs, childDiffs...)
	}
	return diffs
}

func (c *Comparer) compareSlices(path string, expected, actual reflect.Value) []string {
	if actual.Kind() != reflect.Slice {
		return []string{fmt.Sprintf("%s: expected array, got %v", path, actual.Kind())}
	}

	if expected.Len() != actual.Len() {
		return []string{fmt.Sprintf("%s: expected array length %d, got %d", path, expected.Len(), actual.Len())}
	}

	var diffs []string
	for i := 0; i < expected.Len(); i++ {
		childPath := fmt.Sprintf("%s[%d]", path, i)
		childDiffs := c.compareValues(childPath, expected.Index(i).Interface(), actual.Index(i).Interface())
		diffs = append(diffs, childDiffs...)
	}
	return diffs
}
