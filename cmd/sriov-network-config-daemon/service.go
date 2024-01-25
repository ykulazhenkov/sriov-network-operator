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
	serviceCmd.Flags().StringVarP(&phase, "phase", "p", "", fmt.Sprintf("execution phase, supported values are: %s, %s", PhasePre, PhasePost))
}

func runServiceCmd(cmd *cobra.Command, args []string) error {
	if phase != PhasePre && phase != PhasePost {
		return fmt.Errorf("invalid value for required argument \"--phase\", valid values are: %s, %s", PhasePre, PhasePost)
	}
	// init logger
	snolog.InitLog()
	setupLog := log.Log.WithName("sriov-config-service")

	// To help debugging, immediately log version
	setupLog.V(2).Info("sriov-config-service", "version", version.Version)

	setupLog.V(0).Info("Starting sriov-config-service", "phase", phase)

	// Mark that we are running on host
	vars.UsingSystemdMode = true
	vars.InChroot = true

	if err := handleResultFile(setupLog); err != nil {
		return err
	}

	nodeStateSpec, err := systemd.ReadConfFile()
	if err != nil {
		if _, err := os.Stat(systemd.SriovSystemdConfigPath); !errors.Is(err, os.ErrNotExist) {
			setupLog.Error(err, "failed to read the sriov configuration file", "path", systemd.SriovSystemdConfigPath)
			return updateSriovResultErr(fmt.Errorf("failed to read the sriov configuration file in path %s: %v", systemd.SriovSystemdConfigPath, err))
		}
		setupLog.Info("configuration file not found, use default config")
		nodeStateSpec = &systemd.SriovConfig{
			Spec:            sriovv1.SriovNetworkNodeStateSpec{},
			UnsupportedNics: false,
			PlatformType:    consts.Baremetal,
		}
	}

	setupLog.V(2).Info("sriov-config-service", "config", nodeStateSpec)

	if phase == PhasePost {
		if nodeStateSpec.PlatformType == consts.VirtualOpenStack {
			setupLog.Info("skip post phase for virtual cluster")
			return updateResult(consts.SyncStatusSucceeded, "")
		}
	}
	supportedNicIds, err := systemd.ReadSriovSupportedNics()
	if err != nil {
		setupLog.Error(err, "failed to read list of supported nic ids")
		return updateSriovResultErr(fmt.Errorf("failed to read list of supported nic ids: %v", err))
	}
	sriovv1.InitNicIDMapFromList(supportedNicIds)

	hostHelpers, err := helper.NewDefaultHostHelpers()
	if err != nil {
		setupLog.Error(err, "failed to create hostHelpers")
		return updateSriovResultErr(fmt.Errorf("failed to create hostHelpers"))
	}
	platformHelper, err := platforms.NewDefaultPlatformHelper()
	if err != nil {
		setupLog.Error(err, "failed to create platformHelpers")
		return updateSriovResultErr(fmt.Errorf("failed to create platformHelpers"))
	}

	if phase == PhasePre {
		_, err = hostHelpers.TryEnableRdma()
		if err != nil {
			setupLog.Error(err, "warning, failed to enable RDMA")
		}
		hostHelpers.TryEnableTun()
		hostHelpers.TryEnableVhostNet()
	}

	var configPlugin plugin.VendorPlugin
	var ifaceStatuses []sriovv1.InterfaceExt
	if nodeStateSpec.PlatformType == consts.Baremetal {
		// Bare metal support
		vars.DevMode = nodeStateSpec.UnsupportedNics
		ifaceStatuses, err = hostHelpers.DiscoverSriovDevices(hostHelpers)
		if err != nil {
			setupLog.Error(err, "failed to discover sriov devices on the host")
			return updateSriovResultErr(fmt.Errorf("sriov-config-service: failed to discover sriov devices on the host:  %v", err))
		}

		// Create the generic plugin
		configPlugin, err = generic.NewGenericPlugin(hostHelpers, phase == PhasePre)
		if err != nil {
			setupLog.Error(err, "failed to create generic plugin")
			return updateSriovResultErr(fmt.Errorf("sriov-config-service failed to create generic plugin %v", err))
		}
	} else if nodeStateSpec.PlatformType == consts.VirtualOpenStack {
		err = platformHelper.CreateOpenstackDevicesInfo()
		if err != nil {
			setupLog.Error(err, "failed to read OpenStack data")
			return updateSriovResultErr(fmt.Errorf("sriov-config-service failed to read OpenStack data: %v", err))
		}

		ifaceStatuses, err = platformHelper.DiscoverSriovDevicesVirtual()
		if err != nil {
			setupLog.Error(err, "failed to read OpenStack data")
			return updateSriovResultErr(fmt.Errorf("sriov-config-service: failed to read OpenStack data: %v", err))
		}

		// Create the virtual plugin
		configPlugin, err = virtual.NewVirtualPlugin(hostHelpers)
		if err != nil {
			setupLog.Error(err, "failed to create virtual plugin")
			return updateSriovResultErr(fmt.Errorf("sriov-config-service: failed to create virtual plugin %v", err))
		}
	}

	nodeState := &sriovv1.SriovNetworkNodeState{
		Spec:   nodeStateSpec.Spec,
		Status: sriovv1.SriovNetworkNodeStateStatus{Interfaces: ifaceStatuses},
	}

	_, _, err = configPlugin.OnNodeStateChange(nodeState)
	if err != nil {
		setupLog.Error(err, "failed to run OnNodeStateChange to update the generic plugin status")
		return updateSriovResultErr(fmt.Errorf("sriov-config-service: failed to run OnNodeStateChange to update the generic plugin status %v", err))
	}

	err = configPlugin.Apply()
	if err != nil {
		setupLog.Error(err, "failed to run apply node configuration")
		err = updateSriovResultErr(fmt.Errorf("failed to apply configuration: %v", err))
	} else {
		err = updateResult(consts.SyncStatusSucceeded, "")
	}

	setupLog.V(0).Info("shutting down sriov-config-service")
	return err
}

func handleResultFile(setupLog logr.Logger) error {
	if phase == PhasePre {
		// make sure there is no stale result file to avoid situation when we
		// read outdated info in the Post phase when the Pre phase failed
		if err := systemd.RemoveSriovResult(); err != nil {
			err := fmt.Errorf("failed to remove result file: %v", err)
			setupLog.Error(err, "failed to remove result file")
			return updateSriovResultErr(err)
		}
		return nil
	}

	// we are in Post phase, check result file created by Pre phase
	prePhaseResult, err := systemd.ReadSriovResult()
	if err != nil {
		err := fmt.Errorf("failed to read result file: %v", err)
		setupLog.Error(err, "failed to read result file")
		return updateSriovResultErr(err)
	}
	if prePhaseResult.SyncStatus != consts.SyncStatusInProgress {
		err := fmt.Errorf("unexpected result of the pre phase: %s", prePhaseResult.SyncStatus)
		setupLog.Error(err, "unexpected result")
		// if SyncStatus is failed we should keep the original message,
		// otherwise set status and message
		if prePhaseResult.SyncStatus != consts.SyncStatusFailed {
			return updateSriovResultErr(err)
		}
		return err
	}
	setupLog.V(0).Info("read valid result of the Pre phase")
	return nil
}

func updateSriovResultErr(origErr error) error {
	err := updateResult(consts.SyncStatusFailed, fmt.Sprintf("%s: %s", phase, origErr.Error()))
	if err != nil {
		return err
	}
	return origErr
}

func updateResult(result, msg string) error {
	sriovResult := &systemd.SriovResult{
		SyncStatus:    result,
		LastSyncError: msg,
	}
	err := systemd.WriteSriovResult(sriovResult)
	if err != nil {
		log.Log.Error(err, "failed to write sriov result file", "content", *sriovResult)
		return fmt.Errorf("sriov-config-service failed to write sriov result file with content %v error: %v", *sriovResult, err)
	}
	return nil
}
