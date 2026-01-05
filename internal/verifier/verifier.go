// Package verifier provides contract verification functionality.
package verifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jt-chihara/yakusoku/internal/contract"
)

// Config holds verifier configuration.
type Config struct {
	ProviderBaseURL        string
	ProviderStatesSetupURL string
}

// VerificationResult holds the result of a verification.
type VerificationResult struct {
	Success      bool
	Interactions []InteractionResult
}

// InteractionResult holds the result of verifying a single interaction.
type InteractionResult struct {
	Description    string
	Success        bool
	Diff           string
	Error          string
	RequestMethod  string
	RequestPath    string
	ResponseStatus int
	// Details for error reporting
	ProviderState      string
	ExpectedStatus     int
	ExpectedHeaders    map[string]interface{}
	ExpectedBody       interface{}
	ActualHeaders      map[string]string
	ActualBody         interface{}
	ActualBodyRaw      string
	RequestHeaders     map[string]interface{}
	RequestBody        interface{}
}

// Verifier verifies contracts against a provider.
type Verifier struct {
	config         Config
	client         *http.Client
	comparer       *Comparer
	providerStates *ProviderStates
}

// New creates a new Verifier.
func New(config Config) *Verifier {
	return &Verifier{
		config:         config,
		client:         &http.Client{},
		comparer:       NewComparer(),
		providerStates: NewProviderStates(config.ProviderStatesSetupURL),
	}
}

// Verify verifies a contract against the provider.
func (v *Verifier) Verify(c *contract.Contract) (*VerificationResult, error) {
	result := &VerificationResult{
		Success:      true,
		Interactions: make([]InteractionResult, 0, len(c.Interactions)),
	}

	for i := range c.Interactions {
		ir := v.verifyInteraction(&c.Interactions[i])
		result.Interactions = append(result.Interactions, ir)
		if !ir.Success {
			result.Success = false
		}
	}

	return result, nil
}

func (v *Verifier) verifyInteraction(interaction *contract.Interaction) InteractionResult {
	ir := InteractionResult{
		Description:     interaction.Description,
		RequestMethod:   interaction.Request.Method,
		RequestPath:     interaction.Request.Path,
		ProviderState:   interaction.ProviderState,
		ExpectedStatus:  interaction.Response.Status,
		ExpectedHeaders: interaction.Response.Headers,
		ExpectedBody:    interaction.Response.Body,
		RequestHeaders:  interaction.Request.Headers,
		RequestBody:     interaction.Request.Body,
	}

	// Setup provider states
	if interaction.ProviderState != "" {
		if err := v.providerStates.Setup(interaction.ProviderState, nil); err != nil {
			ir.Error = fmt.Sprintf("failed to setup provider state: %v", err)
			return ir
		}
	}
	if len(interaction.ProviderStates) > 0 {
		if err := v.providerStates.SetupMultiple(interaction.ProviderStates); err != nil {
			ir.Error = fmt.Sprintf("failed to setup provider states: %v", err)
			return ir
		}
	}

	// Make request to provider
	url := v.config.ProviderBaseURL + interaction.Request.Path

	// Prepare request body if present
	var bodyReader io.Reader = http.NoBody
	if interaction.Request.Body != nil {
		bodyBytes, err := json.Marshal(interaction.Request.Body)
		if err != nil {
			ir.Error = fmt.Sprintf("failed to marshal request body: %v", err)
			return ir
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(interaction.Request.Method, url, bodyReader)
	if err != nil {
		ir.Error = fmt.Sprintf("failed to create request: %v", err)
		return ir
	}

	// Add headers
	for key, value := range interaction.Request.Headers {
		req.Header.Set(key, fmt.Sprintf("%v", value))
	}

	resp, err := v.client.Do(req)
	if err != nil {
		ir.Error = fmt.Sprintf("connection error: %v", err)
		return ir
	}
	defer resp.Body.Close()

	ir.ResponseStatus = resp.StatusCode

	// Compare response
	var diffs []string

	// Compare status
	statusResult := v.comparer.CompareStatus(interaction.Response.Status, resp.StatusCode)
	if !statusResult.Match {
		diffs = append(diffs, statusResult.Diff)
	}

	// Compare headers
	actualHeaders := make(map[string]string)
	for key := range resp.Header {
		actualHeaders[key] = resp.Header.Get(key)
	}
	ir.ActualHeaders = actualHeaders

	if interaction.Response.Headers != nil {
		headerResult := v.comparer.CompareHeaders(interaction.Response.Headers, actualHeaders)
		if !headerResult.Match {
			diffs = append(diffs, headerResult.Diff)
		}
	}

	// Read and compare body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ir.Error = fmt.Sprintf("failed to read response body: %v", err)
		return ir
	}
	ir.ActualBodyRaw = string(body)

	if len(body) > 0 {
		var actualBody interface{}
		if err := json.Unmarshal(body, &actualBody); err != nil {
			ir.Error = fmt.Sprintf("failed to parse response body: %v", err)
			return ir
		}
		ir.ActualBody = actualBody
	}

	if interaction.Response.Body != nil {
		bodyResult, err := v.comparer.CompareBody(interaction.Response.Body, ir.ActualBody, interaction.Response.MatchingRules.Body)
		if err != nil {
			ir.Error = fmt.Sprintf("failed to compare body: %v", err)
			return ir
		}
		if !bodyResult.Match {
			diffs = append(diffs, bodyResult.Diff)
		}
	}

	if len(diffs) > 0 {
		ir.Diff = strings.Join(diffs, "; ")
	} else {
		ir.Success = true
	}

	return ir
}
