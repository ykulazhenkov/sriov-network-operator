package vdpa

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/k8snetworkplumbingwg/govdpa/pkg/kvdpa"
	"github.com/vishvananda/netlink"
	"sigs.k8s.io/controller-runtime/pkg/log"

	constants "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	netlinkLibPkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/netlink"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/types"
)

type vdpa struct {
	kernel     types.KernelInterface
	netlinkLib netlinkLibPkg.NetlinkLib
}

func New(k types.KernelInterface, netlinkLib netlinkLibPkg.NetlinkLib) types.VdpaInterface {
	return &vdpa{kernel: k, netlinkLib: netlinkLib}
}

// CreateVDPADevice creates VDPA device for VF with required type,
// pciAddr - PCI address of the VF
// vdpaType - type of the VDPA device to create: virtio of vhost
func (v *vdpa) CreateVDPADevice(pciAddr, vdpaType string) error {
	log.Log.V(2).Info("CreateVDPADevice(): create VDPA device for VF",
		"device", pciAddr, "vdpaType", vdpaType)
	expectedDriver := vdpaTypeToDriver(vdpaType)
	if expectedDriver == "" {
		return fmt.Errorf("unknown VDPA device type: %s", vdpaType)
	}
	expectedVDPAName := generateVDPADevName(pciAddr)
	_, err := v.netlinkLib.VDPAGetDevByName(expectedVDPAName)
	if err != nil {
		if !errors.Is(err, syscall.ENODEV) {
			log.Log.Error(err, "CreateVDPADevice(): fail to check if VDPA device exist",
				"device", pciAddr, "vdpaDev", expectedVDPAName)
			return err
		}
		// first try to create VDPA device with MaxVQP parameter set to 32 to exactly match HW offloading use-case with the
		// old swtichdev implementation. Create device without MaxVQP parameter if it is not supported.
		if err := v.netlinkLib.VDPANewDev(expectedVDPAName, constants.BusPci, pciAddr, netlink.VDPANewDevParams{MaxVQP: 32}); err != nil {
			if !errors.Is(err, syscall.ENOTSUP) {
				log.Log.Error(err, "CreateVDPADevice(): fail to create VDPA device with MaxVQP parameter",
					"device", pciAddr, "vdpaDev", expectedVDPAName)
				return err
			}
			log.Log.V(2).Info("failed to create VDPA device with MaxVQP parameter, try without it")
			if err := v.netlinkLib.VDPANewDev(expectedVDPAName, constants.BusPci, pciAddr, netlink.VDPANewDevParams{}); err != nil {
				log.Log.Error(err, "CreateVDPADevice(): fail to create VDPA device without MaxVQP parameter",
					"device", pciAddr, "vdpaDev", expectedVDPAName)
				return err
			}
		}
	}
	err = v.kernel.BindDriverByBusAndDevice(constants.BusVdpa, expectedVDPAName, expectedDriver)
	if err != nil {
		log.Log.Error(err, "CreateVDPADevice(): fail to bind VDPA device to the driver",
			"device", pciAddr, "vdpaDev", expectedVDPAName, "driver", expectedDriver)
		return err
	}
	return nil
}

// DeleteVDPADevice removes VDPA device for provided pci address
// pciAddr - PCI address of the VF
func (v *vdpa) DeleteVDPADevice(pciAddr string) error {
	log.Log.V(2).Info("DeleteVDPADevice(): delete VDPA device for VF",
		"device", pciAddr)
	expectedVDPAName := generateVDPADevName(pciAddr)
	if err := v.netlinkLib.VDPADelDev(expectedVDPAName); err != nil {
		if errors.Is(err, syscall.ENODEV) {
			log.Log.V(2).Info("DeleteVDPADevice(): VDPA device not found",
				"device", pciAddr, "name", expectedVDPAName)
			return nil
		}
		log.Log.Error(err, "DeleteVDPADevice(): fail to remove VDPA device",
			"device", pciAddr, "name", expectedVDPAName)
		return err
	}
	return nil
}

// DiscoverVDPAType returns type of existing VDPA device for VF,
// returns empty string if VDPA device not found or unknown driver is in use
// pciAddr - PCI address of the VF
func (v *vdpa) DiscoverVDPAType(pciAddr string) string {
	expectedVDPADevName := generateVDPADevName(pciAddr)
	_, err := v.netlinkLib.VDPAGetDevByName(expectedVDPADevName)
	if err != nil {
		if errors.Is(err, syscall.ENODEV) {
			log.Log.V(2).Info("DiscoverVDPAType(): VDPA device for VF not found", "device", pciAddr)
			return ""
		}
		log.Log.Error(err, "DiscoverVDPAType(): unable to get VF VDPA devices", "device", pciAddr)
		return ""
	}
	driverName, err := v.kernel.GetDriverByBusAndDevice(constants.BusVdpa, expectedVDPADevName)
	if err != nil {
		log.Log.Error(err, "DiscoverVDPAType(): unable to get driver info for VF VDPA devices", "device", pciAddr)
		return ""
	}
	if driverName == "" {
		log.Log.V(2).Info("DiscoverVDPAType(): VDPA device has no driver", "device", pciAddr)
		return ""
	}
	vdpaType := vdpaDriverToType(driverName)
	if vdpaType == "" {
		log.Log.Error(nil, "DiscoverVDPAType(): WARNING: unknown VDPA device type for VF, ignore",
			"device", pciAddr, "driver", driverName)
	}
	return vdpaType
}

// generates predictable name for VDPA device, example: vpda:0000:03:00.1
func generateVDPADevName(pciAddr string) string {
	return "vdpa:" + pciAddr
}

// vdpa type to driver name conversion
func vdpaTypeToDriver(vdpaType string) string {
	switch vdpaType {
	case constants.VdpaTypeVhost:
		return kvdpa.VhostVdpaDriver
	case constants.VdpaTypeVirtio:
		return kvdpa.VirtioVdpaDriver
	default:
		return ""
	}
}

// vdpa driver name to type conversion
func vdpaDriverToType(driver string) string {
	switch driver {
	case kvdpa.VhostVdpaDriver:
		return constants.VdpaTypeVhost
	case kvdpa.VirtioVdpaDriver:
		return constants.VdpaTypeVirtio
	default:
		return ""
	}
}
