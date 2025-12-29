package yakusoku

import (
	"github.com/jt-chihara/yakusoku/internal/contract"
	"github.com/jt-chihara/yakusoku/internal/mock"
)

// Request represents an expected HTTP request.
type Request struct {
	Method  string
	Path    string
	Query   map[string][]string
	Headers map[string]string
	Body    interface{}
}

// Response represents an expected HTTP response.
type Response struct {
	Status  int
	Headers map[string]string
	Body    interface{}
}

// Pact is the main struct for defining consumer contracts.
type Pact struct {
	config       Config
	interactions []contract.Interaction
	server       *mock.Server
	current      *interactionBuilder
}

type interactionBuilder struct {
	providerState string
	description   string
	request       contract.Request
	response      contract.Response
}

// NewPact creates a new Pact instance.
func NewPact(config Config) *Pact {
	return &Pact{
		config:       config,
		interactions: make([]contract.Interaction, 0),
		server:       mock.NewServer(),
	}
}

// Consumer returns the consumer name.
func (p *Pact) Consumer() string {
	return p.config.Consumer
}

// Provider returns the provider name.
func (p *Pact) Provider() string {
	return p.config.Provider
}

// Given sets the provider state.
func (p *Pact) Given(state string) *Pact {
	if p.current == nil {
		p.current = &interactionBuilder{}
	}
	p.current.providerState = state
	return p
}

// UponReceiving sets the interaction description.
func (p *Pact) UponReceiving(description string) *Pact {
	if p.current == nil {
		p.current = &interactionBuilder{}
	}
	p.current.description = description
	return p
}

// WithRequest sets the expected request.
func (p *Pact) WithRequest(request Request) *Pact {
	if p.current == nil {
		p.current = &interactionBuilder{}
	}
	p.current.request = contract.Request{
		Method:  request.Method,
		Path:    request.Path,
		Query:   request.Query,
		Headers: toInterfaceMap(request.Headers),
		Body:    request.Body,
	}
	return p
}

// WillRespondWith sets the expected response and finalizes the interaction.
func (p *Pact) WillRespondWith(response Response) *Pact {
	if p.current == nil {
		p.current = &interactionBuilder{}
	}
	p.current.response = contract.Response{
		Status:  response.Status,
		Headers: toInterfaceMap(response.Headers),
		Body:    response.Body,
	}

	// Finalize and store the interaction
	interaction := contract.Interaction{
		Description:   p.current.description,
		ProviderState: p.current.providerState,
		Request:       p.current.request,
		Response:      p.current.response,
	}
	p.interactions = append(p.interactions, interaction)

	// Reset current builder
	p.current = nil

	return p
}

// HasInteractions returns true if there are registered interactions.
func (p *Pact) HasInteractions() bool {
	return len(p.interactions) > 0
}

// ServerURL returns the mock server URL.
func (p *Pact) ServerURL() string {
	return p.server.URL()
}

// Verify runs the verification callback with the mock server.
func (p *Pact) Verify(callback func() error) error {
	// Register all interactions with the mock server
	for _, interaction := range p.interactions {
		p.server.RegisterInteraction(interaction)
	}

	// Start the server
	if err := p.server.Start(); err != nil {
		return err
	}

	// Run the callback
	if err := callback(); err != nil {
		return err
	}

	// Write the contract file
	writer := contract.NewWriter()
	c := contract.Contract{
		Consumer:     contract.Pacticipant{Name: p.config.Consumer},
		Provider:     contract.Pacticipant{Name: p.config.Provider},
		Interactions: p.interactions,
		Metadata: contract.Metadata{
			PactSpecification: contract.PactSpec{Version: "3.0.0"},
			Client: &contract.Client{
				Name:    "yakusoku",
				Version: "0.1.0",
			},
		},
	}

	if _, err := writer.WriteToDir(c, p.config.PactDir); err != nil {
		return err
	}

	return nil
}

// Teardown stops the mock server and cleans up.
func (p *Pact) Teardown() {
	if p.server != nil {
		p.server.Stop()
	}
}

func toInterfaceMap(m map[string]string) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
