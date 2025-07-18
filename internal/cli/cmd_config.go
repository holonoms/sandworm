package cli

import (
	"fmt"

	"github.com/holonoms/sandworm/internal/config"
	"github.com/spf13/cobra"
)

// ConfigOption represents a configuration option
type ConfigOption struct {
	Key         string
	Description string
	Default     string
	ValidValues []string // For enumerated values like true/false
	Validator   func(string) error
}

// Registry of all available configuration options
var configOptions = []ConfigOption{
	{
		Key:         "claude.organization_id",
		Description: "The organization ID to use for the Claude API",
		Default:     "",
	},
	{
		Key:         "claude.project_id",
		Description: "The project ID to use for the Claude API",
		Default:     "",
	},
	{
		Key:         "claude.document_id",
		Description: "The document ID to use for the Claude API",
		Default:     "",
	},
}

// MARK: Sub-commands

// newConfigCmd creates the config command and its subcommands
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage project configuration",
	}

	// Add subcommands
	cmd.AddCommand(
		newConfigListCmd(),
		newConfigGetCmd(),
		newConfigSetCmd(),
		newConfigUnsetCmd(),
	)

	return cmd
}

func newConfigListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runConfigList()
		},
	}

	return cmd
}

func runConfigList() error {
	cfg, err := config.New(".")
	if err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	fmt.Println("Available configuration options:")
	fmt.Println()

	for _, option := range configOptions {
		fmt.Printf("  %s\n", option.Key)
		fmt.Printf("    Description: %s\n", option.Description)
		fmt.Printf("    Default: %s\n", option.Default)

		if cfg.Has(option.Key) {
			value := cfg.Get(option.Key)
			fmt.Printf("    Current: %s\n", value)
		} else {
			fmt.Printf("    Current: %s (default)\n", option.Default)
		}
		fmt.Println()
	}

	return nil
}

func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			return runConfigSet(args[0], args[1])
		},
		ValidArgsFunction: func(
			cmd *cobra.Command,
			args []string,
			toComplete string,
		) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return configOptionsKeys(), cobra.ShellCompDirectiveNoFileComp
			}
			// For values, provide common completions based on the key
			if len(args) == 1 {
				option := findConfigOption(args[0])
				if option != nil && len(option.ValidValues) > 0 {
					return option.ValidValues, cobra.ShellCompDirectiveNoFileComp
				}
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func runConfigSet(key, value string) error {
	// Find the configuration option
	option := findConfigOption(key)
	if option == nil {
		return fmt.Errorf("unknown configuration option: %s\n\nRun 'sandworm config list' to see available options", key)
	}

	// Validate the value
	if option.Validator != nil {
		if err := option.Validator(value); err != nil {
			return fmt.Errorf("invalid value for %s: %w", key, err)
		}
	}

	cfg, err := config.New(".")
	if err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	if err := cfg.Set(key, value); err != nil {
		return fmt.Errorf("unable to set config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

func newConfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runConfigGet(args[0])
		},
		ValidArgsFunction: func(
			cmd *cobra.Command,
			args []string,
			toComplete string,
		) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return configOptionsKeys(), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func runConfigGet(key string) error {
	// Validate that the key is a known option
	option := findConfigOption(key)
	if option == nil {
		return fmt.Errorf("unknown configuration option: %s\n\nRun 'sandworm config list' to see available options", key)
	}

	cfg, err := config.New(".")
	if err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	if !cfg.Has(key) {
		fmt.Printf("%s = %s (default)\n", key, option.Default)
		return nil
	}

	value := cfg.Get(key)
	fmt.Printf("%s = %s\n", key, value)
	return nil
}

func newConfigUnsetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset <key>",
		Short: "Unset a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runConfigUnset(args[0])
		},
		ValidArgsFunction: func(
			cmd *cobra.Command,
			args []string,
			toComplete string,
		) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return configOptionsKeys(), cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmd
}

func runConfigUnset(key string) error {
	cfg, err := config.New(".")
	if err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	if err := cfg.Delete(key); err != nil {
		return fmt.Errorf("unable to unset config: %w", err)
	}

	fmt.Printf("Unset %s\n", key)
	return nil
}

// MARK: Helpers

// findConfigOption finds a config option by key
func findConfigOption(key string) *ConfigOption {
	for i := range configOptions {
		if configOptions[i].Key == key {
			return &configOptions[i]
		}
	}
	return nil
}

func configOptionsKeys() []string {
	var keys []string
	for _, option := range configOptions {
		keys = append(keys, option.Key)
	}
	return keys
}

// MARK: Validators

// validateBoolOption validates that a value is either "true" or "false"
func validateBoolOption(value string) error {
	if value != "true" && value != "false" {
		return fmt.Errorf("value must be either 'true' or 'false', got: %s", value)
	}
	return nil
}
