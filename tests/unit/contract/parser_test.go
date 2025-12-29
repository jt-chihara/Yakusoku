package contract_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/jt-chihara/yakusoku/internal/contract"
)

func TestParser_ParseFile(t *testing.T) {
	t.Run("parse valid contract file", func(t *testing.T) {
		content := `{
			"consumer": {"name": "OrderService"},
			"provider": {"name": "UserService"},
			"interactions": [
				{
					"description": "a request for user 1",
					"providerState": "user 1 exists",
					"request": {
						"method": "GET",
						"path": "/users/1",
						"headers": {"Accept": "application/json"}
					},
					"response": {
						"status": 200,
						"headers": {"Content-Type": "application/json"},
						"body": {"id": 1, "name": "John Doe"}
					}
				}
			],
			"metadata": {
				"pactSpecification": {"version": "3.0.0"}
			}
		}`

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "orderservice-userservice.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		parser := contract.NewParser()
		c, err := parser.ParseFile(filePath)
		require.NoError(t, err)

		assert.Equal(t, "OrderService", c.Consumer.Name)
		assert.Equal(t, "UserService", c.Provider.Name)
		assert.Len(t, c.Interactions, 1)
		assert.Equal(t, "a request for user 1", c.Interactions[0].Description)
		assert.Equal(t, "3.0.0", c.Metadata.PactSpecification.Version)
	})

	t.Run("parse file not found", func(t *testing.T) {
		parser := contract.NewParser()
		_, err := parser.ParseFile("/nonexistent/path.json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read contract file")
	})

	t.Run("parse invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(filePath, []byte("not valid json"), 0644)
		require.NoError(t, err)

		parser := contract.NewParser()
		_, err = parser.ParseFile(filePath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse contract JSON")
	})

	t.Run("parse contract with matching rules", func(t *testing.T) {
		content := `{
			"consumer": {"name": "Consumer"},
			"provider": {"name": "Provider"},
			"interactions": [
				{
					"description": "test",
					"request": {"method": "GET", "path": "/test"},
					"response": {
						"status": 200,
						"body": {"id": 1, "email": "test@example.com"},
						"matchingRules": {
							"body": {
								"$.id": {
									"matchers": [{"match": "type"}]
								},
								"$.email": {
									"matchers": [{"match": "regex", "regex": "^[\\w.+-]+@[\\w.-]+\\.\\w+$"}]
								}
							}
						}
					}
				}
			],
			"metadata": {"pactSpecification": {"version": "3.0.0"}}
		}`

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "contract.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		parser := contract.NewParser()
		c, err := parser.ParseFile(filePath)
		require.NoError(t, err)

		mr := c.Interactions[0].Response.MatchingRules
		assert.Equal(t, "type", mr.Body["$.id"].Matchers[0].Match)
		assert.Equal(t, "regex", mr.Body["$.email"].Matchers[0].Match)
		assert.NotEmpty(t, mr.Body["$.email"].Matchers[0].Regex)
	})

	t.Run("parse contract with v3 provider states", func(t *testing.T) {
		content := `{
			"consumer": {"name": "Consumer"},
			"provider": {"name": "Provider"},
			"interactions": [
				{
					"description": "test",
					"providerStates": [
						{"name": "user exists", "params": {"userId": 1}},
						{"name": "user is active"}
					],
					"request": {"method": "GET", "path": "/users/1"},
					"response": {"status": 200}
				}
			],
			"metadata": {"pactSpecification": {"version": "3.0.0"}}
		}`

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "contract.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		parser := contract.NewParser()
		c, err := parser.ParseFile(filePath)
		require.NoError(t, err)

		assert.Len(t, c.Interactions[0].ProviderStates, 2)
		assert.Equal(t, "user exists", c.Interactions[0].ProviderStates[0].Name)
		assert.Equal(t, float64(1), c.Interactions[0].ProviderStates[0].Params["userId"])
		assert.Equal(t, "user is active", c.Interactions[0].ProviderStates[1].Name)
	})
}

func TestParser_ParseBytes(t *testing.T) {
	t.Run("parse valid bytes", func(t *testing.T) {
		content := []byte(`{
			"consumer": {"name": "Consumer"},
			"provider": {"name": "Provider"},
			"interactions": [
				{
					"description": "test",
					"request": {"method": "GET", "path": "/test"},
					"response": {"status": 200}
				}
			],
			"metadata": {"pactSpecification": {"version": "3.0.0"}}
		}`)

		parser := contract.NewParser()
		c, err := parser.ParseBytes(content)
		require.NoError(t, err)

		assert.Equal(t, "Consumer", c.Consumer.Name)
		assert.Equal(t, "Provider", c.Provider.Name)
	})

	t.Run("parse empty bytes", func(t *testing.T) {
		parser := contract.NewParser()
		_, err := parser.ParseBytes([]byte{})
		require.Error(t, err)
	})

	t.Run("parse nil bytes", func(t *testing.T) {
		parser := contract.NewParser()
		_, err := parser.ParseBytes(nil)
		require.Error(t, err)
	})
}
