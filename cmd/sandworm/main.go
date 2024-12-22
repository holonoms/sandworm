package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/umwelt-studio/sandworm/internal/claude"
	"github.com/umwelt-studio/sandworm/internal/config"
	"github.com/umwelt-studio/sandworm/internal/processor"
	"github.com/umwelt-studio/sandworm/internal/util"
)

var (
	version = "0.1.0"
)

type cmdCfg struct {
	command    string
	directory  string
	outputFile string
	ignoreFile string
	keepFile   bool
}

func main() {
	cfg, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Setup Claude client for push or purge commands
	var client *claude.Client
	if cfg.command == "purge" || cfg.command == "push" {
		conf, err := config.New("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load cmdCfg: %v\n", err)
			os.Exit(1)
		}

		client = claude.New(conf.GetSection("claude"))
		ok, err := client.Setup(false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if !ok {
			fmt.Println("Setup did not complete; exiting.")
			os.Exit(1)
		}
	}

	if cfg.command == "purge" {
		count, err := client.PurgeProjectFiles(func(filename string, current, total int) {
			fmt.Printf("%d/%d: Deleting '%s'...\n", current, total, filename)
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
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
		os.Exit(0)
	}

	// Set output file
	if cfg.command == "generate" && cfg.outputFile == "" {
		cfg.outputFile = "sandworm.txt"
		cfg.keepFile = true
	} else if cfg.outputFile == "" {
		cfg.outputFile = fmt.Sprintf(".sandworm-%d.txt", time.Now().Unix())
	}

	// Process files
	fmt.Printf("Generating project file '%s'...\n", cfg.outputFile)
	p, err := processor.New(cfg.directory, cfg.outputFile, cfg.ignoreFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	size, err := p.Process()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if cfg.command == "push" {
		if err := client.Push(cfg.outputFile, "project.txt"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated project file (%s)\n", util.FormatSize(size))
	} else {
		fmt.Printf("Generated '%s' (%s)\n", cfg.outputFile, util.FormatSize(size))
	}

	// Clean up the output file unless we're keeping it
	if !cfg.keepFile {
		os.Remove(cfg.outputFile)
	}
}

func parseArgs() (*cmdCfg, error) {
	cfg := &cmdCfg{
		command: "push", // default command
	}

	// Define flags
	flags := flag.NewFlagSet("sandworm", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Printf("Sandworm v%s - Project file concatenator\n\n", version)
		fmt.Println("Usage: sandworm [command] [options] [directory]")
		fmt.Println("\nCommands:")
		fmt.Println("    generate    Generate concatenated file only")
		fmt.Println("    push        Generate and push to Claude (default)")
		fmt.Println("    purge       Remove all files from Claude project")
		fmt.Println("\nOptions:")
		flags.PrintDefaults()
	}

	flags.StringVar(&cfg.outputFile, "output", "", "Output file (defaults to temp file on push and sandworm.txt for generate)")
	flags.StringVar(&cfg.outputFile, "o", "", "Output file (short flag)")
	flags.StringVar(&cfg.ignoreFile, "ignore", "", "Ignore file (default: .gitignore)")
	flags.BoolVar(&cfg.keepFile, "keep", false, "Keep the generated file after pushing (only affects push)")
	flags.BoolVar(&cfg.keepFile, "k", false, "Keep the generated file (short flag)")

	// Version flag
	var showVersion bool
	flags.BoolVar(&showVersion, "version", false, "Show version")
	flags.BoolVar(&showVersion, "v", false, "Show version (short flag)")

	// Parse args
	if len(os.Args) > 1 {
		if os.Args[1] == "generate" || os.Args[1] == "push" || os.Args[1] == "purge" {
			cfg.command = os.Args[1]
			err := flags.Parse(os.Args[2:])
			if err != nil {
				return nil, err
			}
		} else {
			err := flags.Parse(os.Args[1:])
			if err != nil {
				return nil, err
			}
		}
	}

	if showVersion {
		fmt.Printf("Sandworm version %s\n", version)
		os.Exit(0)
	}

	// Get directory argument or use current directory
	args := flags.Args()
	if len(args) > 0 {
		cfg.directory = args[0]
	} else {
		var err error
		cfg.directory, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	return cfg, nil
}
