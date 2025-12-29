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

		// Check if this is a static file request (has known file extension)
		ext := filepath.Ext(path)
		isStaticFile := false
		switch ext {
		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			isStaticFile = true
		case ".js":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			isStaticFile = true
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			isStaticFile = true
		case ".json":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			isStaticFile = true
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
			isStaticFile = true
		case ".png":
			w.Header().Set("Content-Type", "image/png")
			isStaticFile = true
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
			isStaticFile = true
		}

		// For SPA routing, serve index.html for paths that are not static files
		if !isStaticFile {
			r.URL.Path = "/"
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			r.URL.Path = path
		}

		fileServer.ServeHTTP(w, r)
	})
}
