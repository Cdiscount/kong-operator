package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KongService describe a kong service
type KongService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KongServiceSpec   `json:"spec"`
	Status KongServiceStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KongServiceList list of KongService
type KongServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []KongService `json:"items"`
}

// KongServiceSpec represent kong service spec
type KongServiceSpec struct{}

// KongServiceStatus represent kong service status
type KongServiceStatus struct{}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KongRoute describe a kong route
type KongRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KongServiceSpec   `json:"spec"`
	Status KongServiceStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KongRouteList list of KongRoute
type KongRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []KongRoute `json:"items"`
}

// KongRouteSpec represent kong route spec
type KongRouteSpec struct{}

// KongRouteStatus represent kong route status
type KongRouteStatus struct{}
