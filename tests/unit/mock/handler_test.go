package mock_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/jt-chihara/yakusoku/internal/contract"
	"github.com/jt-chihara/yakusoku/internal/mock"
)

func TestHandler_MatchRequest(t *testing.T) {
	t.Run("matches GET request by method and path", func(t *testing.T) {
		handler := mock.NewHandler()
		handler.RegisterInteraction(contract.Interaction{
			Description: "get user",
			Request:     contract.Request{Method: "GET", Path: "/users/1"},
			Response:    contract.Response{Status: 200},
		})

		req := httptest.NewRequest("GET", "/users/1", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("matches POST request with body", func(t *testing.T) {
		handler := mock.NewHandler()
		handler.RegisterInteraction(contract.Interaction{
			Description: "create user",
			Request: contract.Request{
				Method: "POST",
				Path:   "/users",
				Body:   map[string]interface{}{"name": "John"},
			},
			Response: contract.Response{
				Status: 201,
				Body:   map[string]interface{}{"id": float64(1), "name": "John"},
			},
		})

		body := bytes.NewBufferString(`{"name": "John"}`)
		req := httptest.NewRequest("POST", "/users", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, 201, w.Code)
	})

	t.Run("does not match different method", func(t *testing.T) {
		handler := mock.NewHandler()
		handler.RegisterInteraction(contract.Interaction{
			Description: "get user",
			Request:     contract.Request{Method: "GET", Path: "/users/1"},
			Response:    contract.Response{Status: 200},
		})

		req := httptest.NewRequest("POST", "/users/1", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("does not match different path", func(t *testing.T) {
		handler := mock.NewHandler()
		handler.RegisterInteraction(contract.Interaction{
			Description: "get user",
			Request:     contract.Request{Method: "GET", Path: "/users/1"},
			Response:    contract.Response{Status: 200},
		})

		req := httptest.NewRequest("GET", "/users/2", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("matches request with headers", func(t *testing.T) {
		handler := mock.NewHandler()
		handler.RegisterInteraction(contract.Interaction{
			Description: "get user with auth",
			Request: contract.Request{
				Method: "GET",
				Path:   "/users/1",
				Headers: map[string]interface{}{
					"Authorization": "Bearer token123",
				},
			},
			Response: contract.Response{Status: 200},
		})

		req := httptest.NewRequest("GET", "/users/1", nil)
		req.Header.Set("Authorization", "Bearer token123")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("matches request with query params", func(t *testing.T) {
		handler := mock.NewHandler()
		handler.RegisterInteraction(contract.Interaction{
			Description: "search users",
			Request: contract.Request{
				Method: "GET",
				Path:   "/users",
				Query:  map[string][]string{"status": {"active"}},
			},
			Response: contract.Response{Status: 200},
		})

		req := httptest.NewRequest("GET", "/users?status=active", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})
}

func TestHandler_ResponseBody(t *testing.T) {
	t.Run("returns JSON body", func(t *testing.T) {
		handler := mock.NewHandler()
		handler.RegisterInteraction(contract.Interaction{
			Description: "get user",
			Request:     contract.Request{Method: "GET", Path: "/users/1"},
			Response: contract.Response{
				Status: 200,
				Headers: map[string]interface{}{
					"Content-Type": "application/json",
				},
				Body: map[string]interface{}{
					"id":   float64(1),
					"name": "John Doe",
				},
			},
		})

		req := httptest.NewRequest("GET", "/users/1", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		body, _ := io.ReadAll(w.Body)
		assert.Contains(t, string(body), `"id":1`)
		assert.Contains(t, string(body), `"name":"John Doe"`)
	})

	t.Run("returns array body", func(t *testing.T) {
		handler := mock.NewHandler()
		handler.RegisterInteraction(contract.Interaction{
			Description: "list users",
			Request:     contract.Request{Method: "GET", Path: "/users"},
			Response: contract.Response{
				Status: 200,
				Body: []interface{}{
					map[string]interface{}{"id": float64(1)},
					map[string]interface{}{"id": float64(2)},
				},
			},
		})

		req := httptest.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		body, _ := io.ReadAll(w.Body)
		assert.Contains(t, string(body), `"id":1`)
		assert.Contains(t, string(body), `"id":2`)
	})

	t.Run("returns string body", func(t *testing.T) {
		handler := mock.NewHandler()
		handler.RegisterInteraction(contract.Interaction{
			Description: "get text",
			Request:     contract.Request{Method: "GET", Path: "/text"},
			Response: contract.Response{
				Status: 200,
				Headers: map[string]interface{}{
					"Content-Type": "text/plain",
				},
				Body: "Hello, World!",
			},
		})

		req := httptest.NewRequest("GET", "/text", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		body, _ := io.ReadAll(w.Body)
		assert.Equal(t, "Hello, World!", string(body))
	})
}

func TestHandler_ResponseHeaders(t *testing.T) {
	t.Run("sets response headers", func(t *testing.T) {
		handler := mock.NewHandler()
		handler.RegisterInteraction(contract.Interaction{
			Description: "get user",
			Request:     contract.Request{Method: "GET", Path: "/users/1"},
			Response: contract.Response{
				Status: 200,
				Headers: map[string]interface{}{
					"Content-Type":  "application/json",
					"X-Custom-Header": "custom-value",
				},
			},
		})

		req := httptest.NewRequest("GET", "/users/1", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "custom-value", w.Header().Get("X-Custom-Header"))
	})
}

func TestHandler_MatchFirstMatchingInteraction(t *testing.T) {
	t.Run("uses first matching interaction when multiple match", func(t *testing.T) {
		handler := mock.NewHandler()
		handler.RegisterInteraction(contract.Interaction{
			Description: "first match",
			Request:     contract.Request{Method: "GET", Path: "/test"},
			Response:    contract.Response{Status: 200},
		})
		handler.RegisterInteraction(contract.Interaction{
			Description: "second match",
			Request:     contract.Request{Method: "GET", Path: "/test"},
			Response:    contract.Response{Status: 201},
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})
}

func TestHandler_UnmatchedRequest(t *testing.T) {
	t.Run("returns error for unmatched request", func(t *testing.T) {
		handler := mock.NewHandler()

		req := httptest.NewRequest("GET", "/unknown", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		body, err := io.ReadAll(w.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "no matching interaction")
	})
}
