package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/jt-chihara/yakusoku/internal/contract"
)

// NewListCommand creates the list command
func NewListCommand() *cobra.Command {
	var pactDir string
	var pattern string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contract files in a directory",
		Long:  "List all contract files in the specified directory with optional pattern filtering",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, pactDir, pattern, jsonOutput)
		},
	}

	cmd.Flags().StringVar(&pactDir, "pact-dir", ".", "Directory containing contract files")
	cmd.Flags().StringVar(&pattern, "pattern", "*.json", "Glob pattern for filtering contract files")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

type contractInfo struct {
	File     string `json:"file"`
	Consumer string `json:"consumer"`
	Provider string `json:"provider"`
}

func runList(cmd *cobra.Command, pactDir, pattern string, jsonOutput bool) error {
	// Check if directory exists
	info, err := os.Stat(pactDir)
	if err != nil {
		return fmt.Errorf("directory not found: %s", pactDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", pactDir)
	}

	// Find matching files
	matches, err := filepath.Glob(filepath.Join(pactDir, pattern))
	if err != nil {
		return fmt.Errorf("invalid pattern: %s", pattern)
	}

	// Parse contracts to get consumer/provider names
	parser := contract.NewParser()
	var contracts []contractInfo

	for _, match := range matches {
		c, err := parser.ParseFile(match)
		if err != nil {
			continue // Skip invalid files
		}
		contracts = append(contracts, contractInfo{
			File:     filepath.Base(match),
			Consumer: c.Consumer.Name,
			Provider: c.Provider.Name,
		})
	}

	// Handle no contracts found
	if len(contracts) == 0 {
		if jsonOutput {
			fmt.Fprintln(cmd.OutOrStdout(), "[]")
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "No contracts found")
		}
		return nil
	}

	// Output results
	if jsonOutput {
		result := make([]map[string]interface{}, len(contracts))
		for i, c := range contracts {
			result[i] = map[string]interface{}{
				"file":     c.File,
				"consumer": c.Consumer,
				"provider": c.Provider,
			}
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Contracts:")
		for _, c := range contracts {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", c.File)
			fmt.Fprintf(cmd.OutOrStdout(), "    Consumer: %s\n", c.Consumer)
			fmt.Fprintf(cmd.OutOrStdout(), "    Provider: %s\n", c.Provider)
		}
	}

	return nil
}
