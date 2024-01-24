package sriov

import (
	"fmt"
	"strconv"
	"syscall"

	"github.com/golang/mock/gomock"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	dputilsMockPkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/dputils/mock"
	netlinkMockPkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/netlink/mock"
	hostMockPkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/mock"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/fakefilesystem"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/helpers"
)

var _ = Describe("SRIOV", func() {
	var (
		s                *sriov
		netlinkLibMock   *netlinkMockPkg.MockNetlinkLib
		dputilsLibMock   *dputilsMockPkg.MockDPUtilsLib
		hostMock         *hostMockPkg.MockHostManagerInterface
		sriovPrivateMock *MocksriovPrivateInterface

		testCtrl *gomock.Controller

		testError = fmt.Errorf("test")
	)
	BeforeEach(func() {
		testCtrl = gomock.NewController(GinkgoT())
		netlinkLibMock = netlinkMockPkg.NewMockNetlinkLib(testCtrl)
		dputilsLibMock = dputilsMockPkg.NewMockDPUtilsLib(testCtrl)
		sriovPrivateMock = NewMocksriovPrivateInterface(testCtrl)
		hostMock = hostMockPkg.NewMockHostManagerInterface(testCtrl)
		s = New(nil, hostMock, hostMock, hostMock, hostMock, netlinkLibMock, dputilsLibMock).(*sriov)
		s.sriovHelper = hostMock
		s.private = sriovPrivateMock
	})

	AfterEach(func() {
		testCtrl.Finish()
	})

	Context("SetSriovNumVfs", func() {
		It("set", func() {
			helpers.GinkgoConfigureFakeFS(&fakefilesystem.FS{
				Dirs:  []string{"/sys/bus/pci/devices/0000:d8:00.0"},
				Files: map[string][]byte{"/sys/bus/pci/devices/0000:d8:00.0/sriov_numvfs": {}},
			})
			Expect(s.SetSriovNumVfs("0000:d8:00.0", 5)).NotTo(HaveOccurred())
			helpers.GinkgoAssertFileContentsEquals("/sys/bus/pci/devices/0000:d8:00.0/sriov_numvfs", strconv.Itoa(5))
		})
		It("fail - no such device", func() {
			Expect(s.SetSriovNumVfs("0000:d8:00.0", 5)).To(HaveOccurred())
		})
	})

	Context("GetNicSriovMode", func() {
		It("devlink returns info", func() {
			netlinkLibMock.EXPECT().DevLinkGetDeviceByName("pci", "0000:d8:00.0").Return(
				&netlink.DevlinkDevice{Attrs: netlink.DevlinkDevAttrs{Eswitch: netlink.DevlinkDevEswitchAttr{Mode: "switchdev"}}},
				nil)
			mode, err := s.GetNicSriovMode("0000:d8:00.0")
			Expect(err).NotTo(HaveOccurred())
			Expect(mode).To(Equal("switchdev"))
		})
		It("devlink returns error", func() {
			netlinkLibMock.EXPECT().DevLinkGetDeviceByName("pci", "0000:d8:00.0").Return(nil, testError)
			_, err := s.GetNicSriovMode("0000:d8:00.0")
			Expect(err).To(MatchError(testError))
		})
		It("devlink not supported - fail to get name", func() {
			netlinkLibMock.EXPECT().DevLinkGetDeviceByName("pci", "0000:d8:00.0").Return(nil, syscall.ENODEV)
			mode, err := s.GetNicSriovMode("0000:d8:00.0")
			Expect(err).NotTo(HaveOccurred())
			Expect(mode).To(BeEmpty())
		})
	})

	Context("SetNicSriovMode", func() {
		It("set", func() {
			testDev := &netlink.DevlinkDevice{}
			netlinkLibMock.EXPECT().DevLinkGetDeviceByName("pci", "0000:d8:00.0").Return(&netlink.DevlinkDevice{}, nil)
			netlinkLibMock.EXPECT().DevLinkSetEswitchMode(testDev, "legacy").Return(nil)
			Expect(s.SetNicSriovMode("0000:d8:00.0", "legacy")).NotTo(HaveOccurred())
		})
		It("fail to get dev", func() {
			netlinkLibMock.EXPECT().DevLinkGetDeviceByName("pci", "0000:d8:00.0").Return(nil, testError)
			Expect(s.SetNicSriovMode("0000:d8:00.0", "legacy")).To(MatchError(testError))
		})
		It("fail to set mode", func() {
			testDev := &netlink.DevlinkDevice{}
			netlinkLibMock.EXPECT().DevLinkGetDeviceByName("pci", "0000:d8:00.0").Return(&netlink.DevlinkDevice{}, nil)
			netlinkLibMock.EXPECT().DevLinkSetEswitchMode(testDev, "legacy").Return(testError)
			Expect(s.SetNicSriovMode("0000:d8:00.0", "legacy")).To(MatchError(testError))
		})
	})

	Context("Private functions", func() {
		Context("addUdevRules", func() {
			It("legacy - added", func() {
				hostMock.EXPECT().AddUdevRule("0000:d8:00.0").Return(nil)
				Expect(s.addUdevRules(&sriovnetworkv1.Interface{
					PciAddress: "0000:d8:00.0"})).NotTo(HaveOccurred())
			})
			It("legacy - fail, can't create NM rule", func() {
				hostMock.EXPECT().AddUdevRule("0000:d8:00.0").Return(testError)
				Expect(s.addUdevRules(&sriovnetworkv1.Interface{
					PciAddress: "0000:d8:00.0"})).To(MatchError(testError))
			})
			It("switchdev - added", func() {
				hostMock.EXPECT().AddUdevRule("0000:d8:00.0").Return(nil)
				hostMock.EXPECT().GetPhysPortName("enp216s0f0np0").Return("p0", nil)
				hostMock.EXPECT().GetPhysSwitchID("enp216s0f0np0").Return("7cfe90ff2cc0", nil)
				hostMock.EXPECT().AddVfRepresentorUdevRule("0000:d8:00.0", "enp216s0f0np0",
					"7cfe90ff2cc0", "p0").Return(nil)
				Expect(s.addUdevRules(&sriovnetworkv1.Interface{
					Name:        "enp216s0f0np0",
					PciAddress:  "0000:d8:00.0",
					EswitchMode: "switchdev"})).NotTo(HaveOccurred())
			})
			It("switchdev - fail, can't read port name", func() {
				hostMock.EXPECT().AddUdevRule("0000:d8:00.0").Return(nil)
				hostMock.EXPECT().GetPhysPortName("enp216s0f0np0").Return("", testError)
				Expect(s.addUdevRules(&sriovnetworkv1.Interface{
					Name:        "enp216s0f0np0",
					PciAddress:  "0000:d8:00.0",
					EswitchMode: "switchdev"})).To(MatchError(testError))
			})
			It("switchdev - fail, can't read switch id", func() {
				hostMock.EXPECT().AddUdevRule("0000:d8:00.0").Return(nil)
				hostMock.EXPECT().GetPhysPortName("enp216s0f0np0").Return("p0", nil)
				hostMock.EXPECT().GetPhysSwitchID("enp216s0f0np0").Return("", testError)
				Expect(s.addUdevRules(&sriovnetworkv1.Interface{
					Name:        "enp216s0f0np0",
					PciAddress:  "0000:d8:00.0",
					EswitchMode: "switchdev"})).To(MatchError(testError))
			})
			It("switchdev - fail, can't create VF rep rule", func() {
				hostMock.EXPECT().AddUdevRule("0000:d8:00.0").Return(nil)
				hostMock.EXPECT().GetPhysPortName("enp216s0f0np0").Return("p0", nil)
				hostMock.EXPECT().GetPhysSwitchID("enp216s0f0np0").Return("7cfe90ff2cc0", nil)
				hostMock.EXPECT().AddVfRepresentorUdevRule("0000:d8:00.0", "enp216s0f0np0",
					"7cfe90ff2cc0", "p0").Return(testError)
				Expect(s.addUdevRules(&sriovnetworkv1.Interface{
					Name:        "enp216s0f0np0",
					PciAddress:  "0000:d8:00.0",
					EswitchMode: "switchdev"})).To(MatchError(testError))
			})
		})
		Context("removeUdevRules", func() {
			It("removed", func() {
				hostMock.EXPECT().RemoveUdevRule("0000:d8:00.0").Return(nil)
				hostMock.EXPECT().RemoveVfRepresentorUdevRule("0000:d8:00.0").Return(nil)
				Expect(s.removeUdevRules("0000:d8:00.0")).NotTo(HaveOccurred())
			})
			It("fail, can't remove NM rule", func() {
				hostMock.EXPECT().RemoveUdevRule("0000:d8:00.0").Return(testError)
				Expect(s.removeUdevRules("0000:d8:00.0")).To(MatchError(testError))
			})
			It("fail, can't create VF rep rule", func() {
				hostMock.EXPECT().RemoveUdevRule("0000:d8:00.0").Return(nil)
				hostMock.EXPECT().RemoveVfRepresentorUdevRule("0000:d8:00.0").Return(testError)
				Expect(s.removeUdevRules("0000:d8:00.0")).To(MatchError(testError))
			})
		})
		Context("configureESwitchMode", func() {
			It("already configured", func() {
				Expect(s.configureESwitchMode(
					&sriovnetworkv1.Interface{
						PciAddress:  "0000:d8:00.0",
						EswitchMode: "switchdev"},
					&sriovnetworkv1.InterfaceExt{
						PciAddress:  "0000:d8:00.0",
						EswitchMode: "switchdev",
					})).NotTo(HaveOccurred())
			})
			It("configure", func() {
				sriovPrivateMock.EXPECT().unbindAllVFsOnPF("0000:d8:00.0").Return(nil)
				hostMock.EXPECT().SetNicSriovMode("0000:d8:00.0", "switchdev").Return(nil)
				Expect(s.configureESwitchMode(
					&sriovnetworkv1.Interface{
						PciAddress:  "0000:d8:00.0",
						EswitchMode: "switchdev"},
					&sriovnetworkv1.InterfaceExt{
						PciAddress:  "0000:d8:00.0",
						EswitchMode: "legacy",
					})).NotTo(HaveOccurred())
			})
			It("fail - can't unbind VFs", func() {
				sriovPrivateMock.EXPECT().unbindAllVFsOnPF("0000:d8:00.0").Return(testError)
				Expect(s.configureESwitchMode(
					&sriovnetworkv1.Interface{
						PciAddress:  "0000:d8:00.0",
						EswitchMode: "switchdev"},
					&sriovnetworkv1.InterfaceExt{
						PciAddress:  "0000:d8:00.0",
						EswitchMode: "legacy",
					})).To(MatchError(testError))
			})
			It("fail - can't switch mode", func() {
				sriovPrivateMock.EXPECT().unbindAllVFsOnPF("0000:d8:00.0").Return(nil)
				hostMock.EXPECT().SetNicSriovMode("0000:d8:00.0", "switchdev").Return(testError)
				Expect(s.configureESwitchMode(
					&sriovnetworkv1.Interface{
						PciAddress:  "0000:d8:00.0",
						EswitchMode: "switchdev"},
					&sriovnetworkv1.InterfaceExt{
						PciAddress:  "0000:d8:00.0",
						EswitchMode: "legacy",
					})).To(HaveOccurred())
			})
		})
		Context("createVFs", func() {
			It("already configured", func() {
				Expect(s.createVFs(
					&sriovnetworkv1.Interface{
						PciAddress: "0000:d8:00.0",
						NumVfs:     5,
					},
					&sriovnetworkv1.InterfaceExt{
						PciAddress: "0000:d8:00.0",
						NumVfs:     5,
					})).NotTo(HaveOccurred())
			})
			It("create", func() {
				hostMock.EXPECT().SetSriovNumVfs("0000:d8:00.0", 5).Return(nil)
				Expect(s.createVFs(
					&sriovnetworkv1.Interface{
						PciAddress: "0000:d8:00.0",
						NumVfs:     5,
					},
					&sriovnetworkv1.InterfaceExt{
						PciAddress: "0000:d8:00.0",
						NumVfs:     0,
					})).NotTo(HaveOccurred())
			})
			It("fail - legacy mode, fail to set num VFs", func() {
				hostMock.EXPECT().SetSriovNumVfs("0000:d8:00.0", 5).Return(testError)
				Expect(s.createVFs(
					&sriovnetworkv1.Interface{
						PciAddress: "0000:d8:00.0",
						NumVfs:     5,
					},
					&sriovnetworkv1.InterfaceExt{
						PciAddress: "0000:d8:00.0",
						NumVfs:     0,
					})).To(HaveOccurred())
			})
			It("switchdev mode, can't create VFs, VFs created by switching to legacy mode", func() {
				iface := &sriovnetworkv1.Interface{
					PciAddress:  "0000:d8:00.0",
					NumVfs:      5,
					EswitchMode: "switchdev",
				}
				hostMock.EXPECT().SetSriovNumVfs("0000:d8:00.0", 5).Return(testError)
				sriovPrivateMock.EXPECT().createSwitchdevVFsBySwitchingToLegacy(iface).Return(nil)
				Expect(s.createVFs(
					iface,
					&sriovnetworkv1.InterfaceExt{
						PciAddress:  "0000:d8:00.0",
						NumVfs:      0,
						EswitchMode: "switchdev",
					})).NotTo(HaveOccurred())
			})
			It("switchdev mode, can't create VFs", func() {
				iface := &sriovnetworkv1.Interface{
					PciAddress:  "0000:d8:00.0",
					NumVfs:      5,
					EswitchMode: "switchdev",
				}
				hostMock.EXPECT().SetSriovNumVfs("0000:d8:00.0", 5).Return(testError)
				sriovPrivateMock.EXPECT().createSwitchdevVFsBySwitchingToLegacy(iface).Return(testError)
				Expect(s.createVFs(
					iface,
					&sriovnetworkv1.InterfaceExt{
						PciAddress:  "0000:d8:00.0",
						NumVfs:      0,
						EswitchMode: "switchdev",
					})).To(HaveOccurred())
			})
		})
		Context("createSwitchdevVFsBySwitchingToLegacy", func() {
			It("created", func() {
				sriovPrivateMock.EXPECT().unbindAllVFsOnPF("0000:d8:00.0").Return(nil).Times(2)
				hostMock.EXPECT().SetNicSriovMode("0000:d8:00.0", "legacy").Return(nil)
				hostMock.EXPECT().SetSriovNumVfs("0000:d8:00.0", 5).Return(nil)
				hostMock.EXPECT().SetNicSriovMode("0000:d8:00.0", "switchdev").Return(nil)
				Expect(s.createSwitchdevVFsBySwitchingToLegacy(&sriovnetworkv1.Interface{
					PciAddress:  "0000:d8:00.0",
					NumVfs:      5,
					EswitchMode: "switchdev",
				})).NotTo(HaveOccurred())
			})
			It("fail - can't unbind VFs", func() {
				sriovPrivateMock.EXPECT().unbindAllVFsOnPF("0000:d8:00.0").Return(testError)
				Expect(s.createSwitchdevVFsBySwitchingToLegacy(&sriovnetworkv1.Interface{
					PciAddress:  "0000:d8:00.0",
					NumVfs:      5,
					EswitchMode: "switchdev",
				})).To(HaveOccurred())
			})
			It("fail - can't switch to legacy mode", func() {
				sriovPrivateMock.EXPECT().unbindAllVFsOnPF("0000:d8:00.0").Return(nil)
				hostMock.EXPECT().SetNicSriovMode("0000:d8:00.0", "legacy").Return(testError)
				Expect(s.createSwitchdevVFsBySwitchingToLegacy(&sriovnetworkv1.Interface{
					PciAddress:  "0000:d8:00.0",
					NumVfs:      5,
					EswitchMode: "switchdev",
				})).To(HaveOccurred())
			})
			It("fail - can't configure num VFs", func() {
				sriovPrivateMock.EXPECT().unbindAllVFsOnPF("0000:d8:00.0").Return(nil)
				hostMock.EXPECT().SetNicSriovMode("0000:d8:00.0", "legacy").Return(nil)
				hostMock.EXPECT().SetSriovNumVfs("0000:d8:00.0", 5).Return(testError)
				Expect(s.createSwitchdevVFsBySwitchingToLegacy(&sriovnetworkv1.Interface{
					PciAddress:  "0000:d8:00.0",
					NumVfs:      5,
					EswitchMode: "switchdev",
				})).To(HaveOccurred())
			})
			It("fail - can't unbind VFs in legacy mode", func() {
				sriovPrivateMock.EXPECT().unbindAllVFsOnPF("0000:d8:00.0").Return(nil)
				hostMock.EXPECT().SetNicSriovMode("0000:d8:00.0", "legacy").Return(nil)
				hostMock.EXPECT().SetSriovNumVfs("0000:d8:00.0", 5).Return(nil)
				sriovPrivateMock.EXPECT().unbindAllVFsOnPF("0000:d8:00.0").Return(testError)
				Expect(s.createSwitchdevVFsBySwitchingToLegacy(&sriovnetworkv1.Interface{
					PciAddress:  "0000:d8:00.0",
					NumVfs:      5,
					EswitchMode: "switchdev",
				})).To(HaveOccurred())
			})
			It("fail - can't return to switchdev mode", func() {
				sriovPrivateMock.EXPECT().unbindAllVFsOnPF("0000:d8:00.0").Return(nil).Times(2)
				hostMock.EXPECT().SetNicSriovMode("0000:d8:00.0", "legacy").Return(nil)
				hostMock.EXPECT().SetSriovNumVfs("0000:d8:00.0", 5).Return(nil)
				hostMock.EXPECT().SetNicSriovMode("0000:d8:00.0", "switchdev").Return(testError)
				Expect(s.createSwitchdevVFsBySwitchingToLegacy(&sriovnetworkv1.Interface{
					PciAddress:  "0000:d8:00.0",
					NumVfs:      5,
					EswitchMode: "switchdev",
				})).To(HaveOccurred())
			})
		})
		Context("unbindAllVFsOnPF", func() {
			It("no VFs", func() {
				dputilsLibMock.EXPECT().GetVFList("0000:d8:00.0").Return(nil, nil)
				Expect(s.unbindAllVFsOnPF("0000:d8:00.0")).NotTo(HaveOccurred())
			})
			It("fail - can't list VFs", func() {
				dputilsLibMock.EXPECT().GetVFList("0000:d8:00.0").Return(nil, testError)
				Expect(s.unbindAllVFsOnPF("0000:d8:00.0")).To(HaveOccurred())
			})
			It("unbind", func() {
				dputilsLibMock.EXPECT().GetVFList("0000:d8:00.0").Return(
					[]string{"0000:d8:00.2", "0000:d8:00.3"}, nil)
				hostMock.EXPECT().Unbind("0000:d8:00.2").Return(nil)
				hostMock.EXPECT().Unbind("0000:d8:00.3").Return(nil)
				Expect(s.unbindAllVFsOnPF("0000:d8:00.0")).NotTo(HaveOccurred())
			})
			It("fail - can't unbind VF", func() {
				dputilsLibMock.EXPECT().GetVFList("0000:d8:00.0").Return(
					[]string{"0000:d8:00.2", "0000:d8:00.3"}, nil)
				hostMock.EXPECT().Unbind("0000:d8:00.2").Return(nil)
				hostMock.EXPECT().Unbind("0000:d8:00.3").Return(testError)
				Expect(s.unbindAllVFsOnPF("0000:d8:00.0")).To(HaveOccurred())
			})
		})
	})
})
