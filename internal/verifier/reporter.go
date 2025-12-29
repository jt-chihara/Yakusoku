package verifier

import (
	"fmt"
	"io"
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
		} else {
			failed++
			fmt.Fprintf(r.w, "  ✗ %s - failed\n", ir.Description)
			if ir.Error != "" {
				fmt.Fprintf(r.w, "    Error: %s\n", ir.Error)
			}
			if ir.Diff != "" {
				fmt.Fprintf(r.w, "    Diff: %s\n", ir.Diff)
			}
		}

		if r.verbose {
			if ir.RequestMethod != "" {
				fmt.Fprintf(r.w, "    Request: %s %s\n", ir.RequestMethod, ir.RequestPath)
			}
			if ir.ResponseStatus != 0 {
				fmt.Fprintf(r.w, "    Response: %d\n", ir.ResponseStatus)
			}
		}
	}

	fmt.Fprintf(r.w, "\nSummary: %d passed, %d failed (total: %d)\n", passed, failed, len(result.Interactions))
}
