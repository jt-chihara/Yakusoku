// Package cli provides CLI commands for yakusoku.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jt-chihara/yakusoku/internal/contract"
	"github.com/jt-chihara/yakusoku/internal/verifier"
)

type verifyOptions struct {
	providerBaseURL        string
	pactFile               string
	providerStatesSetupURL string
	verbose                bool
}

// NewVerifyCommand creates the verify command.
func NewVerifyCommand() *cobra.Command {
	opts := &verifyOptions{}

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify a provider against a contract file",
		Long:  "Verify that a provider API satisfies the expectations defined in a Pact contract file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.providerBaseURL, "provider-base-url", "", "Base URL of the provider API (required)")
	cmd.Flags().StringVar(&opts.pactFile, "pact-file", "", "Path to the Pact contract file (required)")
	cmd.Flags().StringVar(&opts.providerStatesSetupURL, "provider-states-setup-url", "", "URL for provider states setup")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", false, "Show detailed output")

	_ = cmd.MarkFlagRequired("provider-base-url")
	_ = cmd.MarkFlagRequired("pact-file")

	return cmd
}

func runVerify(cmd *cobra.Command, opts *verifyOptions) error {
	// Validate flags
	if opts.providerBaseURL == "" {
		return fmt.Errorf("--provider-base-url is required")
	}
	if opts.pactFile == "" {
		return fmt.Errorf("--pact-file is required")
	}

	// Check if file exists
	if _, err := os.Stat(opts.pactFile); os.IsNotExist(err) {
		return fmt.Errorf("pact file not found: %s", opts.pactFile)
	}

	// Parse contract
	parser := contract.NewParser()
	c, err := parser.ParseFile(opts.pactFile)
	if err != nil {
		return fmt.Errorf("failed to parse pact file: %w", err)
	}

	// Verify contract
	v := verifier.New(verifier.Config{
		ProviderBaseURL:        opts.providerBaseURL,
		ProviderStatesSetupURL: opts.providerStatesSetupURL,
	})

	result, err := v.Verify(c)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Report results
	reporter := verifier.NewReporter(cmd.OutOrStdout())
	reporter.SetVerbose(opts.verbose)
	reporter.Report(result)

	// Return error if verification failed (for exit code)
	if !result.Success {
		return fmt.Errorf("verification failed: %d interactions failed", countFailed(result.Interactions))
	}

	return nil
}

func countFailed(interactions []verifier.InteractionResult) int {
	count := 0
	for _, i := range interactions {
		if !i.Success {
			count++
		}
	}
	return count
}
