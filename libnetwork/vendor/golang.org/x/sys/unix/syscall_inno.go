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

// SockaddrNetlink implements the Sockaddr interface for AF_NETLINK type sockets.
type SockaddrNetlink struct {
	Family uint16
	Pad    uint16
	Pid    uint32
	Groups uint32
	raw    RawSockaddrNetlink
}

// NetlinkMessage represents a netlink message.
type NetlinkMessage struct {
	Header NlMsghdr
	Data   []byte
}

// ParseNetlinkMessage parses b as an array of netlink messages and
// returns the slice containing the NetlinkMessage structures.
func ParseNetlinkMessage(b []byte) ([]NetlinkMessage, error) {
	var msgs []NetlinkMessage
	for len(b) >= NLMSG_HDRLEN {
		h, dbuf, dlen, err := netlinkMessageHeaderAndData(b)
		if err != nil {
			return nil, err
		}
		m := NetlinkMessage{Header: *h, Data: dbuf[:int(h.Len)-NLMSG_HDRLEN]}
		msgs = append(msgs, m)
		b = b[dlen:]
	}
	return msgs, nil
}

func netlinkMessageHeaderAndData(b []byte) (*NlMsghdr, []byte, int, error) {
	h := (*NlMsghdr)(unsafe.Pointer(&b[0]))
	l := nlmAlignOf(int(h.Len))
	if int(h.Len) < NLMSG_HDRLEN || l > len(b) {
		return nil, nil, 0, EINVAL
	}
	return h, b[NLMSG_HDRLEN:], l, nil
}

// Round the length of a netlink message up to align it properly.
func nlmAlignOf(msglen int) int {
	return (msglen + NLMSG_ALIGNTO - 1) & ^(NLMSG_ALIGNTO - 1)
}

// Round the length of a netlink route attribute up to align it
// properly.
func rtaAlignOf(attrlen int) int {
	return (attrlen + RTA_ALIGNTO - 1) & ^(RTA_ALIGNTO - 1)
}

// NetlinkRouteAttr represents a netlink route attribute.
type NetlinkRouteAttr struct {
	Attr  RtAttr
	Value []byte
}

// ParseNetlinkRouteAttr parses m's payload as an array of netlink
// route attributes and returns the slice containing the
// NetlinkRouteAttr structures.
func ParseNetlinkRouteAttr(m *NetlinkMessage) ([]NetlinkRouteAttr, error) {
	var b []byte
	switch m.Header.Type {
	case RTM_NEWLINK, RTM_DELLINK:
		b = m.Data[SizeofIfInfomsg:]
	case RTM_NEWADDR, RTM_DELADDR:
		b = m.Data[SizeofIfAddrmsg:]
	case RTM_NEWROUTE, RTM_DELROUTE:
		b = m.Data[SizeofRtMsg:]
	default:
		return nil, EINVAL
	}
	var attrs []NetlinkRouteAttr
	for len(b) >= SizeofRtAttr {
		a, vbuf, alen, err := netlinkRouteAttrAndValue(b)
		if err != nil {
			return nil, err
		}
		ra := NetlinkRouteAttr{Attr: *a, Value: vbuf[:int(a.Len)-SizeofRtAttr]}
		attrs = append(attrs, ra)
		b = b[alen:]
	}
	return attrs, nil
}

func netlinkRouteAttrAndValue(b []byte) (*RtAttr, []byte, int, error) {
	a := (*RtAttr)(unsafe.Pointer(&b[0]))
	if int(a.Len) < SizeofRtAttr || int(a.Len) > len(b) {
		return nil, nil, 0, EINVAL
	}
	return a, b[SizeofRtAttr:], rtaAlignOf(int(a.Len)), nil
}
