package netlink

import (
	"net"

	"github.com/vishvananda/netlink"
)

func New() NetlinkLib {
	return &libWrapper{}
}

type Link interface {
	netlink.Link
}

//go:generate ../../../../../bin/mockgen -destination mock/mock_netlink.go -source netlink.go
type NetlinkLib interface {
	// LinkSetVfNodeGUID sets the node GUID of a vf for the link.
	// Equivalent to: `ip link set dev $link vf $vf node_guid $nodeguid`
	LinkSetVfNodeGUID(link Link, vf int, nodeguid net.HardwareAddr) error
	// LinkSetVfPortGUID sets the port GUID of a vf for the link.
	// Equivalent to: `ip link set dev $link vf $vf port_guid $portguid`
	LinkSetVfPortGUID(link Link, vf int, portguid net.HardwareAddr) error
	// LinkByName finds a link by name and returns a pointer to the object.
	LinkByName(name string) (Link, error)
	// LinkSetVfHardwareAddr sets the hardware address of a vf for the link.
	// Equivalent to: `ip link set $link vf $vf mac $hwaddr`
	LinkSetVfHardwareAddr(link Link, vf int, hwaddr net.HardwareAddr) error
	// LinkSetUp enables the link device.
	// Equivalent to: `ip link set $link up`
	LinkSetUp(link Link) error
	// DevlinkGetDeviceByName provides a pointer to devlink device and nil error,
	// otherwise returns an error code.
	DevLinkGetDeviceByName(bus string, device string) (*netlink.DevlinkDevice, error)
	// DevLinkSetEswitchMode sets eswitch mode if able to set successfully or
	// returns an error code.
	// Equivalent to: `devlink dev eswitch set $dev mode switchdev`
	// Equivalent to: `devlink dev eswitch set $dev mode legacy`
	DevLinkSetEswitchMode(dev *netlink.DevlinkDevice, newMode string) error
}

type libWrapper struct{}

// LinkSetVfNodeGUID sets the node GUID of a vf for the link.
// Equivalent to: `ip link set dev $link vf $vf node_guid $nodeguid`
func (w *libWrapper) LinkSetVfNodeGUID(link Link, vf int, nodeguid net.HardwareAddr) error {
	return netlink.LinkSetVfNodeGUID(link, vf, nodeguid)
}

// LinkSetVfPortGUID sets the port GUID of a vf for the link.
// Equivalent to: `ip link set dev $link vf $vf port_guid $portguid`
func (w *libWrapper) LinkSetVfPortGUID(link Link, vf int, portguid net.HardwareAddr) error {
	return netlink.LinkSetVfPortGUID(link, vf, portguid)
}

// LinkByName finds a link by name and returns a pointer to the object.// LinkByName finds a link by name and returns a pointer to the object.
func (w *libWrapper) LinkByName(name string) (Link, error) {
	return netlink.LinkByName(name)
}

// LinkSetVfHardwareAddr sets the hardware address of a vf for the link.
// Equivalent to: `ip link set $link vf $vf mac $hwaddr`
func (w *libWrapper) LinkSetVfHardwareAddr(link Link, vf int, hwaddr net.HardwareAddr) error {
	return netlink.LinkSetVfHardwareAddr(link, vf, hwaddr)
}

// LinkSetUp enables the link device.
// Equivalent to: `ip link set $link up`
func (w *libWrapper) LinkSetUp(link Link) error {
	return netlink.LinkSetUp(link)
}

// DevlinkGetDeviceByName provides a pointer to devlink device and nil error,
// otherwise returns an error code.
func (w *libWrapper) DevLinkGetDeviceByName(bus string, device string) (*netlink.DevlinkDevice, error) {
	return netlink.DevLinkGetDeviceByName(bus, device)
}

// DevLinkSetEswitchMode sets eswitch mode if able to set successfully or
// returns an error code.
// Equivalent to: `devlink dev eswitch set $dev mode switchdev`
// Equivalent to: `devlink dev eswitch set $dev mode legacy`
func (w *libWrapper) DevLinkSetEswitchMode(dev *netlink.DevlinkDevice, newMode string) error {
	return netlink.DevLinkSetEswitchMode(dev, newMode)
}
