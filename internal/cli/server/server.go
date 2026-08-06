// Package server registers server-specific subcommands for the nest CLI.
package server

import (
	"github.com/spf13/cobra"

	"github.com/puffin/nest/internal/cli"
	"github.com/puffin/nest/internal/netmgmt"
)

func init() {
	cli.RegisterFlavorCmds = register
}

func register(root *cobra.Command, opts *cli.Options) {
	root.AddCommand(netmgmt.NewNetworkCmd())
}