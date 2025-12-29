// Package examples demonstrates how to use the yakusoku Go SDK.
//
// Run with: go test -v ./sdk/go/examples/...
package examples

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/jt-chihara/yakusoku/sdk/go/yakusoku"
)

// User represents a user returned by the UserService API.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserClient is a client for the UserService API.
type UserClient struct {
	BaseURL string
}

// GetUser fetches a user by ID.
func (c *UserClient) GetUser(id int) (*User, error) {
	resp, err := http.Get(c.BaseURL + "/users/1")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// TestGetUser_Success demonstrates a basic consumer contract test.
func TestGetUser_Success(t *testing.T) {
	// 1. Create a Pact instance
	pact := yakusoku.NewPact(yakusoku.Config{
		Consumer: "OrderService",
		Provider: "UserService",
		PactDir:  "./pacts",
	})
	defer pact.Teardown()

	// 2. Define the expected interaction
	pact.
		Given("user 1 exists").
		UponReceiving("a request to get user 1").
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

	// 3. Verify the interaction with actual client code
	err := pact.Verify(func() error {
		client := &UserClient{BaseURL: pact.ServerURL()}
		user, err := client.GetUser(1)
		if err != nil {
			return err
		}

		// Assert the response
		if user.ID != 1 {
			t.Errorf("expected ID 1, got %d", user.ID)
		}
		if user.Name != "John Doe" {
			t.Errorf("expected name 'John Doe', got '%s'", user.Name)
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

// TestGetUser_NotFound demonstrates testing error scenarios.
func TestGetUser_NotFound(t *testing.T) {
	pact := yakusoku.NewPact(yakusoku.Config{
		Consumer: "OrderService",
		Provider: "UserService",
		PactDir:  "./pacts",
	})
	defer pact.Teardown()

	pact.
		Given("user 999 does not exist").
		UponReceiving("a request to get non-existent user").
		WithRequest(yakusoku.Request{
			Method: "GET",
			Path:   "/users/999",
		}).
		WillRespondWith(yakusoku.Response{
			Status: 404,
			Body: map[string]interface{}{
				"error": "User not found",
			},
		})

	err := pact.Verify(func() error {
		resp, err := http.Get(pact.ServerURL() + "/users/999")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 404 {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

// TestCreateUser demonstrates testing POST requests with request bodies.
func TestCreateUser(t *testing.T) {
	pact := yakusoku.NewPact(yakusoku.Config{
		Consumer: "OrderService",
		Provider: "UserService",
		PactDir:  "./pacts",
	})
	defer pact.Teardown()

	pact.
		UponReceiving("a request to create a user").
		WithRequest(yakusoku.Request{
			Method: "POST",
			Path:   "/users",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"name":  "Jane Doe",
				"email": "jane@example.com",
			},
		}).
		WillRespondWith(yakusoku.Response{
			Status: 201,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"id":    2,
				"name":  "Jane Doe",
				"email": "jane@example.com",
			},
		})

	err := pact.Verify(func() error {
		resp, err := http.Post(
			pact.ServerURL()+"/users",
			"application/json",
			nil, // In real code, you'd send the body
		)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			t.Errorf("expected status 201, got %d", resp.StatusCode)
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

// TestMultipleInteractions demonstrates multiple interactions in one test.
func TestMultipleInteractions(t *testing.T) {
	pact := yakusoku.NewPact(yakusoku.Config{
		Consumer: "OrderService",
		Provider: "UserService",
		PactDir:  "./pacts",
	})
	defer pact.Teardown()

	// First interaction: get user
	pact.
		Given("user 1 exists").
		UponReceiving("a request to get user 1").
		WithRequest(yakusoku.Request{
			Method: "GET",
			Path:   "/users/1",
		}).
		WillRespondWith(yakusoku.Response{
			Status: 200,
			Body:   map[string]interface{}{"id": 1, "name": "John"},
		})

	// Second interaction: get user's orders
	pact.
		Given("user 1 has orders").
		UponReceiving("a request to get user 1's orders").
		WithRequest(yakusoku.Request{
			Method: "GET",
			Path:   "/users/1/orders",
		}).
		WillRespondWith(yakusoku.Response{
			Status: 200,
			Body: []interface{}{
				map[string]interface{}{"id": 101, "total": 99.99},
				map[string]interface{}{"id": 102, "total": 149.99},
			},
		})

	err := pact.Verify(func() error {
		// Call first endpoint
		resp1, err := http.Get(pact.ServerURL() + "/users/1")
		if err != nil {
			return err
		}
		defer resp1.Body.Close()
		body1, _ := io.ReadAll(resp1.Body)
		t.Logf("User response: %s", body1)

		// Call second endpoint
		resp2, err := http.Get(pact.ServerURL() + "/users/1/orders")
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		body2, _ := io.ReadAll(resp2.Body)
		t.Logf("Orders response: %s", body2)

		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}
