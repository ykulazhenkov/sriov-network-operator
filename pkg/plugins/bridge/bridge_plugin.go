package main

import (
	"github.com/golang/glog"
	"github.com/vishvananda/netlink"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
)

type BridgePlugin struct {
	PluginName     string
	SpecVersion    string
	DesireState    *sriovnetworkv1.SriovNetworkNodeState
}


var Plugin BridgePlugin

// Initialize our plugin and set up initial values
func init() {
	Plugin = BridgePlugin{
		PluginName:     "bridge_plugin",
		SpecVersion:    "1.0",
	}
}

// Name returns the name of the plugin
func (p *BridgePlugin) Name() string {
	return p.PluginName
}

// Spec returns the version of the spec expected by the plugin
func (p *BridgePlugin) Spec() string {
	return p.SpecVersion
}

// OnNodeStateAdd Invoked when SriovNetworkNodeState CR is created, return if need dain and/or reboot node
func (p *BridgePlugin) OnNodeStateAdd(state *sriovnetworkv1.SriovNetworkNodeState) (needDrain bool, needReboot bool, err error) {
	glog.Info("bridge-plugin OnNodeStateAdd()")
	p.DesireState = state
	return false, false ,nil
}

// OnNodeStateChange Invoked when SriovNetworkNodeState CR is updated, return if need dain and/or reboot node
func (p *BridgePlugin) OnNodeStateChange(_, new *sriovnetworkv1.SriovNetworkNodeState) (needDrain bool, needReboot bool, err error) {
	glog.Info("bridge-plugin OnNodeStateChange()")
	p.DesireState = new
	return false, false ,nil
}

// Apply config change
func (p *BridgePlugin) Apply() error {
	glog.Infof("bridge-plugin Apply(): desiredState=%v", p.DesireState.Spec)
	for _, br := range p.DesireState.Spec.Bridges {
		link := &netlink.Bridge{
			LinkAttrs: netlink.LinkAttrs{
				Name: br.Name,
			},
		}
		err := netlink.LinkAdd(link)
		if err != nil {
			return err
		}
		err = netlink.LinkSetUp(link)
		if err != nil {
			return err
		}
	}
	return nil
}
