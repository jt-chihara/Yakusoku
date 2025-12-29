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

func TestListCommand_Execute(t *testing.T) {
	t.Run("lists contract files in directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test contract files
		createTestContract(t, tmpDir, "consumer1-provider1.json")
		createTestContract(t, tmpDir, "consumer2-provider2.json")

		var stdout, stderr bytes.Buffer
		cmd := cli.NewListCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-dir", tmpDir})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "consumer1-provider1.json")
		assert.Contains(t, output, "consumer2-provider2.json")
	})

	t.Run("lists contracts with glob pattern", func(t *testing.T) {
		tmpDir := t.TempDir()

		createTestContract(t, tmpDir, "orderservice-userservice.json")
		createTestContract(t, tmpDir, "orderservice-paymentservice.json")
		createTestContract(t, tmpDir, "other-service.json")

		var stdout, stderr bytes.Buffer
		cmd := cli.NewListCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-dir", tmpDir, "--pattern", "orderservice-*.json"})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "orderservice-userservice.json")
		assert.Contains(t, output, "orderservice-paymentservice.json")
		assert.NotContains(t, output, "other-service.json")
	})

	t.Run("shows consumer and provider names", func(t *testing.T) {
		tmpDir := t.TempDir()
		createTestContractWithNames(t, tmpDir, "test.json", "OrderService", "UserService")

		var stdout, stderr bytes.Buffer
		cmd := cli.NewListCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-dir", tmpDir})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "OrderService")
		assert.Contains(t, output, "UserService")
	})

	t.Run("outputs JSON format", func(t *testing.T) {
		tmpDir := t.TempDir()
		createTestContractWithNames(t, tmpDir, "test.json", "Consumer", "Provider")

		var stdout, stderr bytes.Buffer
		cmd := cli.NewListCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-dir", tmpDir, "--json"})

		err := cmd.Execute()
		require.NoError(t, err)

		var result []map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Consumer", result[0]["consumer"])
		assert.Equal(t, "Provider", result[0]["provider"])
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		cmd := cli.NewListCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-dir", "/nonexistent/path"})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("shows message when no contracts found", func(t *testing.T) {
		tmpDir := t.TempDir()

		var stdout, stderr bytes.Buffer
		cmd := cli.NewListCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--pact-dir", tmpDir})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "No contracts found")
	})
}

func createTestContract(t *testing.T, dir, filename string) {
	createTestContractWithNames(t, dir, filename, "Consumer", "Provider")
}

func createTestContractWithNames(t *testing.T, dir, filename, consumer, provider string) {
	contract := map[string]interface{}{
		"consumer": map[string]interface{}{"name": consumer},
		"provider": map[string]interface{}{"name": provider},
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
	err := os.WriteFile(filepath.Join(dir, filename), data, 0644)
	require.NoError(t, err)
}
