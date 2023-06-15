//go:build !linux && !darwin && !windows && !inno
// +build !linux,!darwin,!windows,!inno

package cnmallocator

import (
	"github.com/moby/swarmkit/v2/manager/allocator/networkallocator"
)

const initializers = nil

// PredefinedNetworks returns the list of predefined network structures
func PredefinedNetworks() []networkallocator.PredefinedNetworkData {
	return nil
}
