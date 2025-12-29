// Package mock provides a mock HTTP server for contract testing.
package mock

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/jt-chihara/yakusoku/internal/contract"
)

// Server is a mock HTTP server for testing consumer code.
type Server struct {
	handler  *Handler
	server   *http.Server
	listener net.Listener
	url      string
	mu       sync.RWMutex
}

// NewServer creates a new mock server.
func NewServer() *Server {
	return &Server{
		handler: NewHandler(),
	}
}

// Start starts the mock server.
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a listener on a random available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	s.listener = listener

	// Create server mux with health check
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	mux.Handle("/", s.handler)

	srv := &http.Server{
		Handler: mux,
	}
	s.server = srv
	s.url = fmt.Sprintf("http://%s", listener.Addr().String())

	// Start server in goroutine (capture srv to avoid race with Stop())
	go func() { _ = srv.Serve(listener) }()

	return nil
}

// Stop stops the mock server.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		if err := s.server.Close(); err != nil {
			return fmt.Errorf("failed to stop server: %w", err)
		}
		s.server = nil
	}
	return nil
}

// URL returns the mock server's URL.
func (s *Server) URL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.url
}

// RegisterInteraction registers an interaction with the mock server.
func (s *Server) RegisterInteraction(i contract.Interaction) {
	s.handler.RegisterInteraction(i)
}

// ClearInteractions clears all registered interactions.
func (s *Server) ClearInteractions() {
	s.handler.ClearInteractions()
}

// RecordedInteractions returns all recorded interactions.
func (s *Server) RecordedInteractions() []contract.Interaction {
	return s.handler.RecordedInteractions()
}
