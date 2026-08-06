// Package server registers server-specific subcommands for the nest CLI.
package server

import (
	"github.com/spf13/cobra"

	"github.com/puffin/nest/internal/cli"
)

func init() {
	cli.RegisterFlavorCmds = register
}

func register(root *cobra.Command, opts *cli.Options) {
	// Placeholder: server-specific subcommands will be added here.
	_ = opts
}