package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VPNSpec defines the desired state of VPN
type VPNSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	Connections []Connection `json:"connections"`
	NodeName    string       `json:"nodename,omitempty"`
}

type Connection struct {
	Name        string `json:"name"`
	Left        string `json:"left"`
	LeftSubnet  string `json:"leftsubnet"`
	Right       string `json:"right"`
	RightSubnet string `json:"rightsubnet"`
	PSK         string `json:"psk"`

	LeftID          string `json:"leftid,omitempty"`
	RightID         string `json:"rightid,omitempty"`
	IsServiceSubnet bool   `json:"isservicesubnet,omitempty"`
	Keyexchange     string `json:"keyexchange,omitempty"`
	IKELifetime     string `json:"ikelifetime,omitempty"`
	Lifetime        string `json:"lifetime,omitempty"`
	IKE             string `json:"ike,omitempty"`
	ESP             string `json:"esp,omitempty"`
	Also            string `json:"also,omitempty"`
}

// VPNStatus defines the observed state of VPN
type VPNStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VPN is the Schema for the vpns API
type VPN struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VPNSpec   `json:"spec,omitempty"`
	Status VPNStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VPNList contains a list of VPN
type VPNList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VPN `json:"items"`
}
