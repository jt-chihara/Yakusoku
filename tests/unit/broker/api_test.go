package broker_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jt-chihara/yakusoku/internal/broker"
	"github.com/jt-chihara/yakusoku/internal/contract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI_PublishContract(t *testing.T) {
	t.Run("publishes contract successfully", func(t *testing.T) {
		api := broker.NewAPI(broker.NewMemoryStorage())
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		c := createTestContract("Consumer", "Provider", "1.0.0")
		body, _ := json.Marshal(c)

		resp, err := http.Post(
			server.URL+"/pacts/provider/Provider/consumer/Consumer/version/1.0.0",
			"application/json",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		api := broker.NewAPI(broker.NewMemoryStorage())
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		resp, err := http.Post(
			server.URL+"/pacts/provider/Provider/consumer/Consumer/version/1.0.0",
			"application/json",
			bytes.NewReader([]byte("invalid json")),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAPI_GetContract(t *testing.T) {
	t.Run("retrieves contract successfully", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		c := createTestContract("Consumer", "Provider", "1.0.0")
		storage.SaveContract(c)

		api := broker.NewAPI(storage)
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		resp, err := http.Get(server.URL + "/pacts/provider/Provider/consumer/Consumer/version/1.0.0")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrieved contract.Contract
		json.NewDecoder(resp.Body).Decode(&retrieved)
		assert.Equal(t, "Consumer", retrieved.Consumer.Name)
	})

	t.Run("returns 404 for non-existent contract", func(t *testing.T) {
		api := broker.NewAPI(broker.NewMemoryStorage())
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		resp, err := http.Get(server.URL + "/pacts/provider/Provider/consumer/Consumer/version/1.0.0")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("retrieves latest version when version is 'latest'", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		storage.SaveContract(createTestContract("Consumer", "Provider", "1.0.0"))
		storage.SaveContract(createTestContract("Consumer", "Provider", "2.0.0"))

		api := broker.NewAPI(storage)
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		resp, err := http.Get(server.URL + "/pacts/provider/Provider/consumer/Consumer/latest")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestAPI_ListContracts(t *testing.T) {
	t.Run("lists all contracts", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		storage.SaveContract(createTestContract("Consumer1", "Provider1", "1.0.0"))
		storage.SaveContract(createTestContract("Consumer2", "Provider2", "1.0.0"))

		api := broker.NewAPI(storage)
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		resp, err := http.Get(server.URL + "/pacts")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Len(t, result, 2)
	})
}

func TestAPI_GetContractsByProvider(t *testing.T) {
	t.Run("returns contracts for provider", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		storage.SaveContract(createTestContract("Consumer1", "TargetProvider", "1.0.0"))
		storage.SaveContract(createTestContract("Consumer2", "TargetProvider", "1.0.0"))
		storage.SaveContract(createTestContract("Consumer3", "OtherProvider", "1.0.0"))

		api := broker.NewAPI(storage)
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		resp, err := http.Get(server.URL + "/pacts/provider/TargetProvider")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Len(t, result, 2)
	})
}

func TestAPI_DeleteContract(t *testing.T) {
	t.Run("deletes contract successfully", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		storage.SaveContract(createTestContract("Consumer", "Provider", "1.0.0"))

		api := broker.NewAPI(storage)
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		req, _ := http.NewRequest(
			http.MethodDelete,
			server.URL+"/pacts/provider/Provider/consumer/Consumer/version/1.0.0",
			nil,
		)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}

func TestAPI_CanIDeploy(t *testing.T) {
	t.Run("returns true when all contracts verified", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		c := createTestContract("Consumer", "Provider", "1.0.0")
		storage.SaveContract(c)
		storage.RecordVerification("Consumer", "Provider", "1.0.0", true)

		api := broker.NewAPI(storage)
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		resp, err := http.Get(server.URL + "/matrix?pacticipant=Consumer&version=1.0.0")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, true, result["deployable"])
	})

	t.Run("returns false when verification failed", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		c := createTestContract("Consumer", "Provider", "1.0.0")
		storage.SaveContract(c)
		storage.RecordVerification("Consumer", "Provider", "1.0.0", false)

		api := broker.NewAPI(storage)
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		resp, err := http.Get(server.URL + "/matrix?pacticipant=Consumer&version=1.0.0")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, false, result["deployable"])
	})

	t.Run("returns false when no verification exists", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		c := createTestContract("Consumer", "Provider", "1.0.0")
		storage.SaveContract(c)

		api := broker.NewAPI(storage)
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		resp, err := http.Get(server.URL + "/matrix?pacticipant=Consumer&version=1.0.0")
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.Equal(t, false, result["deployable"])
	})
}

func TestAPI_RecordVerification(t *testing.T) {
	t.Run("records successful verification", func(t *testing.T) {
		storage := broker.NewMemoryStorage()
		storage.SaveContract(createTestContract("Consumer", "Provider", "1.0.0"))

		api := broker.NewAPI(storage)
		server := httptest.NewServer(api.Handler())
		defer server.Close()

		body, _ := json.Marshal(map[string]interface{}{
			"success":         true,
			"providerVersion": "2.0.0",
		})

		resp, err := http.Post(
			server.URL+"/pacts/provider/Provider/consumer/Consumer/version/1.0.0/verification-results",
			"application/json",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}
