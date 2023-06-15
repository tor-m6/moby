package ns

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

var (
	NetlinkSocketsTimeout = 3 * time.Second
)

// Init initializes a new network namespace
func Init() {
	logrus.Errorf("ns.Init() not implemented")
}

// SetNamespace sets the initial namespace handler
func SetNamespace() error {
	logrus.Errorf("ns.SetNamespace() not implemented")
	return nil
}

// ParseHandlerInt transforms the namespace handler into an integer
func ParseHandlerInt() int {
	return -1
}

// NlHandle returns the netlink handler
func NlHandle() *netlink.Handle {
	logrus.Errorf("ns.NlHandle() not implemented")
	return nil
}
