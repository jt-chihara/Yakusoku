package verifier

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Reporter reports verification results.
type Reporter struct {
	w       io.Writer
	verbose bool
}

// NewReporter creates a new Reporter.
func NewReporter(w io.Writer) *Reporter {
	return &Reporter{w: w}
}

// SetVerbose sets verbose mode.
func (r *Reporter) SetVerbose(verbose bool) {
	r.verbose = verbose
}

// Report writes the verification result to the output.
func (r *Reporter) Report(result *VerificationResult) {
	passed := 0
	failed := 0

	for _, ir := range result.Interactions {
		if ir.Success {
			passed++
			fmt.Fprintf(r.w, "  ✓ %s - passed\n", ir.Description)
			if r.verbose {
				r.printRequestInfo(&ir)
			}
		} else {
			failed++
			fmt.Fprintf(r.w, "  ✗ %s - failed\n", ir.Description)
			r.printFailureDetails(&ir)
		}
	}

	fmt.Fprintf(r.w, "\nSummary: %d passed, %d failed (total: %d)\n", passed, failed, len(result.Interactions))
}

func (r *Reporter) printRequestInfo(ir *InteractionResult) {
	if ir.RequestMethod != "" {
		fmt.Fprintf(r.w, "    Request: %s %s\n", ir.RequestMethod, ir.RequestPath)
	}
	if ir.ResponseStatus != 0 {
		fmt.Fprintf(r.w, "    Response: %d\n", ir.ResponseStatus)
	}
}

func (r *Reporter) printFailureDetails(ir *InteractionResult) {
	// Always show the diff/error first
	if ir.Error != "" {
		fmt.Fprintf(r.w, "    Error: %s\n", ir.Error)
	}
	if ir.Diff != "" {
		fmt.Fprintf(r.w, "    Diff: %s\n", ir.Diff)
	}

	fmt.Fprintf(r.w, "\n")

	// Provider State
	if ir.ProviderState != "" {
		fmt.Fprintf(r.w, "    Provider State: %s\n", ir.ProviderState)
	}

	// Request details
	fmt.Fprintf(r.w, "    Request:\n")
	fmt.Fprintf(r.w, "      %s %s\n", ir.RequestMethod, ir.RequestPath)
	if len(ir.RequestHeaders) > 0 {
		fmt.Fprintf(r.w, "      Headers:\n")
		for k, v := range ir.RequestHeaders {
			fmt.Fprintf(r.w, "        %s: %v\n", k, v)
		}
	}
	if ir.RequestBody != nil {
		fmt.Fprintf(r.w, "      Body: %s\n", r.formatJSON(ir.RequestBody))
	}

	// Expected response
	fmt.Fprintf(r.w, "    Expected Response:\n")
	fmt.Fprintf(r.w, "      Status: %d\n", ir.ExpectedStatus)
	if len(ir.ExpectedHeaders) > 0 {
		fmt.Fprintf(r.w, "      Headers:\n")
		for k, v := range ir.ExpectedHeaders {
			fmt.Fprintf(r.w, "        %s: %v\n", k, v)
		}
	}
	if ir.ExpectedBody != nil {
		fmt.Fprintf(r.w, "      Body: %s\n", r.formatJSON(ir.ExpectedBody))
	}

	// Actual response
	fmt.Fprintf(r.w, "    Actual Response:\n")
	fmt.Fprintf(r.w, "      Status: %d\n", ir.ResponseStatus)
	if len(ir.ActualHeaders) > 0 {
		fmt.Fprintf(r.w, "      Headers:\n")
		for k, v := range ir.ActualHeaders {
			fmt.Fprintf(r.w, "        %s: %s\n", k, v)
		}
	}
	if ir.ActualBody != nil {
		fmt.Fprintf(r.w, "      Body: %s\n", r.formatJSON(ir.ActualBody))
	} else if ir.ActualBodyRaw != "" {
		fmt.Fprintf(r.w, "      Body (raw): %s\n", ir.ActualBodyRaw)
	}

	fmt.Fprintf(r.w, "\n")
}

func (r *Reporter) formatJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "             ", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	// Remove the leading indentation from the first line
	result := string(b)
	result = strings.TrimPrefix(result, "             ")
	return result
}
