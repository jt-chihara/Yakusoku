package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jt-chihara/yakusoku/internal/cli"
)

func TestShowCommand_Execute(t *testing.T) {
	t.Run("shows contract details", func(t *testing.T) {
		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "test.json")
		createDetailedContract(t, contractPath)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewShowCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-file", contractPath})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "OrderService")
		assert.Contains(t, output, "UserService")
		assert.Contains(t, output, "get user 1")
		assert.Contains(t, output, "GET")
		assert.Contains(t, output, "/users/1")
	})

	t.Run("shows multiple interactions", func(t *testing.T) {
		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "test.json")
		createContractWithMultipleInteractions(t, contractPath)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewShowCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-file", contractPath})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "get user 1")
		assert.Contains(t, output, "create user")
	})

	t.Run("outputs JSON format", func(t *testing.T) {
		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "test.json")
		createDetailedContract(t, contractPath)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewShowCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-file", contractPath, "--json"})

		err := cmd.Execute()
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &result)
		require.NoError(t, err)

		consumer := result["consumer"].(map[string]interface{})
		assert.Equal(t, "OrderService", consumer["name"])
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := cli.NewShowCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-file", "/nonexistent/file.json"})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "invalid.json")
		os.WriteFile(contractPath, []byte("not valid json"), 0644)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewShowCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-file", contractPath})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("shows provider state", func(t *testing.T) {
		tmpDir := t.TempDir()
		contractPath := filepath.Join(tmpDir, "test.json")
		createDetailedContract(t, contractPath)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewShowCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-file", contractPath})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "user 1 exists")
	})
}

func createDetailedContract(t *testing.T, path string) {
	contract := map[string]interface{}{
		"consumer": map[string]interface{}{"name": "OrderService"},
		"provider": map[string]interface{}{"name": "UserService"},
		"interactions": []interface{}{
			map[string]interface{}{
				"description":   "get user 1",
				"providerState": "user 1 exists",
				"request": map[string]interface{}{
					"method": "GET",
					"path":   "/users/1",
				},
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
	data, _ := json.MarshalIndent(contract, "", "  ")
	err := os.WriteFile(path, data, 0644)
	require.NoError(t, err)
}

func createContractWithMultipleInteractions(t *testing.T, path string) {
	contract := map[string]interface{}{
		"consumer": map[string]interface{}{"name": "Consumer"},
		"provider": map[string]interface{}{"name": "Provider"},
		"interactions": []interface{}{
			map[string]interface{}{
				"description": "get user 1",
				"request":     map[string]interface{}{"method": "GET", "path": "/users/1"},
				"response":    map[string]interface{}{"status": 200},
			},
			map[string]interface{}{
				"description": "create user",
				"request":     map[string]interface{}{"method": "POST", "path": "/users"},
				"response":    map[string]interface{}{"status": 201},
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
