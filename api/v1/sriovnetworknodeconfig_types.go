package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SriovNetworkNodeConfigSpec defines the desired state of SriovNetworkNodeConfig
type SriovNetworkNodeConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SriovNetworkNodeConfig. Edit sriovnetworknodeconfig_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// SriovNetworkNodeConfigStatus defines the observed state of SriovNetworkNodeConfig
type SriovNetworkNodeConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
