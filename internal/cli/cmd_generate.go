package cli

import (
	"fmt"

	"github.com/holonoms/sandworm/internal/config"
	"github.com/holonoms/sandworm/internal/processor"
	"github.com/holonoms/sandworm/internal/util"
	"github.com/spf13/cobra"
)

// newGenerateCmd creates the generate command
func newGenerateCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [directory]",
		Short: "Generate concatenated file only",
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Directory = args[0]
			}
			// Default output for generate
			if opts.OutputFile == "" {
				opts.OutputFile = "sandworm.txt"
			}
			opts.KeepFile = true

			fmt.Printf("Generating project '%s'...\n", opts.OutputFile)
			size, err := runGenerate(opts)
			fmt.Printf("Generated '%s' (%s)\n", opts.OutputFile, util.FormatSize(size))
			return err
		},
	}

	return cmd
}

func runGenerate(opts *Options) (int64, error) {
	if opts.Directory == "" {
		opts.Directory = "."
	}

	printLineNumbers := false
	if opts.ShowLineNumbers != nil {
		printLineNumbers = *opts.ShowLineNumbers
	} else {
		// If line-numbers flag wasn't explicitly set, load & check the project's settings.
		cfg, err := config.New(opts.Directory)
		if err != nil {
			return 0, fmt.Errorf("unable to load config: %w", err)
		}

		if cfg.Has("processor.print_line_numbers") {
			value := cfg.Get("processor.print_line_numbers")
			if value == "true" {
				printLineNumbers = true
			}
		}
	}

	p, err := processor.New(opts.Directory, opts.OutputFile, opts.IgnoreFile, printLineNumbers)
	if err != nil {
		return 0, fmt.Errorf("unable to create processor: %w", err)
	}

	size, err := p.Process()
	if err != nil {
		return 0, fmt.Errorf("unable to process files: %w", err)
	}

	return size, nil
}
