package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jt-chihara/yakusoku/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffCommand_Execute(t *testing.T) {
	t.Run("shows no differences for identical contracts", func(t *testing.T) {
		tmpDir := t.TempDir()
		contract1 := filepath.Join(tmpDir, "contract1.json")
		contract2 := filepath.Join(tmpDir, "contract2.json")
		createDiffContract(t, contract1, "Consumer", "Provider", []string{"get user"})
		createDiffContract(t, contract2, "Consumer", "Provider", []string{"get user"})

		var stdout, stderr bytes.Buffer
		cmd := cli.NewDiffCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--old", contract1, "--new", contract2})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "No differences")
	})

	t.Run("detects added interaction", func(t *testing.T) {
		tmpDir := t.TempDir()
		contract1 := filepath.Join(tmpDir, "contract1.json")
		contract2 := filepath.Join(tmpDir, "contract2.json")
		createDiffContract(t, contract1, "Consumer", "Provider", []string{"get user"})
		createDiffContract(t, contract2, "Consumer", "Provider", []string{"get user", "create user"})

		var stdout, stderr bytes.Buffer
		cmd := cli.NewDiffCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--old", contract1, "--new", contract2})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "added")
		assert.Contains(t, output, "create user")
	})

	t.Run("detects removed interaction", func(t *testing.T) {
		tmpDir := t.TempDir()
		contract1 := filepath.Join(tmpDir, "contract1.json")
		contract2 := filepath.Join(tmpDir, "contract2.json")
		createDiffContract(t, contract1, "Consumer", "Provider", []string{"get user", "delete user"})
		createDiffContract(t, contract2, "Consumer", "Provider", []string{"get user"})

		var stdout, stderr bytes.Buffer
		cmd := cli.NewDiffCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--old", contract1, "--new", contract2})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "removed")
		assert.Contains(t, output, "delete user")
	})

	t.Run("detects modified interaction", func(t *testing.T) {
		tmpDir := t.TempDir()
		contract1 := filepath.Join(tmpDir, "contract1.json")
		contract2 := filepath.Join(tmpDir, "contract2.json")
		createDiffContractWithStatus(t, contract1, "get user", 200)
		createDiffContractWithStatus(t, contract2, "get user", 201)

		var stdout, stderr bytes.Buffer
		cmd := cli.NewDiffCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--old", contract1, "--new", contract2})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "modified")
	})

	t.Run("returns error for non-existent old file", func(t *testing.T) {
		tmpDir := t.TempDir()
		contract2 := filepath.Join(tmpDir, "contract2.json")
		createDiffContract(t, contract2, "Consumer", "Provider", []string{"get user"})

		var stdout, stderr bytes.Buffer
		cmd := cli.NewDiffCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--old", "/nonexistent.json", "--new", contract2})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("returns error for non-existent new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		contract1 := filepath.Join(tmpDir, "contract1.json")
		createDiffContract(t, contract1, "Consumer", "Provider", []string{"get user"})

		var stdout, stderr bytes.Buffer
		cmd := cli.NewDiffCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--old", contract1, "--new", "/nonexistent.json"})

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("detects consumer name change", func(t *testing.T) {
		tmpDir := t.TempDir()
		contract1 := filepath.Join(tmpDir, "contract1.json")
		contract2 := filepath.Join(tmpDir, "contract2.json")
		createDiffContract(t, contract1, "OldConsumer", "Provider", []string{"get user"})
		createDiffContract(t, contract2, "NewConsumer", "Provider", []string{"get user"})

		var stdout, stderr bytes.Buffer
		cmd := cli.NewDiffCommand()
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--old", contract1, "--new", contract2})

		err := cmd.Execute()
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "consumer")
		assert.Contains(t, output, "OldConsumer")
		assert.Contains(t, output, "NewConsumer")
	})
}

func createDiffContract(t *testing.T, path, consumer, provider string, interactions []string) {
	interactionList := make([]interface{}, len(interactions))
	for i, desc := range interactions {
		interactionList[i] = map[string]interface{}{
			"description": desc,
			"request":     map[string]interface{}{"method": "GET", "path": "/test"},
			"response":    map[string]interface{}{"status": 200},
		}
	}

	contract := map[string]interface{}{
		"consumer":     map[string]interface{}{"name": consumer},
		"provider":     map[string]interface{}{"name": provider},
		"interactions": interactionList,
		"metadata": map[string]interface{}{
			"pactSpecification": map[string]interface{}{"version": "3.0.0"},
		},
	}
	data, _ := json.Marshal(contract)
	err := os.WriteFile(path, data, 0644)
	require.NoError(t, err)
}

func createDiffContractWithStatus(t *testing.T, path, description string, status int) {
	contract := map[string]interface{}{
		"consumer": map[string]interface{}{"name": "Consumer"},
		"provider": map[string]interface{}{"name": "Provider"},
		"interactions": []interface{}{
			map[string]interface{}{
				"description": description,
				"request":     map[string]interface{}{"method": "GET", "path": "/test"},
				"response":    map[string]interface{}{"status": status},
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
