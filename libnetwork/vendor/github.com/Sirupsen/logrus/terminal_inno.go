// +build linux aix zos inno
// +build !js

package logrus

import "golang.org/x/sys/unix"
import "os"

const ioctlReadTermios = unix.TCGETS

func IsTerminal() bool {
	_, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), ioctlReadTermios)
	return err == nil
}
