package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information (set at build time)
var (
	Version   = "0.1.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// NewVersionCommand creates the version command.
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "yakusoku version %s\n", Version)
			fmt.Fprintf(cmd.OutOrStdout(), "  Git commit: %s\n", GitCommit)
			fmt.Fprintf(cmd.OutOrStdout(), "  Build date: %s\n", BuildDate)
		},
	}
}
