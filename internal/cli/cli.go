// Package cli provides the command-line interface for sandworm.
package cli

import (
	"github.com/spf13/cobra"
)

var (
	// Default version for development/non-release builds
	// GoReleaser overrides this for release builds with the git tag.
	// See .goreleaser.yml
	version = "dev"
)

// NewRootCmd creates the root command with all subcommands
func NewRootCmd() *cobra.Command {
	opts := &Options{}

	rootCmd := &cobra.Command{
		Use:          "sandworm [directory]",
		Short:        "Project file concatenator",
		Version:      version,
		SilenceUsage: true,
		// NB: ArbitraryArgs is required to avoid interpreting the first argument
		// as a subcommand. This is necessary for the use case `sandworm [folder]`,
		// where folder would otherwise be interpreted as a subcommand and fail.
		Args: cobra.ArbitraryArgs,
		// When no subcommand is supplied, execute the push command
		RunE: newPushCmd(opts).RunE,
	}

	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&opts.OutputFile, "output", "o", "", "Output file")
	rootCmd.PersistentFlags().StringVar(&opts.IgnoreFile, "ignore", "", "Ignore file (default: .gitignore)")
	rootCmd.PersistentFlags().BoolVarP(&opts.KeepFile, "keep", "k", false, "Keep the generated file after pushing")

	// Add commands
	rootCmd.AddCommand(
		newGenerateCmd(opts),
		newPushCmd(opts),
		newPurgeCmd(),
		newSetupCmd(),
		newConfigCmd(),
	)

	return rootCmd
}
