package integration_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jt-chihara/yakusoku/sdk/go/yakusoku"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UserClient is a sample consumer client for testing.
type UserClient struct {
	baseURL string
}

func NewUserClient(baseURL string) *UserClient {
	return &UserClient{baseURL: baseURL}
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

func (c *UserClient) GetUser(id int) (*User, error) {
	resp, err := http.Get(fmt.Sprintf("%s/users/%d", c.baseURL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *UserClient) CreateUser(name, email string) (*User, error) {
	resp, err := http.Post(
		fmt.Sprintf("%s/users", c.baseURL),
		"application/json",
		nil,
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func TestConsumerE2E_GetUser(t *testing.T) {
	tmpDir := t.TempDir()

	pact := yakusoku.NewPact(yakusoku.Config{
		Consumer: "OrderService",
		Provider: "UserService",
		PactDir:  tmpDir,
	})
	defer pact.Teardown()

	// Define interaction
	pact.
		Given("user 1 exists").
		UponReceiving("a request for user 1").
		WithRequest(yakusoku.Request{
			Method: "GET",
			Path:   "/users/1",
		}).
		WillRespondWith(yakusoku.Response{
			Status: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"id":    1,
				"name":  "John Doe",
				"email": "john@example.com",
			},
		})

	// Verify with actual consumer code
	err := pact.Verify(func() error {
		client := NewUserClient(pact.ServerURL())
		user, err := client.GetUser(1)
		if err != nil {
			return err
		}

		if user.ID != 1 {
			return fmt.Errorf("expected user ID 1, got %d", user.ID)
		}
		if user.Name != "John Doe" {
			return fmt.Errorf("expected name 'John Doe', got '%s'", user.Name)
		}
		return nil
	})

	require.NoError(t, err)

	// Verify contract file was generated
	contractPath := filepath.Join(tmpDir, "orderservice-userservice.json")
	data, err := os.ReadFile(contractPath)
	require.NoError(t, err)

	var contract map[string]interface{}
	err = json.Unmarshal(data, &contract)
	require.NoError(t, err)

	// Verify contract structure
	assert.Equal(t, "OrderService", contract["consumer"].(map[string]interface{})["name"])
	assert.Equal(t, "UserService", contract["provider"].(map[string]interface{})["name"])

	interactions := contract["interactions"].([]interface{})
	assert.Len(t, interactions, 1)

	interaction := interactions[0].(map[string]interface{})
	assert.Equal(t, "a request for user 1", interaction["description"])
	assert.Equal(t, "user 1 exists", interaction["providerState"])

	request := interaction["request"].(map[string]interface{})
	assert.Equal(t, "GET", request["method"])
	assert.Equal(t, "/users/1", request["path"])

	response := interaction["response"].(map[string]interface{})
	assert.Equal(t, float64(200), response["status"])

	body := response["body"].(map[string]interface{})
	assert.Equal(t, float64(1), body["id"])
	assert.Equal(t, "John Doe", body["name"])
}

func TestConsumerE2E_MultipleInteractions(t *testing.T) {
	tmpDir := t.TempDir()

	pact := yakusoku.NewPact(yakusoku.Config{
		Consumer: "OrderService",
		Provider: "UserService",
		PactDir:  tmpDir,
	})
	defer pact.Teardown()

	// Define multiple interactions
	pact.
		Given("user 1 exists").
		UponReceiving("a request for user 1").
		WithRequest(yakusoku.Request{Method: "GET", Path: "/users/1"}).
		WillRespondWith(yakusoku.Response{
			Status: 200,
			Body:   map[string]interface{}{"id": 1, "name": "John"},
		})

	pact.
		Given("user 2 exists").
		UponReceiving("a request for user 2").
		WithRequest(yakusoku.Request{Method: "GET", Path: "/users/2"}).
		WillRespondWith(yakusoku.Response{
			Status: 200,
			Body:   map[string]interface{}{"id": 2, "name": "Jane"},
		})

	err := pact.Verify(func() error {
		client := NewUserClient(pact.ServerURL())

		user1, err := client.GetUser(1)
		if err != nil {
			return err
		}
		if user1.Name != "John" {
			return fmt.Errorf("expected John, got %s", user1.Name)
		}

		user2, err := client.GetUser(2)
		if err != nil {
			return err
		}
		if user2.Name != "Jane" {
			return fmt.Errorf("expected Jane, got %s", user2.Name)
		}

		return nil
	})

	require.NoError(t, err)

	// Verify both interactions are in the contract
	contractPath := filepath.Join(tmpDir, "orderservice-userservice.json")
	data, _ := os.ReadFile(contractPath)

	var contract map[string]interface{}
	json.Unmarshal(data, &contract)

	interactions := contract["interactions"].([]interface{})
	assert.Len(t, interactions, 2)
}

func TestConsumerE2E_ResponseHeaders(t *testing.T) {
	tmpDir := t.TempDir()

	pact := yakusoku.NewPact(yakusoku.Config{
		Consumer: "TestConsumer",
		Provider: "TestProvider",
		PactDir:  tmpDir,
	})
	defer pact.Teardown()

	pact.
		UponReceiving("a request").
		WithRequest(yakusoku.Request{Method: "GET", Path: "/test"}).
		WillRespondWith(yakusoku.Response{
			Status: 200,
			Headers: map[string]string{
				"Content-Type":    "application/json",
				"X-Custom-Header": "custom-value",
			},
			Body: "OK",
		})

	err := pact.Verify(func() error {
		resp, err := http.Get(pact.ServerURL() + "/test")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.Header.Get("Content-Type") != "application/json" {
			return fmt.Errorf("expected Content-Type application/json")
		}
		if resp.Header.Get("X-Custom-Header") != "custom-value" {
			return fmt.Errorf("expected custom header")
		}

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			return fmt.Errorf("expected body OK, got %s", string(body))
		}
		return nil
	})

	require.NoError(t, err)
}

func TestConsumerE2E_ErrorResponse(t *testing.T) {
	tmpDir := t.TempDir()

	pact := yakusoku.NewPact(yakusoku.Config{
		Consumer: "TestConsumer",
		Provider: "TestProvider",
		PactDir:  tmpDir,
	})
	defer pact.Teardown()

	pact.
		Given("user does not exist").
		UponReceiving("a request for non-existent user").
		WithRequest(yakusoku.Request{Method: "GET", Path: "/users/999"}).
		WillRespondWith(yakusoku.Response{
			Status: 404,
			Body:   map[string]interface{}{"error": "User not found"},
		})

	err := pact.Verify(func() error {
		resp, err := http.Get(pact.ServerURL() + "/users/999")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 404 {
			return fmt.Errorf("expected 404, got %d", resp.StatusCode)
		}
		return nil
	})

	require.NoError(t, err)
}
