package ovs

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
)

var _ = Describe("OVS", func() {
	var ()
	BeforeEach(func() {
	})

	AfterEach(func() {
	})
	Context("TODO", func() {
		It("foo", func() {
			c, err := newOVSManager(context.Background(), "unix:///var/run/openvswitch/db.sock2")
			Expect(err).NotTo(HaveOccurred())
			Expect(c.CreateOVSBridge(context.Background(), &sriovnetworkv1.OVSConfigExt{Name: "foobar22"}, nil)).NotTo(HaveOccurred())
		})
	})
})
