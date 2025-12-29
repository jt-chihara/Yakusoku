package broker_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jt-chihara/yakusoku/internal/broker"
	"github.com/jt-chihara/yakusoku/internal/contract"
)

func TestS3Storage_SaveContract(t *testing.T) {
	t.Run("saves contract to S3", func(t *testing.T) {
		mock := broker.NewMockS3Client()
		storage := broker.NewS3Storage(mock, "test-bucket", "pacts/")

		c := &contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Interactions: []contract.Interaction{
				{
					Description: "test interaction",
					Request:     contract.Request{Method: "GET", Path: "/test"},
					Response:    contract.Response{Status: 200},
				},
			},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "1.0.0"},
			},
		}

		err := storage.SaveContract(c)
		require.NoError(t, err)

		// Verify the contract was saved
		retrieved, err := storage.GetContract("Consumer", "Provider", "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, "Consumer", retrieved.Consumer.Name)
		assert.Equal(t, "Provider", retrieved.Provider.Name)
	})
}

func TestS3Storage_GetContract(t *testing.T) {
	t.Run("retrieves contract from S3", func(t *testing.T) {
		mock := broker.NewMockS3Client()
		storage := broker.NewS3Storage(mock, "test-bucket", "pacts/")

		// First save a contract
		c := &contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "1.0.0"},
			},
		}
		_ = storage.SaveContract(c)

		// Then retrieve it
		retrieved, err := storage.GetContract("Consumer", "Provider", "1.0.0")
		require.NoError(t, err)
		assert.Equal(t, "Consumer", retrieved.Consumer.Name)
	})

	t.Run("returns error for non-existent contract", func(t *testing.T) {
		mock := broker.NewMockS3Client()
		storage := broker.NewS3Storage(mock, "test-bucket", "pacts/")

		_, err := storage.GetContract("Unknown", "Provider", "1.0.0")
		require.Error(t, err)
		assert.Equal(t, broker.ErrNotFound, err)
	})

	t.Run("returns latest version when version is empty", func(t *testing.T) {
		mock := broker.NewMockS3Client()
		storage := broker.NewS3Storage(mock, "test-bucket", "pacts/")

		// Save multiple versions
		c1 := &contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "1.0.0"},
			},
		}
		c2 := &contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "2.0.0"},
			},
		}
		_ = storage.SaveContract(c1)
		_ = storage.SaveContract(c2)

		// Get latest
		retrieved, err := storage.GetContract("Consumer", "Provider", "")
		require.NoError(t, err)
		assert.Equal(t, "2.0.0", retrieved.Metadata.PactSpecification.Version)
	})
}

func TestS3Storage_ListContracts(t *testing.T) {
	t.Run("lists all contracts", func(t *testing.T) {
		mock := broker.NewMockS3Client()
		storage := broker.NewS3Storage(mock, "test-bucket", "pacts/")

		c1 := &contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer1"},
			Provider: contract.Pacticipant{Name: "Provider1"},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "1.0.0"},
			},
		}
		c2 := &contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer2"},
			Provider: contract.Pacticipant{Name: "Provider2"},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "1.0.0"},
			},
		}
		_ = storage.SaveContract(c1)
		_ = storage.SaveContract(c2)

		contracts := storage.ListContracts()
		assert.Len(t, contracts, 2)
	})
}

func TestS3Storage_DeleteContract(t *testing.T) {
	t.Run("deletes contract from S3", func(t *testing.T) {
		mock := broker.NewMockS3Client()
		storage := broker.NewS3Storage(mock, "test-bucket", "pacts/")

		c := &contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "1.0.0"},
			},
		}
		_ = storage.SaveContract(c)

		err := storage.DeleteContract("Consumer", "Provider", "1.0.0")
		require.NoError(t, err)

		_, err = storage.GetContract("Consumer", "Provider", "1.0.0")
		require.Error(t, err)
	})
}

func TestS3Storage_Verification(t *testing.T) {
	t.Run("records and retrieves verification", func(t *testing.T) {
		mock := broker.NewMockS3Client()
		storage := broker.NewS3Storage(mock, "test-bucket", "pacts/")

		err := storage.RecordVerification("Consumer", "Provider", "1.0.0", true)
		require.NoError(t, err)

		success, exists := storage.GetVerification("Consumer", "Provider", "1.0.0")
		assert.True(t, exists)
		assert.True(t, success)
	})

	t.Run("returns false for non-existent verification", func(t *testing.T) {
		mock := broker.NewMockS3Client()
		storage := broker.NewS3Storage(mock, "test-bucket", "pacts/")

		_, exists := storage.GetVerification("Unknown", "Provider", "1.0.0")
		assert.False(t, exists)
	})
}

func TestS3Storage_IsDeployable(t *testing.T) {
	t.Run("returns true when all verifications pass", func(t *testing.T) {
		mock := broker.NewMockS3Client()
		storage := broker.NewS3Storage(mock, "test-bucket", "pacts/")

		c := &contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "1.0.0"},
			},
		}
		_ = storage.SaveContract(c)
		_ = storage.RecordVerification("Consumer", "Provider", "1.0.0", true)

		deployable, _ := storage.IsDeployable("Consumer", "1.0.0")
		assert.True(t, deployable)
	})

	t.Run("returns false when verification failed", func(t *testing.T) {
		mock := broker.NewMockS3Client()
		storage := broker.NewS3Storage(mock, "test-bucket", "pacts/")

		c := &contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "1.0.0"},
			},
		}
		_ = storage.SaveContract(c)
		_ = storage.RecordVerification("Consumer", "Provider", "1.0.0", false)

		deployable, _ := storage.IsDeployable("Consumer", "1.0.0")
		assert.False(t, deployable)
	})
}

// MockS3Client test to ensure interface compliance
func TestMockS3Client_ImplementsInterface(t *testing.T) {
	mock := broker.NewMockS3Client()

	// Test PutObject
	err := mock.PutObject(context.Background(), "bucket", "key", []byte("data"))
	require.NoError(t, err)

	// Test GetObject
	data, err := mock.GetObject(context.Background(), "bucket", "key")
	require.NoError(t, err)
	assert.Equal(t, []byte("data"), data)

	// Test DeleteObject
	err = mock.DeleteObject(context.Background(), "bucket", "key")
	require.NoError(t, err)

	// Test ListObjects
	keys, err := mock.ListObjects(context.Background(), "bucket", "")
	require.NoError(t, err)
	assert.Empty(t, keys) // deleted above
}
