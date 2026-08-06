package netmgmt

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// NewNetworkCmd creates the network command tree for the server CLI.
func NewNetworkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Manage network interfaces, VLANs, and bonds",
		Long: `Manage the network stack on this Puffin server.

Subcommands:
  list     - List all network interfaces in a table
  status   - Show detailed status of a specific interface
  add      - Interactively add a VLAN, bond, or interface
  remove   - Interactively select and remove an interface`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newRemoveCmd())

	return cmd
}

func newListCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all network interfaces",
		Long:  "Print a table of all network interfaces with their type, state, MTU, MAC, addresses, and traffic stats.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ifaces, err := GatherInterfaces()
			if err != nil {
				return fmt.Errorf("gathering interfaces: %w", err)
			}

			if jsonOut {
				out, err := formatJSON(ifaces)
				if err != nil {
					return err
				}
				fmt.Println(out)
				return nil
			}

			fmt.Print(FormatList(ifaces))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output in JSON format")

	return cmd
}

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [interface]",
		Short: "Show detailed status of a network interface",
		Long:  "Show detailed status of a specific network interface, including addresses, statistics, bond/VLAN details.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ifaces, err := GatherInterfaces()
			if err != nil {
				return fmt.Errorf("gathering interfaces: %w", err)
			}

			var found *Interface
			for i := range ifaces {
				if ifaces[i].Name == name {
					found = &ifaces[i]
					break
				}
			}
			if found == nil {
				return fmt.Errorf("interface %q not found", name)
			}

			var bond *BondInfo
			if found.Type == TypeBond {
				bond, _ = GatherBondInfo(name)
			}

			var vlan *VLANInfo
			if found.Type == TypeVLAN {
				vlan, _ = GatherVLANInfo(name)
			}

			fmt.Print(FormatStatus(*found, bond, vlan))
			return nil
		},
	}

	return cmd
}

func newAddCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a VLAN, bond, or network interface",
		Long: `Interactively add a network interface.

A bubbletea form will guide you through:
  - VLAN: name, parent, VLAN ID, optional IP/gateway
  - Bond: name, mode, slave interfaces, optional IP/gateway
  - Interface (dummy): name, optional IP/gateway

Use --dry-run to see the generated commands without executing them.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := RunAddForm()
			if err != nil {
				return fmt.Errorf("add form: %w", err)
			}
			if result == nil {
				fmt.Println("Cancelled.")
				return nil
			}

			cmds := generateAddCommands(result)
			if dryRun {
				fmt.Println("Commands that would be executed:")
				for _, c := range cmds {
					fmt.Printf("  $ %s\n", c)
				}
				return nil
			}

			fmt.Println("Adding interface...")
			for _, c := range cmds {
				fmt.Printf("  $ %s\n", c)
				// TODO: execute via ip command or netlink
			}
			fmt.Println("Note: actual execution not yet implemented. Use --dry-run to preview.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview commands without executing")

	return cmd
}

func newRemoveCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a network interface",
		Long: `Interactively select and remove a network interface.

A bubbletea list shows all interfaces for selection.
Use --dry-run to see the command without executing it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ifaces, err := GatherInterfaces()
			if err != nil {
				return fmt.Errorf("gathering interfaces: %w", err)
			}

			// Filter out loopback and physical interfaces that shouldn't be removed
			var removable []Interface
			for _, iface := range ifaces {
				if iface.Type != TypeLoopback && iface.Type != TypePhysical {
					removable = append(removable, iface)
				}
			}

			if len(removable) == 0 {
				fmt.Println("No removable interfaces found.")
				return nil
			}

			name, err := RunRemoveForm(removable)
			if err != nil {
				return fmt.Errorf("remove form: %w", err)
			}
			if name == "" {
				fmt.Println("Cancelled.")
				return nil
			}

			rmCmd := fmt.Sprintf("ip link delete %s", name)
			if dryRun {
				fmt.Printf("Command that would be executed:\n  $ %s\n", rmCmd)
				return nil
			}

			fmt.Printf("Removing interface %s...\n", name)
			fmt.Printf("  $ %s\n", rmCmd)
			// TODO: execute via ip command or netlink
			fmt.Println("Note: actual execution not yet implemented. Use --dry-run to preview.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview command without executing")

	return cmd
}

// generateAddCommands produces the ip commands for the given add result.
func generateAddCommands(r *AddResult) []string {
	var cmds []string

	switch r.Type {
	case AddVLAN:
		cmds = append(cmds, fmt.Sprintf("ip link add link %s name %s type vlan id %s", r.Parent, r.Name, r.VLANID))
		cmds = append(cmds, fmt.Sprintf("ip link set %s up", r.Name))
		if r.IPAddr != "" {
			cmds = append(cmds, fmt.Sprintf("ip addr add %s dev %s", r.IPAddr, r.Name))
		}
		if r.Gateway != "" {
			cmds = append(cmds, fmt.Sprintf("ip route add default via %s dev %s", r.Gateway, r.Name))
		}
	case AddBond:
		cmds = append(cmds, fmt.Sprintf("ip link add %s type bond mode %s", r.Name, r.Mode))
		for _, slave := range strings.Split(r.Slaves, ",") {
			slave = strings.TrimSpace(slave)
			if slave == "" {
				continue
			}
			cmds = append(cmds, fmt.Sprintf("ip link set %s down", slave))
			cmds = append(cmds, fmt.Sprintf("ip link set %s master %s", slave, r.Name))
		}
		cmds = append(cmds, fmt.Sprintf("ip link set %s up", r.Name))
		if r.IPAddr != "" {
			cmds = append(cmds, fmt.Sprintf("ip addr add %s dev %s", r.IPAddr, r.Name))
		}
		if r.Gateway != "" {
			cmds = append(cmds, fmt.Sprintf("ip route add default via %s dev %s", r.Gateway, r.Name))
		}
	case AddInterface:
		cmds = append(cmds, fmt.Sprintf("ip link add %s type dummy", r.Name))
		cmds = append(cmds, fmt.Sprintf("ip link set %s up", r.Name))
		if r.IPAddr != "" {
			cmds = append(cmds, fmt.Sprintf("ip addr add %s dev %s", r.IPAddr, r.Name))
		}
		if r.Gateway != "" {
			cmds = append(cmds, fmt.Sprintf("ip route add default via %s dev %s", r.Gateway, r.Name))
		}
	}

	return cmds
}

// formatJSON serialises interfaces as indented JSON.
func formatJSON(ifaces []Interface) (string, error) {
	b, err := json.MarshalIndent(ifaces, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}