//go:build linux

package diskmgmt

import (
	"syscall"
)

type unixStatfs = syscall.Statfs_t

func statfsCall(path string, stat *unixStatfs) error {
	return syscall.Statfs(path, stat)
}