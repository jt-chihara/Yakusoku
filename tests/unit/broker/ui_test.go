package broker_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jt-chihara/yakusoku/internal/broker/ui"
)

func TestUI_Handler(t *testing.T) {
	handler := ui.Handler()
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("serves index.html for root path", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "<!doctype html>")
	})

	t.Run("serves index.html for SPA routes (non-file paths)", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/contracts")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "<!doctype html>")
	})

	t.Run("strips /ui prefix and serves content", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/ui")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
	})

	t.Run("strips /ui prefix for nested routes", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/ui/contracts")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
	})

	t.Run("serves SVG file with correct content type", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/vite.svg")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "image/svg+xml", resp.Header.Get("Content-Type"))
	})

	t.Run("serves CSS file from assets with correct content type", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/assets/index-CWyftKQk.css")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/css; charset=utf-8", resp.Header.Get("Content-Type"))
	})

	t.Run("serves JS file from assets with correct content type", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/assets/index-DWa93a-y.js")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/javascript; charset=utf-8", resp.Header.Get("Content-Type"))
	})

	t.Run("handles /ui/ prefix with trailing slash", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/ui/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("serves index.html for unknown paths (SPA fallback)", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/some/nested/route")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should serve index.html for SPA routing
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

		body, _ := io.ReadAll(resp.Body)
		assert.True(t, strings.Contains(string(body), "<!doctype html>"))
	})
}
