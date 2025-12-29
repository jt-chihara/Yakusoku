package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

// NewCanIDeployCommand creates the can-i-deploy command
func NewCanIDeployCommand() *cobra.Command {
	var brokerURL string
	var pacticipant string
	var version string
	var toEnvironment string
	var latest bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "can-i-deploy",
		Short: "Check if a pacticipant version can be deployed",
		Long:  "Query the broker to determine if a pacticipant version is safe to deploy",
		RunE: func(cmd *cobra.Command, args []string) error {
			if brokerURL == "" {
				return fmt.Errorf("--broker-url is required")
			}
			if pacticipant == "" {
				return fmt.Errorf("--pacticipant is required")
			}
			if version == "" && !latest {
				return fmt.Errorf("either --version or --latest is required")
			}

			return runCanIDeploy(cmd, brokerURL, pacticipant, version, toEnvironment, latest, jsonOutput)
		},
	}

	cmd.Flags().StringVar(&brokerURL, "broker-url", "", "URL of the Pact broker (required)")
	cmd.Flags().StringVar(&pacticipant, "pacticipant", "", "Name of the pacticipant (required)")
	cmd.Flags().StringVar(&version, "version", "", "Version of the pacticipant")
	cmd.Flags().StringVar(&toEnvironment, "to-environment", "", "Target environment for deployment")
	cmd.Flags().BoolVar(&latest, "latest", false, "Use the latest version")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func runCanIDeploy(cmd *cobra.Command, brokerURL, pacticipant, version, toEnvironment string, latest, jsonOutput bool) error {
	// Build query parameters
	params := url.Values{}
	params.Set("pacticipant", pacticipant)

	if latest {
		params.Set("latest", "true")
	} else {
		params.Set("version", version)
	}

	if toEnvironment != "" {
		params.Set("environment", toEnvironment)
	}

	// Build URL
	requestURL := fmt.Sprintf("%s/matrix?%s", brokerURL, params.Encode())

	// Send request
	resp, err := http.Get(requestURL)
	if err != nil {
		return fmt.Errorf("failed to query broker: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Deployable bool `json:"deployable"`
		Summary    struct {
			Deployable bool   `json:"deployable"`
			Reason     string `json:"reason"`
		} `json:"summary"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if jsonOutput {
		fmt.Fprintln(cmd.OutOrStdout(), string(body))
	} else {
		if result.Deployable {
			fmt.Fprintf(cmd.OutOrStdout(), "%s version %s can be deployed\n", pacticipant, version)
			fmt.Fprintf(cmd.OutOrStdout(), "Reason: %s\n", result.Summary.Reason)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "%s version %s cannot be deployed\n", pacticipant, version)
			fmt.Fprintf(cmd.OutOrStdout(), "Reason: %s\n", result.Summary.Reason)
		}
	}

	if !result.Deployable {
		return fmt.Errorf("deployment not allowed: %s", result.Summary.Reason)
	}

	return nil
}
