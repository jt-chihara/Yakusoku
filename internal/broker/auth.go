package broker

import (
	"net/http"
	"strings"
)

// AuthMiddleware provides Bearer token authentication
type AuthMiddleware struct {
	token   string
	handler http.Handler
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(token string, handler http.Handler) *AuthMiddleware {
	return &AuthMiddleware{
		token:   token,
		handler: handler,
	}
}

// ServeHTTP implements http.Handler
func (m *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Expect "Bearer <token>" format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	if parts[1] != m.token {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	m.handler.ServeHTTP(w, r)
}

// WrapWithAuth wraps a handler with authentication
func WrapWithAuth(token string, handler http.Handler) http.Handler {
	return NewAuthMiddleware(token, handler)
}
