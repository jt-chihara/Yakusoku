package contract

import (
	"encoding/json"
	"fmt"
	"os"
)

// Parser parses contract files.
type Parser struct{}

// NewParser creates a new Parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile parses a contract file from the given path.
func (p *Parser) ParseFile(path string) (*Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract file: %w", err)
	}
	return p.ParseBytes(data)
}

// ParseBytes parses a contract from raw bytes.
func (p *Parser) ParseBytes(data []byte) (*Contract, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("failed to parse contract JSON: empty data")
	}

	var c Contract
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to parse contract JSON: %w", err)
	}
	return &c, nil
}
