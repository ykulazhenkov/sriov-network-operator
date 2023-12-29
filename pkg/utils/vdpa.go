package utils

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/k8snetworkplumbingwg/govdpa/pkg/kvdpa"
	"sigs.k8s.io/controller-runtime/pkg/log"

	constants "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
)

// DeleteVDPADevice removes VDPA device for provided pci address
func DeleteVDPADevice(pciAddr string) error {
	log.Log.V(2).Info("DeleteVDPADevice(): delete VDPA device for VF",
		"device", pciAddr)
	expectedVDPAName := generateVDPADevName(pciAddr)
	if err := kvdpa.DeleteVdpaDevice(expectedVDPAName); err != nil {
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

// CreateVDPADevice creates VDPA device for VF with required type
func CreateVDPADevice(pciAddr, vdpaType string) error {
	log.Log.V(2).Info("CreateVDPADevice(): create VDPA device for VF",
		"device", pciAddr, "vdpaType", vdpaType)
	var expectedDriver string
	switch vdpaType {
	case constants.VdpaTypeVhost:
		expectedDriver = kvdpa.VhostVdpaDriver
	case constants.VdpaTypeVirtio:
		expectedDriver = kvdpa.VirtioVdpaDriver
	default:
		return fmt.Errorf("unknown VDPA device type: %s", vdpaType)
	}

	expectedVDPAName := generateVDPADevName(pciAddr)
	_, err := kvdpa.GetVdpaDevice(expectedVDPAName)
	if err != nil {
		if !errors.Is(err, syscall.ENODEV) {
			log.Log.Error(err, "CreateVDPADevice(): fail to check if VDPA device exist",
			"device", pciAddr, "vdpaDev", expectedVDPAName)
			return err
		}
		if err := kvdpa.AddVdpaDevice("pci/"+pciAddr, expectedVDPAName); err != nil {
			log.Log.Error(err, "CreateVDPADevice(): fail to create VDPA device",
				"device", pciAddr, "vdpaDev", expectedVDPAName)
			return err
		}
	}
	err = BindDriverByBusAndDevice(busVdpa, expectedVDPAName, expectedDriver)
	if err != nil {
		log.Log.Error(err, "CreateVDPADevice(): fail to bind VDPA device to the driver",
			"device", pciAddr, "vdpaDev", expectedVDPAName, "driver", expectedDriver)
		return err
	}
	return nil
}

// returns existing VDPA device type for VF,
// returns empty string if VDPA device not found or unknown driver is in use
func discoverVDPAType(pciAddr string) string {
	expectedVDPADevName := generateVDPADevName(pciAddr)
	vdpaDev, err := kvdpa.GetVdpaDevice(expectedVDPADevName)
	if err != nil {
		if errors.Is(err, syscall.ENODEV) {
			log.Log.V(2).Info("discoverVDPAType(): VDPA device for VF not found", "device", pciAddr)
			return ""
		}
		log.Log.Error(err, "getVfInfo(): unable to get VF VDPA devices", "device", pciAddr)
		return ""
	}
	driverName := vdpaDev.Driver()
	if driverName == "" {
		log.Log.V(2).Info("discoverVDPAType(): VDPA device has no driver", "device", pciAddr)
		return ""
	}
	var vdpaType string
	switch driverName {
	case kvdpa.VhostVdpaDriver:
		vdpaType = constants.VdpaTypeVhost
	case kvdpa.VirtioVdpaDriver:
		vdpaType = constants.VdpaTypeVirtio
	default:
		log.Log.Error(nil, "getVfInfo(): WARNING: unknown VDPA device driver for VF, ignore",
		"device", pciAddr, "driver", driverName)
	}
	return vdpaType
}

func generateVDPADevName(pciAddr string) string {
	return "vdpa:" + pciAddr
}
