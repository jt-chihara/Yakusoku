package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jt-chihara/yakusoku/internal/contract"
	"github.com/jt-chihara/yakusoku/internal/verifier"
)

func TestVerifyE2E_SuccessfulVerification(t *testing.T) {
	// Create a provider that implements the expected contract
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/users/1" && r.Method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    1,
				"name":  "John Doe",
				"email": "john@example.com",
			})
		case r.URL.Path == "/users" && r.Method == "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   2,
				"name": "Jane Doe",
			})
		default:
			w.WriteHeader(404)
		}
	}))
	defer provider.Close()

	// Create a contract with multiple interactions
	c := contract.Contract{
		Consumer: contract.Pacticipant{Name: "OrderService"},
		Provider: contract.Pacticipant{Name: "UserService"},
		Interactions: []contract.Interaction{
			{
				Description:   "a request for user 1",
				ProviderState: "user 1 exists",
				Request: contract.Request{
					Method: "GET",
					Path:   "/users/1",
				},
				Response: contract.Response{
					Status: 200,
					Headers: map[string]interface{}{
						"Content-Type": "application/json",
					},
					Body: map[string]interface{}{
						"id":    float64(1),
						"name":  "John Doe",
						"email": "john@example.com",
					},
				},
			},
			{
				Description: "create a new user",
				Request: contract.Request{
					Method: "POST",
					Path:   "/users",
				},
				Response: contract.Response{
					Status: 201,
					Body: map[string]interface{}{
						"id":   float64(2),
						"name": "Jane Doe",
					},
				},
			},
		},
		Metadata: contract.Metadata{
			PactSpecification: contract.PactSpec{Version: "3.0.0"},
		},
	}

	v := verifier.New(verifier.Config{
		ProviderBaseURL: provider.URL,
	})

	result, err := v.Verify(&c)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Len(t, result.Interactions, 2)

	for _, ir := range result.Interactions {
		assert.True(t, ir.Success, "interaction %s should pass", ir.Description)
	}
}

func TestVerifyE2E_FailedVerification(t *testing.T) {
	// Provider returns wrong data
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   1,
			"name": "Wrong Name", // Different from expected
		})
	}))
	defer provider.Close()

	c := contract.Contract{
		Consumer: contract.Pacticipant{Name: "Consumer"},
		Provider: contract.Pacticipant{Name: "Provider"},
		Interactions: []contract.Interaction{
			{
				Description: "get user",
				Request:     contract.Request{Method: "GET", Path: "/users/1"},
				Response: contract.Response{
					Status: 200,
					Body:   map[string]interface{}{"id": float64(1), "name": "John Doe"},
				},
			},
		},
	}

	v := verifier.New(verifier.Config{ProviderBaseURL: provider.URL})

	result, err := v.Verify(&c)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.False(t, result.Interactions[0].Success)
	assert.NotEmpty(t, result.Interactions[0].Diff)
}

func TestVerifyE2E_WithProviderStates(t *testing.T) {
	statesReceived := make([]string, 0)

	// Provider states setup endpoint
	statesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if state, ok := body["state"].(string); ok {
			statesReceived = append(statesReceived, state)
		}
		w.WriteHeader(200)
	}))
	defer statesServer.Close()

	// Provider server
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 1})
	}))
	defer provider.Close()

	c := contract.Contract{
		Consumer: contract.Pacticipant{Name: "Consumer"},
		Provider: contract.Pacticipant{Name: "Provider"},
		Interactions: []contract.Interaction{
			{
				Description:   "get user when exists",
				ProviderState: "user 1 exists",
				Request:       contract.Request{Method: "GET", Path: "/users/1"},
				Response:      contract.Response{Status: 200, Body: map[string]interface{}{"id": float64(1)}},
			},
			{
				Description:   "get user when active",
				ProviderState: "user 1 is active",
				Request:       contract.Request{Method: "GET", Path: "/users/1"},
				Response:      contract.Response{Status: 200, Body: map[string]interface{}{"id": float64(1)}},
			},
		},
	}

	v := verifier.New(verifier.Config{
		ProviderBaseURL:        provider.URL,
		ProviderStatesSetupURL: statesServer.URL,
	})

	result, err := v.Verify(&c)
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Check that provider states were called
	assert.Contains(t, statesReceived, "user 1 exists")
	assert.Contains(t, statesReceived, "user 1 is active")
}

func TestVerifyE2E_FromContractFile(t *testing.T) {
	// Create provider
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer provider.Close()

	// Write contract to file
	tmpDir := t.TempDir()
	contractPath := filepath.Join(tmpDir, "consumer-provider.json")

	c := contract.Contract{
		Consumer: contract.Pacticipant{Name: "Consumer"},
		Provider: contract.Pacticipant{Name: "Provider"},
		Interactions: []contract.Interaction{
			{
				Description: "health check",
				Request:     contract.Request{Method: "GET", Path: "/health"},
				Response:    contract.Response{Status: 200, Body: map[string]interface{}{"status": "ok"}},
			},
		},
		Metadata: contract.Metadata{PactSpecification: contract.PactSpec{Version: "3.0.0"}},
	}

	writer := contract.NewWriter()
	err := writer.Write(&c, contractPath)
	require.NoError(t, err)

	// Parse and verify
	parser := contract.NewParser()
	parsed, err := parser.ParseFile(contractPath)
	require.NoError(t, err)

	v := verifier.New(verifier.Config{ProviderBaseURL: provider.URL})
	result, err := v.Verify(parsed)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestVerifyE2E_PartialFailure(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(200)
		case "/failure":
			w.WriteHeader(500) // Wrong status
		}
	}))
	defer provider.Close()

	c := contract.Contract{
		Consumer: contract.Pacticipant{Name: "Consumer"},
		Provider: contract.Pacticipant{Name: "Provider"},
		Interactions: []contract.Interaction{
			{
				Description: "success endpoint",
				Request:     contract.Request{Method: "GET", Path: "/success"},
				Response:    contract.Response{Status: 200},
			},
			{
				Description: "failure endpoint",
				Request:     contract.Request{Method: "GET", Path: "/failure"},
				Response:    contract.Response{Status: 200}, // Expected 200, but gets 500
			},
		},
	}

	v := verifier.New(verifier.Config{ProviderBaseURL: provider.URL})

	result, err := v.Verify(&c)
	require.NoError(t, err)
	assert.False(t, result.Success)

	// First should pass, second should fail
	assert.True(t, result.Interactions[0].Success)
	assert.False(t, result.Interactions[1].Success)
}
