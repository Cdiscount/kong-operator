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
type KongServiceSpec struct {
	Protocol       string `json:"protocol"`
	Path           string `json:"path"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Retries        int    `json:"retries"`
	ConnectTimeout int    `json:"connectTimeout"`
	WriteTimeout   int    `json:"writeTimeout"`
	ReadTimeout    int    `json:"readTimeout"`
}

// KongServiceStatus represent kong service status
type KongServiceStatus struct {
	KongStatus   string `json:"kongStatus"`
	KongID       string `json:"kongId"`
	URL          string `json:"url"`
	CreationDate string `json:"createdAt"`
	UpdateDate   string `json:"updatedAt"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KongRoute describe a kong route
type KongRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KongRouteSpec   `json:"spec"`
	Status KongRouteStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KongRouteList list of KongRoute
type KongRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []KongRoute `json:"items"`
}

// KongRouteSpec represent kong route spec
type KongRouteSpec struct {
	ServiceName  string   `json:"service"`
	Protocols    []string `json:"protocols"`
	Methods      []string `json:"methods"`
	Hosts        []string `json:"hosts"`
	Paths        []string `json:"paths"`
	StripPath    bool     `json:"stripPath"`
	PreserveHost bool     `json:"preserveHost"`
}

// KongRouteStatus represent kong route status
type KongRouteStatus struct {
	KongStatus   string `json:"kongStatus"`
	KongID       string `json:"kongId"`
	ServiceRefID string `json:"serviceRefId"`
	CreationDate string `json:"createdAt"`
	UpdateDate   string `json:"updatedAt"`
}
