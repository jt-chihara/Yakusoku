package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/jt-chihara/yakusoku/internal/contract"
)

// NewShowCommand creates the show command
func NewShowCommand() *cobra.Command {
	var pactFile string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show contract details",
		Long:  "Display detailed information about a contract file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(cmd, pactFile, jsonOutput)
		},
	}

	cmd.Flags().StringVar(&pactFile, "pact-file", "", "Path to the contract file (required)")
	cmd.MarkFlagRequired("pact-file")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func runShow(cmd *cobra.Command, pactFile string, jsonOutput bool) error {
	parser := contract.NewParser()
	c, err := parser.ParseFile(pactFile)
	if err != nil {
		return err
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(c, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	// Pretty print contract details
	fmt.Fprintf(cmd.OutOrStdout(), "Contract: %s -> %s\n", c.Consumer.Name, c.Provider.Name)
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintf(cmd.OutOrStdout(), "Consumer: %s\n", c.Consumer.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "Provider: %s\n", c.Provider.Name)
	fmt.Fprintln(cmd.OutOrStdout(), "")

	fmt.Fprintf(cmd.OutOrStdout(), "Interactions (%d):\n", len(c.Interactions))
	for i, interaction := range c.Interactions {
		fmt.Fprintf(cmd.OutOrStdout(), "\n  [%d] %s\n", i+1, interaction.Description)
		if interaction.ProviderState != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "      Provider State: %s\n", interaction.ProviderState)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "      Request:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "        Method: %s\n", interaction.Request.Method)
		fmt.Fprintf(cmd.OutOrStdout(), "        Path: %s\n", interaction.Request.Path)
		if len(interaction.Request.Headers) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "        Headers:\n")
			for k, v := range interaction.Request.Headers {
				fmt.Fprintf(cmd.OutOrStdout(), "          %s: %v\n", k, v)
			}
		}
		fmt.Fprintf(cmd.OutOrStdout(), "      Response:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "        Status: %d\n", interaction.Response.Status)
		if len(interaction.Response.Headers) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "        Headers:\n")
			for k, v := range interaction.Response.Headers {
				fmt.Fprintf(cmd.OutOrStdout(), "          %s: %v\n", k, v)
			}
		}
		if interaction.Response.Body != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "        Body: %v\n", interaction.Response.Body)
		}
	}

	return nil
}
