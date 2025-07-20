// Package main is the main package for the sandworm CLI.
package main

import (
	"os"

	"github.com/holonoms/sandworm/internal/cli"
)

func main() {
	opts := &cli.Options{}
	if err := cli.NewRootCmd(opts).Execute(); err != nil {
		os.Exit(1)
	}
}
