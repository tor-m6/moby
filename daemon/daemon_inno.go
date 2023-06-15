//go:build inno
// +build inno

package daemon // import "github.com/docker/docker/daemon"

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"strconv"
	coci "github.com/containerd/containerd/oci"
	"github.com/docker/docker/oci"
	"github.com/docker/docker/pkg/stringid"

	"github.com/moby/sys/mountinfo"
	"github.com/docker/docker/daemon/config"
	"github.com/containerd/containerd/containers"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/container"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runc/libcontainer/user"
	dconfig "github.com/docker/docker/daemon/config"
)

func setupResolvConf(config *config.Config) {}

// On Linux, plugins use a static path for storing execution state,
// instead of deriving path from daemon's exec-root. This is because
// plugin socket files are created here and they cannot exceed max
// path length of 108 bytes.
func getPluginExecRoot(_ *config.Config) string {
	return "/run/docker/plugins"
}

func (daemon *Daemon) cleanupMountsByID(in string) error {
	return nil
}

func (daemon *Daemon) cleanupMounts() error {
	return nil
}

// from oci_linux.go

const (
	sharedPropagationOption = "shared:"
	slavePropagationOption  = "master:"
)

// Get the source mount point of directory passed in as argument. Also return
// optional fields.
func getSourceMount(source string) (string, string, error) {
	// Ensure any symlinks are resolved.
	sourcePath, err := filepath.EvalSymlinks(source)
	if err != nil {
		return "", "", err
	}

	mi, err := mountinfo.GetMounts(mountinfo.ParentsFilter(sourcePath))
	if err != nil {
		return "", "", err
	}
	if len(mi) < 1 {
		return "", "", fmt.Errorf("Can't find mount point of %s", source)
	}

	// find the longest mount point
	var idx, maxlen int
	for i := range mi {
		if len(mi[i].Mountpoint) > maxlen {
			maxlen = len(mi[i].Mountpoint)
			idx = i
		}
	}
	return mi[idx].Mountpoint, mi[idx].Optional, nil
}

// hasMountInfoOption checks if any of the passed any of the given option values
// are set in the passed in option string.
func hasMountInfoOption(opts string, vals ...string) bool {
	for _, opt := range strings.Split(opts, " ") {
		for _, val := range vals {
			if strings.HasPrefix(opt, val) {
				return true
			}
		}
	}
	return false
}

// mergeUlimits merge the Ulimits from HostConfig with daemon defaults, and update HostConfig
func (daemon *Daemon) mergeUlimits(c *containertypes.HostConfig) {
	ulimits := c.Ulimits
	// Merge ulimits with daemon defaults
	ulIdx := make(map[string]struct{})
	for _, ul := range ulimits {
		ulIdx[ul.Name] = struct{}{}
	}
	for name, ul := range daemon.configStore.Ulimits {
		if _, exists := ulIdx[name]; !exists {
			ulimits = append(ulimits, ul)
		}
	}
	c.Ulimits = ulimits
}

func (daemon *Daemon) createSpec(ctx context.Context, c *container.Container) (retSpec *specs.Spec, err error) {
	var (
		opts []coci.SpecOpts
		s    = oci.DefaultSpec()
	)
	opts = append(opts,
		WithCommonOptions(daemon, c),
		// WithCgroups(daemon, c),
		WithResources(c),
		// WithSysctls(c),
		WithDevices(daemon, c),
		WithUser(c),
		WithRlimits(daemon, c),
		// WithNamespaces(daemon, c),
		// WithCapabilities(c),
		// WithSeccomp(daemon, c),
		// WithMounts(daemon, c),
		WithLibnetwork(daemon, c),
		// WithApparmor(c),
		// WithSelinux(c),
		// WithOOMScore(&c.HostConfig.OomScoreAdj),
	)
	if c.NoNewPrivileges {
		opts = append(opts, coci.WithNoNewPrivileges)
	}
	if c.Config.Tty {
		opts = append(opts, WithConsoleSize(c))
	}
	// Set the masked and readonly paths with regard to the host config options if they are set.
	if c.HostConfig.MaskedPaths != nil {
		opts = append(opts, coci.WithMaskedPaths(c.HostConfig.MaskedPaths))
	}
	if c.HostConfig.ReadonlyPaths != nil {
		opts = append(opts, coci.WithReadonlyPaths(c.HostConfig.ReadonlyPaths))
	}
	// if daemon.configStore.Rootless {
	// 	opts = append(opts, WithRootless(daemon))
	// }

	var snapshotter, snapshotKey string
	if daemon.UsesSnapshotter() {
		snapshotter = daemon.imageService.StorageDriver()
		snapshotKey = c.ID
	}

	return &s, coci.ApplyOpts(ctx, nil, &containers.Container{
		ID:          c.ID,
		Snapshotter: snapshotter,
		SnapshotKey: snapshotKey,
	}, &s, opts...)
}
// WithCommonOptions sets common docker options
func WithCommonOptions(daemon *Daemon, c *container.Container) coci.SpecOpts {
	return func(ctx context.Context, _ coci.Client, _ *containers.Container, s *coci.Spec) error {
		if c.BaseFS == "" && !daemon.UsesSnapshotter() {
			return errors.New("populateCommonSpec: BaseFS of container " + c.ID + " is unexpectedly empty")
		}
		linkedEnv, err := daemon.setupLinkedContainers(c)
		if err != nil {
			return err
		}
		if !daemon.UsesSnapshotter() {
			s.Root = &specs.Root{
				Path:     c.BaseFS,
				Readonly: c.HostConfig.ReadonlyRootfs,
			}
		}
		if err := c.SetupWorkingDirectory(daemon.idMapping.RootPair()); err != nil {
			return err
		}
		cwd := c.Config.WorkingDir
		if len(cwd) == 0 {
			cwd = "/"
		}
		s.Process.Args = append([]string{c.Path}, c.Args...)

		// only add the custom init if it is specified and the container is running in its
		// own private pid namespace.  It does not make sense to add if it is running in the
		// host namespace or another container's pid namespace where we already have an init
		// if c.HostConfig.PidMode.IsPrivate() {
		// 	if (c.HostConfig.Init != nil && *c.HostConfig.Init) ||
		// 		(c.HostConfig.Init == nil && daemon.configStore.Init) {
		// 		s.Process.Args = append([]string{inContainerInitPath, "--", c.Path}, c.Args...)
		// 		path := daemon.configStore.InitPath
		// 		if path == "" {
		// 			path, err = exec.LookPath(dconfig.DefaultInitBinary)
		// 			if err != nil {
		// 				return err
		// 			}
		// 		}
		// 		s.Mounts = append(s.Mounts, specs.Mount{
		// 			Destination: inContainerInitPath,
		// 			Type:        "bind",
		// 			Source:      path,
		// 			Options:     []string{"bind", "ro"},
		// 		})
		// 	}
		// }
		s.Process.Cwd = cwd
		s.Process.Env = c.CreateDaemonEnvironment(c.Config.Tty, linkedEnv)
		s.Process.Terminal = c.Config.Tty

		s.Hostname = c.Config.Hostname
		// setLinuxDomainname(c, s)

		// Add default sysctls that are generally safe and useful; currently we
		// grant the capabilities to allow these anyway. You can override if
		// you want to restore the original behaviour.
		// We do not set network sysctls if network namespace is host, or if we are
		// joining an existing namespace, only if we create a new net namespace.
		// if c.HostConfig.NetworkMode.IsPrivate() {
		// 	// We cannot set up ping socket support in a user namespace
		// 	userNS := daemon.configStore.RemappedRoot != "" && c.HostConfig.UsernsMode.IsPrivate()
		// 	if !userNS && !userns.RunningInUserNS() && sysctlExists("net.ipv4.ping_group_range") {
		// 		// allow unprivileged ICMP echo sockets without CAP_NET_RAW
		// 		s.Linux.Sysctl["net.ipv4.ping_group_range"] = "0 2147483647"
		// 	}
		// 	// allow opening any port less than 1024 without CAP_NET_BIND_SERVICE
		// 	if sysctlExists("net.ipv4.ip_unprivileged_port_start") {
		// 		s.Linux.Sysctl["net.ipv4.ip_unprivileged_port_start"] = "0"
		// 	}
		// }

		return nil
	}
}

// WithResources applies the container resources
func WithResources(c *container.Container) coci.SpecOpts {
	return func(ctx context.Context, _ coci.Client, _ *containers.Container, s *coci.Spec) error {
		r := c.HostConfig.Resources
		weightDevices, err := getBlkioWeightDevices(r)
		if err != nil {
			return err
		}
		readBpsDevice, err := getBlkioThrottleDevices(r.BlkioDeviceReadBps)
		if err != nil {
			return err
		}
		writeBpsDevice, err := getBlkioThrottleDevices(r.BlkioDeviceWriteBps)
		if err != nil {
			return err
		}
		readIOpsDevice, err := getBlkioThrottleDevices(r.BlkioDeviceReadIOps)
		if err != nil {
			return err
		}
		writeIOpsDevice, err := getBlkioThrottleDevices(r.BlkioDeviceWriteIOps)
		if err != nil {
			return err
		}

		memoryRes := getMemoryResources(r)
		cpuRes, err := getCPUResources(r)
		if err != nil {
			return err
		}
		blkioWeight := r.BlkioWeight

		specResources := &specs.LinuxResources{
			Memory: memoryRes,
			CPU:    cpuRes,
			BlockIO: &specs.LinuxBlockIO{
				Weight:                  &blkioWeight,
				WeightDevice:            weightDevices,
				ThrottleReadBpsDevice:   readBpsDevice,
				ThrottleWriteBpsDevice:  writeBpsDevice,
				ThrottleReadIOPSDevice:  readIOpsDevice,
				ThrottleWriteIOPSDevice: writeIOpsDevice,
			},
			Pids: getPidsLimit(r),
		}

		if s.Linux.Resources != nil && len(s.Linux.Resources.Devices) > 0 {
			specResources.Devices = s.Linux.Resources.Devices
		}

		s.Linux.Resources = specResources
		return nil
	}
}

func getUser(c *container.Container, username string) (specs.User, error) {
	var usr specs.User
	passwdPath, err := resourcePath(c, user.GetPasswdPath)
	if err != nil {
		return usr, err
	}
	groupPath, err := resourcePath(c, user.GetGroupPath)
	if err != nil {
		return usr, err
	}
	execUser, err := user.GetExecUserPath(username, nil, passwdPath, groupPath)
	if err != nil {
		return usr, err
	}
	usr.UID = uint32(execUser.Uid)
	usr.GID = uint32(execUser.Gid)
	usr.AdditionalGids = []uint32{usr.GID}

	var addGroups []int
	if len(c.HostConfig.GroupAdd) > 0 {
		addGroups, err = user.GetAdditionalGroupsPath(c.HostConfig.GroupAdd, groupPath)
		if err != nil {
			return usr, err
		}
	}
	for _, g := range append(execUser.Sgids, addGroups...) {
		usr.AdditionalGids = append(usr.AdditionalGids, uint32(g))
	}
	return usr, nil
}

// WithUser sets the container's user
func WithUser(c *container.Container) coci.SpecOpts {
	return func(ctx context.Context, _ coci.Client, _ *containers.Container, s *coci.Spec) error {
		var err error
		s.Process.User, err = getUser(c, c.Config.User)
		return err
	}
}

// WithDevices sets the container's devices
func WithDevices(daemon *Daemon, c *container.Container) coci.SpecOpts {
	return func(ctx context.Context, _ coci.Client, _ *containers.Container, s *coci.Spec) error {
		// Build lists of devices allowed and created within the container.
		var devs []specs.LinuxDevice
		devPermissions := s.Linux.Resources.Devices

		if c.HostConfig.Privileged {
			// hostDevices, err := coci.HostDevices()
			// if err != nil {
			// 	return err
			// }
			// devs = append(devs, hostDevices...)

			// // adding device mappings in privileged containers
			// for _, deviceMapping := range c.HostConfig.Devices {
			// 	// issue a warning that custom cgroup permissions are ignored in privileged mode
			// 	if deviceMapping.CgroupPermissions != "rwm" {
			// 		logrus.WithField("container", c.ID).Warnf("custom %s permissions for device %s are ignored in privileged mode", deviceMapping.CgroupPermissions, deviceMapping.PathOnHost)
			// 	}
			// 	// issue a warning that the device path already exists via /dev mounting in privileged mode
			// 	if deviceMapping.PathOnHost == deviceMapping.PathInContainer {
			// 		logrus.WithField("container", c.ID).Warnf("path in container %s already exists in privileged mode", deviceMapping.PathInContainer)
			// 		continue
			// 	}
			// 	d, _, err := oci.DevicesFromPath(deviceMapping.PathOnHost, deviceMapping.PathInContainer, "rwm")
			// 	if err != nil {
			// 		return err
			// 	}
			// 	devs = append(devs, d...)
			// }

			// devPermissions = []specs.LinuxDeviceCgroup{
			// 	{
			// 		Allow:  true,
			// 		Access: "rwm",
			// 	},
			// }
		} else {
			for _, deviceMapping := range c.HostConfig.Devices {
				d, dPermissions, err := oci.DevicesFromPath(deviceMapping.PathOnHost, deviceMapping.PathInContainer, "")
				if err != nil {
					return err
				}
				devs = append(devs, d...)
				devPermissions = append(devPermissions, dPermissions...)
			}

			var err error
			devPermissions, err = oci.AppendDevicePermissionsFromCgroupRules(devPermissions, c.HostConfig.DeviceCgroupRules)
			if err != nil {
				return err
			}
		}

		s.Linux.Devices = append(s.Linux.Devices, devs...)
		s.Linux.Resources.Devices = devPermissions

		for _, req := range c.HostConfig.DeviceRequests {
			if err := daemon.handleDevice(req, s); err != nil {
				return err
			}
		}
		return nil
	}
}

const inContainerInitPath = "/sbin/" + dconfig.DefaultInitBinary

// WithRlimits sets the container's rlimits along with merging the daemon's rlimits
func WithRlimits(daemon *Daemon, c *container.Container) coci.SpecOpts {
	return func(ctx context.Context, _ coci.Client, _ *containers.Container, s *coci.Spec) error {
		var rlimits []specs.POSIXRlimit

		// We want to leave the original HostConfig alone so make a copy here
		hostConfig := *c.HostConfig
		// Merge with the daemon defaults
		daemon.mergeUlimits(&hostConfig)
		for _, ul := range hostConfig.Ulimits {
			rlimits = append(rlimits, specs.POSIXRlimit{
				Type: "RLIMIT_" + strings.ToUpper(ul.Name),
				Soft: uint64(ul.Soft),
				Hard: uint64(ul.Hard),
			})
		}

		s.Process.Rlimits = rlimits
		return nil
	}
}

// WithLibnetwork sets the libnetwork hook
func WithLibnetwork(daemon *Daemon, c *container.Container) coci.SpecOpts {
	return func(ctx context.Context, _ coci.Client, _ *containers.Container, s *coci.Spec) error {
		if s.Hooks == nil {
			s.Hooks = &specs.Hooks{}
		}
		for _, ns := range s.Linux.Namespaces {
			if ns.Type == "network" && ns.Path == "" && !c.Config.NetworkDisabled {
				target := filepath.Join("/proc", strconv.Itoa(os.Getpid()), "exe")
				shortNetCtlrID := stringid.TruncateID(daemon.netController.ID())
				s.Hooks.Prestart = append(s.Hooks.Prestart, specs.Hook{
					Path: target,
					Args: []string{
						"libnetwork-setkey",
						"-exec-root=" + daemon.configStore.GetExecRoot(),
						c.ID,
						shortNetCtlrID,
					},
				})
			}
		}
		return nil
	}
}

func resourcePath(c *container.Container, getPath func() (string, error)) (string, error) {
	p, err := getPath()
	if err != nil {
		return "", err
	}
	return c.GetResourcePath(p)
}
