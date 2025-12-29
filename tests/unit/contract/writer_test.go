package contract_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jt-chihara/yakusoku/internal/contract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter_Write(t *testing.T) {
	t.Run("writes contract to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test-contract.json")

		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Interactions: []contract.Interaction{
				{
					Description: "test interaction",
					Request: contract.Request{
						Method: "GET",
						Path:   "/test",
					},
					Response: contract.Response{
						Status: 200,
					},
				},
			},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "3.0.0"},
			},
		}

		writer := contract.NewWriter()
		err := writer.Write(&c, filePath)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(filePath)
		require.NoError(t, err)

		// Verify content
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)

		var result contract.Contract
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "Consumer", result.Consumer.Name)
		assert.Equal(t, "Provider", result.Provider.Name)
		assert.Len(t, result.Interactions, 1)
	})

	t.Run("creates directory if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "subdir", "test-contract.json")

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
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "3.0.0"},
			},
		}

		writer := contract.NewWriter()
		err := writer.Write(&c, filePath)
		require.NoError(t, err)

		_, err = os.Stat(filePath)
		require.NoError(t, err)
	})

	t.Run("writes pretty-printed JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test-contract.json")

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
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "3.0.0"},
			},
		}

		writer := contract.NewWriter()
		err := writer.Write(&c, filePath)
		require.NoError(t, err)

		data, err := os.ReadFile(filePath)
		require.NoError(t, err)

		// Pretty-printed JSON should have newlines
		assert.Contains(t, string(data), "\n")
	})

	t.Run("generates filename from consumer and provider", func(t *testing.T) {
		tmpDir := t.TempDir()

		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "OrderService"},
			Provider: contract.Pacticipant{Name: "UserService"},
			Interactions: []contract.Interaction{
				{
					Description: "test",
					Request:     contract.Request{Method: "GET", Path: "/test"},
					Response:    contract.Response{Status: 200},
				},
			},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "3.0.0"},
			},
		}

		writer := contract.NewWriter()
		filePath, err := writer.WriteToDir(&c, tmpDir)
		require.NoError(t, err)

		assert.Equal(t, filepath.Join(tmpDir, "orderservice-userservice.json"), filePath)

		_, err = os.Stat(filePath)
		require.NoError(t, err)
	})

	t.Run("handles special characters in names", func(t *testing.T) {
		tmpDir := t.TempDir()

		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "Order Service"},
			Provider: contract.Pacticipant{Name: "User_Service"},
			Interactions: []contract.Interaction{
				{
					Description: "test",
					Request:     contract.Request{Method: "GET", Path: "/test"},
					Response:    contract.Response{Status: 200},
				},
			},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "3.0.0"},
			},
		}

		writer := contract.NewWriter()
		filePath, err := writer.WriteToDir(&c, tmpDir)
		require.NoError(t, err)

		// Names should be lowercased and spaces replaced
		assert.Equal(t, filepath.Join(tmpDir, "order_service-user_service.json"), filePath)
	})
}

func TestWriter_WriteBytes(t *testing.T) {
	t.Run("returns contract as bytes", func(t *testing.T) {
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
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "3.0.0"},
			},
		}

		writer := contract.NewWriter()
		data, err := writer.WriteBytes(&c)
		require.NoError(t, err)

		var result contract.Contract
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "Consumer", result.Consumer.Name)
	})
}
