package cli_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/jt-chihara/yakusoku/internal/cli"
)

func TestVerifyCommand_Execute(t *testing.T) {
	t.Run("verifies contract against provider", func(t *testing.T) {
		// Create test provider
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/users/1" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				w.Write([]byte(`{"id":1,"name":"John"}`))
			}
		}))
		defer provider.Close()

		// Create test contract file
		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "consumer-provider.json")
		contract := map[string]interface{}{
			"consumer": map[string]interface{}{"name": "Consumer"},
			"provider": map[string]interface{}{"name": "Provider"},
			"interactions": []interface{}{
				map[string]interface{}{
					"description": "get user 1",
					"request":     map[string]interface{}{"method": "GET", "path": "/users/1"},
					"response": map[string]interface{}{
						"status": 200,
						"body":   map[string]interface{}{"id": 1, "name": "John"},
					},
				},
			},
			"metadata": map[string]interface{}{
				"pactSpecification": map[string]interface{}{"version": "3.0.0"},
			},
		}
		data, _ := json.Marshal(contract)
		os.WriteFile(contractPath, data, 0644)

		// Execute verify command
		var stdout, stderr bytes.Buffer
		cmd := cli.NewVerifyCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--provider-base-url", provider.URL,
			"--pact-file", contractPath,
		})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "get user 1")
		assert.Contains(t, output, "passed")
	})

	t.Run("returns error for missing pact file", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := cli.NewVerifyCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--provider-base-url", "http://localhost:8080",
			"--pact-file", "/nonexistent/file.json",
		})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("returns error for missing provider-base-url", func(t *testing.T) {
		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "test.json")
		os.WriteFile(contractPath, []byte("{}"), 0644)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewVerifyCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--pact-file", contractPath,
		})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("reports failure when verification fails", func(t *testing.T) {
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404) // Wrong status
		}))
		defer provider.Close()

		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "test.json")
		contract := map[string]interface{}{
			"consumer": map[string]interface{}{"name": "Consumer"},
			"provider": map[string]interface{}{"name": "Provider"},
			"interactions": []interface{}{
				map[string]interface{}{
					"description": "get user",
					"request":     map[string]interface{}{"method": "GET", "path": "/users/1"},
					"response":    map[string]interface{}{"status": 200},
				},
			},
			"metadata": map[string]interface{}{
				"pactSpecification": map[string]interface{}{"version": "3.0.0"},
			},
		}
		data, _ := json.Marshal(contract)
		os.WriteFile(contractPath, data, 0644)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewVerifyCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--provider-base-url", provider.URL,
			"--pact-file", contractPath,
		})

		err := cmd.Execute()
		// Command should succeed but report failure in output
		require.Error(t, err) // Exit code 1 for verification failure

		output := stdout.String()
		assert.Contains(t, output, "failed")
	})

	t.Run("uses provider states setup URL", func(t *testing.T) {
		statesCalled := false
		statesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			statesCalled = true
			w.WriteHeader(200)
		}))
		defer statesServer.Close()

		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte(`{"id":1}`))
		}))
		defer provider.Close()

		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "test.json")
		contract := map[string]interface{}{
			"consumer": map[string]interface{}{"name": "Consumer"},
			"provider": map[string]interface{}{"name": "Provider"},
			"interactions": []interface{}{
				map[string]interface{}{
					"description":   "get user",
					"providerState": "user exists",
					"request":       map[string]interface{}{"method": "GET", "path": "/users/1"},
					"response": map[string]interface{}{
						"status": 200,
						"body":   map[string]interface{}{"id": 1},
					},
				},
			},
			"metadata": map[string]interface{}{
				"pactSpecification": map[string]interface{}{"version": "3.0.0"},
			},
		}
		data, _ := json.Marshal(contract)
		os.WriteFile(contractPath, data, 0644)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewVerifyCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--provider-base-url", provider.URL,
			"--pact-file", contractPath,
			"--provider-states-setup-url", statesServer.URL,
		})

		cmd.Execute()
		assert.True(t, statesCalled)
	})

	t.Run("verbose flag shows detailed output", func(t *testing.T) {
		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte(`{"id":1}`))
		}))
		defer provider.Close()

		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "test.json")
		contract := map[string]interface{}{
			"consumer": map[string]interface{}{"name": "Consumer"},
			"provider": map[string]interface{}{"name": "Provider"},
			"interactions": []interface{}{
				map[string]interface{}{
					"description": "get user",
					"request":     map[string]interface{}{"method": "GET", "path": "/users/1"},
					"response":    map[string]interface{}{"status": 200, "body": map[string]interface{}{"id": 1}},
				},
			},
			"metadata": map[string]interface{}{
				"pactSpecification": map[string]interface{}{"version": "3.0.0"},
			},
		}
		data, _ := json.Marshal(contract)
		os.WriteFile(contractPath, data, 0644)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewVerifyCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--provider-base-url", provider.URL,
			"--pact-file", contractPath,
			"--verbose",
		})

		cmd.Execute()
		output := stdout.String()
		assert.Contains(t, output, "GET")
		assert.Contains(t, output, "/users/1")
	})
}
