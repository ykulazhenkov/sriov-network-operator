package k8s

import (
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/log"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/helper"
	hostTypes "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/types"
	plugins "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/plugins"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
)

var PluginName = "k8s"

type K8sPlugin struct {
	PluginName  string
	SpecVersion string
	hostHelper  helper.HostHelpersInterface

	openVSwitchService      *hostTypes.Service
	sriovService            *hostTypes.Service
	sriovPostNetworkService *hostTypes.Service

	updateTarget *k8sUpdateTarget
}

type k8sUpdateTarget struct {
	sriovScript            bool
	sriovPostNetworkScript bool
	systemServices         []*hostTypes.Service
}

func (u *k8sUpdateTarget) needUpdate() bool {
	return u.sriovScript ||
		u.sriovPostNetworkScript ||
		len(u.systemServices) > 0
}

func (u *k8sUpdateTarget) needReboot() bool {
	return u.sriovScript || u.sriovPostNetworkScript
}

func (u *k8sUpdateTarget) reset() {
	u.sriovScript = false
	u.sriovPostNetworkScript = false
	u.systemServices = []*hostTypes.Service{}
}

func (u *k8sUpdateTarget) String() string {
	var updateList []string
	if u.sriovScript {
		updateList = append(updateList, "sriov-config.service")
	}
	if u.sriovPostNetworkScript {
		updateList = append(updateList, "sriov-config-post-network.service")
	}
	for _, s := range u.systemServices {
		updateList = append(updateList, s.Name)
	}
	return strings.Join(updateList, ",")
}

const (
	bindataManifestPath      = "bindata/manifests/"
	switchdevManifestPath    = bindataManifestPath + "switchdev-config/"
	switchdevUnits           = switchdevManifestPath + "switchdev-units/"
	sriovUnits               = bindataManifestPath + "sriov-config-service/kubernetes/"
	sriovUnitFile            = sriovUnits + "sriov-config-service.yaml"
	sriovPostNetworkUnitFile = sriovUnits + "sriov-config-post-network-service.yaml"
	ovsUnitFile              = switchdevManifestPath + "ovs-units/ovs-vswitchd.service.yaml"
)

// Initialize our plugin and set up initial values
func NewK8sPlugin(helper helper.HostHelpersInterface) (plugins.VendorPlugin, error) {
	k8sPluging := &K8sPlugin{
		PluginName:   PluginName,
		SpecVersion:  "1.0",
		hostHelper:   helper,
		updateTarget: &k8sUpdateTarget{},
	}

	return k8sPluging, k8sPluging.readManifestFiles()
}

// Name returns the name of the plugin
func (p *K8sPlugin) Name() string {
	return p.PluginName
}

// Spec returns the version of the spec expected by the plugin
func (p *K8sPlugin) Spec() string {
	return p.SpecVersion
}

// OnNodeStateChange Invoked when SriovNetworkNodeState CR is created or updated, return if need dain and/or reboot node
func (p *K8sPlugin) OnNodeStateChange(new *sriovnetworkv1.SriovNetworkNodeState) (needDrain bool, needReboot bool, err error) {
	log.Log.Info("k8s plugin OnNodeStateChange()")
	needDrain = false
	needReboot = false

	p.updateTarget.reset()
	// TODO add check for enableOvsOffload in OperatorConfig later
	// Update services if switchdev required
	if !vars.UsingSystemdMode && !sriovnetworkv1.IsSwitchdevModeSpec(new.Spec) {
		return
	}

	if sriovnetworkv1.IsSwitchdevModeSpec(new.Spec) {
		// Check services
		err = p.systemServicesStateUpdate()
		if err != nil {
			log.Log.Error(err, "k8s plugin OnNodeStateChange(): failed")
			return
		}
	}

	if vars.UsingSystemdMode {
		// Check sriov service
		err = p.sriovServicesStateUpdate()
		if err != nil {
			log.Log.Error(err, "k8s plugin OnNodeStateChange(): failed")
			return
		}
	}

	if p.updateTarget.needUpdate() {
		needDrain = true
		if p.updateTarget.needReboot() {
			needReboot = true
			log.Log.Info("k8s plugin OnNodeStateChange(): needReboot to update", "target", p.updateTarget)
		} else {
			log.Log.Info("k8s plugin OnNodeStateChange(): needDrain to update", "target", p.updateTarget)
		}
	}

	return
}

// Apply config change
func (p *K8sPlugin) Apply() error {
	log.Log.Info("k8s plugin Apply()")
	if vars.UsingSystemdMode {
		if err := p.updateSriovServices(); err != nil {
			return err
		}
	}

	for _, systemService := range p.updateTarget.systemServices {
		if err := p.hostHelper.UpdateSystemService(systemService); err != nil {
			return err
		}
	}

	return nil
}

func (p *K8sPlugin) readOpenVSwitchdManifest() error {
	openVSwitchService, err := p.hostHelper.ReadServiceInjectionManifestFile(ovsUnitFile)
	if err != nil {
		return err
	}

	p.openVSwitchService = openVSwitchService
	return nil
}

func (p *K8sPlugin) readSriovServiceManifest() error {
	sriovService, err := p.hostHelper.ReadServiceManifestFile(sriovUnitFile)
	if err != nil {
		return err
	}

	p.sriovService = sriovService
	return nil
}

func (p *K8sPlugin) readSriovPostNetworkServiceManifest() error {
	sriovService, err := p.hostHelper.ReadServiceManifestFile(sriovPostNetworkUnitFile)
	if err != nil {
		return err
	}

	p.sriovPostNetworkService = sriovService
	return nil
}

func (p *K8sPlugin) readManifestFiles() error {
	if err := p.readOpenVSwitchdManifest(); err != nil {
		return err
	}

	if err := p.readSriovServiceManifest(); err != nil {
		return err
	}

	if err := p.readSriovPostNetworkServiceManifest(); err != nil {
		return err
	}

	return nil
}

func (p *K8sPlugin) sriovServicesStateUpdate() error {
	log.Log.Info("sriovServicesStateUpdate()")

	for _, s := range []struct {
		srv    *hostTypes.Service
		update *bool
	}{
		{srv: p.sriovService, update: &p.updateTarget.sriovScript},
		{srv: p.sriovPostNetworkService, update: &p.updateTarget.sriovPostNetworkScript},
	} {
		isServiceEnabled, err := p.hostHelper.IsServiceEnabled(s.srv.Path)
		if err != nil {
			return err
		}
		// create and enable the service if it doesn't exist or is not enabled
		if !isServiceEnabled {
			*s.update = true
		} else {
			*s.update = p.isSystemDServiceNeedUpdate(s.srv)
		}
	}

	return nil
}

func (p *K8sPlugin) systemServicesStateUpdate() error {
	var services []*hostTypes.Service
	for _, systemService := range []*hostTypes.Service{p.openVSwitchService} {
		exist, err := p.hostHelper.IsServiceExist(systemService.Path)
		if err != nil {
			return err
		}
		if !exist {
			log.Log.Info("k8s plugin systemServicesStateUpdate(): WARNING! system service not found, skip update",
				"service", p.openVSwitchService.Name)
			return nil
		}
		if p.isSystemDServiceNeedUpdate(systemService) {
			services = append(services, systemService)
		}
	}

	p.updateTarget.systemServices = services
	return nil
}

func (p *K8sPlugin) isSystemDServiceNeedUpdate(serviceObj *hostTypes.Service) bool {
	log.Log.Info("isSystemDServiceNeedUpdate()")
	systemService, err := p.hostHelper.ReadService(serviceObj.Path)
	if err != nil {
		log.Log.Error(err, "k8s plugin isSystemDServiceNeedUpdate(): failed to read service file, ignoring",
			"path", serviceObj.Path)
		return false
	}
	if systemService != nil {
		needChange, err := p.hostHelper.CompareServices(systemService, serviceObj)
		if err != nil {
			log.Log.Error(err, "k8s plugin isSystemDServiceNeedUpdate(): failed to compare service, ignoring")
			return false
		}
		return needChange
	}

	return false
}

func (p *K8sPlugin) updateSriovServices() error {
	for _, s := range []struct {
		srv    *hostTypes.Service
		update bool
	}{
		{srv: p.sriovService, update: p.updateTarget.sriovScript},
		{srv: p.sriovPostNetworkService, update: p.updateTarget.sriovPostNetworkScript},
	} {
		if s.update {
			err := p.hostHelper.EnableService(s.srv)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
