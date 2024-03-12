package ovs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/ovn-org/libovsdb/client"
	"github.com/ovn-org/libovsdb/model"
	"github.com/ovn-org/libovsdb/ovsdb"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sriovnetworkv1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/store"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/host/types"
	"github.com/k8snetworkplumbingwg/sriov-network-operator/pkg/vars"
)

// New creates new instance of the OVSInterface
func New() types.OVSInterface {
	return &ovs{socketPath: vars.OVSDBSocketPath}
}

type ovs struct {
	socketPath string
}

// GetOVSManager return new instance of OVS client
func (o *ovs) GetOVSManager(ctx context.Context) (types.OVSManagerInterface, error) {
	return newOVSManager(ctx, o.socketPath)
}

// OpenvSwitchEntry defines schema of the object in the Open_vSwitch table
type OpenvSwitchEntry struct {
	UUID    string   `ovsdb:"_uuid"`
	Bridges []string `ovsdb:"bridges"`
}

// BridgeEntry defines schema of the object in the Bridge table
type BridgeEntry struct {
	UUID         string            `ovsdb:"_uuid"`
	Name         string            `ovsdb:"name"`
	DatapathType string            `ovsdb:"datapath_type"`
	ExternalIDs  map[string]string `ovsdb:"external_ids"`
	OtherConfig  map[string]string `ovsdb:"other_config"`
	Ports        []string          `ovsdb:"ports"`
}

// HasPort returns true if portUUID is found in Ports slice
func (b *BridgeEntry) HasPort(portUUID string) bool {
	return slices.Contains(b.Ports, portUUID)
}

// InterfaceEntry defines schema of the object in the Interface table
type InterfaceEntry struct {
	UUID        string            `ovsdb:"_uuid"`
	Name        string            `ovsdb:"name"`
	Type        string            `ovsdb:"type"`
	Error       *string           `ovsdb:"error"`
	Options     map[string]string `ovsdb:"options"`
	ExternalIDs map[string]string `ovsdb:"external_ids"`
	OtherConfig map[string]string `ovsdb:"other_config"`
}

// PortEntry defines schema of the object in the Port table
type PortEntry struct {
	UUID       string   `ovsdb:"_uuid"`
	Name       string   `ovsdb:"name"`
	Interfaces []string `ovsdb:"interfaces"`
}

func newOVSManager(ctx context.Context, ovsSocket string) (*ovsManager, error) {
	openvSwitchEntry := &OpenvSwitchEntry{}
	bridgeEntry := &BridgeEntry{}
	interfaceEntry := &InterfaceEntry{}
	portEntry := &PortEntry{}
	clientDBModel, err := model.NewClientDBModel("Open_vSwitch",
		map[string]model.Model{
			"Open_vSwitch": openvSwitchEntry,
			"Bridge":       bridgeEntry,
			"Interface":    interfaceEntry,
			"Port":         portEntry,
		})
	if err != nil {
		return nil, fmt.Errorf("can't create client DB model: %v", err)
	}

	dbClient, err := client.NewOVSDBClient(clientDBModel, client.WithEndpoint(ovsSocket))
	if err != nil {
		return nil, fmt.Errorf("can't create DB client: %v", err)
	}

	err = dbClient.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't connect to ovsdb server: %v", err)
	}
	_, err = dbClient.Monitor(ctx, dbClient.NewMonitor(
		client.WithTable(openvSwitchEntry,
			&openvSwitchEntry.UUID,
			&openvSwitchEntry.Bridges,
		),
		client.WithTable(bridgeEntry,
			&bridgeEntry.UUID,
			&bridgeEntry.Name,
			&bridgeEntry.DatapathType,
			&bridgeEntry.ExternalIDs,
			&bridgeEntry.OtherConfig,
			&bridgeEntry.Ports,
		),
		client.WithTable(interfaceEntry,
			&interfaceEntry.UUID,
			&interfaceEntry.Name,
			&interfaceEntry.Type,
			&interfaceEntry.Error,
			&interfaceEntry.Options,
			&interfaceEntry.ExternalIDs,
			&interfaceEntry.OtherConfig,
		),
		client.WithTable(portEntry,
			&portEntry.UUID,
			&portEntry.Name,
			&portEntry.Interfaces,
		),
	))
	if err != nil {
		dbClient.Close()
		return nil, fmt.Errorf("can't start monitor: %v", err)
	}
	var rootUUID string
	for uuid := range dbClient.Cache().Table("Open_vSwitch").Rows() {
		rootUUID = uuid
	}
	if rootUUID == "" {
		return nil, fmt.Errorf("can't retrieve object uuid from Open_vSwitch table")
	}
	return &ovsManager{
		client:  dbClient,
		rootObj: &OpenvSwitchEntry{UUID: rootUUID}}, nil
}

type ovsManager struct {
	client  client.Client
	rootObj *OpenvSwitchEntry
}

// Close connection to ovsdb
func (c *ovsManager) Close() {
	c.client.Close()
}

// CreateOVSBridge creates OVS bridge from the provided config,
// does nothing if OVS bridge with the right config already exist,
// if OVS bridge exist with different config it will be removed and re-created
func (c *ovsManager) CreateOVSBridge(ctx context.Context, conf *sriovnetworkv1.OVSConfigExt, storeManager store.ManagerInterface) error {
	if len(conf.Uplinks) != 1 {
		return fmt.Errorf("unsupported configuration, uplinks list must contain one element")
	}
	funcLog := log.Log.WithValues("bridge", conf.Name, "ifaceAddr", conf.Uplinks[0].PciAddress, "ifaceName", conf.Uplinks[0].Name)
	funcLog.V(2).Info("CreateOVSBridge(): start configuration of the OVS bridge")
	knownConfig := storeManager.GetManagedOVSBridge(conf.Name)
	if knownConfig == nil || !reflect.DeepEqual(conf, knownConfig) {
		funcLog.V(2).Info("CreateOVSBridge(): save current configuration to the store")
		// config in store manager is not found or it is not the same config as passed with conf arg,
		// update config in the store manager
		if err := storeManager.AddManagedOVSBridge(conf); err != nil {
			funcLog.Error(err, "CreateOVSBridge(): failed to save current configuration to the store")
			return err
		}
	}
	keepBridge := false
	if knownConfig != nil {
		funcLog.V(2).Info("CreateOVSBridge(): configuration for the bridge found in the store")
		// use knownConfig to query current state
		currentState, err := c.getCurrentBridgeState(ctx, knownConfig)
		if err != nil {
			funcLog.Error(err, "CreateOVSBridge(): failed to query current bridge state")
			return err
		}
		if reflect.DeepEqual(conf, currentState) {
			// bridge already exist with the right config
			funcLog.V(2).Info("CreateOVSBridge(): bridge state already match current configuration, no actions required")
			return nil
		}
		funcLog.V(2).Info("CreateOVSBridge(): bridge state differs from the current configuration, reconfiguration required")
		keepBridge = reflect.DeepEqual(conf.Bridge, currentState.Bridge)
		if !keepBridge {
			funcLog.V(2).Info("CreateOVSBridge(): remove existing bridge")
			if err := c.deleteBridge(ctx, conf.Name); err != nil {
				funcLog.Error(err, "CreateOVSBridge(): failed to remove existing bridge")
				return err
			}
		}
	} else {
		funcLog.V(2).Info("CreateOVSBridge(): configuration for the bridge not found in the store, create the bridge")
	}
	funcLog.V(2).Info("CreateOVSBridge(): ensure uplink is not attached to any bridge")
	// removal of the bridge should also remove all interfaces that are attached to it.
	// we need to remove interface with additional call even if keepBridge is false to make
	// sure that the interface is not attached to a different OVS bridge
	if err := c.deleteInterface(ctx, conf.Uplinks[0].Name); err != nil {
		funcLog.Error(err, "CreateOVSBridge(): failed to remove uplink interface")
		return err
	}
	if !keepBridge {
		funcLog.V(2).Info("CreateOVSBridge(): create OVS bridge")
		if err := c.createBridge(ctx, &BridgeEntry{
			Name:         conf.Name,
			DatapathType: conf.Bridge.DatapathType,
			ExternalIDs:  conf.Bridge.ExternalIDs,
			OtherConfig:  conf.Bridge.ExternalIDs,
		}); err != nil {
			return err
		}
	}
	bridge, err := c.getBridgeByName(ctx, conf.Name)
	if err != nil {
		funcLog.Error(err, "CreateOVSBridge(): failed to retrieve information about created bridge from OVSDB")
		return err
	}
	if bridge == nil {
		err = fmt.Errorf("can't retrieve bridge after creation")
		funcLog.Error(err, "CreateOVSBridge(): failed to get bridge after creation")
		return err
	}
	funcLog.V(2).Info("CreateOVSBridge(): add uplink interface to the bridge")
	if err := c.addInterface(ctx, bridge, &InterfaceEntry{
		Name:        conf.Uplinks[0].Name,
		Type:        conf.Uplinks[0].Interface.Type,
		Options:     conf.Uplinks[0].Interface.Options,
		ExternalIDs: conf.Uplinks[0].Interface.ExternalIDs,
		OtherConfig: conf.Uplinks[0].Interface.OtherConfig,
	}); err != nil {
		funcLog.Error(err, "CreateOVSBridge(): failed to add uplink interface to the bridge")
		return err
	}
	return nil
}

// GetOVSBridges returns configuration for all managed bridges
func (c *ovsManager) GetOVSBridges(ctx context.Context, storeManager store.ManagerInterface) ([]sriovnetworkv1.OVSConfigExt, error) {
	funcLog := log.Log
	funcLog.V(2).Info("GetOVSBridges(): get managed OVS bridges")
	knownConfigs := storeManager.GetManagedOVSBridges()
	result := make([]sriovnetworkv1.OVSConfigExt, 0, len(knownConfigs))
	for _, knownConfig := range knownConfigs {
		currentState, err := c.getCurrentBridgeState(ctx, knownConfig)
		if err != nil {
			funcLog.Error(err, "GetOVSBridges(): failed to get state for the managed bridge", "bridge", knownConfig.Name)
			return nil, err
		}
		result = append(result, *currentState)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	if funcLog.V(3).Enabled() {
		data, _ := json.Marshal(&result)
		funcLog.V(3).Info("GetOVSBridges()", "result", string(data))
	}
	return result, nil
}

// RemoveOVSBridge removes OVS bridge to which PF with ifaceAddr is attached
func (c *ovsManager) RemoveOVSBridge(ctx context.Context, ifaceAddr string, storeManager store.ManagerInterface) error {
	funcLog := log.Log.WithValues("ifaceAddr", ifaceAddr)
	funcLog.V(2).Info("RemoveOVSBridge(): remove managed bridge for the interface")
	knownConfigs := storeManager.GetManagedOVSBridges()
	foundBridge := false
	for _, knownConfig := range knownConfigs {
		if len(knownConfig.Uplinks) != 1 {
			// unexpected config
			continue
		}
		if knownConfig.Uplinks[0].PciAddress != ifaceAddr {
			continue
		}
		foundBridge = true
		funcLog.V(2).Info("RemoveOVSBridge(): interface belongs to the managed bridge", "bridge", knownConfig.Name)
		currentState, err := c.getCurrentBridgeState(ctx, knownConfig)
		if err != nil {
			funcLog.Error(err, "RemoveOVSBridge(): failed to get state of the managed bridge", "bridge", knownConfig.Name)
			return err
		}
		if currentState != nil {
			funcLog.V(2).Info("RemoveOVSBridge(): remove managed bridge", "bridge", knownConfig.Name)
			if err := c.deleteBridge(ctx, knownConfig.Name); err != nil {
				funcLog.Error(err, "RemoveOVSBridge(): failed to remove managed bridge", "bridge", knownConfig.Name)
				return err
			}
		} else {
			funcLog.V(2).Info("RemoveOVSBridge(): managed bridge not exist", "bridge", knownConfig.Name)
		}

		funcLog.V(2).Info("RemoveOVSBridge(): remove information about the bridge from the store", "bridge", knownConfig.Name)
		if err := storeManager.RemoveManagedOVSBridge(knownConfig.Name); err != nil {
			funcLog.Error(err, "RemoveOVSBridge(): failed to remove information from the store", "bridge", knownConfig.Name)
			return err
		}
	}
	if !foundBridge {
		funcLog.V(2).Info("RemoveOVSBridge(): interface doesn't belong to any managed OVS bridge")
	}
	return nil
}

func (c *ovsManager) getBridgeByName(ctx context.Context, name string) (*BridgeEntry, error) {
	br := &BridgeEntry{Name: name}
	if err := c.client.Get(ctx, br); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("get call for the bridge %s failed: %v", name, err)
		}
	}
	return br, nil
}

func (c *ovsManager) getInterfaceByName(ctx context.Context, name string) (*InterfaceEntry, error) {
	iface := &InterfaceEntry{Name: name}
	if err := c.client.Get(ctx, iface); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("get call for the interfaces %s failed: %v", name, err)
		}
	}
	return iface, nil
}

func (c *ovsManager) getPortByInterface(ctx context.Context, iface *InterfaceEntry) (*PortEntry, error) {
	portEntry := &PortEntry{}
	portEntryList := []*PortEntry{}
	err := c.client.WhereAll(portEntry, model.Condition{
		Field:    &portEntry.Interfaces,
		Function: ovsdb.ConditionIncludes,
		Value:    []string{iface.UUID},
	}).List(ctx, &portEntryList)
	if err != nil {
		return nil, fmt.Errorf("failed to list ports related to interface %s: %v", iface.Name, err)
	}
	if len(portEntryList) == 0 {
		return nil, nil
	}
	return portEntryList[0], nil
}

func (c *ovsManager) getBridgeByPort(ctx context.Context, port *PortEntry) (*BridgeEntry, error) {
	brEntry := &BridgeEntry{}
	brEntryList := []*BridgeEntry{}
	err := c.client.WhereAll(brEntry, model.Condition{
		Field:    &brEntry.Ports,
		Function: ovsdb.ConditionIncludes,
		Value:    []string{port.UUID},
	}).List(ctx, &brEntryList)
	if err != nil {
		return nil, fmt.Errorf("failed to list bridges related to port %s: %v", port.Name, err)
	}
	if len(brEntryList) == 0 {
		return nil, nil
	}
	return brEntryList[0], nil
}

// create bridge with provided configuration
func (c *ovsManager) createBridge(ctx context.Context, br *BridgeEntry) error {
	brCreateOps, err := c.client.Create(br)
	if err != nil {
		return fmt.Errorf("failed to prepare operation for bridge creation: %v", err)
	}
	ovsMutateOps, err := c.client.Where(c.rootObj).Mutate(c.rootObj, model.Mutation{
		Field:   &c.rootObj.Bridges,
		Mutator: ovsdb.MutateOperationInsert,
		Value:   []string{br.UUID},
	})
	if err != nil {
		return fmt.Errorf("failed to create mutate operation for Open_vSwitch table: %v", err)
	}
	if err := c.execTransaction(ctx, brCreateOps, ovsMutateOps); err != nil {
		return fmt.Errorf("bridge creation failed: %v", err)
	}
	return nil
}

// add interface with provided configuration to the provided bridge
// and check that interface has no error for the next 2 seconds
func (c *ovsManager) addInterface(ctx context.Context, br *BridgeEntry, iface *InterfaceEntry) error {
	addInterfaceOPs, err := c.client.Create(iface)
	if err != nil {
		return fmt.Errorf("failed to prepare operation for interface creation: %v", err)
	}
	port := &PortEntry{Name: iface.Name, UUID: uuid.NewString(), Interfaces: []string{iface.UUID}}
	addPortOPs, err := c.client.Create(port)
	if err != nil {
		return fmt.Errorf("failed to prepare operation for port creation: %v", err)
	}
	bridgeMutateOps, err := c.client.Where(br).Mutate(br, model.Mutation{
		Field:   &br.Ports,
		Mutator: ovsdb.MutateOperationInsert,
		Value:   []string{port.UUID},
	})
	if err != nil {
		return fmt.Errorf("failed to prepare operation for bridge mutate: %v", err)
	}
	if err := c.execTransaction(ctx, addInterfaceOPs, addPortOPs, bridgeMutateOps); err != nil {
		return fmt.Errorf("bridge deletion failed: %v", err)
	}
	// check after ~2 seconds that interface has no error
	for i := 0; i < 2; i++ {
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
		}
		if err := c.client.Get(ctx, iface); err != nil {
			return fmt.Errorf("failed to read interface after creation: %v", err)
		}
		if iface.Error != nil {
			return fmt.Errorf("created interface is in error state: %s", *iface.Error)
		}
	}
	return nil
}

// delete bridge by the name
func (c *ovsManager) deleteBridge(ctx context.Context, brName string) error {
	br, err := c.getBridgeByName(ctx, brName)
	if err != nil {
		return err
	}
	if br == nil {
		return nil
	}
	brDeleteOps, err := c.client.Where(br).Delete()
	if err != nil {
		return fmt.Errorf("failed to prepare operation for bridge deletion: %v", err)
	}
	ovsMutateOps, err := c.client.Where(c.rootObj).Mutate(c.rootObj, model.Mutation{
		Field:   &c.rootObj.Bridges,
		Mutator: ovsdb.MutateOperationDelete,
		Value:   []string{br.UUID},
	})
	if err != nil {
		return fmt.Errorf("failed to create mutate operation for Open_vSwitch table: %v", err)
	}
	if err := c.execTransaction(ctx, brDeleteOps, ovsMutateOps); err != nil {
		return fmt.Errorf("bridge deletion failed: %v", err)
	}
	return nil
}

// delete interface by the name
func (c *ovsManager) deleteInterface(ctx context.Context, ifaceName string) error {
	var operations [][]ovsdb.Operation
	iface, err := c.getInterfaceByName(ctx, ifaceName)
	if err != nil {
		return err
	}
	if iface == nil {
		return nil
	}
	delIfaceOPs, err := c.client.Where(iface).Delete()
	if err != nil {
		return fmt.Errorf("failed to prepare operation for interface deletion: %v", err)
	}
	operations = append(operations, delIfaceOPs)

	port, err := c.getPortByInterface(ctx, iface)
	if err != nil {
		return err
	}
	if port != nil {
		delPortOPs, err := c.client.Where(port).Delete()
		if err != nil {
			return fmt.Errorf("failed to prepare operation for port deletion: %v", err)
		}
		operations = append(operations, delPortOPs)

		bridge, err := c.getBridgeByPort(ctx, port)
		if err != nil {
			return err
		}
		if bridge != nil {
			bridgeMutateOps, err := c.client.Where(bridge).Mutate(bridge, model.Mutation{
				Field:   &bridge.Ports,
				Mutator: ovsdb.MutateOperationDelete,
				Value:   []string{port.UUID},
			})
			if err != nil {
				return fmt.Errorf("failed to prepare operation for bridge mutate: %v", err)
			}
			operations = append(operations, bridgeMutateOps)
		}
	}
	if err := c.execTransaction(ctx, operations...); err != nil {
		return fmt.Errorf("failed to remove interface %s: %v", iface.Name, err)
	}
	return nil
}

// execute multiple prepared OVSDB operations as a single transaction
func (c *ovsManager) execTransaction(ctx context.Context, ops ...[]ovsdb.Operation) error {
	var operations []ovsdb.Operation
	for _, o := range ops {
		operations = append(operations, o...)
	}
	result, err := c.client.Transact(ctx, operations...)
	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}
	operationsErr, err := ovsdb.CheckOperationResults(result, operations)
	if err != nil || len(operationsErr) > 0 {
		return fmt.Errorf("operation failed: %v, %v", err, operationsErr)
	}
	return nil
}

// return current state of the bridge and of the uplink interface.
// uses knownConfig to check which fields are managed by the operator (other fields can be updated OVS itself or by other programs,
// we should not take them into account)
func (c *ovsManager) getCurrentBridgeState(ctx context.Context, knownConfig *sriovnetworkv1.OVSConfigExt) (*sriovnetworkv1.OVSConfigExt, error) {
	bridge, err := c.getBridgeByName(ctx, knownConfig.Name)
	if err != nil {
		return nil, err
	}
	if bridge == nil {
		return nil, nil
	}
	currentConfig := &sriovnetworkv1.OVSConfigExt{
		Name: bridge.Name,
		Bridge: sriovnetworkv1.OVSBridgeConfig{
			DatapathType: bridge.DatapathType,
			// for ExternalIDs and OtherConfig maps we take into account only field which
			// were set by the operator
			ExternalIDs: updateMap(knownConfig.Bridge.ExternalIDs, bridge.ExternalIDs),
			OtherConfig: updateMap(knownConfig.Bridge.OtherConfig, bridge.ExternalIDs),
		},
	}
	if len(knownConfig.Uplinks) == 0 {
		return currentConfig, nil
	}
	knownConfigUplink := knownConfig.Uplinks[0]
	iface, err := c.getInterfaceByName(ctx, knownConfigUplink.Name)
	if err != nil {
		return nil, err
	}
	if iface == nil {
		return currentConfig, nil
	}

	if iface.Error != nil {
		// interface has an error, do not report info about it to let the operator try to recreate it
		return currentConfig, nil
	}

	port, err := c.getPortByInterface(ctx, iface)
	if err != nil {
		return nil, err
	}
	if port == nil {
		return currentConfig, nil
	}

	if !bridge.HasPort(port.UUID) {
		// interface belongs to a wrong bridge, do not include uplink config to
		// the current bridge state to let the operator try to fix this
		return currentConfig, nil
	}
	currentConfig.Uplinks = []sriovnetworkv1.OVSUplinkConfigExt{{
		PciAddress: knownConfigUplink.PciAddress,
		Name:       knownConfigUplink.Name,
		Interface: sriovnetworkv1.OVSInterfaceConfig{
			Type:        iface.Type,
			ExternalIDs: updateMap(knownConfigUplink.Interface.ExternalIDs, iface.ExternalIDs),
			Options:     updateMap(knownConfigUplink.Interface.Options, iface.Options),
			OtherConfig: updateMap(knownConfigUplink.Interface.OtherConfig, iface.OtherConfig),
		},
	}}
	return currentConfig, nil
}

// resulting map contains keys from the old map with values from the new map.
// if key from the old map not found in the new map it will not be added to resulting map
func updateMap(old, new map[string]string) map[string]string {
	result := map[string]string{}
	for k := range old {
		val, found := new[k]
		if found {
			result[k] = val
		}
	}
	return result
}
