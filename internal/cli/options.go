// Package cli provides the command-line interface for sandworm
package cli

import (
	"fmt"
	"time"
)

// Options holds the command-line options shared across commands
type Options struct {
	// OutputFile specifies the path where the concatenated project file will be written.
	// If empty, defaults are applied based on the command context.
	OutputFile string

	// IgnoreFile specifies a custom ignore file to use instead of the default .gitignore.
	// If empty, sandworm will:
	// - look for .sandwormignore first, then
	// - fall back to .gitignore if present
	// - use a sane internal list of ignore patterns (see internal/processor.go)
	IgnoreFile string

	// KeepFile determines whether to retain the generated file after pushing to Claude.
	// For generate command, this is always true. For push command, this is controlled by the --keep flag.
	KeepFile bool

	// Directory specifies the root directory to process.
	// If empty, defaults to the current directory (".").
	Directory string

	// ShowLineNumbers determines whether to show line numbers in the output.
	// If nil, the value from config will be used. If set, it overrides the config.
	ShowLineNumbers *bool
}

// SetDefaults sets default values for options based on the command context
func (o *Options) SetDefaults(command string) {
	if o.Directory == "" {
		o.Directory = "."
	}

	switch command {
	case "generate":
		if o.OutputFile == "" {
			o.OutputFile = "sandworm.txt"
		}
		o.KeepFile = true
	case "push":
		if o.OutputFile == "" {
			o.OutputFile = fmt.Sprintf(".sandworm-%d.txt", time.Now().Unix())
		}
	}
}
