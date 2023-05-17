//go:build linux || freebsd || darwin || openbsd || inno
// +build linux freebsd darwin openbsd inno

package layer // import "github.com/docker/docker/layer"

import "github.com/docker/docker/pkg/stringid"

func (ls *layerStore) mountID(name string) string {
	return stringid.GenerateRandomID()
}
