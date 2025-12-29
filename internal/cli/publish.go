package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/jt-chihara/yakusoku/internal/contract"
)

// NewPublishCommand creates the publish command
func NewPublishCommand() *cobra.Command {
	var brokerURL string
	var pactFile string
	var pactDir string
	var consumerVersion string
	var tags []string

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish contracts to a broker",
		Long:  "Publish contract files to a Pact broker for centralized management",
		RunE: func(cmd *cobra.Command, args []string) error {
			if brokerURL == "" {
				return fmt.Errorf("--broker-url is required")
			}
			if consumerVersion == "" {
				return fmt.Errorf("--consumer-version is required")
			}

			var files []string
			if pactFile != "" {
				files = append(files, pactFile)
			} else if pactDir != "" {
				matches, err := filepath.Glob(filepath.Join(pactDir, "*.json"))
				if err != nil {
					return fmt.Errorf("failed to find contracts: %w", err)
				}
				files = matches
			} else {
				return fmt.Errorf("either --pact-file or --pact-dir is required")
			}

			return runPublish(cmd, brokerURL, files, consumerVersion, tags)
		},
	}

	cmd.Flags().StringVar(&brokerURL, "broker-url", "", "URL of the Pact broker (required)")
	cmd.Flags().StringVar(&pactFile, "pact-file", "", "Path to a contract file")
	cmd.Flags().StringVar(&pactDir, "pact-dir", "", "Directory containing contract files")
	cmd.Flags().StringVar(&consumerVersion, "consumer-version", "", "Version of the consumer (required)")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Tags to apply to the published contracts")

	return cmd
}

func runPublish(cmd *cobra.Command, brokerURL string, files []string, version string, tags []string) error {
	parser := contract.NewParser()

	for _, file := range files {
		c, err := parser.ParseFile(file)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", file, err)
		}

		// Build URL
		url := fmt.Sprintf("%s/pacts/provider/%s/consumer/%s/version/%s",
			brokerURL, c.Provider.Name, c.Consumer.Name, version)

		// Read file content
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		// Create request
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		// Send request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to publish %s: %w", file, err)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to publish %s: %s (status %d)", file, string(body), resp.StatusCode)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Contract %s published successfully\n", filepath.Base(file))

		// Apply tags if specified
		for _, tag := range tags {
			tagURL := fmt.Sprintf("%s/pacticipants/%s/versions/%s/tags/%s",
				brokerURL, c.Consumer.Name, version, tag)
			tagReq, _ := http.NewRequest(http.MethodPut, tagURL, nil)
			tagResp, err := http.DefaultClient.Do(tagReq)
			if err == nil {
				tagResp.Body.Close()
			}
		}
	}

	return nil
}
