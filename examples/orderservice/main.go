// OrderService is a sample Consumer application for contract testing demonstration.
//
// Run with: go run ./examples/orderservice
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// UserClient is a client for the UserService API.
type UserClient struct {
	BaseURL string
}

// User represents a user from UserService.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetUser fetches a user by ID from UserService.
func (c *UserClient) GetUser(id int) (*User, error) {
	resp, err := http.Get(fmt.Sprintf("%s/users/%d", c.BaseURL, id))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

// Order represents an order in the system.
type Order struct {
	ID       int     `json:"id"`
	UserID   int     `json:"userId"`
	UserName string  `json:"userName"`
	Product  string  `json:"product"`
	Total    float64 `json:"total"`
}

var userClient *UserClient

func main() {
	userClient = &UserClient{
		BaseURL: "http://localhost:8080", // UserService URL
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /orders/{id}", handleGetOrder)

	port := "8081"
	fmt.Printf("OrderService (Consumer) starting on http://localhost:%s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  GET /orders/{id} - Get order by ID (fetches user from UserService)")
	fmt.Println("")
	fmt.Println("Make sure UserService is running on http://localhost:8080")

	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handleGetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int
	fmt.Sscanf(idStr, "%d", &id)

	// Sample order data
	orderData := map[int]struct {
		UserID  int
		Product string
		Total   float64
	}{
		1: {UserID: 1, Product: "Laptop", Total: 999.99},
		2: {UserID: 2, Product: "Phone", Total: 599.99},
	}

	order, exists := orderData[id]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Order not found"})
		return
	}

	// Fetch user from UserService
	user, err := userClient.GetUser(order.UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	response := Order{
		ID:       id,
		UserID:   order.UserID,
		UserName: user.Name,
		Product:  order.Product,
		Total:    order.Total,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
