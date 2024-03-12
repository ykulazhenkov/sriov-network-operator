package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/renameio/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/consts"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/utils"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
)

// Contains all the file storing on the host
//
//go:generate ../../../bin/mockgen -destination mock/mock_store.go -source store.go
type ManagerInterface interface {
	ClearPCIAddressFolder() error
	SaveLastPfAppliedStatus(PfInfo *sriovnetworkv1.Interface) error
	LoadPfsStatus(pciAddress string) (*sriovnetworkv1.Interface, bool, error)

	GetCheckPointNodeState() (*sriovnetworkv1.SriovNetworkNodeState, error)
	WriteCheckpointFile(*sriovnetworkv1.SriovNetworkNodeState) error

	// GetManagedOVSBridges returns map with saved information about managed OVS bridges.
	// Bridge name is a key in the map
	GetManagedOVSBridges() map[string]*sriovnetworkv1.OVSConfigExt
	// GetManagedOVSBridge returns saved information about managed OVS bridge
	GetManagedOVSBridge(name string) *sriovnetworkv1.OVSConfigExt
	// AddManagedOVSBridge save information about the OVS bridge
	AddManagedOVSBridge(br *sriovnetworkv1.OVSConfigExt) error
	// RemoveManagedOVSBridge removes saved information about the OVS bridge
	RemoveManagedOVSBridge(name string) error
}

type manager struct {
	ovsStore
}

// NewManager: create the initial folders needed to store the info about the PF
// and return a manager struct that implements the ManagerInterface interface
func NewManager() (ManagerInterface, error) {
	if err := createOperatorConfigFolderIfNeeded(); err != nil {
		return nil, err
	}
	ovs, err := newOVSStore()
	if err != nil {
		return nil, err
	}
	return &manager{ovsStore: *ovs}, nil
}

// createOperatorConfigFolderIfNeeded: create the operator base folder on the host
// together with the pci folder to save the PF status objects as json files
func createOperatorConfigFolderIfNeeded() error {
	hostExtension := utils.GetHostExtension()
	SriovConfBasePathUse := filepath.Join(hostExtension, consts.SriovConfBasePath)
	_, err := os.Stat(SriovConfBasePathUse)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(SriovConfBasePathUse, os.ModeDir)
			if err != nil {
				return fmt.Errorf("failed to create the sriov config folder on host in path %s: %v", SriovConfBasePathUse, err)
			}
		} else {
			return fmt.Errorf("failed to check if the sriov config folder on host in path %s exist: %v", SriovConfBasePathUse, err)
		}
	}

	PfAppliedConfigUse := filepath.Join(hostExtension, consts.PfAppliedConfig)
	_, err = os.Stat(PfAppliedConfigUse)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(PfAppliedConfigUse, os.ModeDir)
			if err != nil {
				return fmt.Errorf("failed to create the pci folder on host in path %s: %v", PfAppliedConfigUse, err)
			}
		} else {
			return fmt.Errorf("failed to check if the pci folder on host in path %s exist: %v", PfAppliedConfigUse, err)
		}
	}

	return nil
}

// ClearPCIAddressFolder: removes all the PFs storage information
func (s *manager) ClearPCIAddressFolder() error {
	hostExtension := utils.GetHostExtension()
	PfAppliedConfigUse := filepath.Join(hostExtension, consts.PfAppliedConfig)
	_, err := os.Stat(PfAppliedConfigUse)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to check the pci address folder path %s: %v", PfAppliedConfigUse, err)
	}

	err = os.RemoveAll(PfAppliedConfigUse)
	if err != nil {
		return fmt.Errorf("failed to remove the PCI address folder on path %s: %v", PfAppliedConfigUse, err)
	}

	err = os.Mkdir(PfAppliedConfigUse, os.ModeDir)
	if err != nil {
		return fmt.Errorf("failed to create the pci folder on host in path %s: %v", PfAppliedConfigUse, err)
	}

	return nil
}

// SaveLastPfAppliedStatus will save the PF object as a json into the /etc/sriov-operator/pci/<pci-address>
// this function must be called after running the chroot function
func (s *manager) SaveLastPfAppliedStatus(PfInfo *sriovnetworkv1.Interface) error {
	data, err := json.Marshal(PfInfo)
	if err != nil {
		log.Log.Error(err, "failed to marshal PF status", "status", *PfInfo)
		return err
	}

	hostExtension := utils.GetHostExtension()
	pathFile := filepath.Join(hostExtension, consts.PfAppliedConfig, PfInfo.PciAddress)
	err = renameio.WriteFile(pathFile, data, 0644)
	return err
}

// LoadPfsStatus convert the /etc/sriov-operator/pci/<pci-address> json to pfstatus
// returns false if the file doesn't exist.
func (s *manager) LoadPfsStatus(pciAddress string) (*sriovnetworkv1.Interface, bool, error) {
	hostExtension := utils.GetHostExtension()
	pathFile := filepath.Join(hostExtension, consts.PfAppliedConfig, pciAddress)
	pfStatus := &sriovnetworkv1.Interface{}
	data, err := os.ReadFile(pathFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		log.Log.Error(err, "failed to read PF status", "path", pathFile)
		return nil, false, err
	}

	err = json.Unmarshal(data, pfStatus)
	if err != nil {
		log.Log.Error(err, "failed to unmarshal PF status", "data", string(data))
		return nil, false, err
	}

	return pfStatus, true, nil
}

func (s *manager) GetCheckPointNodeState() (*sriovnetworkv1.SriovNetworkNodeState, error) {
	log.Log.Info("getCheckPointNodeState()")
	configdir := filepath.Join(vars.Destdir, consts.CheckpointFileName)
	file, err := os.OpenFile(configdir, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()
	if err = json.NewDecoder(file).Decode(&sriovnetworkv1.InitialState); err != nil {
		return nil, err
	}

	return &sriovnetworkv1.InitialState, nil
}

func (s *manager) WriteCheckpointFile(ns *sriovnetworkv1.SriovNetworkNodeState) error {
	configdir := filepath.Join(vars.Destdir, consts.CheckpointFileName)
	file, err := os.OpenFile(configdir, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	log.Log.Info("WriteCheckpointFile(): try to decode the checkpoint file")
	if err = json.NewDecoder(file).Decode(&sriovnetworkv1.InitialState); err != nil {
		log.Log.V(2).Error(err, "WriteCheckpointFile(): fail to decode, writing new file instead")
		log.Log.Info("WriteCheckpointFile(): write checkpoint file")
		if err = file.Truncate(0); err != nil {
			return err
		}
		if _, err = file.Seek(0, 0); err != nil {
			return err
		}
		if err = json.NewEncoder(file).Encode(*ns); err != nil {
			return err
		}
		sriovnetworkv1.InitialState = *ns
	}
	return nil
}

func newOVSStore() (*ovsStore, error) {
	s := &ovsStore{
		lock:          &sync.RWMutex{},
		storeFilePath: utils.GetHostExtensionPath(consts.ManagedOVSBridgesPath),
	}
	var err error
	s.cache, err = s.readStoreFile()
	if err != nil {
		return nil, err
	}
	return s, nil
}

type ovsStore struct {
	lock          *sync.RWMutex
	cache         map[string]sriovnetworkv1.OVSConfigExt
	storeFilePath string
}

// GetManagedOVSBridges returns map with saved information about managed OVS bridges.
// Bridge name is a key in the map
func (s *ovsStore) GetManagedOVSBridges() map[string]*sriovnetworkv1.OVSConfigExt {
	funcLog := log.Log
	funcLog.V(2).Info("GetManagedOVSBridges(): get information about all managed OVS bridges from the store")
	s.lock.RLock()
	defer s.lock.RUnlock()
	result := make(map[string]*sriovnetworkv1.OVSConfigExt, len(s.cache))
	for k, v := range s.cache {
		result[k] = v.DeepCopy()
	}
	if funcLog.V(3).Enabled() {
		data, _ := json.Marshal(result)
		funcLog.V(3).Info("GetManagedOVSBridges()", "result", string(data))
	}
	return result
}

// GetManagedOVSBridge returns saved information about managed OVS bridge
func (s *ovsStore) GetManagedOVSBridge(name string) *sriovnetworkv1.OVSConfigExt {
	funcLog := log.Log.WithValues("name", name)
	funcLog.V(2).Info("GetManagedOVSBridge(): get information about managed OVS bridge from the store")
	s.lock.RLock()
	defer s.lock.RUnlock()
	b, found := s.cache[name]
	if !found {
		funcLog.V(2).Info("GetManagedOVSBridge(): bridge info not found")
		return nil
	}
	if funcLog.V(3).Enabled() {
		data, _ := json.Marshal(&b)
		funcLog.V(3).Info("GetManagedOVSBridge()", "result", string(data))
	}
	return b.DeepCopy()
}

// AddManagedOVSBridge save information about the OVS bridge
func (s *ovsStore) AddManagedOVSBridge(br *sriovnetworkv1.OVSConfigExt) error {
	log.Log.V(2).Info("AddManagedOVSBridge(): add information about managed OVS bridge to the store", "name", br.Name)
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cache[br.Name] = *br.DeepCopy()
	return s.writeStoreFile()
}

// RemoveManagedOVSBridge removes saved information about the OVS bridge
func (s *ovsStore) RemoveManagedOVSBridge(name string) error {
	log.Log.V(2).Info("RemoveManagedOVSBridge(): remove information about managed OVS bridge from the store", "name", name)
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.cache, name)
	return s.writeStoreFile()
}

func (s *ovsStore) readStoreFile() (map[string]sriovnetworkv1.OVSConfigExt, error) {
	funcLog := log.Log.WithValues("storeFilePath", s.storeFilePath)
	funcLog.V(2).Info("readStoreFile(): read OVS store file")
	result := map[string]sriovnetworkv1.OVSConfigExt{}
	data, err := os.ReadFile(s.storeFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			funcLog.V(2).Info("readStoreFile(): OVS store file not found")
			return result, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &result); err != nil {
		funcLog.Error(err, "readStoreFile(): failed to unmarshal content of the OVS store file")
		return nil, err
	}
	return result, nil
}

func (s *ovsStore) writeStoreFile() error {
	funcLog := log.Log.WithValues("storeFilePath", s.storeFilePath)
	data, err := json.Marshal(s.cache)
	if err != nil {
		funcLog.Error(err, "writeStoreFile(): can't serialize cached info about managed OVS bridges")
		return err
	}
	if err := renameio.WriteFile(s.storeFilePath, data, 0644); err != nil {
		funcLog.Error(err, "writeStoreFile(): can't write info about managed OVS bridge to disk")
		return err
	}
	return nil
}
