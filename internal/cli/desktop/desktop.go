// Package desktop registers desktop-specific subcommands for the nest CLI.
package desktop

import (
	"github.com/spf13/cobra"

	"github.com/puffin/nest/internal/cli"
)

func init() {
	cli.RegisterFlavorCmds = register
}

func register(root *cobra.Command, opts *cli.Options) {
	// Placeholder: desktop-specific subcommands will be added here.
	_ = opts
}