//go:build !windows
// +build !windows

package system // import "github.com/docker/docker/pkg/system"

import (
	// "golang.org/x/sys/unix"
)

// Mkdev is used to build the value of linux devices (in /dev/) which specifies major
// and minor number of the newly created device special file.
// Linux device nodes are a bit weird due to backwards compat with 16 bit device nodes.
// They are, from low to high: the lower 8 bits of the minor, then 12 bits of the major,
// then the top 12 bits of the minor.
func Mkdev(major int64, minor int64) uint32 {
	// return uint32(syscall.Mkdev(uint32(major), uint32(minor)))
	return uint32(mkdev((major), (minor)))
}

func mkdev(major int64, minor int64) uint32 {
	return uint32(((minor & 0xfff00) << 12) | ((major & 0xfff) << 8) | (minor & 0xff))
}

