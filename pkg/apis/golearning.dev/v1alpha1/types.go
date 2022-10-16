package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ClusterID",type=string,JSONPath=`.status.klusterID`
// +kubebuilder:printcolumn:name="Progress",type=string,JSONPath=`.status.progress`
type Kluster struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec klusterSpec
}

type klusterSpec struct {
	Name    string
	Region  string
	Version string

	NodePools []NodePool
}

type NodePool struct {
	Size  string
	Name  string
	Count int
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type KlusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Kluster `json:"items,omitempty"`
}
