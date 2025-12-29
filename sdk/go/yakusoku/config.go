// Package yakusoku provides a Go SDK for consumer-driven contract testing.
package yakusoku

// Config holds configuration for a Pact instance.
type Config struct {
	Consumer string
	Provider string
	PactDir  string
}
