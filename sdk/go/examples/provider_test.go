package examples

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jt-chihara/yakusoku/internal/contract"
	"github.com/jt-chihara/yakusoku/internal/verifier"
)

// This file demonstrates how to verify a provider against consumer contracts.
//
// In a real scenario, you would:
// 1. Run consumer tests to generate contract files
// 2. Use `yakusoku verify` CLI or this code to verify your provider

// createSampleContract creates a sample contract for testing
func createSampleContract(t *testing.T, dir string) string {
	c := contract.Contract{
		Consumer: contract.Pacticipant{Name: "OrderService"},
		Provider: contract.Pacticipant{Name: "UserService"},
		Interactions: []contract.Interaction{
			{
				Description:   "a request to get user 1",
				ProviderState: "user 1 exists",
				Request: contract.Request{
					Method: "GET",
					Path:   "/users/1",
				},
				Response: contract.Response{
					Status: 200,
					Body: map[string]interface{}{
						"id":   1,
						"name": "John Doe",
					},
				},
			},
		},
		Metadata: contract.Metadata{
			PactSpecification: contract.PactSpec{Version: "3.0.0"},
		},
	}

	path := filepath.Join(dir, "orderservice-userservice.json")
	data, _ := json.MarshalIndent(c, "", "  ")
	os.WriteFile(path, data, 0644)
	return path
}

// TestProviderVerification demonstrates programmatic provider verification.
func TestProviderVerification(t *testing.T) {
	// 1. Create a mock provider (in real usage, this would be your actual provider)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/1":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   1,
				"name": "John Doe",
			})
		case "/provider-states":
			// Handle provider state setup
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer provider.Close()

	// 2. Create a sample contract
	tmpDir := t.TempDir()
	contractPath := createSampleContract(t, tmpDir)

	// 3. Parse the contract
	parser := contract.NewParser()
	c, err := parser.ParseFile(contractPath)
	if err != nil {
		t.Fatalf("failed to parse contract: %v", err)
	}

	// 4. Create verifier and verify
	v := verifier.New(verifier.Config{
		ProviderBaseURL: provider.URL,
	})

	result, err := v.Verify(*c)
	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}

	// 5. Check results
	if !result.Success {
		t.Errorf("expected verification to succeed")
		for _, ir := range result.Interactions {
			if !ir.Success {
				t.Logf("Failed interaction: %s", ir.Description)
				if ir.Diff != "" {
					t.Logf("  Diff: %s", ir.Diff)
				}
				if ir.Error != "" {
					t.Logf("  Error: %s", ir.Error)
				}
			}
		}
	}
}

// TestProviderWithStates demonstrates verification with provider states.
func TestProviderWithStates(t *testing.T) {
	var setupCalled bool

	// Provider with state setup endpoint
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/provider-states":
			setupCalled = true
			var state struct {
				State string `json:"state"`
			}
			json.NewDecoder(r.Body).Decode(&state)
			t.Logf("Provider state setup called: %s", state.State)
			w.WriteHeader(http.StatusOK)
		case "/users/1":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   1,
				"name": "John Doe",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer provider.Close()

	// Create contract
	tmpDir := t.TempDir()
	contractPath := createSampleContract(t, tmpDir)

	parser := contract.NewParser()
	c, _ := parser.ParseFile(contractPath)

	// Verify with provider states URL
	v := verifier.New(verifier.Config{
		ProviderBaseURL:        provider.URL,
		ProviderStatesSetupURL: provider.URL + "/provider-states",
	})

	result, err := v.Verify(*c)
	if err != nil {
		t.Fatalf("verification failed: %v", err)
	}

	if !setupCalled {
		t.Error("expected provider state setup to be called")
	}

	if !result.Success {
		t.Error("expected verification to succeed")
	}
}
