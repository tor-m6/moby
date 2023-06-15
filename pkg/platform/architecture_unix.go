//go:build !windows
// +build !windows

// Package platform provides helper function to get the runtime architecture
// for different platforms.
package platform // import "github.com/docker/docker/pkg/platform"

import (
	"syscall"
	"golang.org/x/sys/unix"
)

// runtimeArchitecture gets the name of the current architecture (x86, x86_64, i86pc, sun4v, ...)
func runtimeArchitecture() (string, error) {
	utsname := &syscall.Utsname{}
	if err := unix.Uname(utsname); err != nil {
		return "", err
	}
	b := make([]byte, len(utsname.Machine))
    for i, v := range utsname.Machine {
        b[i] = byte(v)
    }

	return unix.ByteSliceToString(b[:]), nil
}
