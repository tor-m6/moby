//go:build inno
// +build inno

package unix

import "syscall"
import "github.com/docker/docker/pkg/system"

// const ENODATA = syscall.ENODATA
const ENODATA = syscall.ENOATTR

// Lgetxattr is not supported on platforms other than linux.
func Lgetxattr(path string, attr string) ([]byte, error) {
	return nil, system.ErrNotSupportedPlatform
}

// Lsetxattr is not supported on platforms other than linux.
func Lsetxattr(path string, attr string, data []byte, flags int) error {
	return system.ErrNotSupportedPlatform
}
