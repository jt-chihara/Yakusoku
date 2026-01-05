// UserService is a sample Provider API for contract testing demonstration.
//
// Run with: go run ./examples/userservice
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// User represents a user in the system.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Order represents an order in the system.
type Order struct {
	ID    int     `json:"id"`
	Total float64 `json:"total"`
}

// In-memory user database
var users = map[int]User{
	1: {ID: 1, Name: "John Doe", Email: "john@example.com"},
	2: {ID: 2, Name: "Jane Doe", Email: "jane@example.com"},
	3: {ID: 3, Name: "Charlie", Email: "charlie@example.com"},
}

// In-memory orders database
var orders = map[int][]Order{
	1: {{ID: 101, Total: 99.99}, {ID: 102, Total: 149.99}},
}

var nextUserID = 3

func main() {
	mux := http.NewServeMux()

	// GET /users/:id - Get user by ID
	mux.HandleFunc("GET /users/{id}", handleGetUser)

	// GET /users - List all users
	mux.HandleFunc("GET /users", handleListUsers)

	// POST /users - Create a new user
	mux.HandleFunc("POST /users", handleCreateUser)

	// GET /users/:id/orders - Get user's orders
	mux.HandleFunc("GET /users/{id}/orders", handleGetUserOrders)

	// POST /provider-states - Provider state setup for contract testing
	mux.HandleFunc("POST /provider-states", handleProviderStates)

	port := "8080"
	fmt.Printf("UserService (Provider) starting on http://localhost:%s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /users      - List all users")
	fmt.Println("  GET  /users/{id} - Get user by ID")
	fmt.Println("  POST /provider-states - Provider state setup (for testing)")

	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)

	user, exists := users[id]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "User not found",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	userList := make([]User, 0, len(users))
	for _, u := range users {
		userList = append(userList, u)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(userList)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nextUserID++
	user := User{
		ID:    nextUserID,
		Name:  input.Name,
		Email: input.Email,
	}
	users[user.ID] = user

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

func handleGetUserOrders(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)

	userOrders, exists := orders[id]
	if !exists {
		userOrders = []Order{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(userOrders)
}

func handleProviderStates(w http.ResponseWriter, r *http.Request) {
	var state struct {
		State  string                 `json:"state"`
		Params map[string]interface{} `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Setting up provider state: %s", state.State)

	// Handle different provider states
	switch {
	case strings.Contains(state.State, "user 1 exists"):
		users[1] = User{ID: 1, Name: "John Doe", Email: "john@example.com"}
	case strings.Contains(state.State, "user 999 does not exist"):
		delete(users, 999)
	case strings.Contains(state.State, "no users exist"):
		users = make(map[int]User)
		nextUserID = 1 // Reset so next created user gets ID 2
	case strings.Contains(state.State, "user 1 has orders"):
		users[1] = User{ID: 1, Name: "John Doe", Email: "john@example.com"}
		orders[1] = []Order{{ID: 101, Total: 99.99}, {ID: 102, Total: 149.99}}
	default:
		// Reset state for creating new users (nextUserID = 1 so next created user gets ID 2)
		nextUserID = 1
	}

	w.WriteHeader(http.StatusOK)
}
