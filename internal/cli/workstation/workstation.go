// Package workstation registers workstation-specific subcommands for the nest CLI.
package workstation

import (
	"github.com/spf13/cobra"

	"github.com/puffin/nest/internal/cli"
)

func init() {
	cli.RegisterFlavorCmds = register
}

func register(root *cobra.Command, opts *cli.Options) {
	// Placeholder: workstation-specific subcommands will be added here.
	_ = opts
}