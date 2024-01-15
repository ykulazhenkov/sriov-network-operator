package internal

import (
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
)

const (
	HostPathFromDaemon    = consts.Host
	RedHatReleaseFile     = "/etc/redhat-release"
	RHELRDMAConditionFile = "/usr/libexec/rdma-init-kernel"
	RHELRDMAServiceName   = "rdma"
	RHELPackageManager    = "yum"

	UbuntuRDMAConditionFile = "/usr/sbin/rdma-ndd"
	UbuntuRDMAServiceName   = "rdma-ndd"
	UbuntuPackageManager    = "apt-get"

	GenericOSReleaseFile = "/etc/os-release"
)
