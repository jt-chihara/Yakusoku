// Package contract provides types and utilities for Pact v3 contract files.
package contract

// Contract represents a Pact v3 contract file.
type Contract struct {
	Consumer     Pacticipant   `json:"consumer"`
	Provider     Pacticipant   `json:"provider"`
	Interactions []Interaction `json:"interactions"`
	Messages     []Message     `json:"messages,omitempty"`
	Metadata     Metadata      `json:"metadata"`
}

// Pacticipant represents a service participating in a contract (consumer or provider).
type Pacticipant struct {
	Name string `json:"name"`
}

// Interaction represents a single request/response pair.
type Interaction struct {
	Description    string          `json:"description"`
	ProviderState  string          `json:"providerState,omitempty"`
	ProviderStates []ProviderState `json:"providerStates,omitempty"`
	Request        Request         `json:"request"`
	Response       Response        `json:"response"`
}

// ProviderState represents a provider state with optional parameters (Pact v3).
type ProviderState struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// Request represents an expected HTTP request.
type Request struct {
	Method        string                 `json:"method"`
	Path          string                 `json:"path"`
	Query         map[string][]string    `json:"query,omitempty"`
	Headers       map[string]interface{} `json:"headers,omitempty"`
	Body          interface{}            `json:"body,omitempty"`
	MatchingRules MatchingRules          `json:"matchingRules,omitempty"`
	Generators    Generators             `json:"generators,omitempty"`
}

// Response represents an expected HTTP response.
type Response struct {
	Status        int                    `json:"status"`
	Headers       map[string]interface{} `json:"headers,omitempty"`
	Body          interface{}            `json:"body,omitempty"`
	MatchingRules MatchingRules          `json:"matchingRules,omitempty"`
	Generators    Generators             `json:"generators,omitempty"`
}

// MatchingRules defines matching rules for different parts of request/response.
type MatchingRules struct {
	Body    map[string]MatcherSet `json:"body,omitempty"`
	Headers map[string]MatcherSet `json:"headers,omitempty"`
	Path    MatcherSet            `json:"path,omitempty"`
	Query   map[string]MatcherSet `json:"query,omitempty"`
}

// MatcherSet is a set of matchers with an optional combine strategy.
type MatcherSet struct {
	Matchers []Matcher `json:"matchers"`
	Combine  string    `json:"combine,omitempty"`
}

// Matcher represents a single matching rule.
type Matcher struct {
	Match string      `json:"match"`
	Regex string      `json:"regex,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Min   *int        `json:"min,omitempty"`
	Max   *int        `json:"max,omitempty"`
}

// Generators defines value generators (Pact v3).
type Generators struct {
	Body    map[string]Generator `json:"body,omitempty"`
	Headers map[string]Generator `json:"headers,omitempty"`
	Path    Generator            `json:"path,omitempty"`
	Query   map[string]Generator `json:"query,omitempty"`
}

// Generator represents a value generator.
type Generator struct {
	Type   string                 `json:"type"`
	Format string                 `json:"format,omitempty"`
	Min    *int                   `json:"min,omitempty"`
	Max    *int                   `json:"max,omitempty"`
	Digits *int                   `json:"digits,omitempty"`
	Values map[string]interface{} `json:"values,omitempty"`
}

// Metadata contains contract file metadata.
type Metadata struct {
	PactSpecification PactSpec `json:"pactSpecification"`
	Client            *Client  `json:"client,omitempty"`
}

// PactSpec contains Pact specification version.
type PactSpec struct {
	Version string `json:"version"`
}

// Client contains information about the tool that generated the contract.
type Client struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Message represents an async message (Pact v3).
type Message struct {
	Description    string                 `json:"description"`
	ProviderStates []ProviderState        `json:"providerStates,omitempty"`
	Contents       interface{}            `json:"contents"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	MatchingRules  MatchingRules          `json:"matchingRules,omitempty"`
	Generators     Generators             `json:"generators,omitempty"`
}
