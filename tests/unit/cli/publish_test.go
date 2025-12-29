package cli_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jt-chihara/yakusoku/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishCommand_Execute(t *testing.T) {
	t.Run("publishes contract to broker", func(t *testing.T) {
		var receivedContract map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPut || r.Method == http.MethodPost {
				json.NewDecoder(r.Body).Decode(&receivedContract)
				w.WriteHeader(http.StatusCreated)
			}
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "contract.json")
		createPublishContract(t, contractPath)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewPublishCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--broker-url", server.URL,
			"--pact-file", contractPath,
			"--consumer-version", "1.0.0",
		})

		err := cmd.Execute()
		require.NoError(t, err)

		assert.NotNil(t, receivedContract)
		assert.Contains(t, stdout.String(), "published")
	})

	t.Run("publishes multiple contracts from directory", func(t *testing.T) {
		publishCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPut || r.Method == http.MethodPost {
				publishCount++
				w.WriteHeader(http.StatusCreated)
			}
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		createPublishContract(t, filepath.Join(tmpDir, "contract1.json"))
		createPublishContract(t, filepath.Join(tmpDir, "contract2.json"))

		var stdout, stderr bytes.Buffer
		cmd := cli.NewPublishCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--broker-url", server.URL,
			"--pact-dir", tmpDir,
			"--consumer-version", "1.0.0",
		})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Equal(t, 2, publishCount)
	})

	t.Run("returns error for missing broker URL", func(t *testing.T) {
		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "contract.json")
		createPublishContract(t, contractPath)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewPublishCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--pact-file", contractPath,
			"--consumer-version", "1.0.0",
		})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("returns error for missing version", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "contract.json")
		createPublishContract(t, contractPath)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewPublishCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--broker-url", server.URL,
			"--pact-file", contractPath,
		})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("returns error when broker returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "contract.json")
		createPublishContract(t, contractPath)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewPublishCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--broker-url", server.URL,
			"--pact-file", contractPath,
			"--consumer-version", "1.0.0",
		})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("adds tags to published contract", func(t *testing.T) {
		var requestPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestPath = r.URL.Path
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "contract.json")
		createPublishContract(t, contractPath)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewPublishCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--broker-url", server.URL,
			"--pact-file", contractPath,
			"--consumer-version", "1.0.0",
			"--tag", "main",
		})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.NotEmpty(t, requestPath)
	})
}

func createPublishContract(t *testing.T, path string) {
	contract := map[string]interface{}{
		"consumer": map[string]interface{}{"name": "Consumer"},
		"provider": map[string]interface{}{"name": "Provider"},
		"interactions": []interface{}{
			map[string]interface{}{
				"description": "test",
				"request":     map[string]interface{}{"method": "GET", "path": "/test"},
				"response":    map[string]interface{}{"status": 200},
			},
		},
		"metadata": map[string]interface{}{
			"pactSpecification": map[string]interface{}{"version": "3.0.0"},
		},
	}
	data, _ := json.Marshal(contract)
	err := os.WriteFile(path, data, 0644)
	require.NoError(t, err)
}
