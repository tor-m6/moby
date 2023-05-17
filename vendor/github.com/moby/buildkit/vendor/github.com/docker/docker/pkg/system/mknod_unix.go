//go:build !freebsd && !windows
// +build !freebsd,!windows

package system // import "github.com/docker/docker/pkg/system"

import (
	// "golang.org/x/sys/unix"
	"syscall"
)

// Mknod creates a filesystem node (file, device special file or named pipe) named path
// with attributes specified by mode and dev.
func Mknod(path string, mode uint32, dev int) error {
	return syscall.Mknod(path, mode, dev)
}
