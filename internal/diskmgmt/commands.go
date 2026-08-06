package diskmgmt

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// NewDiskCmd creates the disk command tree for the server CLI.
func NewDiskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disk",
		Short: "Manage disks, partitions, and filesystems",
		Long: `Manage block devices on this Puffin server.

Subcommands:
  list      - List all block devices in a table
  status    - Show detailed status of a specific device
  mounts    - List all mounted filesystems with usage
  mount     - Mount a block device at a mountpoint
  unmount   - Unmount a block device or mountpoint
  format    - Format a block device with a filesystem
  expand    - Grow partition and resize filesystem to fill available space`,
	}

	cmd.AddCommand(newDiskListCmd())
	cmd.AddCommand(newDiskStatusCmd())
	cmd.AddCommand(newMountsCmd())
	cmd.AddCommand(newDiskMountCmd())
	cmd.AddCommand(newDiskUnmountCmd())
	cmd.AddCommand(newDiskFormatCmd())
	cmd.AddCommand(newDiskExpandCmd())

	return cmd
}

func newDiskListCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all block devices",
		Long:  "Print a table of all block devices with their type, size, filesystem, label, and mountpoint.",
		RunE: func(cmd *cobra.Command, args []string) error {
			devices, err := GatherDisks()
			if err != nil {
				return fmt.Errorf("gathering disks: %w", err)
			}

			if jsonOut {
				out, err := json.MarshalIndent(devices, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			fmt.Print(FormatDiskList(devices))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	return cmd
}

func newDiskStatusCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "status [device]",
		Short: "Show detailed status of a block device",
		Long:  "Show detailed status of a specific block device, including partitions, filesystem, and mount info.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			devices, err := GatherDisks()
			if err != nil {
				return fmt.Errorf("gathering disks: %w", err)
			}

			dev := FindDevice(devices, path)
			if dev == nil {
				return fmt.Errorf("device %q not found", path)
			}

			if jsonOut {
				out, err := json.MarshalIndent(dev, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			fmt.Print(FormatDiskStatus(*dev))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	return cmd
}

func newMountsCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "mounts",
		Short: "List all mounted filesystems with usage",
		Long:  "Print a table of all mounted filesystems with their size, used space, available space, and usage percentage.",
		RunE: func(cmd *cobra.Command, args []string) error {
			mounts, err := GatherMounts()
			if err != nil {
				return fmt.Errorf("gathering mounts: %w", err)
			}

			if jsonOut {
				out, err := json.MarshalIndent(mounts, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(out))
				return nil
			}

			fmt.Print(FormatMountList(mounts))
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	return cmd
}

func newDiskMountCmd() *cobra.Command {
	var options []string

	cmd := &cobra.Command{
		Use:   "mount <device> <mountpoint>",
		Short: "Mount a block device at a mountpoint",
		Long: `Mount a block device at the specified mountpoint.

The mountpoint directory will be created if it does not exist.
Use --options to pass comma-separated mount options (e.g. --options ro,noatime).`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			device := args[0]
			mountpoint := args[1]

			if err := MountDevice(device, mountpoint, options); err != nil {
				return err
			}
			fmt.Printf("Mounted %s at %s\n", device, mountpoint)
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&options, "options", "o", nil, "comma-separated mount options (e.g. ro,noatime)")
	return cmd
}

func newDiskUnmountCmd() *cobra.Command {
	var lazy bool

	cmd := &cobra.Command{
		Use:   "unmount <device|mountpoint>",
		Short: "Unmount a block device or mountpoint",
		Long: `Unmount a block device or mountpoint.

Use --lazy for a lazy unmount (detaches the filesystem now, cleans up when
the last reference is released).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			if err := UnmountDevice(target, lazy); err != nil {
				return err
			}
			fmt.Printf("Unmounted %s\n", target)
			return nil
		},
	}

	cmd.Flags().BoolVar(&lazy, "lazy", false, "lazy unmount (detach now, clean up later)")
	return cmd
}

func newDiskFormatCmd() *cobra.Command {
	var fstype string
	var label string

	cmd := &cobra.Command{
		Use:   "format <device>",
		Short: "Format a block device with a filesystem",
		Long: `Format a block device with the specified filesystem type.

Supported types: ext4, ext3, xfs, btrfs, fat32, swap.
Use --label to set the filesystem label.

WARNING: This destroys all data on the device.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			device := args[0]

			if fstype == "" {
				fstype = "ext4"
			}

			if err := FormatDevice(device, fstype, label); err != nil {
				return err
			}
			fmt.Printf("Formatted %s as %s\n", device, fstype)
			return nil
		},
	}

	cmd.Flags().StringVarP(&fstype, "type", "t", "ext4", "filesystem type (ext4, ext3, xfs, btrfs, fat32, swap)")
	cmd.Flags().StringVarP(&label, "label", "l", "", "filesystem label")
	return cmd
}

func newDiskExpandCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "expand <device>",
		Short: "Grow partition and resize filesystem to fill available space",
		Long: `Expand a block device's partition and filesystem to use all available space.

This runs growpart to grow the partition, then resizes the filesystem:
  - ext4/ext3: uses resize2fs
  - xfs: uses xfs_growfs (device must be mounted)
  - btrfs: uses btrfs filesystem resize (device must be mounted)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			device := args[0]

			if err := ExpandDevice(device); err != nil {
				return err
			}
			fmt.Printf("Expanded %s\n", device)
			return nil
		},
	}
	return cmd
}