package k8s

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/coreos/go-systemd/v22/unit"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	plugins "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/plugins"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/service"
)

var PluginName = "k8s_plugin"

type K8sPlugin struct {
	PluginName          string
	SpecVersion         string
	serviceManager      service.ServiceManager
	switchdevUdevScript *service.ScriptManifestFile
	openVSwitchService  *service.Service
	sriovService        *service.Service
	updateTarget        *k8sUpdateTarget
	useSystemdService   bool
}

type k8sUpdateTarget struct {
	switchdevUdevScript bool
	sriovScript         bool
	systemServices      []*service.Service
}

func (u *k8sUpdateTarget) needUpdate() bool {
	return u.switchdevUdevScript || u.sriovScript || len(u.systemServices) > 0
}

func (u *k8sUpdateTarget) needReboot() bool {
	return u.switchdevUdevScript || u.sriovScript
}

func (u *k8sUpdateTarget) reset() {
	u.switchdevUdevScript = false
	u.sriovScript = false
	u.systemServices = []*service.Service{}
}

func (u *k8sUpdateTarget) String() string {
	var updateList []string
	if u.switchdevUdevScript {
		updateList = append(updateList, "SwitchdevUdevScript")
	}
	for _, s := range u.systemServices {
		updateList = append(updateList, s.Name)
	}

	return strings.Join(updateList, ",")
}

const (
	bindataManifestPath         = "bindata/manifests/"
	switchdevManifestPath       = bindataManifestPath + "switchdev-config/"
	sriovUnits                  = bindataManifestPath + "sriov-config-service/kubernetes/"
	sriovUnitFile               = sriovUnits + "sriov-config-service.yaml"
	ovsUnitFile                 = switchdevManifestPath + "ovs-units/ovs-vswitchd.service.yaml"
	switchdevRenamingUdevScript = switchdevManifestPath + "files/switchdev-vf-link-name.sh.yaml"

	chroot = "/host"
)

// Initialize our plugin and set up initial values
func NewK8sPlugin(useSystemdService bool) (plugins.VendorPlugin, error) {
	k8sPlugin := &K8sPlugin{
		PluginName:        PluginName,
		SpecVersion:       "1.0",
		serviceManager:    service.NewServiceManager(chroot),
		updateTarget:      &k8sUpdateTarget{},
		useSystemdService: useSystemdService,
	}

	return k8sPlugin, k8sPlugin.readManifestFiles()
}

// Name returns the name of the plugin
func (p *K8sPlugin) Name() string {
	return p.PluginName
}

// Spec returns the version of the spec expected by the plugin
func (p *K8sPlugin) Spec() string {
	return p.SpecVersion
}

// OnNodeStateChange Invoked when SriovNetworkNodeState CR is created or updated, return if need drain and/or reboot node
func (p *K8sPlugin) OnNodeStateChange(new *sriovnetworkv1.SriovNetworkNodeState) (needDrain bool, needReboot bool, err error) {
	log.Log.Info("k8s-plugin OnNodeStateChange()")
	needDrain = false
	needReboot = false

	p.updateTarget.reset()
	// TODO add clean up logic
	if !p.useSystemdService {
		return
	}

	// check switchdev related files
	err = p.switchDevServicesStateUpdate()
	if err != nil {
		log.Log.Error(err, "k8s-plugin OnNodeStateChange(): failed")
		return
	}
	// Check sriov service
	err = p.sriovServiceStateUpdate()
	if err != nil {
		log.Log.Error(err, "k8s-plugin OnNodeStateChange(): failed")
		return
	}

	if p.updateTarget.needUpdate() {
		needDrain = true
		if p.updateTarget.needReboot() {
			needReboot = true
			log.Log.Info("k8s-plugin OnNodeStateChange(): needReboot to update", "target", p.updateTarget)
		} else {
			log.Log.Info("k8s-plugin OnNodeStateChange(): needDrain to update", "target", p.updateTarget)
		}
	}

	return
}

// Apply config change
func (p *K8sPlugin) Apply() error {
	log.Log.Info("k8s-plugin Apply()")
	if err := p.updateSwitchdevService(); err != nil {
		return err
	}

	if p.useSystemdService {
		if err := p.updateSriovService(); err != nil {
			return err
		}
	}

	for _, systemService := range p.updateTarget.systemServices {
		if err := p.updateSystemService(systemService); err != nil {
			return err
		}
	}

	return nil
}

func (p *K8sPlugin) readSwitchdevManifest() error {
	// Read switchdev udev script
	switchdevUdevScript, err := service.ReadScriptManifestFile(switchdevRenamingUdevScript)
	if err != nil {
		return err
	}
	p.switchdevUdevScript = switchdevUdevScript

	return nil
}

func (p *K8sPlugin) readOpenVSwitchdManifest() error {
	openVSwitchService, err := service.ReadServiceInjectionManifestFile(ovsUnitFile)
	if err != nil {
		return err
	}

	p.openVSwitchService = openVSwitchService
	return nil
}

func (p *K8sPlugin) readSriovServiceManifest() error {
	sriovService, err := service.ReadServiceManifestFile(sriovUnitFile)
	if err != nil {
		return err
	}

	p.sriovService = sriovService
	return nil
}

func (p *K8sPlugin) readManifestFiles() error {
	if err := p.readSwitchdevManifest(); err != nil {
		return err
	}

	if err := p.readOpenVSwitchdManifest(); err != nil {
		return err
	}

	if err := p.readSriovServiceManifest(); err != nil {
		return err
	}

	return nil
}

func (p *K8sPlugin) switchdevServiceStateUpdate() error {
	// Check switchdev udev script
	needUpdate, err := p.isSwitchdevScriptNeedUpdate(p.switchdevUdevScript)
	if err != nil {
		return err
	}
	p.updateTarget.switchdevUdevScript = needUpdate

	return nil
}

func (p *K8sPlugin) sriovServiceStateUpdate() error {
	log.Log.Info("sriovServiceStateUpdate()")
	isServiceEnabled, err := p.serviceManager.IsServiceEnabled(p.sriovService.Path)
	if err != nil {
		return err
	}

	// create and enable the service if it doesn't exist or is not enabled
	if !isServiceEnabled {
		p.updateTarget.sriovScript = true
	} else {
		p.updateTarget.sriovScript = p.isSystemServiceNeedUpdate(p.sriovService)
	}

	if p.updateTarget.sriovScript {
		p.updateTarget.systemServices = append(p.updateTarget.systemServices, p.sriovService)
	}
	return nil
}

func (p *K8sPlugin) getSwitchDevSystemServices() []*service.Service {
	return []*service.Service{p.openVSwitchService}
}

func (p *K8sPlugin) isSwitchdevScriptNeedUpdate(scriptObj *service.ScriptManifestFile) (needUpdate bool, err error) {
	data, err := os.ReadFile(path.Join(chroot, scriptObj.Path))
	if err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
		return true, nil
	} else if string(data) != scriptObj.Contents.Inline {
		return true, nil
	}
	return false, nil
}

func (p *K8sPlugin) isSystemServiceNeedUpdate(serviceObj *service.Service) bool {
	log.Log.Info("isSystemServiceNeedUpdate()")
	systemService, err := p.serviceManager.ReadService(serviceObj.Path)
	if err != nil {
		log.Log.Error(err, "k8s-plugin isSystemServiceNeedUpdate(): failed to read sriov-config service file, ignoring",
			"path", serviceObj.Path)
		return false
	}
	if systemService != nil {
		needChange, err := service.CompareServices(systemService, serviceObj)
		if err != nil {
			log.Log.Error(err, "k8s-plugin isSystemServiceNeedUpdate(): failed to compare sriov-config service, ignoring")
			return false
		}
		return needChange
	}

	return false
}

func (p *K8sPlugin) systemServicesStateUpdate() error {
	var services []*service.Service
	for _, systemService := range p.getSwitchDevSystemServices() {
		exist, err := p.serviceManager.IsServiceExist(systemService.Path)
		if err != nil {
			return err
		}
		if !exist {
			return fmt.Errorf("k8s-plugin systemServicesStateUpdate(): %q not found", systemService.Name)
		}
		if p.isSystemServiceNeedUpdate(systemService) {
			services = append(services, systemService)
		}
	}

	p.updateTarget.systemServices = services
	return nil
}

func (p *K8sPlugin) switchDevServicesStateUpdate() error {
	// Check switchdev
	err := p.switchdevServiceStateUpdate()
	if err != nil {
		return err
	}

	// Check system services
	err = p.systemServicesStateUpdate()
	if err != nil {
		return err
	}

	return nil
}

func (p *K8sPlugin) updateSriovService() error {
	if p.updateTarget.sriovScript {
		err := p.serviceManager.EnableService(p.sriovService)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *K8sPlugin) updateSwitchdevService() error {
	if p.updateTarget.switchdevUdevScript {
		err := os.WriteFile(path.Join(chroot, p.switchdevUdevScript.Path),
			[]byte(p.switchdevUdevScript.Contents.Inline), 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *K8sPlugin) updateSystemService(serviceObj *service.Service) error {
	systemService, err := p.serviceManager.ReadService(serviceObj.Path)
	if err != nil {
		return err
	}
	if systemService == nil {
		// Invalid case to reach here
		return fmt.Errorf("k8s-plugin updateSystemService(): can't update non-existing service %q", serviceObj.Name)
	}
	serviceOptions, err := unit.Deserialize(strings.NewReader(serviceObj.Content))
	if err != nil {
		return err
	}
	updatedService, err := service.AppendToService(systemService, serviceOptions...)
	if err != nil {
		return err
	}

	return p.serviceManager.EnableService(updatedService)
}

func (p *K8sPlugin) SetSystemdFlag() {
	p.useSystemdService = true
}

func (p *K8sPlugin) IsSystemService() bool {
	return p.useSystemdService
}
