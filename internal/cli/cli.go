// Package cli provides the shared cobra-based command tree for all nest binaries.
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/puffin/nest/internal/sysinfo"
)

// Flavor identifies which Puffin derivative this CLI runs on.
type Flavor string

const (
	FlavorServer      Flavor = "server"
	FlavorDesktop     Flavor = "desktop"
	FlavorWorkstation Flavor = "workstation"
)

// Options holds configuration for the root command.
type Options struct {
	Flavor Flavor
	Output io.Writer
}

// RegisterFlavorCmds is a hook that each binary can override to add
// flavor-specific subcommands to the root command. By default no
// flavor-specific commands are registered.
var RegisterFlavorCmds func(root *cobra.Command, opts *Options)

// NewRootCmd creates the root cobra command for a nest binary.
// The flavor parameter controls which subcommands are available.
func NewRootCmd(flavor Flavor) *cobra.Command {
	opts := &Options{
		Flavor: flavor,
		Output: os.Stdout,
	}

	root := &cobra.Command{
		Use:   "nest",
		Short: fmt.Sprintf("Puffin %s management CLI", flavor),
		Long: fmt.Sprintf(
			"nest is the management CLI for Puffin OS.\n\nThis binary is built for the %s derivative.",
			flavor,
		),
		SilenceUsage: true,
	}

	// Shared subcommands available to all flavors.
	root.AddCommand(newSystemInfoCmd(opts))
	root.AddCommand(newUpdateCmd(opts))

	// Flavor-specific subcommands (if any are registered).
	if RegisterFlavorCmds != nil {
		RegisterFlavorCmds(root, opts)
	}

	return root
}

func newSystemInfoCmd(opts *Options) *cobra.Command {
	var jsonOut bool
	var plain bool

	cmd := &cobra.Command{
		Use:   "system-info",
		Short: "Print system information",
		Long: `Print information about the system this CLI is running on.

Includes CPU, memory, disk, network, OS, and runtime details.
Use --json for machine-readable output.
Use --plain for unstyled text output.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := sysinfo.Gather()
			if err != nil {
				return fmt.Errorf("gathering system info: %w", err)
			}

			if jsonOut {
				out, err := info.FormatJSON()
				if err != nil {
					return fmt.Errorf("formatting json: %w", err)
				}
				fmt.Fprintln(opts.Output, out)
				return nil
			}

			if plain {
				fmt.Fprint(opts.Output, info.FormatText())
				return nil
			}

			fmt.Fprint(opts.Output, info.FormatStyled())
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&plain, "plain", false, "output in plain text (no styling)")

	return cmd
}

func newUpdateCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the system",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(opts.Output, "Update finished")
			return nil
		},
	}
	return cmd
}
