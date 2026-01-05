package broker_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jt-chihara/yakusoku/internal/broker"
)

func TestAuthMiddleware(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	t.Run("rejects request without Authorization header", func(t *testing.T) {
		middleware := broker.NewAuthMiddleware("secret-token", dummyHandler)
		server := httptest.NewServer(middleware)
		defer server.Close()

		resp, err := http.Get(server.URL)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("rejects request with invalid token", func(t *testing.T) {
		middleware := broker.NewAuthMiddleware("secret-token", dummyHandler)
		server := httptest.NewServer(middleware)
		defer server.Close()

		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		req.Header.Set("Authorization", "Bearer wrong-token")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("allows request with valid token", func(t *testing.T) {
		middleware := broker.NewAuthMiddleware("secret-token", dummyHandler)
		server := httptest.NewServer(middleware)
		defer server.Close()

		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		req.Header.Set("Authorization", "Bearer secret-token")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("rejects request with invalid format", func(t *testing.T) {
		middleware := broker.NewAuthMiddleware("secret-token", dummyHandler)
		server := httptest.NewServer(middleware)
		defer server.Close()

		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		req.Header.Set("Authorization", "Basic abc123")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("is case insensitive for Bearer keyword", func(t *testing.T) {
		middleware := broker.NewAuthMiddleware("secret-token", dummyHandler)
		server := httptest.NewServer(middleware)
		defer server.Close()

		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		req.Header.Set("Authorization", "bearer secret-token")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestWrapWithAuth(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("wraps handler with auth middleware", func(t *testing.T) {
		handler := broker.WrapWithAuth("secret", dummyHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		// Without token should fail
		resp, err := http.Get(server.URL)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("allows request with valid token", func(t *testing.T) {
		handler := broker.WrapWithAuth("secret", dummyHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		req.Header.Set("Authorization", "Bearer secret")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
