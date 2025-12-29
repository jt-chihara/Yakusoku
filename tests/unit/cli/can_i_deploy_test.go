package cli_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jt-chihara/yakusoku/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanIDeployCommand_Execute(t *testing.T) {
	t.Run("returns success when deployable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"deployable": true,
				"summary": map[string]interface{}{
					"deployable": true,
					"reason":     "All required verification results are published and successful",
				},
			})
		}))
		defer server.Close()

		var stdout, stderr bytes.Buffer
		cmd := cli.NewCanIDeployCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--broker-url", server.URL,
			"--pacticipant", "Consumer",
			"--version", "1.0.0",
		})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "can be deployed")
	})

	t.Run("returns error when not deployable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"deployable": false,
				"summary": map[string]interface{}{
					"deployable": false,
					"reason":     "Verification failed",
				},
			})
		}))
		defer server.Close()

		var stdout, stderr bytes.Buffer
		cmd := cli.NewCanIDeployCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--broker-url", server.URL,
			"--pacticipant", "Consumer",
			"--version", "1.0.0",
		})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("checks deployment to specific environment", func(t *testing.T) {
		var requestedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestedPath = r.URL.RawQuery
			json.NewEncoder(w).Encode(map[string]interface{}{
				"deployable": true,
			})
		}))
		defer server.Close()

		var stdout, stderr bytes.Buffer
		cmd := cli.NewCanIDeployCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--broker-url", server.URL,
			"--pacticipant", "Consumer",
			"--version", "1.0.0",
			"--to-environment", "production",
		})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, requestedPath, "environment=production")
	})

	t.Run("returns error for missing broker URL", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := cli.NewCanIDeployCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--pacticipant", "Consumer",
			"--version", "1.0.0",
		})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("returns error for missing pacticipant", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		var stdout, stderr bytes.Buffer
		cmd := cli.NewCanIDeployCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--broker-url", server.URL,
			"--version", "1.0.0",
		})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("outputs JSON format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"deployable": true,
				"summary": map[string]interface{}{
					"deployable": true,
					"reason":     "Success",
				},
			})
		}))
		defer server.Close()

		var stdout, stderr bytes.Buffer
		cmd := cli.NewCanIDeployCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--broker-url", server.URL,
			"--pacticipant", "Consumer",
			"--version", "1.0.0",
			"--json",
		})

		err := cmd.Execute()
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, true, result["deployable"])
	})

	t.Run("uses latest version when --latest flag is set", func(t *testing.T) {
		var requestedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestedPath = r.URL.RawQuery
			json.NewEncoder(w).Encode(map[string]interface{}{
				"deployable": true,
			})
		}))
		defer server.Close()

		var stdout, stderr bytes.Buffer
		cmd := cli.NewCanIDeployCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{
			"--broker-url", server.URL,
			"--pacticipant", "Consumer",
			"--latest",
		})

		err := cmd.Execute()
		require.NoError(t, err)
		assert.Contains(t, requestedPath, "latest=true")
	})
}
