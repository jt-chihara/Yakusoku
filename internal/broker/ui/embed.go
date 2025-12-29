package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// Handler returns an http.Handler that serves the embedded UI files
func Handler() http.Handler {
	// Strip the "dist" prefix from the embedded filesystem
	subFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(subFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Remove /ui prefix if present
		if strings.HasPrefix(path, "/ui") {
			path = strings.TrimPrefix(path, "/ui")
			if path == "" {
				path = "/"
			}
		}

		// Set correct Content-Type based on file extension
		ext := filepath.Ext(path)
		switch ext {
		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case ".json":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
		}

		// For SPA routing, serve index.html for paths that don't match a file
		if path != "/" && !strings.Contains(path, ".") {
			r.URL.Path = "/"
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			r.URL.Path = path
		}

		fileServer.ServeHTTP(w, r)
	})
}
