package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewPurgeCmd creates the purge command
func NewPurgeCmd() *cobra.Command {
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
