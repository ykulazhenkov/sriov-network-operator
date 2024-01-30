/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sriovv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/helper"
	snolog "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/log"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/platforms"
	plugin "github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/plugins"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/plugins/generic"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/plugins/virtual"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/systemd"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/version"
)

const (
	PhasePre  = "pre"
	PhasePost = "post"
)

var (
	serviceCmd = &cobra.Command{
		Use:   "service",
		Short: "Starts SR-IOV service Config",
		Long:  "",
		RunE:  runServiceCmd,
	}
	phase string
)

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.Flags().StringVarP(&phase, "phase", "p", "", fmt.Sprintf("configuration phase, supported values are: %s, %s", PhasePre, PhasePost))
}

// The service has required argument "phase"
// Two phases are supported now:
// * pre - before the NetworkManager or systemd-networkd
// * post - after the NetworkManager or systemd-networkd
// "sriov-config" systemd unit is responsible for starting the service in the "pre" phase mode.
// "sriov-config-post-network" systemd unit starts the service in the "post" phase mode.
// The service may use different plugins for each phase and call different initialization flows.
// The "post" phase checks the completion status of the "pre" phase by reading the sriov result file.
// The "pre" phase should set "InProgress" status if it succeeds or "Failed" otherwise.
// If the result of the "pre" phase is different than "InProgress", then the "post" phase will not be executed
// and the execution result will be forcefully set to "Failed".
func runServiceCmd(cmd *cobra.Command, args []string) error {
	if phase != PhasePre && phase != PhasePost {
		return fmt.Errorf("invalid value for required argument \"--phase\", valid values are: %s, %s", PhasePre, PhasePost)
	}
	// init logger
	snolog.InitLog()
	setupLog := log.Log.WithName("sriov-config-service").WithValues("phase", phase)

	// To help debugging, immediately log version
	setupLog.V(2).Info("sriov-config-service", "version", version.Version)

	setupLog.V(0).Info("Starting sriov-config-service")

	// Mark that we are running on host
	vars.UsingSystemdMode = true
	vars.InChroot = true

	if err := handleResultFile(setupLog); err != nil {
		return updateSriovResultErr(setupLog, err)
	}

	nodeStateSpec, err := systemd.ReadConfFile()
	if err != nil {
		if _, err := os.Stat(systemd.SriovSystemdConfigPath); !errors.Is(err, os.ErrNotExist) {
			return updateSriovResultErr(setupLog, fmt.Errorf("failed to read the sriov configuration file in path %s: %v", systemd.SriovSystemdConfigPath, err))
		}
		setupLog.Info("configuration file not found, use default config")
		nodeStateSpec = &systemd.SriovConfig{
			Spec:            sriovv1.SriovNetworkNodeStateSpec{},
			UnsupportedNics: false,
			PlatformType:    consts.Baremetal,
		}
	}

	setupLog.V(2).Info("sriov-config-service", "config", nodeStateSpec)

	supportedNicIds, err := systemd.ReadSriovSupportedNics()
	if err != nil {
		return updateSriovResultErr(setupLog, fmt.Errorf("failed to read list of supported nic ids: %v", err))
	}
	sriovv1.InitNicIDMapFromList(supportedNicIds)

	hostHelpers, err := helper.NewDefaultHostHelpers()
	if err != nil {
		return updateSriovResultErr(setupLog, fmt.Errorf("failed to create hostHelpers"))
	}

	if phase == PhasePre {
		_, err = hostHelpers.TryEnableRdma()
		if err != nil {
			setupLog.Error(err, "warning, failed to enable RDMA")
		}
		hostHelpers.TryEnableTun()
		hostHelpers.TryEnableVhostNet()
	}

	configPlugin, nodeState, err := getPluginAndState(setupLog, hostHelpers, nodeStateSpec)
	if err != nil {
		return updateSriovResultErr(setupLog, err)
	}

	if configPlugin == nil || nodeState == nil {
		setupLog.V(0).Info("no plugin for the platform for the current phase", "platform", nodeStateSpec.PlatformType)
		return updateSriovResultSucceed(setupLog)
	}

	_, _, err = configPlugin.OnNodeStateChange(nodeState)
	if err != nil {
		return updateSriovResultErr(setupLog,
			fmt.Errorf("sriov-config-service: failed to run OnNodeStateChange to update the plugin status %v", err))
	}

	err = configPlugin.Apply()
	if err != nil {
		err = updateSriovResultErr(setupLog, fmt.Errorf("failed to apply configuration: %v", err))
	} else {
		err = updateSriovResultSucceed(setupLog)
	}

	setupLog.V(0).Info("shutting down sriov-config-service")
	return err
}

func getPluginAndState(setupLog logr.Logger, hostHelpers helper.HostHelpersInterface, conf *systemd.SriovConfig) (
	plugin.VendorPlugin, *sriovv1.SriovNetworkNodeState, error) {
	var (
		configPlugin  plugin.VendorPlugin
		ifaceStatuses []sriovv1.InterfaceExt
		err           error
	)
	switch conf.PlatformType {
	case consts.Baremetal:
		if phase == PhasePost {
			// TODO add initialization for the generic plugin for the post phase
			setupLog.Info("post phase is not implemented for generic plugin yet, skip")
			return nil, nil, nil
		}
		vars.DevMode = conf.UnsupportedNics
		ifaceStatuses, err = hostHelpers.DiscoverSriovDevices(hostHelpers)
		if err != nil {
			return nil, nil, fmt.Errorf("sriov-config-service: failed to discover sriov devices on the host:  %v", err)
		}
		configPlugin, err = generic.NewGenericPlugin(hostHelpers)
		if err != nil {
			return nil, nil, fmt.Errorf("sriov-config-service failed to create generic plugin %v", err)
		}
	case consts.VirtualOpenStack:
		if phase == PhasePost {
			setupLog.Info("skip post configuration phase for virtual cluster")
			return nil, nil, nil
		}
		platformHelper, err := platforms.NewDefaultPlatformHelper()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create platformHelpers")
		}
		err = platformHelper.CreateOpenstackDevicesInfo()
		if err != nil {
			return nil, nil, fmt.Errorf("sriov-config-service failed to read OpenStack data: %v", err)
		}
		ifaceStatuses, err = platformHelper.DiscoverSriovDevicesVirtual()
		if err != nil {
			return nil, nil, fmt.Errorf("sriov-config-service: failed to read OpenStack data: %v", err)
		}
		configPlugin, err = virtual.NewVirtualPlugin(hostHelpers)
		if err != nil {
			return nil, nil, fmt.Errorf("sriov-config-service: failed to create virtual plugin %v", err)
		}
	}
	return configPlugin, &sriovv1.SriovNetworkNodeState{
		Spec:   conf.Spec,
		Status: sriovv1.SriovNetworkNodeStateStatus{Interfaces: ifaceStatuses},
	}, nil
}

func handleResultFile(setupLog logr.Logger) error {
	if phase == PhasePre {
		// make sure there is no stale result file to avoid situation when we
		// read outdated info in the Post phase when the Pre silently failed (should not happen)
		if err := systemd.RemoveSriovResult(); err != nil {
			return fmt.Errorf("failed to remove result file: %v", err)
		}
		return nil
	}
	// we are in the Post phase, check result file created by the Pre phase
	setupLog.V(0).Info("check result of the Pre phase")
	prePhaseResult, err := systemd.ReadSriovResult()
	if err != nil {
		return fmt.Errorf("failed to read result of the pre phase: %v", err)
	}
	if prePhaseResult.SyncStatus != consts.SyncStatusInProgress {
		return fmt.Errorf("unexpected result of the pre phase: %s, syncError: %s", prePhaseResult.SyncStatus, prePhaseResult.LastSyncError)
	}
	setupLog.V(0).Info("Pre phase succeed, continue execution")
	return nil
}

func updateSriovResultErr(setupLog logr.Logger, origErr error) error {
	setupLog.Error(origErr, "service call failed")
	err := updateResult(setupLog, consts.SyncStatusFailed, fmt.Sprintf("%s: %s", phase, origErr.Error()))
	if err != nil {
		return err
	}
	return origErr
}

func updateSriovResultSucceed(setupLog logr.Logger) error {
	setupLog.V(0).Info("service call succeed")
	syncStatus := consts.SyncStatusSucceeded
	if phase == PhasePre {
		syncStatus = consts.SyncStatusInProgress
	}
	return updateResult(setupLog, syncStatus, "")
}

func updateResult(setupLog logr.Logger, result, msg string) error {
	sriovResult := &systemd.SriovResult{
		SyncStatus:    result,
		LastSyncError: msg,
	}
	err := systemd.WriteSriovResult(sriovResult)
	if err != nil {
		setupLog.Error(err, "failed to write sriov result file", "content", *sriovResult)
		return fmt.Errorf("sriov-config-service failed to write sriov result file with content %v error: %v", *sriovResult, err)
	}
	setupLog.V(0).Info("result file updated", "SyncStatus", sriovResult.SyncStatus, "LastSyncError", msg)
	return nil
}
