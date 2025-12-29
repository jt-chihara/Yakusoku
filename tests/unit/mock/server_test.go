package mock_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/jt-chihara/yakusoku/internal/contract"
	"github.com/jt-chihara/yakusoku/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockServer_Start(t *testing.T) {
	t.Run("starts and accepts connections", func(t *testing.T) {
		server := mock.NewServer()
		err := server.Start()
		require.NoError(t, err)
		defer server.Stop()

		assert.NotEmpty(t, server.URL())

		resp, err := http.Get(server.URL() + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("returns URL with http scheme", func(t *testing.T) {
		server := mock.NewServer()
		err := server.Start()
		require.NoError(t, err)
		defer server.Stop()

		assert.Contains(t, server.URL(), "http://")
	})
}

func TestMockServer_Stop(t *testing.T) {
	t.Run("stops gracefully", func(t *testing.T) {
		server := mock.NewServer()
		err := server.Start()
		require.NoError(t, err)

		url := server.URL()
		err = server.Stop()
		require.NoError(t, err)

		// Server should no longer accept connections
		_, err = http.Get(url + "/health")
		assert.Error(t, err)
	})
}

func TestMockServer_RegisterInteraction(t *testing.T) {
	t.Run("registers and matches interaction", func(t *testing.T) {
		server := mock.NewServer()
		err := server.Start()
		require.NoError(t, err)
		defer server.Stop()

		interaction := contract.Interaction{
			Description: "a request for user 1",
			Request: contract.Request{
				Method: "GET",
				Path:   "/users/1",
			},
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
		}

		server.RegisterInteraction(&interaction)

		resp, err := http.Get(server.URL() + "/users/1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "John Doe")
	})

	t.Run("returns 500 for unregistered path", func(t *testing.T) {
		server := mock.NewServer()
		err := server.Start()
		require.NoError(t, err)
		defer server.Stop()

		resp, err := http.Get(server.URL() + "/unknown")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("registers multiple interactions", func(t *testing.T) {
		server := mock.NewServer()
		err := server.Start()
		require.NoError(t, err)
		defer server.Stop()

		server.RegisterInteraction(&contract.Interaction{
			Description: "get user 1",
			Request:     contract.Request{Method: "GET", Path: "/users/1"},
			Response:    contract.Response{Status: 200, Body: map[string]interface{}{"id": float64(1)}},
		})
		server.RegisterInteraction(&contract.Interaction{
			Description: "get user 2",
			Request:     contract.Request{Method: "GET", Path: "/users/2"},
			Response:    contract.Response{Status: 200, Body: map[string]interface{}{"id": float64(2)}},
		})

		resp1, err := http.Get(server.URL() + "/users/1")
		require.NoError(t, err)
		defer resp1.Body.Close()
		assert.Equal(t, 200, resp1.StatusCode)

		resp2, err := http.Get(server.URL() + "/users/2")
		require.NoError(t, err)
		defer resp2.Body.Close()
		assert.Equal(t, 200, resp2.StatusCode)
	})
}

func TestMockServer_ClearInteractions(t *testing.T) {
	t.Run("clears all registered interactions", func(t *testing.T) {
		server := mock.NewServer()
		err := server.Start()
		require.NoError(t, err)
		defer server.Stop()

		server.RegisterInteraction(&contract.Interaction{
			Description: "get user 1",
			Request:     contract.Request{Method: "GET", Path: "/users/1"},
			Response:    contract.Response{Status: 200},
		})

		// Verify interaction works
		resp, err := http.Get(server.URL() + "/users/1")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)

		// Clear and verify it's gone
		server.ClearInteractions()

		resp, err = http.Get(server.URL() + "/users/1")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestMockServer_RecordedInteractions(t *testing.T) {
	t.Run("records matched interactions", func(t *testing.T) {
		server := mock.NewServer()
		err := server.Start()
		require.NoError(t, err)
		defer server.Stop()

		server.RegisterInteraction(&contract.Interaction{
			Description: "get user 1",
			Request:     contract.Request{Method: "GET", Path: "/users/1"},
			Response:    contract.Response{Status: 200},
		})

		_, err = http.Get(server.URL() + "/users/1")
		require.NoError(t, err)

		recorded := server.RecordedInteractions()
		assert.Len(t, recorded, 1)
		assert.Equal(t, "get user 1", recorded[0].Description)
	})
}
