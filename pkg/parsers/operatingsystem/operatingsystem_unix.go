//go:build freebsd || darwin || inno
// +build freebsd darwin inno

package operatingsystem // import "github.com/docker/docker/pkg/parsers/operatingsystem"

import (
	"errors"
	"syscall"

	"golang.org/x/sys/unix"
)

// GetOperatingSystem gets the name of the current operating system.
func GetOperatingSystem() (string, error) {
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

// GetOperatingSystemVersion gets the version of the current operating system, as a string.
func GetOperatingSystemVersion() (string, error) {
	// there's no standard unix way of getting this, sadly...
	return "", errors.New("Unsupported on generic unix")
}

// IsContainerized returns true if we are running inside a container.
// No-op on FreeBSD and Darwin, always returns false.
func IsContainerized() (bool, error) {
	// TODO: Implement jail detection for freeBSD
	return false, errors.New("Cannot detect if we are in container")
}
