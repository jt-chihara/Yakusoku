package verifier_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jt-chihara/yakusoku/internal/contract"
	"github.com/jt-chihara/yakusoku/internal/verifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifier_Verify(t *testing.T) {
	t.Run("verifies successful interaction", func(t *testing.T) {
		// Create a test provider server
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/users/1" && r.Method == "GET" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				w.Write([]byte(`{"id":1,"name":"John Doe"}`))
			}
		}))
		defer provider.Close()

		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Interactions: []contract.Interaction{
				{
					Description: "get user 1",
					Request: contract.Request{
						Method: "GET",
						Path:   "/users/1",
					},
					Response: contract.Response{
						Status: 200,
						Body: map[string]interface{}{
							"id":   float64(1),
							"name": "John Doe",
						},
					},
				},
			},
		}

		v := verifier.New(verifier.Config{
			ProviderBaseURL: provider.URL,
		})

		result, err := v.Verify(c)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Len(t, result.Interactions, 1)
		assert.True(t, result.Interactions[0].Success)
	})

	t.Run("detects response mismatch", func(t *testing.T) {
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte(`{"id":1,"name":"Jane Doe"}`)) // Different name
		}))
		defer provider.Close()

		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Interactions: []contract.Interaction{
				{
					Description: "get user 1",
					Request:     contract.Request{Method: "GET", Path: "/users/1"},
					Response: contract.Response{
						Status: 200,
						Body:   map[string]interface{}{"id": float64(1), "name": "John Doe"},
					},
				},
			},
		}

		v := verifier.New(verifier.Config{ProviderBaseURL: provider.URL})

		result, err := v.Verify(c)
		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.False(t, result.Interactions[0].Success)
		assert.NotEmpty(t, result.Interactions[0].Diff)
	})

	t.Run("detects status code mismatch", func(t *testing.T) {
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404) // Wrong status
		}))
		defer provider.Close()

		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Interactions: []contract.Interaction{
				{
					Description: "get user 1",
					Request:     contract.Request{Method: "GET", Path: "/users/1"},
					Response:    contract.Response{Status: 200},
				},
			},
		}

		v := verifier.New(verifier.Config{ProviderBaseURL: provider.URL})

		result, err := v.Verify(c)
		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Interactions[0].Diff, "status")
	})

	t.Run("verifies multiple interactions", func(t *testing.T) {
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/users/1":
				w.WriteHeader(200)
				w.Write([]byte(`{"id":1}`))
			case "/users/2":
				w.WriteHeader(200)
				w.Write([]byte(`{"id":2}`))
			}
		}))
		defer provider.Close()

		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Interactions: []contract.Interaction{
				{
					Description: "get user 1",
					Request:     contract.Request{Method: "GET", Path: "/users/1"},
					Response:    contract.Response{Status: 200, Body: map[string]interface{}{"id": float64(1)}},
				},
				{
					Description: "get user 2",
					Request:     contract.Request{Method: "GET", Path: "/users/2"},
					Response:    contract.Response{Status: 200, Body: map[string]interface{}{"id": float64(2)}},
				},
			},
		}

		v := verifier.New(verifier.Config{ProviderBaseURL: provider.URL})

		result, err := v.Verify(c)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Len(t, result.Interactions, 2)
	})

	t.Run("handles connection error", func(t *testing.T) {
		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Interactions: []contract.Interaction{
				{
					Description: "test",
					Request:     contract.Request{Method: "GET", Path: "/test"},
					Response:    contract.Response{Status: 200},
				},
			},
		}

		v := verifier.New(verifier.Config{ProviderBaseURL: "http://localhost:99999"})

		result, err := v.Verify(c)
		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Interactions[0].Error, "connection")
	})
}

func TestVerifier_ProviderStates(t *testing.T) {
	t.Run("calls provider states setup URL", func(t *testing.T) {
		statesCalled := make([]string, 0)
		statesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				statesCalled = append(statesCalled, r.URL.Path)
				w.WriteHeader(200)
			}
		}))
		defer statesServer.Close()

		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte(`{"id":1}`))
		}))
		defer provider.Close()

		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Interactions: []contract.Interaction{
				{
					Description:   "get user 1",
					ProviderState: "user 1 exists",
					Request:       contract.Request{Method: "GET", Path: "/users/1"},
					Response:      contract.Response{Status: 200, Body: map[string]interface{}{"id": float64(1)}},
				},
			},
		}

		v := verifier.New(verifier.Config{
			ProviderBaseURL:        provider.URL,
			ProviderStatesSetupURL: statesServer.URL + "/provider-states",
		})

		_, err := v.Verify(c)
		require.NoError(t, err)
		assert.Contains(t, statesCalled, "/provider-states")
	})
}
