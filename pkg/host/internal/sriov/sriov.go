package sriov

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jaypipes/ghw"
	"github.com/vishvananda/netlink"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	dputilsPkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/dputils"
	netlinkPkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/netlink"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/store"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/types"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/utils"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
	mlx "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vendors/mellanox"
)

type sriov struct {
	private       sriovPrivateInterface
	sriovHelper   types.SriovInterface
	utilsHelper   utils.CmdInterface
	kernelHelper  types.KernelInterface
	networkHelper types.NetworkInterface
	udevHelper    types.UdevInterface
	vdpaHelper    types.VdpaInterface
	netlinkLib    netlinkPkg.NetlinkLib
	dputilsLib    dputilsPkg.DPUtilsLib
}

// interface contains internal functions of the sriov struct,
// required for mocking
//
//go:generate ../../../../bin/mockgen -package sriov -destination mock_private.go -source sriov.go
type sriovPrivateInterface interface {
	addUdevRules(iface *sriovnetworkv1.Interface) error
	removeUdevRules(pciAddress string) error
	createVFs(iface *sriovnetworkv1.Interface, ifaceStatus *sriovnetworkv1.InterfaceExt) error
	createSwitchdevVFs(iface *sriovnetworkv1.Interface) error
	unbindAllVFsOnPF(addr string) error
	configSriovPFDevice(iface *sriovnetworkv1.Interface, ifaceStatus *sriovnetworkv1.InterfaceExt) error
	configSriovVFDevices(iface *sriovnetworkv1.Interface, ifaceStatus *sriovnetworkv1.InterfaceExt) error
	checkExternallyManagedPF(iface *sriovnetworkv1.Interface, ifaceStatus *sriovnetworkv1.InterfaceExt) error
}

func New(utilsHelper utils.CmdInterface,
	kernelHelper types.KernelInterface,
	networkHelper types.NetworkInterface,
	udevHelper types.UdevInterface,
	vdpaHelper types.VdpaInterface,
	netlinkLib netlinkPkg.NetlinkLib,
	dputilsLib dputilsPkg.DPUtilsLib) types.SriovInterface {
	s := &sriov{utilsHelper: utilsHelper,
		kernelHelper:  kernelHelper,
		networkHelper: networkHelper,
		udevHelper:    udevHelper,
		vdpaHelper:    vdpaHelper,
		netlinkLib:    netlinkLib,
		dputilsLib:    dputilsLib,
	}
	// use self links to be able to mock
	// public and private functions
	s.private = s
	s.sriovHelper = s
	return s
}

func (s *sriov) SetSriovNumVfs(pciAddr string, numVfs int) error {
	log.Log.V(2).Info("SetSriovNumVfs(): set NumVfs", "device", pciAddr, "numVfs", numVfs)
	numVfsFilePath := filepath.Join(vars.FilesystemRoot, consts.SysBusPciDevices, pciAddr, consts.NumVfsFile)
	bs := []byte(strconv.Itoa(numVfs))
	err := os.WriteFile(numVfsFilePath, []byte("0"), os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "SetSriovNumVfs(): fail to reset NumVfs file", "path", numVfsFilePath)
		return err
	}
	if numVfs == 0 {
		return nil
	}
	err = os.WriteFile(numVfsFilePath, bs, os.ModeAppend)
	if err != nil {
		log.Log.Error(err, "SetSriovNumVfs(): fail to set NumVfs file", "path", numVfsFilePath)
		return err
	}
	return nil
}

func (s *sriov) ResetSriovDevice(ifaceStatus sriovnetworkv1.InterfaceExt) error {
	log.Log.V(2).Info("ResetSriovDevice(): reset SRIOV device", "address", ifaceStatus.PciAddress)
	if err := s.sriovHelper.SetSriovNumVfs(ifaceStatus.PciAddress, 0); err != nil {
		return err
	}
	if ifaceStatus.LinkType == consts.LinkTypeETH {
		var mtu int
		is := sriovnetworkv1.InitialState.GetInterfaceStateByPciAddress(ifaceStatus.PciAddress)
		if is != nil {
			mtu = is.Mtu
		} else {
			mtu = 1500
		}
		log.Log.V(2).Info("ResetSriovDevice(): reset mtu", "value", mtu)
		if err := s.networkHelper.SetNetdevMTU(ifaceStatus.PciAddress, mtu); err != nil {
			return err
		}
	} else if ifaceStatus.LinkType == consts.LinkTypeIB {
		if err := s.networkHelper.SetNetdevMTU(ifaceStatus.PciAddress, 2048); err != nil {
			return err
		}
	}
	return nil
}

func (s *sriov) GetVfInfo(pciAddr string, devices []*ghw.PCIDevice) sriovnetworkv1.VirtualFunction {
	driver, err := s.dputilsLib.GetDriverName(pciAddr)
	if err != nil {
		log.Log.Error(err, "getVfInfo(): unable to parse device driver", "device", pciAddr)
	}
	id, err := s.dputilsLib.GetVFID(pciAddr)
	if err != nil {
		log.Log.Error(err, "getVfInfo(): unable to get VF index", "device", pciAddr)
	}
	vf := sriovnetworkv1.VirtualFunction{
		PciAddress: pciAddr,
		Driver:     driver,
		VfID:       id,
	}

	if mtu := s.networkHelper.GetNetdevMTU(pciAddr); mtu > 0 {
		vf.Mtu = mtu
	}
	if name := s.networkHelper.TryGetInterfaceName(pciAddr); name != "" {
		vf.Name = name
		vf.Mac = s.networkHelper.GetNetDevMac(name)
	}

	for _, device := range devices {
		if pciAddr == device.Address {
			vf.Vendor = device.Vendor.ID
			vf.DeviceID = device.Product.ID
			break
		}
		continue
	}
	return vf
}

func (s *sriov) SetVfGUID(vfAddr string, pfLink netlink.Link) error {
	log.Log.Info("SetVfGUID()", "vf", vfAddr)
	vfID, err := s.dputilsLib.GetVFID(vfAddr)
	if err != nil {
		log.Log.Error(err, "SetVfGUID(): unable to get VF id", "address", vfAddr)
		return err
	}
	guid := utils.GenerateRandomGUID()
	if err := s.netlinkLib.LinkSetVfNodeGUID(pfLink, vfID, guid); err != nil {
		return err
	}
	if err := s.netlinkLib.LinkSetVfPortGUID(pfLink, vfID, guid); err != nil {
		return err
	}
	if err = s.kernelHelper.Unbind(vfAddr); err != nil {
		return err
	}

	return nil
}

func (s *sriov) VFIsReady(pciAddr string) (netlink.Link, error) {
	log.Log.Info("VFIsReady()", "device", pciAddr)
	var err error
	var vfLink netlink.Link
	err = wait.PollImmediate(time.Second, 10*time.Second, func() (bool, error) {
		vfName := s.networkHelper.TryGetInterfaceName(pciAddr)
		vfLink, err = s.netlinkLib.LinkByName(vfName)
		if err != nil {
			log.Log.Error(err, "VFIsReady(): unable to get VF link", "device", pciAddr)
		}
		return err == nil, nil
	})
	if err != nil {
		return vfLink, err
	}
	return vfLink, nil
}

func (s *sriov) SetVfAdminMac(vfAddr string, pfLink, vfLink netlink.Link) error {
	log.Log.Info("SetVfAdminMac()", "vf", vfAddr)

	vfID, err := s.dputilsLib.GetVFID(vfAddr)
	if err != nil {
		log.Log.Error(err, "SetVfAdminMac(): unable to get VF id", "address", vfAddr)
		return err
	}

	if err := s.netlinkLib.LinkSetVfHardwareAddr(pfLink, vfID, vfLink.Attrs().HardwareAddr); err != nil {
		return err
	}

	return nil
}

func (s *sriov) DiscoverSriovDevices(storeManager store.ManagerInterface) ([]sriovnetworkv1.InterfaceExt, error) {
	log.Log.V(2).Info("DiscoverSriovDevices")
	pfList := []sriovnetworkv1.InterfaceExt{}

	pci, err := ghw.PCI()
	if err != nil {
		return nil, fmt.Errorf("DiscoverSriovDevices(): error getting PCI info: %v", err)
	}

	devices := pci.ListDevices()
	if len(devices) == 0 {
		return nil, fmt.Errorf("DiscoverSriovDevices(): could not retrieve PCI devices")
	}

	for _, device := range devices {
		devClass, err := strconv.ParseInt(device.Class.ID, 16, 64)
		if err != nil {
			log.Log.Error(err, "DiscoverSriovDevices(): unable to parse device class, skipping",
				"device", device)
			continue
		}
		if devClass != consts.NetClass {
			// Not network device
			continue
		}

		// TODO: exclude devices used by host system

		if s.dputilsLib.IsSriovVF(device.Address) {
			continue
		}

		driver, err := s.dputilsLib.GetDriverName(device.Address)
		if err != nil {
			log.Log.Error(err, "DiscoverSriovDevices(): unable to parse device driver for device, skipping", "device", device)
			continue
		}

		deviceNames, err := s.dputilsLib.GetNetNames(device.Address)
		if err != nil {
			log.Log.Error(err, "DiscoverSriovDevices(): unable to get device names for device, skipping", "device", device)
			continue
		}

		if len(deviceNames) == 0 {
			// no network devices found, skipping device
			continue
		}

		if !vars.DevMode {
			if !sriovnetworkv1.IsSupportedModel(device.Vendor.ID, device.Product.ID) {
				log.Log.Info("DiscoverSriovDevices(): unsupported device", "device", device)
				continue
			}
		}

		iface := sriovnetworkv1.InterfaceExt{
			PciAddress: device.Address,
			Driver:     driver,
			Vendor:     device.Vendor.ID,
			DeviceID:   device.Product.ID,
		}
		if mtu := s.networkHelper.GetNetdevMTU(device.Address); mtu > 0 {
			iface.Mtu = mtu
		}
		if name := s.networkHelper.TryGetInterfaceName(device.Address); name != "" {
			iface.Name = name
			iface.Mac = s.networkHelper.GetNetDevMac(name)
			iface.LinkSpeed = s.networkHelper.GetNetDevLinkSpeed(name)
		}
		iface.LinkType = s.sriovHelper.GetLinkType(iface)

		pfStatus, exist, err := storeManager.LoadPfsStatus(iface.PciAddress)
		if err != nil {
			log.Log.Error(err, "DiscoverSriovDevices(): failed to load PF status from disk")
		} else {
			if exist {
				iface.ExternallyManaged = pfStatus.ExternallyManaged
			}
		}

		if s.dputilsLib.IsSriovPF(device.Address) {
			iface.TotalVfs = s.dputilsLib.GetSriovVFcapacity(device.Address)
			iface.NumVfs = s.dputilsLib.GetVFconfigured(device.Address)
			if iface.EswitchMode, err = s.sriovHelper.GetNicSriovMode(device.Address); err != nil {
				log.Log.Error(err, "DiscoverSriovDevices(): warning, unable to get device eswitch mode",
					"device", device.Address)
			}
			if s.dputilsLib.SriovConfigured(device.Address) {
				vfs, err := s.dputilsLib.GetVFList(device.Address)
				if err != nil {
					log.Log.Error(err, "DiscoverSriovDevices(): unable to parse VFs for device, skipping",
						"device", device)
					continue
				}
				for _, vf := range vfs {
					instance := s.sriovHelper.GetVfInfo(vf, devices)
					iface.VFs = append(iface.VFs, instance)
				}
			}
		}
		pfList = append(pfList, iface)
	}

	return pfList, nil
}

func (s *sriov) configSriovPFDevice(iface *sriovnetworkv1.Interface, ifaceStatus *sriovnetworkv1.InterfaceExt) error {
	log.Log.V(2).Info("ConfigSriovPFDevice(): configure PF sriov device",
		"device", iface.PciAddress)

	if iface.NumVfs > ifaceStatus.TotalVfs {
		err := fmt.Errorf("cannot config SRIOV device: NumVfs (%d) is larger than TotalVfs (%d)", iface.NumVfs, ifaceStatus.TotalVfs)
		log.Log.Error(err, "configSriovPFDevice(): fail to set NumVfs for device", "device", iface.PciAddress)
		return err
	}
	err := s.private.addUdevRules(iface)
	if err != nil {
		log.Log.Error(err, "configSriovPFDevice(): fail to set add udev rules", "device", iface.PciAddress)
		return err
	}
	err = s.private.createVFs(iface, ifaceStatus)
	if err != nil {
		log.Log.Error(err, "configSriovPFDevice(): fail to set NumVfs for device", "device", iface.PciAddress)
		errRemove := s.private.removeUdevRules(iface.PciAddress)
		if errRemove != nil {
			log.Log.Error(errRemove, "configSriovPFDevice(): fail to remove udev rule", "device", iface.PciAddress)
		}
		return err
	}
	// set PF mtu
	if iface.Mtu > 0 && iface.Mtu > ifaceStatus.Mtu {
		err = s.networkHelper.SetNetdevMTU(iface.PciAddress, iface.Mtu)
		if err != nil {
			log.Log.Error(err, "configSriovPFDevice(): fail to set mtu for PF", "device", iface.PciAddress)
			return err
		}
	}
	return nil
}

func (s *sriov) checkExternallyManagedPF(iface *sriovnetworkv1.Interface, ifaceStatus *sriovnetworkv1.InterfaceExt) error {
	log.Log.V(2).Info("checkExternallyManagedPF(): configure PF sriov device",
		"device", iface.PciAddress)
	if iface.NumVfs > ifaceStatus.NumVfs {
		errMsg := fmt.Sprintf("checkExternallyManagedPF(): number of request virtual functions %d is not equal to configured virtual "+
			"functions %d but the policy is configured as ExternallyManaged for device %s",
			iface.NumVfs, ifaceStatus.NumVfs, iface.PciAddress)
		log.Log.Error(nil, errMsg)
		return fmt.Errorf(errMsg)
	}
	if sriovnetworkv1.NeedToUpdateInterfaceEswitchMode(iface, ifaceStatus) {
		errMsg := fmt.Sprintf("checkExternallyManagedPF(): requested ESwitchMode mode \"%s\" is not equal to configured \"%s\" "+
			"but the policy is configured as ExternallyManaged for device %s",
			sriovnetworkv1.GetEswitchModeFromSpec(iface), sriovnetworkv1.GetEswitchModeFromStatus(ifaceStatus), iface.PciAddress)
		log.Log.Error(nil, errMsg)
		return fmt.Errorf(errMsg)
	}
	if iface.Mtu > 0 && iface.Mtu > ifaceStatus.Mtu {
		err := fmt.Errorf("checkExternallyManagedPF(): requested MTU(%d) is greater than configured MTU(%d) for device %s. cannot change MTU as policy is configured as ExternallyManaged",
			iface.Mtu, ifaceStatus.Mtu, iface.PciAddress)
		log.Log.Error(nil, err.Error())
		return err
	}
	return nil
}

func (s *sriov) configSriovVFDevices(iface *sriovnetworkv1.Interface, ifaceStatus *sriovnetworkv1.InterfaceExt) error {
	log.Log.V(2).Info("configSriovVFDevices(): configure PF sriov device",
		"device", iface.PciAddress)
	if iface.NumVfs > 0 {
		vfAddrs, err := s.dputilsLib.GetVFList(iface.PciAddress)
		if err != nil {
			log.Log.Error(err, "configSriovVFDevices(): unable to parse VFs for device", "device", iface.PciAddress)
		}
		pfLink, err := s.netlinkLib.LinkByName(iface.Name)
		if err != nil {
			log.Log.Error(err, "configSriovVFDevices(): unable to get PF link for device", "device", iface)
			return err
		}

		for _, addr := range vfAddrs {
			var group *sriovnetworkv1.VfGroup

			vfID, err := s.dputilsLib.GetVFID(addr)
			if err != nil {
				log.Log.Error(err, "configSriovVFDevices(): unable to get VF id", "device", iface.PciAddress)
				return err
			}

			for i := range iface.VfGroups {
				if sriovnetworkv1.IndexInRange(vfID, iface.VfGroups[i].VfRange) {
					group = &iface.VfGroups[i]
					break
				}
			}

			// VF group not found.
			if group == nil {
				continue
			}

			// only set GUID and MAC for VF with default driver
			// for userspace drivers like vfio we configure the vf mac using the kernel nic mac address
			// before we switch to the userspace driver
			if yes, d := s.kernelHelper.HasDriver(addr); yes && !sriovnetworkv1.StringInArray(d, vars.DpdkDrivers) {
				// LinkType is an optional field. Let's fallback to current link type
				// if nothing is specified in the SriovNodePolicy
				linkType := iface.LinkType
				if linkType == "" {
					linkType = ifaceStatus.LinkType
				}
				if strings.EqualFold(linkType, consts.LinkTypeIB) {
					if err = s.sriovHelper.SetVfGUID(addr, pfLink); err != nil {
						return err
					}
				} else {
					vfLink, err := s.sriovHelper.VFIsReady(addr)
					if err != nil {
						log.Log.Error(err, "configSriovVFDevices(): VF link is not ready", "address", addr)
						err = s.kernelHelper.RebindVfToDefaultDriver(addr)
						if err != nil {
							log.Log.Error(err, "configSriovVFDevices(): failed to rebind VF", "address", addr)
							return err
						}

						// Try to check the VF status again
						vfLink, err = s.sriovHelper.VFIsReady(addr)
						if err != nil {
							log.Log.Error(err, "configSriovVFDevices(): VF link is not ready", "address", addr)
							return err
						}
					}
					if err = s.sriovHelper.SetVfAdminMac(addr, pfLink, vfLink); err != nil {
						log.Log.Error(err, "configSriovVFDevices(): fail to configure VF admin mac", "device", addr)
						return err
					}
				}
			}

			if err = s.kernelHelper.UnbindDriverIfNeeded(addr, group.IsRdma); err != nil {
				return err
			}

			if sriovnetworkv1.GetEswitchModeFromSpec(iface) == sriovnetworkv1.ESwithModeSwitchDev && group.VdpaType == "" {
				if err := s.vdpaHelper.DeleteVDPADevice(addr); err != nil {
					log.Log.Error(err, "configSriovVFDevices(): fail to delete VDPA device",
						"device", addr)
					return err
				}
			}
			if !sriovnetworkv1.StringInArray(group.DeviceType, vars.DpdkDrivers) {
				if err := s.kernelHelper.BindDefaultDriver(addr); err != nil {
					log.Log.Error(err, "configSriovVFDevices(): fail to bind default driver for device", "device", addr)
					return err
				}
				// only set MTU for VF with default driver
				if group.Mtu > 0 {
					if err := s.networkHelper.SetNetdevMTU(addr, group.Mtu); err != nil {
						log.Log.Error(err, "configSriovVFDevices(): fail to set mtu for VF", "address", addr)
						return err
					}
				}
				if sriovnetworkv1.GetEswitchModeFromSpec(iface) == sriovnetworkv1.ESwithModeSwitchDev && group.VdpaType != "" {
					if err := s.vdpaHelper.CreateVDPADevice(addr, group.VdpaType); err != nil {
						log.Log.Error(err, "configSriovVFDevices(): fail to create VDPA device",
							"vdpaType", group.VdpaType, "device", addr)
						return err
					}
				}
			} else {
				if err := s.kernelHelper.BindDpdkDriver(addr, group.DeviceType); err != nil {
					log.Log.Error(err, "configSriovVFDevices(): fail to bind driver for device",
						"driver", group.DeviceType, "device", addr)
					return err
				}
			}
		}
	}
	return nil
}

func (s *sriov) ConfigSriovDevice(iface *sriovnetworkv1.Interface, ifaceStatus *sriovnetworkv1.InterfaceExt, preConfig bool) error {
	log.Log.V(2).Info("ConfigSriovDevice(): configure sriov device",
		"device", iface.PciAddress, "config", iface, "preConfig", preConfig)
	if !iface.ExternallyManaged {
		if err := s.private.configSriovPFDevice(iface, ifaceStatus); err != nil {
			return err
		}
	}
	if preConfig {
		log.Log.V(2).Info("ConfigSriovDevice(): preConfig is true, unbind all VFs from drivers",
			"device", iface.PciAddress)
		return s.private.unbindAllVFsOnPF(iface.PciAddress)
	}
	// we don't need to validate externally managed PFs during preConfig stage because
	// PFs may not be configured at this stage yet (preConfig stage is executed before NetworkManager, netplan)
	if iface.ExternallyManaged {
		if err := s.private.checkExternallyManagedPF(iface, ifaceStatus); err != nil {
			return err
		}
	}
	if err := s.private.configSriovVFDevices(iface, ifaceStatus); err != nil {
		return err
	}
	// Set PF link up
	pfLink, err := s.netlinkLib.LinkByName(ifaceStatus.Name)
	if err != nil {
		return err
	}
	if pfLink.Attrs().OperState != netlink.OperUp {
		err = s.netlinkLib.LinkSetUp(pfLink)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *sriov) ConfigSriovInterfaces(storeManager store.ManagerInterface,
	interfaces []sriovnetworkv1.Interface, ifaceStatuses []sriovnetworkv1.InterfaceExt, pfsToConfig map[string]bool, preConfig bool) error {
	if s.kernelHelper.IsKernelLockdownMode() && mlx.HasMellanoxInterfacesInSpec(ifaceStatuses, interfaces) {
		log.Log.Error(nil, "cannot use mellanox devices when in kernel lockdown mode")
		return fmt.Errorf("cannot use mellanox devices when in kernel lockdown mode")
	}

	for _, ifaceStatus := range ifaceStatuses {
		configured := false
		for _, iface := range interfaces {
			if iface.PciAddress == ifaceStatus.PciAddress {
				configured = true

				if skip := pfsToConfig[iface.PciAddress]; skip {
					break
				}

				if !sriovnetworkv1.NeedToUpdateSriov(&iface, &ifaceStatus) {
					log.Log.V(2).Info("syncNodeState(): no need update interface", "address", iface.PciAddress)

					// Save the PF status to the host
					err := storeManager.SaveLastPfAppliedStatus(&iface)
					if err != nil {
						log.Log.Error(err, "SyncNodeState(): failed to save PF applied config to host")
						return err
					}

					break
				}
				if err := s.sriovHelper.ConfigSriovDevice(&iface, &ifaceStatus, preConfig); err != nil {
					log.Log.Error(err, "SyncNodeState(): fail to configure sriov interface. resetting interface.", "address", iface.PciAddress)
					if iface.ExternallyManaged {
						log.Log.Info("SyncNodeState(): skipping device reset as the nic is marked as externally created")
					} else {
						if resetErr := s.sriovHelper.ResetSriovDevice(ifaceStatus); resetErr != nil {
							log.Log.Error(resetErr, "SyncNodeState(): failed to reset on error SR-IOV interface")
						}
					}
					return err
				}

				// Save the PF status to the host
				err := storeManager.SaveLastPfAppliedStatus(&iface)
				if err != nil {
					log.Log.Error(err, "SyncNodeState(): failed to save PF applied config to host")
					return err
				}
				break
			}
		}
		if !configured && ifaceStatus.NumVfs > 0 {
			if skip := pfsToConfig[ifaceStatus.PciAddress]; skip {
				continue
			}

			// load the PF info
			pfStatus, exist, err := storeManager.LoadPfsStatus(ifaceStatus.PciAddress)
			if err != nil {
				log.Log.Error(err, "SyncNodeState(): failed to load info about PF status for device",
					"address", ifaceStatus.PciAddress)
				return err
			}

			if !exist {
				log.Log.Info("SyncNodeState(): PF name with pci address has VFs configured but they weren't created by the sriov operator. Skipping the device reset",
					"pf-name", ifaceStatus.Name,
					"address", ifaceStatus.PciAddress)
				continue
			}

			if pfStatus.ExternallyManaged {
				log.Log.Info("SyncNodeState(): PF name with pci address was externally created skipping the device reset",
					"pf-name", ifaceStatus.Name,
					"address", ifaceStatus.PciAddress)
				continue
			} else {
				err = s.private.removeUdevRules(ifaceStatus.PciAddress)
				if err != nil {
					return err
				}
			}

			if err = s.sriovHelper.ResetSriovDevice(ifaceStatus); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *sriov) ConfigSriovDeviceVirtual(iface *sriovnetworkv1.Interface) error {
	log.Log.V(2).Info("ConfigSriovDeviceVirtual(): config interface", "address", iface.PciAddress, "config", iface)
	// Config VFs
	if iface.NumVfs > 0 {
		if iface.NumVfs > 1 {
			log.Log.Error(nil, "ConfigSriovDeviceVirtual(): in a virtual environment, only one VF per interface",
				"numVfs", iface.NumVfs)
			return errors.New("NumVfs > 1")
		}
		if len(iface.VfGroups) != 1 {
			log.Log.Error(nil, "ConfigSriovDeviceVirtual(): missing VFGroup")
			return errors.New("NumVfs != 1")
		}
		addr := iface.PciAddress
		log.Log.V(2).Info("ConfigSriovDeviceVirtual()", "address", addr)
		driver := ""
		vfID := 0
		for _, group := range iface.VfGroups {
			log.Log.V(2).Info("ConfigSriovDeviceVirtual()", "group", group)
			if sriovnetworkv1.IndexInRange(vfID, group.VfRange) {
				log.Log.V(2).Info("ConfigSriovDeviceVirtual()", "indexInRange", vfID)
				if sriovnetworkv1.StringInArray(group.DeviceType, vars.DpdkDrivers) {
					log.Log.V(2).Info("ConfigSriovDeviceVirtual()", "driver", group.DeviceType)
					driver = group.DeviceType
				}
				break
			}
		}
		if driver == "" {
			log.Log.V(2).Info("ConfigSriovDeviceVirtual(): bind default")
			if err := s.kernelHelper.BindDefaultDriver(addr); err != nil {
				log.Log.Error(err, "ConfigSriovDeviceVirtual(): fail to bind default driver", "device", addr)
				return err
			}
		} else {
			log.Log.V(2).Info("ConfigSriovDeviceVirtual(): bind driver", "driver", driver)
			if err := s.kernelHelper.BindDpdkDriver(addr, driver); err != nil {
				log.Log.Error(err, "ConfigSriovDeviceVirtual(): fail to bind driver for device",
					"driver", driver, "device", addr)
				return err
			}
		}
	}
	return nil
}

func (s *sriov) GetNicSriovMode(pciAddress string) (string, error) {
	log.Log.V(2).Info("GetNicSriovMode()", "device", pciAddress)

	devLink, err := s.netlinkLib.DevLinkGetDeviceByName("pci", pciAddress)
	if err != nil {
		if errors.Is(err, syscall.ENODEV) {
			// the device doesn't support devlink
			return "", nil
		}
		return "", err
	}

	return devLink.Attrs.Eswitch.Mode, nil
}

func (s *sriov) SetNicSriovMode(pciAddress string, mode string) error {
	log.Log.V(2).Info("SetNicSriovMode()", "device", pciAddress, "mode", mode)

	dev, err := s.netlinkLib.DevLinkGetDeviceByName("pci", pciAddress)
	if err != nil {
		return err
	}
	return s.netlinkLib.DevLinkSetEswitchMode(dev, mode)
}

func (s *sriov) GetLinkType(ifaceStatus sriovnetworkv1.InterfaceExt) string {
	log.Log.V(2).Info("GetLinkType()", "device", ifaceStatus.PciAddress)
	if ifaceStatus.Name != "" {
		link, err := s.netlinkLib.LinkByName(ifaceStatus.Name)
		if err != nil {
			log.Log.Error(err, "GetLinkType(): failed to get link", "device", ifaceStatus.Name)
			return ""
		}
		linkType := link.Attrs().EncapType
		if linkType == "ether" {
			return consts.LinkTypeETH
		} else if linkType == "infiniband" {
			return consts.LinkTypeIB
		}
	}

	return ""
}

// create required udev rules for PF:
// * rule to disable NetworkManager for VFs - for all modes
// * rule to rename VF representors - only for switchdev mode
func (s *sriov) addUdevRules(iface *sriovnetworkv1.Interface) error {
	log.Log.V(2).Info("addUdevRules(): add udev rules for device",
		"device", iface.PciAddress)
	if err := s.udevHelper.AddUdevRule(iface.PciAddress); err != nil {
		return err
	}
	if sriovnetworkv1.GetEswitchModeFromSpec(iface) == sriovnetworkv1.ESwithModeSwitchDev {
		portName, err := s.networkHelper.GetPhysPortName(iface.Name)
		if err != nil {
			return err
		}
		switchID, err := s.networkHelper.GetPhysSwitchID(iface.Name)
		if err != nil {
			return err
		}
		if err := s.udevHelper.AddVfRepresentorUdevRule(iface.PciAddress, iface.Name, switchID, portName); err != nil {
			return err
		}
	}
	return nil
}

// remove all udev rules for PF created by the operator
func (s *sriov) removeUdevRules(pciAddress string) error {
	log.Log.V(2).Info("removeUdevRules(): remove udev rules for device",
		"device", pciAddress)
	if err := s.udevHelper.RemoveUdevRule(pciAddress); err != nil {
		return err
	}
	return s.udevHelper.RemoveVfRepresentorUdevRule(pciAddress)
}

func (s *sriov) createVFs(iface *sriovnetworkv1.Interface, ifaceStatus *sriovnetworkv1.InterfaceExt) error {
	expectedEswitchMode := sriovnetworkv1.GetEswitchModeFromSpec(iface)
	log.Log.V(2).Info("createVFs(): configure VFs for device",
		"device", iface.PciAddress, "count", iface.NumVfs, "mode", expectedEswitchMode)

	if iface.NumVfs == ifaceStatus.NumVfs && !sriovnetworkv1.NeedToUpdateInterfaceEswitchMode(iface, ifaceStatus) {
		log.Log.V(2).Info("createVFs(): device is already configured",
			"device", iface.PciAddress, "count", iface.NumVfs, "mode", expectedEswitchMode)
		return nil
	}
	if expectedEswitchMode == sriovnetworkv1.ESwithModeSwitchDev {
		return s.private.createSwitchdevVFs(iface)
	}
	if err := s.private.unbindAllVFsOnPF(iface.PciAddress); err != nil {
		return err
	}
	if err := s.sriovHelper.SetNicSriovMode(iface.PciAddress, expectedEswitchMode); err != nil {
		return fmt.Errorf("failed to switch NIC to SRIOV %s mode: %v", expectedEswitchMode, err)
	}
	return s.sriovHelper.SetSriovNumVfs(iface.PciAddress, iface.NumVfs)
}

// some drivers may not support VF creation in switchdev mode,
// try to switch NIC to the legacy mode first, then create VFs and after that switch the NIC
// back to the switchdev mode
func (s *sriov) createSwitchdevVFs(iface *sriovnetworkv1.Interface) error {
	log.Log.V(2).Info("createSwitchdevVFs(): configure VFs for device",
		"device", iface.PciAddress, "count", iface.NumVfs)

	if err := s.private.unbindAllVFsOnPF(iface.PciAddress); err != nil {
		return err
	}
	if err := s.sriovHelper.SetNicSriovMode(iface.PciAddress, sriovnetworkv1.ESwithModeLegacy); err != nil {
		return fmt.Errorf("failed to switch NIC to SRIOV legacy mode: %v", err)
	}

	if err := s.sriovHelper.SetSriovNumVfs(iface.PciAddress, iface.NumVfs); err != nil {
		return err
	}
	if err := s.private.unbindAllVFsOnPF(iface.PciAddress); err != nil {
		return err
	}
	if err := s.sriovHelper.SetNicSriovMode(iface.PciAddress, sriovnetworkv1.ESwithModeSwitchDev); err != nil {
		return fmt.Errorf("failed to switch NIC to SRIOV switchdev mode: %v", err)
	}
	return nil
}

// retrieve all VFs for the PF and unbind them from a driver
func (s *sriov) unbindAllVFsOnPF(addr string) error {
	log.Log.V(2).Info("unbindAllVFsOnPF(): unbind all VFs on PF", "device", addr)
	vfAddrs, err := s.dputilsLib.GetVFList(addr)
	if err != nil {
		return fmt.Errorf("failed to read VF list: %v", err)
	}
	for _, vfAddr := range vfAddrs {
		if err := s.kernelHelper.Unbind(vfAddr); err != nil {
			return fmt.Errorf("failed to unbind VF from the driver: %v", err)
		}
	}
	return nil
}
