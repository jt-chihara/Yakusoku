package mock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/jt-chihara/yakusoku/internal/contract"
)

// Handler is an HTTP handler that matches requests against registered interactions.
type Handler struct {
	mu           sync.RWMutex
	interactions []contract.Interaction
	recorded     []contract.Interaction
}

// NewHandler creates a new mock handler.
func NewHandler() *Handler {
	return &Handler{
		interactions: make([]contract.Interaction, 0),
		recorded:     make([]contract.Interaction, 0),
	}
}

// RegisterInteraction registers an interaction.
func (h *Handler) RegisterInteraction(i contract.Interaction) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.interactions = append(h.interactions, i)
}

// ClearInteractions clears all interactions.
func (h *Handler) ClearInteractions() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.interactions = make([]contract.Interaction, 0)
	h.recorded = make([]contract.Interaction, 0)
}

// ServeHTTP handles HTTP requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Find matching interaction
	for _, interaction := range h.interactions {
		if h.matchRequest(r, interaction.Request) {
			h.recorded = append(h.recorded, interaction)
			h.writeResponse(w, interaction.Response)
			return
		}
	}

	// No matching interaction found
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `{"error": "no matching interaction found for %s %s"}`, r.Method, r.URL.Path)
}

// RecordedInteractions returns recorded interactions.
func (h *Handler) RecordedInteractions() []contract.Interaction {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]contract.Interaction, len(h.recorded))
	copy(result, h.recorded)
	return result
}

func (h *Handler) matchRequest(r *http.Request, expected contract.Request) bool {
	// Match method
	if !strings.EqualFold(r.Method, expected.Method) {
		return false
	}

	// Match path
	if r.URL.Path != expected.Path {
		return false
	}

	// Match query params if specified
	if len(expected.Query) > 0 {
		for key, values := range expected.Query {
			actualValues := r.URL.Query()[key]
			if !sliceEqual(values, actualValues) {
				return false
			}
		}
	}

	// Match headers if specified
	if len(expected.Headers) > 0 {
		for key, value := range expected.Headers {
			actualValue := r.Header.Get(key)
			if actualValue != fmt.Sprintf("%v", value) {
				return false
			}
		}
	}

	return true
}

func (h *Handler) writeResponse(w http.ResponseWriter, resp contract.Response) {
	// Set headers
	for key, value := range resp.Headers {
		w.Header().Set(key, fmt.Sprintf("%v", value))
	}

	// Write status
	w.WriteHeader(resp.Status)

	// Write body
	if resp.Body != nil {
		switch body := resp.Body.(type) {
		case string:
			io.WriteString(w, body)
		default:
			json.NewEncoder(w).Encode(body)
		}
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
