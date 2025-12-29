package verifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jt-chihara/yakusoku/internal/contract"
)

// ProviderStates handles provider state setup.
type ProviderStates struct {
	setupURL string
	client   *http.Client
}

// NewProviderStates creates a new ProviderStates.
func NewProviderStates(setupURL string) *ProviderStates {
	return &ProviderStates{
		setupURL: setupURL,
		client:   &http.Client{},
	}
}

// Setup sets up a single provider state.
func (ps *ProviderStates) Setup(state string, params map[string]interface{}) error {
	if ps.setupURL == "" {
		return nil
	}

	body := map[string]interface{}{
		"state": state,
	}
	if params != nil {
		body["params"] = params
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal provider state: %w", err)
	}

	resp, err := ps.client.Post(ps.setupURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to call provider states setup: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("provider states setup failed with status %d", resp.StatusCode)
	}

	return nil
}

// SetupMultiple sets up multiple provider states (v3).
func (ps *ProviderStates) SetupMultiple(states []contract.ProviderState) error {
	for _, state := range states {
		if err := ps.Setup(state.Name, state.Params); err != nil {
			return err
		}
	}
	return nil
}
