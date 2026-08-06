// Package diskmgmt provides disk management commands for the server CLI.
// It supports listing, inspecting, mounting, unmounting, formatting,
// and expanding disks and partitions using lsblk, findmnt, and mount.
package diskmgmt

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// DiskType categorises a block device.
type DiskType string

const (
	TypeDisk       DiskType = "disk"
	TypePart       DiskType = "part"
	TypeLVM        DiskType = "lvm"
	TypeCrypt      DiskType = "crypt"
	TypeROM        DiskType = "rom"
	TypeLoop       DiskType = "loop"
	TypeUnknownDev DiskType = "unknown"
)

// BlockDevice describes a block device discovered via lsblk.
type BlockDevice struct {
	Path       string   `json:"path"`
	Name       string   `json:"name"`
	Type       DiskType `json:"type"`
	Size       int64    `json:"size"`
	ReadOnly   bool     `json:"read_only"`
	Model      string   `json:"model"`
	Vendor     string   `json:"vendor"`
	FSType     string   `json:"fstype"`
	FsType     string   `json:"fs_type"` // alias for display
	Label      string   `json:"label"`
	UUID       string   `json:"uuid"`
	Mountpoint string   `json:"mountpoint"`
	Children   []BlockDevice `json:"children,omitempty"`
}

// MountInfo describes a mounted filesystem from findmnt.
type MountInfo struct {
	Source      string `json:"source"`
	Target      string `json:"target"`
	FSType      string `json:"fstype"`
	Options     string `json:"options"`
	SizeBytes   int64  `json:"size_bytes"`
	UsedBytes   int64  `json:"used_bytes"`
	AvailBytes  int64  `json:"avail_bytes"`
	UsePercent  int    `json:"use_percent"`
}

// GatherDisks runs lsblk and returns a flat list of all block devices.
func GatherDisks() ([]BlockDevice, error) {
	out, err := exec.Command("lsblk", "-J", "-b", "-o",
		"PATH,NAME,TYPE,SIZE,RO,MODEL,VENDOR,FSTYPE,LABEL,UUID,MOUNTPOINT").Output()
	if err != nil {
		return nil, fmt.Errorf("running lsblk: %w", err)
	}

	var data struct {
		BlockDevices []BlockDevice `json:"blockdevices"`
	}
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("parsing lsblk JSON: %w", err)
	}

	return flattenDevices(data.BlockDevices), nil
}

// flattenDevices walks the lsblk tree and returns a flat slice.
func flattenDevices(devs []BlockDevice) []BlockDevice {
	var result []BlockDevice
	for _, d := range devs {
		result = append(result, d)
		if len(d.Children) > 0 {
			result = append(result, flattenDevices(d.Children)...)
		}
	}
	return result
}

// FindDevice returns the block device with the given path, or nil.
func FindDevice(devices []BlockDevice, path string) *BlockDevice {
	for i := range devices {
		if devices[i].Path == path {
			return &devices[i]
		}
	}
	return nil
}

// GatherMounts runs findmnt and returns mount info for all mounted filesystems.
func GatherMounts() ([]MountInfo, error) {
	out, err := exec.Command("findmnt", "-J").Output()
	if err != nil {
		return nil, fmt.Errorf("running findmnt: %w", err)
	}

	var data struct {
		Filesystems []struct {
			Source    string `json:"source"`
			Target    string `json:"target"`
			FSType    string `json:"fstype"`
			Options   string `json:"options"`
			Children  []struct {
				Source  string `json:"source"`
				Target  string `json:"target"`
				FSType  string `json:"fstype"`
				Options string `json:"options"`
			} `json:"children,omitempty"`
		} `json:"filesystems"`
	}
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("parsing findmnt JSON: %w", err)
	}

	var result []MountInfo
	for _, fs := range data.Filesystems {
		result = append(result, buildMountInfo(fs.Source, fs.Target, fs.FSType, fs.Options))
		for _, child := range fs.Children {
			result = append(result, buildMountInfo(child.Source, child.Target, child.FSType, child.Options))
		}
	}

	return result, nil
}

func buildMountInfo(source, target, fstype, options string) MountInfo {
	mi := MountInfo{
		Source:  source,
		Target:  target,
		FSType:  fstype,
		Options: options,
	}

	// Fill in usage stats from statfs
	var stat unixStatfs
	if err := statfsCall(target, &stat); err == nil {
		mi.SizeBytes = int64(stat.Bsize) * int64(stat.Blocks)
		mi.AvailBytes = int64(stat.Bsize) * int64(stat.Bavail)
		mi.UsedBytes = mi.SizeBytes - mi.AvailBytes
		if mi.SizeBytes > 0 {
			mi.UsePercent = int(mi.UsedBytes * 100 / mi.SizeBytes)
		}
	}

	return mi
}

// IsMounted returns true if the given device path has a mountpoint.
func IsMounted(path string) (bool, error) {
	devices, err := GatherDisks()
	if err != nil {
		return false, err
	}
	dev := FindDevice(devices, path)
	if dev == nil {
		return false, fmt.Errorf("device %q not found", path)
	}
	return dev.Mountpoint != "", nil
}

// MountDevice mounts a block device at the given mountpoint.
func MountDevice(path, mountpoint string, options []string) error {
	if err := os.MkdirAll(mountpoint, 0o755); err != nil {
		return fmt.Errorf("creating mount point %s: %w", mountpoint, err)
	}

	args := []string{}
	if len(options) > 0 {
		args = append(args, "-o", strings.Join(options, ","))
	}
	args = append(args, path, mountpoint)

	if err := exec.Command("mount", args...).Run(); err != nil {
		return fmt.Errorf("mounting %s at %s: %w", path, mountpoint, err)
	}
	return nil
}

// UnmountDevice unmounts a block device or mountpoint.
func UnmountDevice(path string, lazy bool) error {
	args := []string{}
	if lazy {
		args = append(args, "-l")
	}
	args = append(args, path)

	if err := exec.Command("umount", args...).Run(); err != nil {
		return fmt.Errorf("unmounting %s: %w", path, err)
	}
	return nil
}

// FormatDevice formats a block device with the given filesystem type.
func FormatDevice(path, fstype, label string) error {
	var cmd *exec.Cmd
	switch fstype {
	case "ext4":
		args := []string{"-F"}
		if label != "" {
			args = append(args, "-L", label)
		}
		args = append(args, path)
		cmd = exec.Command("mkfs.ext4", args...)
	case "ext3":
		args := []string{"-F"}
		if label != "" {
			args = append(args, "-L", label)
		}
		args = append(args, path)
		cmd = exec.Command("mkfs.ext3", args...)
	case "xfs":
		args := []string{}
		if label != "" {
			args = append(args, "-L", label)
		}
		args = append(args, path)
		cmd = exec.Command("mkfs.xfs", args...)
	case "btrfs":
		args := []string{"-f"}
		if label != "" {
			args = append(args, "-L", label)
		}
		args = append(args, path)
		cmd = exec.Command("mkfs.btrfs", args...)
	case "fat32":
		args := []string{"-F", "32"}
		if label != "" {
			args = append(args, "-n", label)
		}
		args = append(args, path)
		cmd = exec.Command("mkfs.fat", args...)
	case "swap":
		if label != "" {
			cmd = exec.Command("mkswap", "-L", label, path)
		} else {
			cmd = exec.Command("mkswap", path)
		}
	default:
		return fmt.Errorf("unsupported filesystem type %q (supported: ext4, ext3, xfs, btrfs, fat32, swap)", fstype)
	}

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("formatting %s as %s: %w\n%s", path, fstype, err, string(out))
	}
	return nil
}

// ExpandDevice expands the partition and filesystem to fill available space.
func ExpandDevice(path string) error {
	// Grow the partition (best-effort with growpart)
	if err := exec.Command("growpart", path, "1").Run(); err != nil {
		// growpart may not be available or partition may already be max size;
		// continue to try filesystem resize anyway.
	}

	// Detect the filesystem type and resize accordingly
	devices, err := GatherDisks()
	if err != nil {
		return fmt.Errorf("listing devices after growpart: %w", err)
	}
	dev := FindDevice(devices, path)
	if dev == nil {
		return fmt.Errorf("device %q not found after growpart", path)
	}

	switch dev.FSType {
	case "ext4", "ext3":
		if err := exec.Command("resize2fs", path).Run(); err != nil {
			return fmt.Errorf("resizing ext4 filesystem on %s: %w", path, err)
		}
	case "xfs":
		// XFS must be mounted to grow
		mounted := dev.Mountpoint != ""
		if !mounted {
			return fmt.Errorf("xfs filesystem on %s must be mounted to expand", path)
		}
		if err := exec.Command("xfs_growfs", dev.Mountpoint).Run(); err != nil {
			return fmt.Errorf("resizing xfs filesystem on %s: %w", path, err)
		}
	case "btrfs":
		mounted := dev.Mountpoint != ""
		if !mounted {
			return fmt.Errorf("btrfs filesystem on %s must be mounted to expand", path)
		}
		if err := exec.Command("btrfs", "filesystem", "resize", "max", dev.Mountpoint).Run(); err != nil {
			return fmt.Errorf("resizing btrfs filesystem on %s: %w", path, err)
		}
	default:
		return fmt.Errorf("cannot expand filesystem type %q on %s", dev.FSType, path)
	}

	return nil
}

// PartitionByLabel finds the partition path with the given PARTLABEL on the
// specified disk.
func PartitionByLabel(disk, label string) (string, error) {
	out, err := exec.Command("lsblk", "-nrpo", "PATH,PARTLABEL", disk).Output()
	if err != nil {
		return "", fmt.Errorf("running lsblk on %s: %w", disk, err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.SplitN(strings.TrimSpace(line), " ", 2)
		if len(fields) == 2 && fields[1] == label {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("partition %s not found on %s", label, disk)
}

// ParseSizeBytes converts a human size string (e.g. "32G") to bytes.
func ParseSizeBytes(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	multiplier := int64(1)
	last := s[len(s)-1]
	switch last {
	case 'K', 'k':
		multiplier = 1024
		s = s[:len(s)-1]
	case 'M', 'm':
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	case 'G', 'g':
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	case 'T', 't':
		multiplier = 1024 * 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}

	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}
	return n * multiplier, nil
}

// FormatSizeBytes converts bytes to a human-readable string.
func FormatSizeBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// SupportedFSTypes returns the list of filesystem types that can be created.
func SupportedFSTypes() []string {
	return []string{"ext4", "ext3", "xfs", "btrfs", "fat32", "swap"}
}

// joinPath is a thin wrapper for testability.
func joinPath(parts ...string) string {
	return filepath.Join(parts...)
}