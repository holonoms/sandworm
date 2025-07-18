// Package main is the main package for the sandworm CLI.
package main

import (
	"os"

	"github.com/holonoms/sandworm/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
