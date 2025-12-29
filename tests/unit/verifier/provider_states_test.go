package verifier_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jt-chihara/yakusoku/internal/contract"
	"github.com/jt-chihara/yakusoku/internal/verifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderStates_Setup(t *testing.T) {
	t.Run("sends provider state to setup URL", func(t *testing.T) {
		var receivedState string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var state map[string]interface{}
			json.Unmarshal(body, &state)
			receivedState = state["state"].(string)
			w.WriteHeader(200)
		}))
		defer server.Close()

		ps := verifier.NewProviderStates(server.URL)
		err := ps.Setup("user 1 exists", nil)
		require.NoError(t, err)
		assert.Equal(t, "user 1 exists", receivedState)
	})

	t.Run("sends provider state with params", func(t *testing.T) {
		var receivedBody map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedBody)
			w.WriteHeader(200)
		}))
		defer server.Close()

		ps := verifier.NewProviderStates(server.URL)
		err := ps.Setup("user exists", map[string]interface{}{"userId": float64(1)})
		require.NoError(t, err)

		assert.Equal(t, "user exists", receivedBody["state"])
		params := receivedBody["params"].(map[string]interface{})
		assert.Equal(t, float64(1), params["userId"])
	})

	t.Run("handles setup failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			w.Write([]byte("Internal error"))
		}))
		defer server.Close()

		ps := verifier.NewProviderStates(server.URL)
		err := ps.Setup("test state", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("handles connection error", func(t *testing.T) {
		ps := verifier.NewProviderStates("http://localhost:99999")
		err := ps.Setup("test state", nil)
		require.Error(t, err)
	})

	t.Run("handles v3 provider states", func(t *testing.T) {
		var receivedStates []map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			json.Unmarshal(body, &req)
			receivedStates = append(receivedStates, req)
			w.WriteHeader(200)
		}))
		defer server.Close()

		ps := verifier.NewProviderStates(server.URL)
		states := []contract.ProviderState{
			{Name: "user exists", Params: map[string]interface{}{"userId": float64(1)}},
			{Name: "user is active"},
		}
		err := ps.SetupMultiple(states)
		require.NoError(t, err)
		assert.Len(t, receivedStates, 2)
	})

	t.Run("skips setup when URL is empty", func(t *testing.T) {
		ps := verifier.NewProviderStates("")
		err := ps.Setup("test state", nil)
		require.NoError(t, err) // Should not error, just skip
	})
}
