package matcher_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jt-chihara/yakusoku/internal/matcher"
)

func TestEqualityMatcher_Match(t *testing.T) {
	t.Run("equal strings match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match("hello", "hello")
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("different strings do not match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match("hello", "world")
		require.NoError(t, err)
		assert.False(t, result.Matched)
		assert.Contains(t, result.Diff, "expected")
	})

	t.Run("equal integers match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match(42, 42)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("different integers do not match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match(42, 43)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("equal floats match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match(3.14, 3.14)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("different floats do not match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match(3.14, 3.15)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("equal booleans match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match(true, true)
		require.NoError(t, err)
		assert.True(t, result.Matched)

		result, err = m.Match(false, false)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("different booleans do not match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match(true, false)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("null values match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match(nil, nil)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("null and non-null do not match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match(nil, "value")
		require.NoError(t, err)
		assert.False(t, result.Matched)

		result, err = m.Match("value", nil)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("equal arrays match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		expected := []interface{}{"a", "b", "c"}
		actual := []interface{}{"a", "b", "c"}

		result, err := m.Match(expected, actual)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("different arrays do not match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		expected := []interface{}{"a", "b", "c"}
		actual := []interface{}{"a", "b", "d"}

		result, err := m.Match(expected, actual)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("arrays with different lengths do not match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		expected := []interface{}{"a", "b"}
		actual := []interface{}{"a", "b", "c"}

		result, err := m.Match(expected, actual)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("equal maps match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		expected := map[string]interface{}{"id": float64(1), "name": "John"}
		actual := map[string]interface{}{"id": float64(1), "name": "John"}

		result, err := m.Match(expected, actual)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("maps with different values do not match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		expected := map[string]interface{}{"id": float64(1), "name": "John"}
		actual := map[string]interface{}{"id": float64(1), "name": "Jane"}

		result, err := m.Match(expected, actual)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("maps with different keys do not match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		expected := map[string]interface{}{"id": float64(1)}
		actual := map[string]interface{}{"id": float64(1), "name": "John"}

		result, err := m.Match(expected, actual)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("nested structures match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

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

		result, err := m.Match(expected, actual)
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("nested structures with differences do not match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		expected := map[string]interface{}{
			"user": map[string]interface{}{
				"id":   float64(1),
				"name": "John",
			},
		}
		actual := map[string]interface{}{
			"user": map[string]interface{}{
				"id":   float64(2),
				"name": "John",
			},
		}

		result, err := m.Match(expected, actual)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("different types do not match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match("42", 42)
		require.NoError(t, err)
		assert.False(t, result.Matched)
	})

	t.Run("empty strings match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match("", "")
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("empty arrays match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match([]interface{}{}, []interface{}{})
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})

	t.Run("empty maps match", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()

		result, err := m.Match(map[string]interface{}{}, map[string]interface{}{})
		require.NoError(t, err)
		assert.True(t, result.Matched)
	})
}

func TestEqualityMatcher_Name(t *testing.T) {
	t.Run("returns equality", func(t *testing.T) {
		m := matcher.NewEqualityMatcher()
		assert.Equal(t, "equality", m.Name())
	})
}
