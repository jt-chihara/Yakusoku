package contract

import (
	"fmt"
	"strings"
)

// Valid HTTP methods.
var validMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"PATCH":   true,
	"HEAD":    true,
	"OPTIONS": true,
}

// Validator validates contract structures.
type Validator struct{}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a complete contract.
func (v *Validator) Validate(c *Contract) error {
	if err := v.validatePacticipant(c.Consumer, "consumer"); err != nil {
		return err
	}
	if err := v.validatePacticipant(c.Provider, "provider"); err != nil {
		return err
	}
	if len(c.Interactions) == 0 {
		return fmt.Errorf("at least one interaction is required")
	}
	for i := range c.Interactions {
		if err := v.validateInteraction(&c.Interactions[i], i); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) validatePacticipant(p Pacticipant, role string) error {
	if p.Name == "" {
		return fmt.Errorf("%s name is required", role)
	}
	if len(p.Name) > 255 {
		return fmt.Errorf("%s name must be 255 characters or less", role)
	}
	return nil
}

func (v *Validator) validateInteraction(i *Interaction, index int) error {
	if i.Description == "" {
		return fmt.Errorf("interaction %d: description is required", index)
	}
	if err := v.ValidateRequest(&i.Request); err != nil {
		return fmt.Errorf("interaction %d: %w", index, err)
	}
	if err := v.ValidateResponse(&i.Response); err != nil {
		return fmt.Errorf("interaction %d: %w", index, err)
	}
	return nil
}

// ValidateRequest validates a request structure.
func (v *Validator) ValidateRequest(r *Request) error {
	method := strings.ToUpper(r.Method)
	if !validMethods[method] {
		return fmt.Errorf("invalid HTTP method: %s", r.Method)
	}
	if r.Path == "" {
		return fmt.Errorf("request path is required")
	}
	if !strings.HasPrefix(r.Path, "/") {
		return fmt.Errorf("request path must start with /")
	}
	return nil
}

// ValidateResponse validates a response structure.
func (v *Validator) ValidateResponse(r *Response) error {
	if r.Status < 100 || r.Status > 599 {
		return fmt.Errorf("invalid HTTP status code: %d (must be 100-599)", r.Status)
	}
	return nil
}
