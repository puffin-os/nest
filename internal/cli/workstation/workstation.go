// Package workstation registers workstation-specific subcommands for the nest CLI.
package workstation

import (
	"github.com/spf13/cobra"

	"github.com/puffin/nest/internal/cli"
	"github.com/puffin/nest/internal/diskmgmt"
	"github.com/puffin/nest/internal/netmgmt"
	"github.com/puffin/nest/internal/svcmgt"
)

func init() {
	cli.RegisterFlavorCmds = register
}

func register(root *cobra.Command, opts *cli.Options) {
	_ = opts
	root.AddCommand(netmgmt.NewNetworkCmd())
	root.AddCommand(diskmgmt.NewDiskCmd())
	root.AddCommand(svcmgt.NewServiceCmd())
}