//go:build !windows
// +build !windows

package worker

import (
	// "syscall"
	"golang.org/x/sys/unix"
)

func detectDefaultGCCap(root string) int64 {
	var st unix.Statfs_t
	if err := unix.Statfs(root, &st); err != nil {
		return defaultCap
	}
	diskSize := int64(st.Bsize) * int64(st.Blocks) //nolint unconvert
	avail := diskSize / 10
	return (avail/(1<<30) + 1) * 1e9 // round up
}
