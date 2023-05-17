// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Linux system calls.
// This file is compiled as ordinary Go code,
// but it is also input to mksyscall,
// which parses the //sys lines and generates system call stubs.
// Note that sometimes we use a lowercase //sys name and
// wrap it in our own nicer implementation.

package unix

import (
	"unsafe"
	"syscall"
)

// Automatically generated wrapper for raw_ioctl_ptr/__go_ioctl_ptr
//go:noescape
//extern __go_ioctl_ptr
func c___go_ioctl_ptr(fd _C_int, cmd _C_int, val unsafe.Pointer) _C_int
func ioctl(fd int, req uint, arg uintptr) (err error) {
	syscall.Entersyscall()
	_r := c___go_ioctl_ptr(_C_int(fd), _C_int(req), unsafe.Pointer(arg))
	var errno syscall.Errno
	setErrno := false
	if _r < 0 {
		errno = syscall.GetErrno()
		setErrno = true
	}
	syscall.Exitsyscall()
	if setErrno {
		err = error(errno)
	}
	return
}

// Automatically generated wrapper for raw_ioctl_ptr/__go_ioctl_ptr
func ioctlPtr(fd int, req uint, arg unsafe.Pointer) (err error) {
        syscall.Entersyscall()
        _r := c___go_ioctl_ptr(_C_int(fd), _C_int(req), arg)
        var errno syscall.Errno
        setErrno := false
        if _r < 0 {
                errno = syscall.GetErrno()
                setErrno = true
        }
        syscall.Exitsyscall()
        if setErrno {
                err = error(errno)
        }
        return
}

// Automatically generated wrapper for Shutdown/shutdown
//go:noescape
//extern shutdown
func c_flock(fd _C_int, how _C_int) _C_int
func Flock(fd int, how int) (err error) {
	syscall.Entersyscall()
	_r := c_flock(_C_int(fd), _C_int(how))
	var errno syscall.Errno
	setErrno := false
	if _r < 0 {
		errno = syscall.GetErrno()
		setErrno = true
	}
	syscall.Exitsyscall()
	if setErrno {
		err = errno
	}
	return
}

