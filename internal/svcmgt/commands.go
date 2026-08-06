package svcmgt

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// NewServiceCmd creates the service command tree for the server CLI.
func NewServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Aliases: []string{"svc"},
		Short: "Manage systemd services",
		Long: `Manage systemd services on this Puffin server.

Subcommands:
  list      - List all services in a table
  status    - Show detailed status of a specific service
  start     - Start a service
  stop      - Stop a service
  restart   - Restart a service
  reload    - Reload a service's configuration
  enable    - Enable a service to start at boot
  disable   - Disable a service from starting at boot
  mask      - Mask a service so it cannot be started
  unmask    - Unmask a previously masked service
  logs      - View recent journal logs for a service`,
	}

	cmd.AddCommand(newSvcListCmd())
	cmd.AddCommand(newSvcStatusCmd())
	cmd.AddCommand(newSvcStartCmd())
	cmd.AddCommand(newSvcStopCmd())
	cmd.AddCommand(newSvcRestartCmd())
	cmd.AddCommand(newSvcReloadCmd())
	cmd.AddCommand(newSvcEnableCmd())
	cmd.AddCommand(newSvcDisableCmd())
	cmd.AddCommand(newSvcMaskCmd())
	cmd.AddCommand(newSvcUnmaskCmd())
	cmd.AddCommand(newSvcLogsCmd())

	return cmd
}

func newSvcListCmd() *cobra.Command {
	var jsonOut bool
	var state string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all systemd services",
		Long:  "Print a table of all systemd service units with their load state, active state, sub state, and enabled status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := GatherServices()
			if err != nil {
				return fmt.Errorf("gathering services: %w", err)
			}

			if state != "" {
				filtered := services[:0]
				for _, s := range services {
					if string(s.ActiveState) == state {
						filtered = append(filtered, s)
					}
				}
				services = filtered
			}

			if jsonOut {
				out, err := json.MarshalIndent(services, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			fmt.Print(FormatServiceList(services))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	cmd.Flags().StringVarP(&state, "state", "s", "", "filter by active state (active, inactive, failed)")
	return cmd
}

func newSvcStatusCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "status <service>",
		Short: "Show detailed status of a systemd service",
		Long:  "Show detailed status of a specific systemd service, including PID, memory, CPU, environment, and timestamps.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			unit := args[0]
			detail, err := GatherServiceDetail(unit)
			if err != nil {
				return err
			}

			if jsonOut {
				out, err := json.MarshalIndent(detail, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			fmt.Print(FormatServiceDetail(detail))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	return cmd
}

func newSvcStartCmd() *cobra.Command {
	return newSimpleActionCmd("start", "Start a systemd service", StartService)
}

func newSvcStopCmd() *cobra.Command {
	return newSimpleActionCmd("stop", "Stop a systemd service", StopService)
}

func newSvcRestartCmd() *cobra.Command {
	return newSimpleActionCmd("restart", "Restart a systemd service", RestartService)
}

func newSvcReloadCmd() *cobra.Command {
	return newSimpleActionCmd("reload", "Reload a service's configuration", ReloadService)
}

func newSvcEnableCmd() *cobra.Command {
	return newSimpleActionCmd("enable", "Enable a service to start at boot", EnableService)
}

func newSvcDisableCmd() *cobra.Command {
	return newSimpleActionCmd("disable", "Disable a service from starting at boot", DisableService)
}

func newSvcMaskCmd() *cobra.Command {
	return newSimpleActionCmd("mask", "Mask a service so it cannot be started", MaskService)
}

func newSvcUnmaskCmd() *cobra.Command {
	return newSimpleActionCmd("unmask", "Unmask a previously masked service", UnmaskService)
}

func newSimpleActionCmd(use, short string, fn func(string) error) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <service>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			unit := args[0]
			if err := fn(unit); err != nil {
				return err
			}
			fmt.Printf("%s: %s\n", use, unit)
			return nil
		},
	}
}

func newSvcLogsCmd() *cobra.Command {
	var lines int
	var follow bool

	cmd := &cobra.Command{
		Use:   "logs <service>",
		Short: "View journal logs for a service",
		Long: `View recent journal logs for a systemd service.

Use --lines to control how many lines to show (default 50).
Use --follow to follow logs in real time (Ctrl+C to stop).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			unit := args[0]
			logs, err := GatherLogs(unit, lines, follow)
			if err != nil {
				return err
			}
			fmt.Print(logs)
			return nil
		},
	}

	cmd.Flags().IntVarP(&lines, "lines", "n", 50, "number of log lines to show")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow logs in real time")
	return cmd
}