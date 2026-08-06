package quadlet

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// NewAppsCmd creates the apps (quadlet) command tree for the server CLI.
func NewAppsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "Manage Podman Quadlets (containerized services)",
		Long: `Manage Podman Quadlet unit files on this Puffin server.

Quadlets are systemd unit files that describe containerized services.
They mount on the host after the network is ready.

Subcommands:
  list      - List all quadlets in a table
  status    - Show detailed status of a specific quadlet
  create    - Interactive wizard to create a new container quadlet
  delete    - Delete a quadlet and stop its service
  start     - Start a quadlet's service
  stop      - Stop a quadlet's service
  restart   - Restart a quadlet's service
  logs      - View journal logs for a quadlet's service

All subcommands support --user to manage user-level quadlets.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newStartCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newRestartCmd())
	cmd.AddCommand(newLogsCmd())

	return cmd
}

func scopeFromFlag(user bool) Scope {
	if user {
		return ScopeUser
	}
	return ScopeSystem
}

func newListCmd() *cobra.Command {
	var jsonOut bool
	var user bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all quadlets",
		Long:  "Print a table of all quadlet unit files with their type, image, active state, and scope.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// List both system and user if --user is set, otherwise system only
			scope := scopeFromFlag(user)

			quadlets, err := GatherQuadlets(scope)
			if err != nil {
				return fmt.Errorf("gathering quadlets: %w", err)
			}

			if jsonOut {
				out, err := json.MarshalIndent(quadlets, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			fmt.Print(FormatQuadletList(quadlets))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&user, "user", false, "list user-level quadlets")
	return cmd
}

func newStatusCmd() *cobra.Command {
	var jsonOut bool
	var user bool

	cmd := &cobra.Command{
		Use:   "status <name>",
		Short: "Show detailed status of a quadlet",
		Long:  "Show detailed status of a specific quadlet, including unit file path, image, and systemd state.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope := scopeFromFlag(user)

			quadlets, err := GatherQuadlets(scope)
			if err != nil {
				return fmt.Errorf("gathering quadlets: %w", err)
			}

			q := FindQuadlet(quadlets, args[0])
			if q == nil {
				return fmt.Errorf("quadlet %q not found", args[0])
			}

			if jsonOut {
				out, err := json.MarshalIndent(q, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			fmt.Print(FormatQuadletDetail(q))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&user, "user", false, "manage user-level quadlets")
	return cmd
}

func newCreateCmd() *cobra.Command {
	var user bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new container quadlet (interactive wizard)",
		Long: `Launch an interactive wizard to create a new container quadlet.

The wizard will prompt for:
  - Name (used as the systemd unit name)
  - Image (container image)
  - Description
  - Volumes (optional, comma-separated)
  - Ports (optional, comma-separated)
  - Environment variables (optional, comma-separated)
  - Network (select: host, bridge, none, slirp4netns)
  - Restart policy (select: always, on-failure, no, on-abnormal, on-watchdog)
  - Auto-update from registry (select: no, yes)

The quadlet will mount on the host after the network is ready.
Use --dry-run to preview the generated unit file without writing it.
Use --user to create a user-level quadlet.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := RunCreateForm()
			if err != nil {
				return fmt.Errorf("create form: %w", err)
			}
			if spec == nil {
				fmt.Println("Cancelled.")
				return nil
			}

			scope := scopeFromFlag(user)

			if dryRun {
				fmt.Println("Quadlet that would be created:")
				fmt.Printf("  Name:    %s\n", spec.Name)
				fmt.Printf("  Image:   %s\n", spec.Image)
				fmt.Printf("  Desc:    %s\n", spec.Description)
				fmt.Printf("  Volumes: %v\n", spec.Volumes)
				fmt.Printf("  Ports:   %v\n", spec.Ports)
				fmt.Printf("  Env:     %v\n", spec.Environment)
				fmt.Printf("  Network: %s\n", spec.Network)
				fmt.Printf("  Restart: %s\n", spec.Restart)
				fmt.Printf("  Update:  %v\n", spec.AutoUpdate)
				fmt.Printf("  Scope:   %s\n", scope)
				return nil
			}

			if err := CreateQuadlet(spec, scope); err != nil {
				return err
			}
			fmt.Printf("Created quadlet %s (%s scope)\n", spec.Name, scope)
			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "create a user-level quadlet")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview without writing")
	return cmd
}

func newDeleteCmd() *cobra.Command {
	var user bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a quadlet and stop its service",
		Long: `Delete a quadlet unit file, stop and disable its systemd service.

WARNING: This permanently removes the quadlet configuration.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope := scopeFromFlag(user)

			if err := DeleteQuadlet(args[0], scope); err != nil {
				return err
			}
			fmt.Printf("Deleted quadlet %s (%s scope)\n", args[0], scope)
			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "manage user-level quadlets")
	return cmd
}

func newStartCmd() *cobra.Command {
	var user bool

	cmd := &cobra.Command{
		Use:   "start <name>",
		Short: "Start a quadlet's service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope := scopeFromFlag(user)

			if err := StartQuadlet(args[0], scope); err != nil {
				return err
			}
			fmt.Printf("Started quadlet %s (%s scope)\n", args[0], scope)
			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "manage user-level quadlets")
	return cmd
}

func newStopCmd() *cobra.Command {
	var user bool

	cmd := &cobra.Command{
		Use:   "stop <name>",
		Short: "Stop a quadlet's service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope := scopeFromFlag(user)

			if err := StopQuadlet(args[0], scope); err != nil {
				return err
			}
			fmt.Printf("Stopped quadlet %s (%s scope)\n", args[0], scope)
			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "manage user-level quadlets")
	return cmd
}

func newRestartCmd() *cobra.Command {
	var user bool

	cmd := &cobra.Command{
		Use:   "restart <name>",
		Short: "Restart a quadlet's service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope := scopeFromFlag(user)

			if err := RestartQuadlet(args[0], scope); err != nil {
				return err
			}
			fmt.Printf("Restarted quadlet %s (%s scope)\n", args[0], scope)
			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "manage user-level quadlets")
	return cmd
}

func newLogsCmd() *cobra.Command {
	var lines int
	var follow bool
	var user bool

	cmd := &cobra.Command{
		Use:   "logs <name>",
		Short: "View journal logs for a quadlet's service",
		Long: `View recent journal logs for a quadlet's systemd service.

Use --lines to control how many lines to show (default 50).
Use --follow to follow logs in real time (Ctrl+C to stop).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope := scopeFromFlag(user)

			logs, err := GatherQuadletLogs(args[0], lines, follow, scope)
			if err != nil {
				return err
			}
			fmt.Print(logs)
			return nil
		},
	}

	cmd.Flags().IntVarP(&lines, "lines", "n", 50, "number of log lines to show")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow logs in real time")
	cmd.Flags().BoolVar(&user, "user", false, "manage user-level quadlets")
	return cmd
}