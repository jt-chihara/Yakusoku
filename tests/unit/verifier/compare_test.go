package verifier_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/jt-chihara/yakusoku/internal/verifier"
)

func TestCompare_Status(t *testing.T) {
	t.Run("matching status codes", func(t *testing.T) {
		cmp := verifier.NewComparer()
		result := cmp.CompareStatus(200, 200)
		assert.True(t, result.Match)
	})

	t.Run("different status codes", func(t *testing.T) {
		cmp := verifier.NewComparer()
		result := cmp.CompareStatus(200, 404)
		assert.False(t, result.Match)
		assert.Contains(t, result.Diff, "expected status 200")
		assert.Contains(t, result.Diff, "got 404")
	})
}

func TestCompare_Headers(t *testing.T) {
	t.Run("matching headers", func(t *testing.T) {
		cmp := verifier.NewComparer()
		expected := map[string]interface{}{"Content-Type": "application/json"}
		actual := map[string]string{"Content-Type": "application/json"}
		result := cmp.CompareHeaders(expected, actual)
		assert.True(t, result.Match)
	})

	t.Run("missing header", func(t *testing.T) {
		cmp := verifier.NewComparer()
		expected := map[string]interface{}{"Content-Type": "application/json"}
		actual := map[string]string{}
		result := cmp.CompareHeaders(expected, actual)
		assert.False(t, result.Match)
		assert.Contains(t, result.Diff, "Content-Type")
	})

	t.Run("different header value", func(t *testing.T) {
		cmp := verifier.NewComparer()
		expected := map[string]interface{}{"Content-Type": "application/json"}
		actual := map[string]string{"Content-Type": "text/plain"}
		result := cmp.CompareHeaders(expected, actual)
		assert.False(t, result.Match)
	})

	t.Run("extra headers in actual are allowed", func(t *testing.T) {
		cmp := verifier.NewComparer()
		expected := map[string]interface{}{"Content-Type": "application/json"}
		actual := map[string]string{
			"Content-Type": "application/json",
			"X-Extra":      "ignored",
		}
		result := cmp.CompareHeaders(expected, actual)
		assert.True(t, result.Match)
	})

	t.Run("nil expected headers match any actual", func(t *testing.T) {
		cmp := verifier.NewComparer()
		actual := map[string]string{"Content-Type": "application/json"}
		result := cmp.CompareHeaders(nil, actual)
		assert.True(t, result.Match)
	})
}

func TestCompare_Body(t *testing.T) {
	t.Run("matching object bodies", func(t *testing.T) {
		cmp := verifier.NewComparer()
		expected := map[string]interface{}{"id": float64(1), "name": "John"}
		actual := map[string]interface{}{"id": float64(1), "name": "John"}
		result, err := cmp.CompareBody(expected, actual, nil)
		require.NoError(t, err)
		assert.True(t, result.Match)
	})

	t.Run("different body values", func(t *testing.T) {
		cmp := verifier.NewComparer()
		expected := map[string]interface{}{"id": float64(1), "name": "John"}
		actual := map[string]interface{}{"id": float64(1), "name": "Jane"}
		result, err := cmp.CompareBody(expected, actual, nil)
		require.NoError(t, err)
		assert.False(t, result.Match)
		assert.Contains(t, result.Diff, "name")
	})

	t.Run("missing field in actual", func(t *testing.T) {
		cmp := verifier.NewComparer()
		expected := map[string]interface{}{"id": float64(1), "name": "John"}
		actual := map[string]interface{}{"id": float64(1)}
		result, err := cmp.CompareBody(expected, actual, nil)
		require.NoError(t, err)
		assert.False(t, result.Match)
	})

	t.Run("matching array bodies", func(t *testing.T) {
		cmp := verifier.NewComparer()
		expected := []interface{}{
			map[string]interface{}{"id": float64(1)},
			map[string]interface{}{"id": float64(2)},
		}
		actual := []interface{}{
			map[string]interface{}{"id": float64(1)},
			map[string]interface{}{"id": float64(2)},
		}
		result, err := cmp.CompareBody(expected, actual, nil)
		require.NoError(t, err)
		assert.True(t, result.Match)
	})

	t.Run("nil expected body matches any actual", func(t *testing.T) {
		cmp := verifier.NewComparer()
		actual := map[string]interface{}{"id": float64(1)}
		result, err := cmp.CompareBody(nil, actual, nil)
		require.NoError(t, err)
		assert.True(t, result.Match)
	})

	t.Run("extra fields in actual are allowed", func(t *testing.T) {
		cmp := verifier.NewComparer()
		expected := map[string]interface{}{"id": float64(1)}
		actual := map[string]interface{}{"id": float64(1), "extra": "field"}
		result, err := cmp.CompareBody(expected, actual, nil)
		require.NoError(t, err)
		assert.True(t, result.Match)
	})
}
