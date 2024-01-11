package host

import (
	"fmt"
	"syscall"

	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	constants "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	govdpaMock "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/govdpa/mock"
	hostMock "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/mock"
)

const (
	testVdpaVFPciAddr = "0000:d8:00.2"
)

var _ = Describe("VDPA", func() {
	var (
		v           VdpaInterface
		libMock     *govdpaMock.MockGoVdpaLib
		vdpaDevMock *govdpaMock.MockVdpaDevice
		kernelMock  *hostMock.MockKernelInterface

		testCtrl *gomock.Controller

		expectedName   = "vdpa:" + testVdpaVFPciAddr
		expectedDriver = "vhost_vdpa"
		testErr        = fmt.Errorf("test-error")
	)
	BeforeEach(func() {
		testCtrl = gomock.NewController(GinkgoT())
		libMock = govdpaMock.NewMockGoVdpaLib(testCtrl)
		vdpaDevMock = govdpaMock.NewMockVdpaDevice(testCtrl)
		kernelMock = hostMock.NewMockKernelInterface(testCtrl)
		v = newVdpaInterface(kernelMock, libMock)
	})
	AfterEach(func() {
		testCtrl.Finish()
	})
	Context("CreateVDPADevice", func() {
		callFunc := func() error {
			return v.CreateVDPADevice(testVdpaVFPciAddr, constants.VdpaTypeVhost)
		}
		It("Created", func() {
			libMock.EXPECT().GetVdpaDevice(expectedName).Return(nil, syscall.ENODEV)
			libMock.EXPECT().AddVdpaDevice("pci/"+testVdpaVFPciAddr, expectedName).Return(nil)
			kernelMock.EXPECT().BindDriverByBusAndDevice(consts.BusVdpa, expectedName, expectedDriver).Return(nil)
			Expect(callFunc()).NotTo(HaveOccurred())
		})
		It("Already exist", func() {
			libMock.EXPECT().GetVdpaDevice(expectedName).Return(vdpaDevMock, nil)
			kernelMock.EXPECT().BindDriverByBusAndDevice(consts.BusVdpa, expectedName, expectedDriver).Return(nil)
			Expect(callFunc()).NotTo(HaveOccurred())
		})
		It("Fail to Get device", func() {
			libMock.EXPECT().GetVdpaDevice(expectedName).Return(nil, testErr)
			Expect(callFunc()).To(MatchError(testErr))
		})
		It("Fail to Create device", func() {
			libMock.EXPECT().GetVdpaDevice(expectedName).Return(nil, syscall.ENODEV)
			libMock.EXPECT().AddVdpaDevice("pci/"+testVdpaVFPciAddr, expectedName).Return(testErr)
			Expect(callFunc()).To(MatchError(testErr))
		})
		It("Fail to Bind device", func() {
			libMock.EXPECT().GetVdpaDevice(expectedName).Return(vdpaDevMock, nil)
			kernelMock.EXPECT().BindDriverByBusAndDevice(consts.BusVdpa, expectedName, expectedDriver).Return(testErr)
			Expect(callFunc()).To(MatchError(testErr))
		})
	})
	Context("DeleteVDPADevice", func() {
		callFunc := func() error {
			return v.DeleteVDPADevice(testVdpaVFPciAddr)
		}
		It("Removed", func() {
			libMock.EXPECT().DeleteVdpaDevice(expectedName).Return(nil)
			Expect(callFunc()).NotTo(HaveOccurred())
		})
		It("Not found", func() {
			libMock.EXPECT().DeleteVdpaDevice(expectedName).Return(syscall.ENODEV)
			Expect(callFunc()).NotTo(HaveOccurred())
		})
		It("Fail to delete device", func() {
			libMock.EXPECT().DeleteVdpaDevice(expectedName).Return(testErr)
			Expect(callFunc()).To(MatchError(testErr))
		})
	})
	Context("DiscoverVDPAType", func() {
		callFunc := func() string {
			return v.DiscoverVDPAType(testVdpaVFPciAddr)
		}
		It("No device", func() {
			libMock.EXPECT().GetVdpaDevice(expectedName).Return(nil, syscall.ENODEV)
			Expect(callFunc()).To(BeEmpty())
		})
		It("No driver", func() {
			vdpaDevMock.EXPECT().Driver().Return("")
			libMock.EXPECT().GetVdpaDevice(expectedName).Return(vdpaDevMock, nil)
			Expect(callFunc()).To(BeEmpty())
		})
		It("Unknown driver", func() {
			vdpaDevMock.EXPECT().Driver().Return("something")
			libMock.EXPECT().GetVdpaDevice(expectedName).Return(vdpaDevMock, nil)
			Expect(callFunc()).To(BeEmpty())
		})
		It("Vhost driver", func() {
			vdpaDevMock.EXPECT().Driver().Return("vhost_vdpa")
			libMock.EXPECT().GetVdpaDevice(expectedName).Return(vdpaDevMock, nil)
			Expect(callFunc()).To(Equal("vhost"))
		})
		It("Virtio driver", func() {
			vdpaDevMock.EXPECT().Driver().Return("virtio_vdpa")
			libMock.EXPECT().GetVdpaDevice(expectedName).Return(vdpaDevMock, nil)
			Expect(callFunc()).To(Equal("virtio"))
		})
	})
})
