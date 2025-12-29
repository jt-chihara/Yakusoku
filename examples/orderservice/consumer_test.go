package main

import (
	"testing"

	"github.com/jt-chihara/yakusoku/sdk/go/yakusoku"
)

// TestUserClient_GetUser tests the UserClient against a mock UserService.
// This generates the contract file that will be used to verify UserService.
func TestUserClient_GetUser(t *testing.T) {
	// Create a Pact instance
	pact := yakusoku.NewPact(yakusoku.Config{
		Consumer: "OrderService",
		Provider: "UserService",
		PactDir:  "../../pacts", // Save contracts to project root/pacts
	})
	defer pact.Teardown()

	t.Run("get existing user", func(t *testing.T) {
		// Define expected interaction
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
					"name":  "Alice",
					"email": "alice@example.com",
				},
			})

		// Test with mock server
		err := pact.Verify(func() error {
			client := &UserClient{BaseURL: pact.ServerURL()}
			user, err := client.GetUser(1)
			if err != nil {
				return err
			}

			// Verify response
			if user.ID != 1 {
				t.Errorf("expected ID 1, got %d", user.ID)
			}
			if user.Name != "Alice" {
				t.Errorf("expected name 'Alice', got '%s'", user.Name)
			}
			if user.Email != "alice@example.com" {
				t.Errorf("expected email 'alice@example.com', got '%s'", user.Email)
			}
			return nil
		})

		if err != nil {
			t.Fatal(err)
		}
	})
}

// TestUserClient_GetUser_NotFound tests the 404 case.
func TestUserClient_GetUser_NotFound(t *testing.T) {
	pact := yakusoku.NewPact(yakusoku.Config{
		Consumer: "OrderService",
		Provider: "UserService",
		PactDir:  "../../pacts",
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
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"error": "User not found",
			},
		})

	err := pact.Verify(func() error {
		client := &UserClient{BaseURL: pact.ServerURL()}
		_, err := client.GetUser(999)
		if err == nil {
			t.Error("expected error for non-existent user")
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}
