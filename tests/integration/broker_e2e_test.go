package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/jt-chihara/yakusoku/internal/broker"
	"github.com/jt-chihara/yakusoku/internal/contract"
)

func TestBrokerE2E_PublishAndRetrieve(t *testing.T) {
	storage := broker.NewMemoryStorage()
	api := broker.NewAPI(storage)
	server := httptest.NewServer(api.Handler())
	defer server.Close()

	// Publish a contract
	c := contract.Contract{
		Consumer: contract.Pacticipant{Name: "OrderService"},
		Provider: contract.Pacticipant{Name: "UserService"},
		Interactions: []contract.Interaction{
			{
				Description: "get user",
				Request:     contract.Request{Method: "GET", Path: "/users/1"},
				Response:    contract.Response{Status: 200},
			},
		},
		Metadata: contract.Metadata{
			PactSpecification: contract.PactSpec{Version: "3.0.0"},
		},
	}

	body, _ := json.Marshal(c)
	resp, err := http.Post(
		server.URL+"/pacts/provider/UserService/consumer/OrderService/version/1.0.0",
		"application/json",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Retrieve the contract
	resp, err = http.Get(server.URL + "/pacts/provider/UserService/consumer/OrderService/version/1.0.0")
	require.NoError(t, err)
	defer resp.Body.Close()

	var retrieved contract.Contract
	json.NewDecoder(resp.Body).Decode(&retrieved)
	assert.Equal(t, "OrderService", retrieved.Consumer.Name)
	assert.Equal(t, "UserService", retrieved.Provider.Name)
	assert.Len(t, retrieved.Interactions, 1)
}

func TestBrokerE2E_VerificationWorkflow(t *testing.T) {
	storage := broker.NewMemoryStorage()
	api := broker.NewAPI(storage)
	server := httptest.NewServer(api.Handler())
	defer server.Close()

	// 1. Consumer publishes contract
	c := contract.Contract{
		Consumer: contract.Pacticipant{Name: "OrderService"},
		Provider: contract.Pacticipant{Name: "UserService"},
		Interactions: []contract.Interaction{
			{
				Description: "get user",
				Request:     contract.Request{Method: "GET", Path: "/users/1"},
				Response:    contract.Response{Status: 200},
			},
		},
		Metadata: contract.Metadata{
			PactSpecification: contract.PactSpec{Version: "1.0.0"},
		},
	}

	body, _ := json.Marshal(c)
	resp, _ := http.Post(
		server.URL+"/pacts/provider/UserService/consumer/OrderService/version/1.0.0",
		"application/json",
		bytes.NewReader(body),
	)
	resp.Body.Close()

	// 2. Check can-i-deploy before verification (should be false)
	resp, _ = http.Get(server.URL + "/matrix?pacticipant=OrderService&version=1.0.0")
	var canDeploy map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&canDeploy)
	resp.Body.Close()
	assert.Equal(t, false, canDeploy["deployable"])

	// 3. Provider records successful verification
	verificationBody, _ := json.Marshal(map[string]interface{}{
		"success":         true,
		"providerVersion": "2.0.0",
	})
	resp, _ = http.Post(
		server.URL+"/pacts/provider/UserService/consumer/OrderService/version/1.0.0/verification-results",
		"application/json",
		bytes.NewReader(verificationBody),
	)
	resp.Body.Close()

	// 4. Check can-i-deploy after verification (should be true)
	resp, _ = http.Get(server.URL + "/matrix?pacticipant=OrderService&version=1.0.0")
	json.NewDecoder(resp.Body).Decode(&canDeploy)
	resp.Body.Close()
	assert.Equal(t, true, canDeploy["deployable"])
}

func TestBrokerE2E_MultipleConsumers(t *testing.T) {
	storage := broker.NewMemoryStorage()
	api := broker.NewAPI(storage)
	server := httptest.NewServer(api.Handler())
	defer server.Close()

	// Multiple consumers publish contracts for the same provider
	consumers := []string{"OrderService", "PaymentService", "NotificationService"}

	for _, consumer := range consumers {
		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: consumer},
			Provider: contract.Pacticipant{Name: "UserService"},
			Interactions: []contract.Interaction{
				{
					Description: "get user from " + consumer,
					Request:     contract.Request{Method: "GET", Path: "/users/1"},
					Response:    contract.Response{Status: 200},
				},
			},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: "1.0.0"},
			},
		}

		body, _ := json.Marshal(c)
		resp, _ := http.Post(
			server.URL+"/pacts/provider/UserService/consumer/"+consumer+"/version/1.0.0",
			"application/json",
			bytes.NewReader(body),
		)
		resp.Body.Close()
	}

	// Get all contracts for provider
	resp, _ := http.Get(server.URL + "/pacts/provider/UserService")
	defer resp.Body.Close()

	var contracts []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&contracts)
	assert.Len(t, contracts, 3)
}

func TestBrokerE2E_VersionManagement(t *testing.T) {
	storage := broker.NewMemoryStorage()
	api := broker.NewAPI(storage)
	server := httptest.NewServer(api.Handler())
	defer server.Close()

	// Publish multiple versions
	versions := []string{"1.0.0", "1.1.0", "2.0.0"}

	for _, version := range versions {
		c := contract.Contract{
			Consumer: contract.Pacticipant{Name: "Consumer"},
			Provider: contract.Pacticipant{Name: "Provider"},
			Interactions: []contract.Interaction{
				{
					Description: "interaction v" + version,
					Request:     contract.Request{Method: "GET", Path: "/test"},
					Response:    contract.Response{Status: 200},
				},
			},
			Metadata: contract.Metadata{
				PactSpecification: contract.PactSpec{Version: version},
			},
		}

		body, _ := json.Marshal(c)
		resp, _ := http.Post(
			server.URL+"/pacts/provider/Provider/consumer/Consumer/version/"+version,
			"application/json",
			bytes.NewReader(body),
		)
		resp.Body.Close()
	}

	// Get latest version
	resp, _ := http.Get(server.URL + "/pacts/provider/Provider/consumer/Consumer/latest")
	defer resp.Body.Close()

	var latest contract.Contract
	json.NewDecoder(resp.Body).Decode(&latest)
	assert.Equal(t, "2.0.0", latest.Metadata.PactSpecification.Version)
}
