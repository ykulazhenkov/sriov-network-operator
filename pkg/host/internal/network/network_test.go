package network

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/golang/mock/gomock"

	hostMockPkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/helper/mock"
	dputilsMockPkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/dputils/mock"
	ethtoolMockPkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/ethtool/mock"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/types"
)

var _ = Describe("Network", func() {
	var (
		n              types.NetworkInterface
		ethtoolLibMock *ethtoolMockPkg.MockEthtoolLib
		dputilsLibMock *dputilsMockPkg.MockDPUtilsLib
		hostMock       *hostMockPkg.MockHostHelpersInterface

		testCtrl *gomock.Controller
		testErr  = fmt.Errorf("test")
	)
	BeforeEach(func() {
		testCtrl = gomock.NewController(GinkgoT())
		ethtoolLibMock = ethtoolMockPkg.NewMockEthtoolLib(testCtrl)
		dputilsLibMock = dputilsMockPkg.NewMockDPUtilsLib(testCtrl)
		hostMock = hostMockPkg.NewMockHostHelpersInterface(testCtrl)

		n = New(hostMock, dputilsLibMock, ethtoolLibMock)
	})

	AfterEach(func() {
		testCtrl.Finish()
	})
	Context("EnableHwTcOffload", func() {
		It("Enabled", func() {
			ethtoolLibMock.EXPECT().FeatureNames("enp216s0f0np0").Return(map[string]uint{"hw-tc-offload": 42}, nil)
			ethtoolLibMock.EXPECT().Features("enp216s0f0np0").Return(map[string]bool{"hw-tc-offload": false}, nil)
			ethtoolLibMock.EXPECT().Change("enp216s0f0np0", map[string]bool{"hw-tc-offload": true}).Return(nil)
			ethtoolLibMock.EXPECT().Features("enp216s0f0np0").Return(map[string]bool{"hw-tc-offload": true}, nil)
			Expect(n.EnableHwTcOffload("enp216s0f0np0")).NotTo(HaveOccurred())
		})
		It("Feature unknown", func() {
			ethtoolLibMock.EXPECT().FeatureNames("enp216s0f0np0").Return(map[string]uint{}, nil)
			Expect(n.EnableHwTcOffload("enp216s0f0np0")).NotTo(HaveOccurred())
		})
		It("Already enabled", func() {
			ethtoolLibMock.EXPECT().FeatureNames("enp216s0f0np0").Return(map[string]uint{"hw-tc-offload": 42}, nil)
			ethtoolLibMock.EXPECT().Features("enp216s0f0np0").Return(map[string]bool{"hw-tc-offload": true}, nil)
			Expect(n.EnableHwTcOffload("enp216s0f0np0")).NotTo(HaveOccurred())
		})
		It("not supported", func() {
			ethtoolLibMock.EXPECT().FeatureNames("enp216s0f0np0").Return(map[string]uint{"hw-tc-offload": 42}, nil)
			ethtoolLibMock.EXPECT().Features("enp216s0f0np0").Return(map[string]bool{"hw-tc-offload": false}, nil)
			ethtoolLibMock.EXPECT().Change("enp216s0f0np0", map[string]bool{"hw-tc-offload": true}).Return(nil)
			ethtoolLibMock.EXPECT().Features("enp216s0f0np0").Return(map[string]bool{"hw-tc-offload": false}, nil)
			Expect(n.EnableHwTcOffload("enp216s0f0np0")).NotTo(HaveOccurred())
		})
		It("fail - can't list supported", func() {
			ethtoolLibMock.EXPECT().FeatureNames("enp216s0f0np0").Return(nil, testErr)
			Expect(n.EnableHwTcOffload("enp216s0f0np0")).To(MatchError(testErr))
		})
		It("fail - can't get features", func() {
			ethtoolLibMock.EXPECT().FeatureNames("enp216s0f0np0").Return(map[string]uint{"hw-tc-offload": 42}, nil)
			ethtoolLibMock.EXPECT().Features("enp216s0f0np0").Return(nil, testErr)
			Expect(n.EnableHwTcOffload("enp216s0f0np0")).To(MatchError(testErr))
		})
		It("fail - can't change features", func() {
			ethtoolLibMock.EXPECT().FeatureNames("enp216s0f0np0").Return(map[string]uint{"hw-tc-offload": 42}, nil)
			ethtoolLibMock.EXPECT().Features("enp216s0f0np0").Return(map[string]bool{"hw-tc-offload": false}, nil)
			ethtoolLibMock.EXPECT().Change("enp216s0f0np0", map[string]bool{"hw-tc-offload": true}).Return(testErr)
			Expect(n.EnableHwTcOffload("enp216s0f0np0")).To(MatchError(testErr))
		})
		It("fail - can't reread features", func() {
			ethtoolLibMock.EXPECT().FeatureNames("enp216s0f0np0").Return(map[string]uint{"hw-tc-offload": 42}, nil)
			ethtoolLibMock.EXPECT().Features("enp216s0f0np0").Return(map[string]bool{"hw-tc-offload": false}, nil)
			ethtoolLibMock.EXPECT().Change("enp216s0f0np0", map[string]bool{"hw-tc-offload": true}).Return(nil)
			ethtoolLibMock.EXPECT().Features("enp216s0f0np0").Return(nil, testErr)
			Expect(n.EnableHwTcOffload("enp216s0f0np0")).To(MatchError(testErr))
		})
	})
})
