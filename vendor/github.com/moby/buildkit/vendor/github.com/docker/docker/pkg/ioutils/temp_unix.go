//go:build !windows
// +build !windows

package ioutils // import "github.com/docker/docker/pkg/ioutils"

import 	"github.com/docker/docker/myos"


// TempDir on Unix systems is equivalent to os.MkdirTemp.
func TempDir(dir, prefix string) (string, error) {
	return myos.MkdirTemp(dir, prefix)
}
