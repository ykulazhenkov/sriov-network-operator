package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/log"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
)

var DpdkDrivers = []string{"igb_uio", "vfio-pci", "uio_pci_generic"}

// Unbind unbind driver for one device
func Unbind(pciAddr string) error {
	log.Log.V(2).Info("Unbind(): unbind device driver for device", "device", pciAddr)
	return UnbindDriverByBusAndDevice(busPci, pciAddr)
}

// BindDpdkDriver bind dpdk driver for one device
// Bind the device given by "pciAddr" to the driver "driver"
func BindDpdkDriver(pciAddr, driver string) error {
	log.Log.V(2).Info("BindDpdkDriver(): bind device to driver",
		"device", pciAddr, "driver", driver)

	curDriver, err := getDriverByBusAndDevice(busPci, pciAddr)
	if err != nil {
		return err
	}
	if curDriver != "" {
		if curDriver == driver {
			log.Log.V(2).Info("BindDpdkDriver(): device already bound to driver",
				"device", pciAddr, "driver", driver)
			return nil
		}

		if err := UnbindDriverByBusAndDevice(busPci, pciAddr); err != nil {
			return err
		}
	}
	if err := setDriverOverride(busPci, pciAddr, driver); err != nil {
		return err
	}
	if err := bindDriver(busPci, pciAddr, driver); err != nil {
		_, innerErr := os.Readlink(filepath.Join(sysBusPciDevices, pciAddr, "iommu_group"))
		if innerErr != nil {
			log.Log.Error(err, "Could not read IOMMU group for device", "device", pciAddr)
			return fmt.Errorf(
				"cannot bind driver %s to device %s, make sure IOMMU is enabled in BIOS. %w", driver, pciAddr, innerErr)
		}
		return err
	}
	if err := setDriverOverride(busPci, pciAddr, ""); err != nil {
		return err
	}

	return nil
}

// BindDefaultDriver bind driver for one device
// Bind the device given by "pciAddr" to the default driver
func BindDefaultDriver(pciAddr string) error {
	log.Log.V(2).Info("BindDefaultDriver(): bind device to default driver", "device", pciAddr)

	curDriver, err := getDriverByBusAndDevice(busPci, pciAddr)
	if err != nil {
		return err
	}

	if curDriver != "" {
		if !sriovnetworkv1.StringInArray(curDriver, DpdkDrivers) {
			log.Log.V(2).Info("BindDefaultDriver(): device already bound to default driver",
				"device", pciAddr, "driver", curDriver)
			return nil
		}
		if err := UnbindDriverByBusAndDevice(busPci, pciAddr); err != nil {
			return err
		}
	}
	if err := setDriverOverride(busPci, pciAddr, ""); err != nil {
		return err
	}
	if err := probeDriver(busPci, pciAddr); err != nil {
		return err
	}
	return nil
}

// BindDriverByBusAndDevice binds device to the provided driver
// bus - the bus path in the sysfs, e.g. "pci" or "vdpa"
// device - the name of the device on the bus, e.g. 0000:85:1e.5 for PCI or vpda1 for VDPA
// driver - the name of the driver, e.g. vfio-pci or vhost_vdpa.
func BindDriverByBusAndDevice(bus, device, driver string) error {
	log.Log.V(2).Info("BindDriverByBusAndDevice(): bind device to driver",
		"bus", bus, "device", device, "driver", driver)

	curDriver, err := getDriverByBusAndDevice(bus, device)
	if err != nil {
		return err
	}
	if curDriver != "" {
		if curDriver == driver {
			log.Log.V(2).Info("BindDriverByBusAndDevice(): device already bound to driver",
				"bus", bus, "device", device, "driver", driver)
			return nil
		}
		if err := UnbindDriverByBusAndDevice(bus, device); err != nil {
			return err
		}
	}
	if err := setDriverOverride(bus, device, driver); err != nil {
		return err
	}
	if err := bindDriver(bus, device, driver); err != nil {
		return err
	}
	return setDriverOverride(bus, device, "")
}

// UnbindDriverByBusAndDevice unbind device identified by bus and device ID from the driver
// bus - the bus path in the sysfs, e.g. "pci" or "vdpa"
// device - the name of the device on the bus, e.g. 0000:85:1e.5 for PCI or vpda1 for VDPA
func UnbindDriverByBusAndDevice(bus, device string) error {
	log.Log.V(2).Info("UnbindDriverByBusAndDevice(): unbind device driver for device", "bus", bus, "device", device)
	driver, err := getDriverByBusAndDevice(bus, device)
	if err != nil {
		return err
	}
	if driver == "" {
		log.Log.V(2).Info("UnbindDriverByBusAndDevice(): device has no driver", "bus", bus, "device", device)
		return nil
	}
	return unbindDriver(bus, device, driver)
}

// returns driver for device on the bus
func getDriverByBusAndDevice(bus, device string) (string, error) {
	driverLink := filepath.Join(sysBus, bus, "devices", device, "driver")
	driverInfo, err := os.Readlink(driverLink)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		log.Log.Error(err, "getDriverByBusAndDevice(): error getting driver info for device", "bus", bus, "device", device)
		return "", err
	}
	log.Log.V(2).Info("getDriverByBusAndDevice(): driver for device", "bus", bus, "device", device, "driver", driverInfo)
	return filepath.Base(driverInfo), nil
}

// binds device to the provide driver
func bindDriver(bus, device, driver string) error {
	bindPath := filepath.Join(sysBus, bus, "drivers", driver, "bind")
	err := os.WriteFile(bindPath, []byte(device), os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "bindDriver(): failed to bind driver", "bus", bus, "device", device, "driver", driver)
		return err
	}
	return nil
}

// unbind device from the driver
func unbindDriver(bus, device, driver string) error {
	unbindPath := filepath.Join(sysBus, bus, "drivers", driver, "unbind")
	err := os.WriteFile(unbindPath, []byte(device), os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "unbindDriver(): failed to unbind driver", "bus", bus, "device", device, "driver", driver)
		return err
	}
	return nil
}

// probes driver for device on the bus
func probeDriver(bus, device string) error {
	probePath := filepath.Join(sysBus, bus, "drivers_probe")
	err := os.WriteFile(probePath, []byte(device), os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "probeDriver(): failed to trigger driver probe", "bus", bus, "device", device)
		return err
	}
	return nil
}

// set driver override for the bus/device,
// resets overrride if override arg is """
func setDriverOverride(bus, device, override string) error {
	driverOverridePath := filepath.Join(sysBus, bus, "devices", device, "driver_override")
	if override != "" {
		log.Log.V(2).Info("setDriverOverride(): configure driver override for device", "bus", bus, "device", device, "driver", override)
	} else {
		log.Log.V(2).Info("setDriverOverride(): reset driver override for device", "bus", bus, "device", device)
	}
	err := os.WriteFile(driverOverridePath, []byte(override), os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "setDriverOverride(): fail to write driver_override for device",
			"bus", bus, "device", device, "driver", override)
		return err
	}
	return nil
}

func hasDriver(pciAddr string) (bool, string) {
	driver, err := getDriverByBusAndDevice(busPci, pciAddr)
	if err != nil {
		log.Log.V(2).Info("hasDriver(): device driver is empty for device", "device", pciAddr)
		return false, ""
	}
	if driver != "" {
		log.Log.V(2).Info("hasDriver(): device driver for device", "device", pciAddr, "driver", driver)
		return true, driver
	}
	return false, ""
}
