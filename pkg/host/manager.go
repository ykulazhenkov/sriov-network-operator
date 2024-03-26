package host

import (
	"fmt"

	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/kernel"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/dputils"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/ethtool"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/lib/netlink"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/network"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/ovs"
	ovsStorePkg "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/ovs/store"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/service"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/sriov"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/udev"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/internal/vdpa"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/types"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/utils"
)

// Contains all the host manipulation functions
//
//go:generate ../../bin/mockgen -destination mock/mock_host.go -source manager.go
type HostManagerInterface interface {
	types.KernelInterface
	types.NetworkInterface
	types.ServiceInterface
	types.UdevInterface
	types.SriovInterface
	types.VdpaInterface
	types.OVSInterface
}

type hostManager struct {
	utils.CmdInterface
	types.KernelInterface
	types.NetworkInterface
	types.ServiceInterface
	types.UdevInterface
	types.SriovInterface
	types.VdpaInterface
	types.OVSInterface
}

func NewHostManager(utilsInterface utils.CmdInterface) (HostManagerInterface, error) {
	dpUtils := dputils.New()
	netlinkLib := netlink.New()
	ethtoolLib := ethtool.New()
	k := kernel.New(utilsInterface)
	n := network.New(utilsInterface, dpUtils, netlinkLib, ethtoolLib)
	sv := service.New(utilsInterface)
	u := udev.New(utilsInterface)
	v := vdpa.New(k, netlinkLib)
	ovsStore, err := ovsStorePkg.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OVS store manager: %v", err)
	}
	o := ovs.New(ovsStore)
	sr := sriov.New(utilsInterface, k, n, u, v, o, netlinkLib, dpUtils)

	return &hostManager{
		utilsInterface,
		k,
		n,
		sv,
		u,
		sr,
		v,
		o,
	}, nil
}
