//go:build !windows
// +build !windows

package daemon // import "github.com/docker/docker/daemon"

import (
	// "fmt"
	"os"
	"sort"
	// "strconv"
	"strings"

	// mounttypes "github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/container"
	"github.com/docker/docker/pkg/fileutils"
	volumemounts "github.com/docker/docker/volume/mounts"
	"github.com/moby/sys/mount"
)

// sortMounts sorts an array of mounts in lexicographic order. This ensure that
// when mounting, the mounts don't shadow other mounts. For example, if mounting
// /etc and /etc/resolv.conf, /etc/resolv.conf must not be mounted first.
func sortMounts(m []container.Mount) []container.Mount {
	sort.Sort(mounts(m))
	return m
}

// setBindModeIfNull is platform specific processing to ensure the
// shared mode is set to 'z' if it is null. This is called in the case
// of processing a named volume and not a typical bind.
func setBindModeIfNull(bind *volumemounts.MountPoint) {
	if bind.Mode == "" {
		bind.Mode = "z"
	}
}

func (daemon *Daemon) mountVolumes(container *container.Container) error {
	mounts, err := daemon.setupMounts(container)
	if err != nil {
		return err
	}

	for _, m := range mounts {
		dest, err := container.GetResourcePath(m.Destination)
		if err != nil {
			return err
		}

		var stat os.FileInfo
		stat, err = os.Stat(m.Source)
		if err != nil {
			return err
		}
		if err = fileutils.CreateIfNotExists(dest, stat.IsDir()); err != nil {
			return err
		}

		bindMode := "rbind"
		if m.NonRecursive {
			bindMode = "bind"
		}
		writeMode := "ro"
		if m.Writable {
			writeMode = "rw"
		}

		// mountVolumes() seems to be called for temporary mounts
		// outside the container. Soon these will be unmounted with
		// lazy unmount option and given we have mounted the rbind,
		// all the submounts will propagate if these are shared. If
		// daemon is running in host namespace and has / as shared
		// then these unmounts will propagate and unmount original
		// mount as well. So make all these mounts rprivate.
		// Do not use propagation property of volume as that should
		// apply only when mounting happens inside the container.
		opts := strings.Join([]string{bindMode, writeMode, "rprivate"}, ",")
		if err := mount.Mount(m.Source, dest, "", opts); err != nil {
			return err
		}
	}

	return nil
}
