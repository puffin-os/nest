// Package server registers server-specific subcommands for the nest CLI.
package server

import (
	"github.com/spf13/cobra"

	"github.com/puffin/nest/internal/cli"
	"github.com/puffin/nest/internal/diskmgmt"
	"github.com/puffin/nest/internal/netmgmt"
	"github.com/puffin/nest/internal/quadlet"
	"github.com/puffin/nest/internal/svcmgt"
)

func init() {
	cli.RegisterFlavorCmds = register
}

func register(root *cobra.Command, opts *cli.Options) {
	root.AddCommand(netmgmt.NewNetworkCmd())
	root.AddCommand(diskmgmt.NewDiskCmd())
	root.AddCommand(svcmgt.NewServiceCmd())
	root.AddCommand(quadlet.NewAppsCmd())
}