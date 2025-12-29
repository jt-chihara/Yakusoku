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

// In-memory user database
var users = map[int]User{
	1: {ID: 1, Name: "Alice", Email: "alice@example.com"},
	2: {ID: 2, Name: "Bob", Email: "bob@example.com"},
	3: {ID: 3, Name: "Charlie", Email: "charlie@example.com"},
}

func main() {
	mux := http.NewServeMux()

	// GET /users/:id - Get user by ID
	mux.HandleFunc("GET /users/{id}", handleGetUser)

	// GET /users - List all users
	mux.HandleFunc("GET /users", handleListUsers)

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
		json.NewEncoder(w).Encode(map[string]string{
			"error": "User not found",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	userList := make([]User, 0, len(users))
	for _, u := range users {
		userList = append(userList, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userList)
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
		users[1] = User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	case strings.Contains(state.State, "user 999 does not exist"):
		delete(users, 999)
	case strings.Contains(state.State, "no users exist"):
		users = make(map[int]User)
	}

	w.WriteHeader(http.StatusOK)
}
