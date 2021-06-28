package main

import (
	"github.com/golang/glog"
	"os"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/service"
)

type RDMAPlugin struct {
	PluginName      string
	SpecVersion     string
	DesireState     *sriovnetworkv1.SriovNetworkNodeState
	serviceManager  service.ServiceManager
	rdmaModeService *service.Service
	updateRequired  bool
}

const (
	rdmaManifestPath = "bindata/manifests/rdma-mode/"
	rdmaUnits        = rdmaManifestPath + "rdma-mode-units/"
	rdmaUnitFile     = rdmaUnits + "rdma-mode-configuration.yaml"

	chroot = "/host"
)

var (
	Plugin RDMAPlugin
)

// Initialize our plugin and set up initial values
func init() {
	Plugin = RDMAPlugin{
		PluginName:     "rdma_plugin",
		SpecVersion:    "1.0",
		serviceManager: service.NewServiceManager(chroot),
	}

	// Read manifest files for plugin
	if err := Plugin.readManifestFile(); err != nil {
		panic(err)
	}
}

func (p *RDMAPlugin) readManifestFile() error {
	// Read switchdev service
	var err error
	p.rdmaModeService, err = service.ReadServiceManifestFile(rdmaUnitFile)
	return err
}

// Name returns the name of the plugin
func (p *RDMAPlugin) Name() string {
	return p.PluginName
}

// Spec returns the version of the spec expected by the plugin
func (p *RDMAPlugin) Spec() string {
	return p.SpecVersion
}

// OnNodeStateAdd Invoked when SriovNetworkNodeState CR is created, return if need dain and/or reboot node
func (p *RDMAPlugin) OnNodeStateAdd(state *sriovnetworkv1.SriovNetworkNodeState) (needDrain bool, needReboot bool, err error) {
	glog.Info("rdma-plugin OnNodeStateAdd()")
	return p.OnNodeStateChange(nil, state)
}

// OnNodeStateChange Invoked when SriovNetworkNodeState CR is updated, return if need dain and/or reboot node
func (p *RDMAPlugin) OnNodeStateChange(_, new *sriovnetworkv1.SriovNetworkNodeState) (needDrain bool, needReboot bool, err error) {
	glog.Info("rdma-plugin OnNodeStateChange()")
	p.DesireState = new
	err = p.rdmaModeServiceStateCheck(new.Spec.NodeSettings.RDMAExclusiveMode)
	if err != nil {
		glog.Errorf("rdmaMode OnNodeStateChange(): failed : %v", err)
		return
	}

	return p.updateRequired, p.updateRequired, nil
}

// Apply config change
func (p *RDMAPlugin) Apply() error {
	glog.Info("rdma-plugin Apply()")
	if p.updateRequired {
		if p.DesireState.Spec.NodeSettings.RDMAExclusiveMode {
			glog.Info("rdma-plugin Apply(): enable service")
			err := p.serviceManager.EnableService(p.rdmaModeService)
			if err != nil {
				return err
			}
		} else {
			glog.Info("rdma-plugin Apply(): disable service")
			err := p.serviceManager.DisableService(p.rdmaModeService)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *RDMAPlugin) rdmaModeServiceStateCheck(shouldExist bool) error {
	glog.Info("rdma-plugin rdmaModeServiceStateCheck()", "shouldExist ", shouldExist)
	// Check rdma mode service
	serviceExist := true
	rdmaModeService, err := p.serviceManager.ReadService(p.rdmaModeService.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		serviceExist = false
	}
	if !shouldExist {
		p.updateRequired = serviceExist
		return nil
	}

	if !serviceExist {
		p.updateRequired = true
		return nil
	}

	needChange, err := service.CompareServices(rdmaModeService, p.rdmaModeService)
	if err != nil {
		return err
	}
	p.updateRequired = needChange

	return nil
}
