package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SriovNetworkNodeConfigSpec defines the desired state of SriovNetworkNodeConfig
type SriovNetworkNodeConfigSpec struct {
	// RDMAExclusiveMode enables RDMA exclusive mode for the selected nodes
	RDMAExclusiveMode []RDMAExclusiveModeConfig `json:"rdmaExclusiveMode,omitempty"`
	// OvsHardwareOffload describes the OVS HWOL configuration for selected Nodes
	OvsHardwareOffload []OvsHardwareOffloadConfig `json:"ovsHardwareOffload,omitempty"`
}

type RDMAExclusiveModeConfig struct {
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type OvsHardwareOffloadConfig struct {
	// On Kubernetes:
	// NodeSelector selects Kubernetes Nodes to be configured with OVS HWOL configurations
	// OVS HWOL configurations are generated automatically by Operator
	// Labels in NodeSelector are ANDed when selecting Kubernetes Nodes
	// On OpenShift:
	// NodeSelector matches on Labels defined in MachineConfigPoolSpec.NodeSelector
	// OVS HWOL MachineConfigs are generated and applied to Nodes in MachineConfigPool
	// Labels in NodeSelector are ANDed when matching on MachineConfigPoolSpec.NodeSelector
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

type RDMAExclusiveModeConfigStatus struct {
	Nodes []string `json:"nodes,omitempty"`
}

type OvsHardwareOffloadConfigStatus struct {
	// On Kubernetes:
	// Nodes shows the selected names of Kubernetes Nodes that are configured with OVS HWOL
	// On OpenShift:
	// Nodes shows the selected names of MachineConfigPools that are configured with OVS HWOL
	Nodes []string `json:"nodes,omitempty"`
}

// SriovNetworkNodeConfigStatus defines the observed state of SriovNetworkNodeConfig
type SriovNetworkNodeConfigStatus struct {
	// Show the runtime status of OvsHardwareOffload
	OvsHardwareOffload []OvsHardwareOffloadConfigStatus `json:"ovsHardwareOffload,omitempty"`
	RDMAExclusiveMode  []RDMAExclusiveModeConfigStatus  `json:"rdmaExclusiveMode,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SriovNetworkNodeConfig is the Schema for the sriovnetworknodeconfigs API
type SriovNetworkNodeConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SriovNetworkNodeConfigSpec   `json:"spec,omitempty"`
	Status SriovNetworkNodeConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SriovNetworkNodeConfigList contains a list of SriovNetworkNodeConfig
type SriovNetworkNodeConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SriovNetworkNodeConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SriovNetworkNodeConfig{}, &SriovNetworkNodeConfigList{})
}
