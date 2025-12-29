package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root command for yakusoku CLI.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yakusoku",
		Short: "Contract testing CLI tool",
		Long: `Yakusoku is a consumer-driven contract testing CLI tool.
It is compatible with Pact Specification v3/v4.

Use yakusoku to:
  - Verify provider APIs against contract files
  - Manage contract files (list, show, diff)
  - Publish contracts to a broker
  - Check deployment safety with can-i-deploy`,
	}

	// Add subcommands
	cmd.AddCommand(NewVerifyCommand())
	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewShowCommand())
	cmd.AddCommand(NewDiffCommand())
	cmd.AddCommand(NewPublishCommand())
	cmd.AddCommand(NewCanIDeployCommand())

	return cmd
}
