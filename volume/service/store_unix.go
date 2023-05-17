//go:build linux || freebsd || darwin || inno
// +build linux freebsd darwin inno

package service // import "github.com/docker/docker/volume/service"

// normalizeVolumeName is a platform specific function to normalize the name
// of a volume. This is a no-op on Unix-like platforms
func normalizeVolumeName(name string) string {
	return name
}
