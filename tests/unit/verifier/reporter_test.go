package verifier_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jt-chihara/yakusoku/internal/verifier"
)

func TestReporter_Report(t *testing.T) {
	t.Run("reports successful verification", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: true,
			Interactions: []verifier.InteractionResult{
				{
					Description: "get user 1",
					Success:     true,
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "get user 1")
		assert.Contains(t, output, "passed")
	})

	t.Run("reports failed verification", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{
					Description: "get user 1",
					Success:     false,
					Diff:        "expected status 200, got 404",
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "get user 1")
		assert.Contains(t, output, "failed")
		assert.Contains(t, output, "expected status 200")
	})

	t.Run("reports mixed results", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{Description: "interaction 1", Success: true},
				{Description: "interaction 2", Success: false, Diff: "body mismatch"},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "interaction 1")
		assert.Contains(t, output, "interaction 2")
	})

	t.Run("reports summary", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: true,
			Interactions: []verifier.InteractionResult{
				{Description: "interaction 1", Success: true},
				{Description: "interaction 2", Success: true},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "2")
	})

	t.Run("reports connection error", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{
					Description: "get user 1",
					Success:     false,
					Error:       "connection refused",
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "connection refused")
	})
}

func TestReporter_VerboseMode(t *testing.T) {
	t.Run("shows request details in verbose mode", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)
		reporter.SetVerbose(true)

		result := verifier.VerificationResult{
			Success: true,
			Interactions: []verifier.InteractionResult{
				{
					Description:    "get user 1",
					Success:        true,
					RequestMethod:  "GET",
					RequestPath:    "/users/1",
					ResponseStatus: 200,
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "GET")
		assert.Contains(t, output, "/users/1")
		assert.Contains(t, output, "200")
	})
}

func TestReporter_FailureDetails(t *testing.T) {
	t.Run("shows provider state in failure details", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{
					Description:   "get user 1",
					Success:       false,
					ProviderState: "user 1 exists",
					RequestMethod: "GET",
					RequestPath:   "/users/1",
					Diff:          "status mismatch",
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "Provider State")
		assert.Contains(t, output, "user 1 exists")
	})

	t.Run("shows request headers in failure details", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{
					Description:   "get user 1",
					Success:       false,
					RequestMethod: "GET",
					RequestPath:   "/users/1",
					RequestHeaders: map[string]interface{}{
						"Authorization": "Bearer token123",
						"Content-Type":  "application/json",
					},
					Diff: "status mismatch",
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "Headers")
		assert.Contains(t, output, "Authorization")
	})

	t.Run("shows request body in failure details", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{
					Description:   "create user",
					Success:       false,
					RequestMethod: "POST",
					RequestPath:   "/users",
					RequestBody: map[string]interface{}{
						"name":  "John",
						"email": "john@example.com",
					},
					Diff: "status mismatch",
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "Body")
		assert.Contains(t, output, "John")
	})

	t.Run("shows expected headers in failure details", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{
					Description:   "get user 1",
					Success:       false,
					RequestMethod: "GET",
					RequestPath:   "/users/1",
					ExpectedStatus: 200,
					ExpectedHeaders: map[string]interface{}{
						"Content-Type": "application/json",
					},
					Diff: "header mismatch",
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "Expected Response")
		assert.Contains(t, output, "Content-Type")
	})

	t.Run("shows expected body in failure details", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{
					Description:    "get user 1",
					Success:        false,
					RequestMethod:  "GET",
					RequestPath:    "/users/1",
					ExpectedStatus: 200,
					ExpectedBody: map[string]interface{}{
						"id":   float64(1),
						"name": "John",
					},
					Diff: "body mismatch",
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "Expected Response")
		assert.Contains(t, output, "John")
	})

	t.Run("shows actual headers in failure details", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{
					Description:    "get user 1",
					Success:        false,
					RequestMethod:  "GET",
					RequestPath:    "/users/1",
					ResponseStatus: 200,
					ActualHeaders: map[string]string{
						"Content-Type": "text/plain",
					},
					Diff: "header mismatch",
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "Actual Response")
		assert.Contains(t, output, "text/plain")
	})

	t.Run("shows actual body in failure details", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{
					Description:    "get user 1",
					Success:        false,
					RequestMethod:  "GET",
					RequestPath:    "/users/1",
					ResponseStatus: 200,
					ActualBody: map[string]interface{}{
						"id":   float64(1),
						"name": "Jane",
					},
					Diff: "body mismatch",
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "Actual Response")
		assert.Contains(t, output, "Jane")
	})

	t.Run("shows raw body when actual body is nil but raw exists", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{
					Description:    "get user 1",
					Success:        false,
					RequestMethod:  "GET",
					RequestPath:    "/users/1",
					ResponseStatus: 500,
					ActualBody:     nil,
					ActualBodyRaw:  "Internal Server Error",
					Diff:           "status mismatch",
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "Body (raw)")
		assert.Contains(t, output, "Internal Server Error")
	})

	t.Run("shows full failure details with all fields", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := verifier.NewReporter(&buf)

		result := verifier.VerificationResult{
			Success: false,
			Interactions: []verifier.InteractionResult{
				{
					Description:   "create user",
					Success:       false,
					ProviderState: "no users exist",
					RequestMethod: "POST",
					RequestPath:   "/users",
					RequestHeaders: map[string]interface{}{
						"Content-Type": "application/json",
					},
					RequestBody: map[string]interface{}{
						"name": "John",
					},
					ExpectedStatus: 201,
					ExpectedHeaders: map[string]interface{}{
						"Content-Type": "application/json",
					},
					ExpectedBody: map[string]interface{}{
						"id":   float64(1),
						"name": "John",
					},
					ResponseStatus: 500,
					ActualHeaders: map[string]string{
						"Content-Type": "text/plain",
					},
					ActualBody:    nil,
					ActualBodyRaw: "Server Error",
					Diff:          "status: expected 201, got 500",
					Error:         "server returned error",
				},
			},
		}

		reporter.Report(&result)
		output := buf.String()

		assert.Contains(t, output, "create user")
		assert.Contains(t, output, "failed")
		assert.Contains(t, output, "Provider State")
		assert.Contains(t, output, "no users exist")
		assert.Contains(t, output, "Request")
		assert.Contains(t, output, "POST")
		assert.Contains(t, output, "/users")
		assert.Contains(t, output, "Expected Response")
		assert.Contains(t, output, "201")
		assert.Contains(t, output, "Actual Response")
		assert.Contains(t, output, "500")
		assert.Contains(t, output, "Server Error")
	})
}
