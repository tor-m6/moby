// +build !windows,!plan9,!solaris

package bolt

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
	"golang.org/x/sys/unix"
)

// flock acquires an advisory lock on a file descriptor.
func flock(db *DB, mode os.FileMode, exclusive bool, timeout time.Duration) error {
	var t time.Time
	for {
		// If we're beyond our timeout then return an error.
		// This can only occur after we've attempted a flock once.
		if t.IsZero() {
			t = time.Now()
		} else if timeout > 0 && time.Since(t) > timeout {
			return ErrTimeout
		}
		flag := syscall.LOCK_SH
		if exclusive {
			flag = syscall.LOCK_EX
		}

		// Otherwise attempt to obtain an exclusive lock.
		err := unix.Flock(int(db.file.Fd()), flag|syscall.LOCK_NB)
		if err == nil {
			return nil
		} else if err != syscall.EWOULDBLOCK {
			return err
		}

		// Wait for a bit and try again.
		time.Sleep(50 * time.Millisecond)
	}
}

// funlock releases an advisory lock on a file descriptor.
func funlock(db *DB) error {
	return unix.Flock(int(db.file.Fd()), syscall.LOCK_UN)
}

// mmap memory maps a DB's data file.
func mmap(db *DB, sz int) error {
	// Map the data file to memory.
	b, err := syscall.Mmap(int(db.file.Fd()), 0, sz, syscall.PROT_READ, syscall.MAP_SHARED|db.MmapFlags)
	if err != nil {
		return err
	}

	// Advise the kernel that the mmap is accessed randomly.
	if err := madvise(b, syscall.MADV_RANDOM); err != nil {
		return fmt.Errorf("madvise: %s", err)
	}

	// Save the original byte slice and convert to a byte array pointer.
	db.dataref = b
	db.data = (*[maxMapSize]byte)(unsafe.Pointer(&b[0]))
	db.datasz = sz
	return nil
}

// munmap unmaps a DB's data file from memory.
func munmap(db *DB) error {
	// Ignore the unmap if we have no mapped data.
	if db.dataref == nil {
		return nil
	}

	// Unmap using the original byte slice.
	err := syscall.Munmap(db.dataref)
	db.dataref = nil
	db.data = nil
	db.datasz = 0
	return err
}

var _zero uintptr

// NOTE: This function is copied from stdlib because it is not available on darwin.
func madvise(b []byte, advice int) (err error) {
	var _p2 *byte
	if len(b) > 0 {
		_p2 = (*byte)(unsafe.Pointer(&b[0]))
	} else {
		_p2 = (*byte)(unsafe.Pointer(&_zero))
	}
	syscall.Entersyscall()
	_r := c_madvise(unsafe.Pointer(_p2), uintptr(len(b)), int32(advice))
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

//extern madvise
func c_madvise(addr unsafe.Pointer, n uintptr, flags int32) int32
