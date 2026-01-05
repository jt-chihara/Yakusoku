package matcher_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jt-chihara/yakusoku/internal/contract"
	"github.com/jt-chihara/yakusoku/internal/matcher"
)

func TestNewComparator(t *testing.T) {
	t.Run("creates comparator with default matchers", func(t *testing.T) {
		c := matcher.NewComparator()
		require.NotNil(t, c)

		// Should have equality matcher registered by default
		m, ok := c.GetMatcher("equality")
		assert.True(t, ok)
		assert.NotNil(t, m)
		assert.Equal(t, "equality", m.Name())
	})
}

func TestComparator_RegisterMatcher(t *testing.T) {
	t.Run("registers custom matcher", func(t *testing.T) {
		c := matcher.NewComparator()

		// Create and register a custom matcher
		custom := matcher.NewEqualityMatcher()
		c.RegisterMatcher(custom)

		m, ok := c.GetMatcher("equality")
		assert.True(t, ok)
		assert.NotNil(t, m)
	})
}

func TestComparator_GetMatcher(t *testing.T) {
	t.Run("returns matcher when exists", func(t *testing.T) {
		c := matcher.NewComparator()

		m, ok := c.GetMatcher("equality")
		assert.True(t, ok)
		assert.NotNil(t, m)
	})

	t.Run("returns false when matcher does not exist", func(t *testing.T) {
		c := matcher.NewComparator()

		m, ok := c.GetMatcher("nonexistent")
		assert.False(t, ok)
		assert.Nil(t, m)
	})
}

func TestComparator_Compare(t *testing.T) {
	t.Run("uses equality matching when no rules", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := contract.MatchingRules{}

		result, err := c.Compare("hello", "hello", rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("returns mismatch when values differ with no rules", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := contract.MatchingRules{}

		result, err := c.Compare("hello", "world", rules)
		require.NoError(t, err)
		assert.False(t, result.Matched)
		assert.Contains(t, result.Diff, "expected")
	})

	t.Run("uses equality matching with body rules (fallback behavior)", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := contract.MatchingRules{
			Body: map[string]contract.MatcherSet{
				"$.name": {
					Matchers: []contract.Matcher{
						{Match: "type"},
					},
				},
			},
		}

		// Even with rules, current implementation falls back to equality
		result, err := c.Compare("test", "test", rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("compares complex objects", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := contract.MatchingRules{}

		expected := map[string]interface{}{
			"id":   float64(1),
			"name": "John",
		}
		actual := map[string]interface{}{
			"id":   float64(1),
			"name": "John",
		}

		result, err := c.Compare(expected, actual, rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("detects differences in complex objects", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := contract.MatchingRules{}

		expected := map[string]interface{}{
			"id":   float64(1),
			"name": "John",
		}
		actual := map[string]interface{}{
			"id":   float64(1),
			"name": "Jane",
		}

		result, err := c.Compare(expected, actual, rules)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})
}

func TestComparator_CompareBody(t *testing.T) {
	t.Run("uses equality matching when no rules", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		result, err := c.CompareBody("hello", "hello", rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("returns mismatch when values differ", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		result, err := c.CompareBody("hello", "world", rules)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("uses equality matching with rules (fallback behavior)", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{
			"$.name": {
				Matchers: []contract.Matcher{
					{Match: "type"},
				},
			},
		}

		result, err := c.CompareBody("test", "test", rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("compares complex body objects", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		expected := map[string]interface{}{
			"user": map[string]interface{}{
				"id":   float64(1),
				"name": "John",
			},
		}
		actual := map[string]interface{}{
			"user": map[string]interface{}{
				"id":   float64(1),
				"name": "John",
			},
		}

		result, err := c.CompareBody(expected, actual, rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("compares arrays in body", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		expected := []interface{}{"a", "b", "c"}
		actual := []interface{}{"a", "b", "c"}

		result, err := c.CompareBody(expected, actual, rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("detects array differences", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		expected := []interface{}{"a", "b", "c"}
		actual := []interface{}{"a", "b", "d"}

		result, err := c.CompareBody(expected, actual, rules)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("handles nil bodies", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		result, err := c.CompareBody(nil, nil, rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})
}

func TestComparator_CompareHeaders(t *testing.T) {
	t.Run("both nil headers match", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		result, err := c.CompareHeaders(nil, nil, rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("matching headers pass", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		expected := map[string]interface{}{
			"Content-Type": "application/json",
		}
		actual := map[string]interface{}{
			"Content-Type": "application/json",
		}

		result, err := c.CompareHeaders(expected, actual, rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("missing header fails", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		expected := map[string]interface{}{
			"Content-Type": "application/json",
		}
		actual := map[string]interface{}{}

		result, err := c.CompareHeaders(expected, actual, rules)
		require.NoError(t, err)
		assert.False(t, result.Matched)
		assert.Contains(t, result.Diff, "missing header")
		assert.Contains(t, result.Diff, "Content-Type")
	})

	t.Run("different header value fails", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		expected := map[string]interface{}{
			"Content-Type": "application/json",
		}
		actual := map[string]interface{}{
			"Content-Type": "text/plain",
		}

		result, err := c.CompareHeaders(expected, actual, rules)
		require.NoError(t, err)
		assert.False(t, result.Matched)
		assert.Contains(t, result.Diff, "Content-Type")
	})

	t.Run("extra headers in actual are allowed", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		expected := map[string]interface{}{
			"Content-Type": "application/json",
		}
		actual := map[string]interface{}{
			"Content-Type":   "application/json",
			"X-Extra-Header": "extra-value",
		}

		result, err := c.CompareHeaders(expected, actual, rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("multiple expected headers", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		expected := map[string]interface{}{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token",
		}
		actual := map[string]interface{}{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token",
		}

		result, err := c.CompareHeaders(expected, actual, rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("one of multiple headers mismatch fails", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		expected := map[string]interface{}{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token",
		}
		actual := map[string]interface{}{
			"Content-Type":  "application/json",
			"Authorization": "Bearer wrong-token",
		}

		result, err := c.CompareHeaders(expected, actual, rules)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("empty expected headers always match", func(t *testing.T) {
		c := matcher.NewComparator()
		rules := map[string]contract.MatcherSet{}

		expected := map[string]interface{}{}
		actual := map[string]interface{}{
			"Content-Type": "application/json",
		}

		result, err := c.CompareHeaders(expected, actual, rules)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})
}
