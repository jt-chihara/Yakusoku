package cli

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/jt-chihara/yakusoku/internal/contract"
)

// NewDiffCommand creates the diff command
func NewDiffCommand() *cobra.Command {
	var oldFile string
	var newFile string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two contract files",
		Long:  "Show differences between two contract files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(cmd, oldFile, newFile, jsonOutput)
		},
	}

	cmd.Flags().StringVar(&oldFile, "old", "", "Path to the old contract file (required)")
	cmd.Flags().StringVar(&newFile, "new", "", "Path to the new contract file (required)")
	_ = cmd.MarkFlagRequired("old")
	_ = cmd.MarkFlagRequired("new")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

type diffResult struct {
	HasDifferences bool           `json:"hasDifferences"`
	Metadata       []metadataDiff `json:"metadata,omitempty"`
	Added          []string       `json:"added,omitempty"`
	Removed        []string       `json:"removed,omitempty"`
	Modified       []string       `json:"modified,omitempty"`
}

type metadataDiff struct {
	Field    string `json:"field"`
	OldValue string `json:"oldValue"`
	NewValue string `json:"newValue"`
}

func runDiff(cmd *cobra.Command, oldFile, newFile string, jsonOutput bool) error {
	parser := contract.NewParser()

	oldContract, err := parser.ParseFile(oldFile)
	if err != nil {
		return fmt.Errorf("failed to parse old contract: %w", err)
	}

	newContract, err := parser.ParseFile(newFile)
	if err != nil {
		return fmt.Errorf("failed to parse new contract: %w", err)
	}

	result := compareContracts(oldContract, newContract)

	if jsonOutput {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	// Pretty print differences
	if !result.HasDifferences {
		fmt.Fprintln(cmd.OutOrStdout(), "No differences found")
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Differences found:")

	for _, m := range result.Metadata {
		fmt.Fprintf(cmd.OutOrStdout(), "\n  %s changed:\n", m.Field)
		fmt.Fprintf(cmd.OutOrStdout(), "    - %s (old)\n", m.OldValue)
		fmt.Fprintf(cmd.OutOrStdout(), "    + %s (new)\n", m.NewValue)
	}

	if len(result.Added) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "\n  Interactions added:")
		for _, desc := range result.Added {
			fmt.Fprintf(cmd.OutOrStdout(), "    + %s\n", desc)
		}
	}

	if len(result.Removed) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "\n  Interactions removed:")
		for _, desc := range result.Removed {
			fmt.Fprintf(cmd.OutOrStdout(), "    - %s\n", desc)
		}
	}

	if len(result.Modified) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "\n  Interactions modified:")
		for _, desc := range result.Modified {
			fmt.Fprintf(cmd.OutOrStdout(), "    ~ %s\n", desc)
		}
	}

	return nil
}

func compareContracts(oldContract, newContract *contract.Contract) diffResult {
	result := diffResult{}

	// Compare consumer name
	if oldContract.Consumer.Name != newContract.Consumer.Name {
		result.HasDifferences = true
		result.Metadata = append(result.Metadata, metadataDiff{
			Field:    "consumer",
			OldValue: oldContract.Consumer.Name,
			NewValue: newContract.Consumer.Name,
		})
	}

	// Compare provider name
	if oldContract.Provider.Name != newContract.Provider.Name {
		result.HasDifferences = true
		result.Metadata = append(result.Metadata, metadataDiff{
			Field:    "provider",
			OldValue: oldContract.Provider.Name,
			NewValue: newContract.Provider.Name,
		})
	}

	// Build maps of interactions by description
	oldInteractions := make(map[string]*contract.Interaction)
	for i := range oldContract.Interactions {
		oldInteractions[oldContract.Interactions[i].Description] = &oldContract.Interactions[i]
	}

	newInteractions := make(map[string]*contract.Interaction)
	for i := range newContract.Interactions {
		newInteractions[newContract.Interactions[i].Description] = &newContract.Interactions[i]
	}

	// Find added interactions
	for desc := range newInteractions {
		if _, exists := oldInteractions[desc]; !exists {
			result.HasDifferences = true
			result.Added = append(result.Added, desc)
		}
	}

	// Find removed interactions
	for desc := range oldInteractions {
		if _, exists := newInteractions[desc]; !exists {
			result.HasDifferences = true
			result.Removed = append(result.Removed, desc)
		}
	}

	// Find modified interactions
	for desc, oldInt := range oldInteractions {
		if newInt, exists := newInteractions[desc]; exists {
			if !interactionsEqual(oldInt, newInt) {
				result.HasDifferences = true
				result.Modified = append(result.Modified, desc)
			}
		}
	}

	return result
}

func interactionsEqual(a, b *contract.Interaction) bool {
	// Compare request
	if a.Request.Method != b.Request.Method ||
		a.Request.Path != b.Request.Path {
		return false
	}

	// Compare response status
	if a.Response.Status != b.Response.Status {
		return false
	}

	// Compare headers
	if !reflect.DeepEqual(a.Request.Headers, b.Request.Headers) {
		return false
	}
	if !reflect.DeepEqual(a.Response.Headers, b.Response.Headers) {
		return false
	}

	// Compare body
	if !reflect.DeepEqual(a.Response.Body, b.Response.Body) {
		return false
	}

	// Compare provider state
	if a.ProviderState != b.ProviderState {
		return false
	}

	return true
}
