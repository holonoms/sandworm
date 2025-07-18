package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSetupCmd creates the setup command
func NewSetupCmd() *cobra.Command {
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
