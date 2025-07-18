package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/holonoms/sandworm/internal/claude"
	"github.com/holonoms/sandworm/internal/config"
	"github.com/holonoms/sandworm/internal/util"
	"github.com/spf13/cobra"
)

// newPushCmd creates the push command
func newPushCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push [directory]",
		Short: "Generate and push to Claude",
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Directory = args[0]
			}
			// Default output for push
			if opts.OutputFile == "" {
				opts.OutputFile = fmt.Sprintf(".sandworm-%d.txt", time.Now().Unix())
			}
			return runPush(opts)
		},
	}

	return cmd
}

func runPush(opts *Options) error {
	client, err := setupClaudeClient(false)
	if err != nil {
		return err
	}

	defer func() {
		// Clean up unless keepFile is true
		if !opts.KeepFile {
			_ = os.Remove(opts.OutputFile)
		}
	}()

	fmt.Println("Generating project file...")
	var size int64
	if size, err = runGenerate(opts); err != nil {
		return err
	}

	if err := client.Push(opts.OutputFile, "project.txt"); err != nil {
		return fmt.Errorf("unable to push: %w", err)
	}

	fmt.Printf("Updated project file (%s)\n", util.FormatSize(size))

	return nil
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
