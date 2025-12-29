package verifier_test

import (
	"bytes"
	"testing"

	"github.com/jt-chihara/yakusoku/internal/verifier"
	"github.com/stretchr/testify/assert"
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
