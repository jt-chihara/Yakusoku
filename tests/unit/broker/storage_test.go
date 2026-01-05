package broker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jt-chihara/yakusoku/internal/broker"
	"github.com/jt-chihara/yakusoku/internal/contract"
)

func TestStorage_SaveContract(t *testing.T) {
	t.Run("saves contract successfully", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		c := createTestContract("Consumer", "Provider", "1.0.0")

		err := storage.SaveContract(c)
		require.NoError(t, err)

		retrieved, err := storage.GetContract("Consumer", "Provider", "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, "Consumer", retrieved.Consumer.Name)
		assert.Equal(t, "Provider", retrieved.Provider.Name)
	})

	t.Run("overwrites existing version", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		c1 := createTestContract("Consumer", "Provider", "1.0.0")
		c1.Interactions = []contract.Interaction{{Description: "old"}}

		c2 := createTestContract("Consumer", "Provider", "1.0.0")
		c2.Interactions = []contract.Interaction{{Description: "new"}}

		_ = storage.SaveContract(c1)
		_ = storage.SaveContract(c2)

		retrieved, _ := storage.GetContract("Consumer", "Provider", "1.0.0")
		assert.Equal(t, "new", retrieved.Interactions[0].Description)
	})

	t.Run("stores multiple versions", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		storage.SaveContract(createTestContract("Consumer", "Provider", "1.0.0"))
		storage.SaveContract(createTestContract("Consumer", "Provider", "2.0.0"))

		v1, _ := storage.GetContract("Consumer", "Provider", "1.0.0")
		v2, _ := storage.GetContract("Consumer", "Provider", "2.0.0")

		assert.NotNil(t, v1)
		assert.NotNil(t, v2)
	})
}

func TestStorage_GetContract(t *testing.T) {
	t.Run("returns error for non-existent contract", func(t *testing.T) {
		storage := broker.NewMemoryStorage()

		_, err := storage.GetContract("Unknown", "Provider", "1.0.0")
		require.Error(t, err)
	})

	t.Run("returns latest version when version is empty", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		storage.SaveContract(createTestContract("Consumer", "Provider", "1.0.0"))
		storage.SaveContract(createTestContract("Consumer", "Provider", "2.0.0"))

		retrieved, err := storage.GetContract("Consumer", "Provider", "")
		require.NoError(t, err)
		assert.Equal(t, "2.0.0", retrieved.Metadata.PactSpecification.Version)
	})
}

func TestStorage_ListContracts(t *testing.T) {
	t.Run("lists all contracts", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		storage.SaveContract(createTestContract("Consumer1", "Provider1", "1.0.0"))
		storage.SaveContract(createTestContract("Consumer2", "Provider2", "1.0.0"))

		contracts := storage.ListContracts()
		assert.Len(t, contracts, 2)
	})

	t.Run("returns empty list when no contracts", func(t *testing.T) {
		storage := broker.NewMemoryStorage()

		contracts := storage.ListContracts()
		assert.Empty(t, contracts)
	})
}

func TestStorage_GetContractsByProvider(t *testing.T) {
	t.Run("returns contracts for specific provider", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		storage.SaveContract(createTestContract("Consumer1", "SharedProvider", "1.0.0"))
		storage.SaveContract(createTestContract("Consumer2", "SharedProvider", "1.0.0"))
		storage.SaveContract(createTestContract("Consumer3", "OtherProvider", "1.0.0"))

		contracts := storage.GetContractsByProvider("SharedProvider")
		assert.Len(t, contracts, 2)
	})
}

func TestStorage_GetContractsByConsumer(t *testing.T) {
	t.Run("returns contracts for specific consumer", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		storage.SaveContract(createTestContract("SharedConsumer", "Provider1", "1.0.0"))
		storage.SaveContract(createTestContract("SharedConsumer", "Provider2", "1.0.0"))
		storage.SaveContract(createTestContract("OtherConsumer", "Provider3", "1.0.0"))

		contracts := storage.GetContractsByConsumer("SharedConsumer")
		assert.Len(t, contracts, 2)
	})
}

func TestStorage_DeleteContract(t *testing.T) {
	t.Run("deletes contract successfully", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		storage.SaveContract(createTestContract("Consumer", "Provider", "1.0.0"))

		err := storage.DeleteContract("Consumer", "Provider", "1.0.0")
		require.NoError(t, err)

		_, err = storage.GetContract("Consumer", "Provider", "1.0.0")
		require.Error(t, err)
	})

	t.Run("returns error for non-existent contract", func(t *testing.T) {
		storage := broker.NewMemoryStorage()

		err := storage.DeleteContract("Unknown", "Provider", "1.0.0")
		require.Error(t, err)
	})
}

func TestStorage_RecordAndGetVerification(t *testing.T) {
	t.Run("records and retrieves verification", func(t *testing.T) {
		storage := broker.NewMemoryStorage()

		err := storage.RecordVerification("Consumer", "Provider", "1.0.0", true)
		require.NoError(t, err)

		success, exists := storage.GetVerification("Consumer", "Provider", "1.0.0")
		assert.True(t, exists)
		assert.True(t, success)
	})

	t.Run("returns false for non-existent verification", func(t *testing.T) {
		storage := broker.NewMemoryStorage()

		_, exists := storage.GetVerification("Unknown", "Provider", "1.0.0")
		assert.False(t, exists)
	})

	t.Run("records failed verification", func(t *testing.T) {
		storage := broker.NewMemoryStorage()

		err := storage.RecordVerification("Consumer", "Provider", "1.0.0", false)
		require.NoError(t, err)

		success, exists := storage.GetVerification("Consumer", "Provider", "1.0.0")
		assert.True(t, exists)
		assert.False(t, success)
	})
}

func createTestContract(consumer, provider, version string) *contract.Contract {
	return &contract.Contract{
		Consumer: contract.Pacticipant{Name: consumer},
		Provider: contract.Pacticipant{Name: provider},
		Interactions: []contract.Interaction{
			{
				Description: "test interaction",
				Request:     contract.Request{Method: "GET", Path: "/test"},
				Response:    contract.Response{Status: 200},
			},
		},
		Metadata: contract.Metadata{
			PactSpecification: contract.PactSpec{Version: version},
		},
	}
}
