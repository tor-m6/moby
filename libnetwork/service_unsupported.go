// +build !linux,!windows

package libnetwork

import (
	"fmt"
	"net"
)

func (c *controller) cleanupServiceBindings(nid string) {
}

func (c *controller) addServiceBinding(svcName, svcID, nID, eID, containerName string, vip net.IP, ingressPorts []*PortConfig, serviceAliases, taskAliases []string, ip net.IP, method string) error {
	return fmt.Errorf("not supported")
}

func (c *controller) rmServiceBinding(svcName, svcID, nID, eID, containerName string, vip net.IP, ingressPorts []*PortConfig, serviceAliases []string, taskAliases []string, ip net.IP, method string, deleteSvcRecords bool, fullRemove bool) error {
	return fmt.Errorf("not supported")
}

func (sb *sandbox) populateLoadbalancers(ep *endpoint) {
}

func arrangeIngressFilterRule() {
}

func (c *controller) addContainerNameResolution(nID, eID, containerName string, taskAliases []string, ip net.IP, method string) error {
		return nil
}

func (c *controller) getLBIndex(sid, nid string, ingressPorts []*PortConfig) int {
	return 0
}

func (c *controller) delContainerNameResolution(nID, eID, containerName string, taskAliases []string, ip net.IP, method string) error {
	return nil
}

// cleanupServiceDiscovery when the network is being deleted, erase all the associated service discovery records
func (c *controller) cleanupServiceDiscovery(cleanupNID string) {
}