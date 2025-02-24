package main

import (
	"fmt"
	"os"
	"time"

	"github.com/holonoms/sandworm/internal/claude"
	"github.com/holonoms/sandworm/internal/config"
	"github.com/holonoms/sandworm/internal/processor"
	"github.com/holonoms/sandworm/internal/util"
	"github.com/spf13/cobra"
)

var (
	// Default version for development/non-release builds
	// GoReleaser overrides this for release builds with the git tag.
	// See .goreleaser.yml
	version = "dev"
)

type cmdOptions struct {
	outputFile string
	ignoreFile string
	keepFile   bool
	directory  string
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	opts := &cmdOptions{}

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
	rootCmd.PersistentFlags().StringVarP(&opts.outputFile, "output", "o", "", "Output file")
	rootCmd.PersistentFlags().StringVar(&opts.ignoreFile, "ignore", "", "Ignore file (default: .gitignore)")
	rootCmd.PersistentFlags().BoolVarP(&opts.keepFile, "keep", "k", false, "Keep the generated file after pushing")

	// Add commands
	rootCmd.AddCommand(
		newGenerateCmd(opts),
		newPushCmd(opts),
		newPurgeCmd(),
		newSetupCmd(),
	)

	return rootCmd
}

// MARK: Generate command

func newGenerateCmd(opts *cmdOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [directory]",
		Short: "Generate concatenated file only",
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.directory = args[0]
			}
			// Default output for generate
			if opts.outputFile == "" {
				opts.outputFile = "sandworm.txt"
			}
			opts.keepFile = true

			fmt.Printf("Generating project '%s'...\n", opts.outputFile)
			size, err := runGenerate(opts)
			fmt.Printf("Generated '%s' (%s)\n", opts.outputFile, util.FormatSize(size))
			return err
		},
	}

	return cmd
}

func runGenerate(opts *cmdOptions) (int64, error) {
	if opts.directory == "" {
		opts.directory = "."
	}

	p, err := processor.New(opts.directory, opts.outputFile, opts.ignoreFile)
	if err != nil {
		return 0, fmt.Errorf("unable to create processor: %w", err)
	}

	size, err := p.Process()
	if err != nil {
		return 0, fmt.Errorf("unable to process files: %w", err)
	}

	return size, nil
}

// MARK: Push command

func newPushCmd(opts *cmdOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push [directory]",
		Short: "Generate and push to Claude",
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.directory = args[0]
			}
			// Default output for push
			if opts.outputFile == "" {
				opts.outputFile = fmt.Sprintf(".sandworm-%d.txt", time.Now().Unix())
			}
			return runPush(opts)
		},
	}

	return cmd
}

func runPush(opts *cmdOptions) error {
	client, err := setupClaudeClient(false)
	if err != nil {
		return err
	}

	defer func() {
		// Clean up unless keepFile is true
		if !opts.keepFile {
			os.Remove(opts.outputFile)
		}
	}()

	fmt.Println("Generating project file...")
	var size int64
	if size, err = runGenerate(opts); err != nil {
		return err
	}

	fmt.Println("Syncing with Claude project...")
	if err := client.Push(opts.outputFile, "project.txt"); err != nil {
		return fmt.Errorf("unable to push: %w", err)
	}

	fmt.Printf("Project file synced (%s)\n", util.FormatSize(size))

	return nil
}

// MARK: Purge command

func newPurgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Remove all files from Claude project",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runPurge()
		},
	}

	return cmd
}

func runPurge() error {
	client, err := setupClaudeClient(false)
	if err != nil {
		return err
	}

	count, err := client.PurgeProjectFiles(func(filename string, current, total int) {
		fmt.Printf("%d/%d: Deleting '%s'...\n", current, total, filename)
	})
	if err != nil {
		return err
	}

	if count == 0 {
		fmt.Println("No files to delete.")
	} else {
		suffix := ""
		if count > 1 {
			suffix = "s"
		}
		fmt.Printf("Done! Removed %d file%s\n", count, suffix)
	}

	return nil
}

// MARK: Setup command

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure Claude project",
		RunE: func(_ *cobra.Command, _ []string) error {
			_, err := setupClaudeClient(true)
			if err != nil {
				return err
			}

			fmt.Println("\nSetup complete! Run 'sandworm push' to generate and push your project file.")
			return nil
		},
	}

	return cmd
}

func setupClaudeClient(force bool) (*claude.Client, error) {
	conf, err := config.New(".")
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	client := claude.New(conf)
	ok, err := client.Setup(force)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("setup did not complete")
	}

	return client, nil
}
