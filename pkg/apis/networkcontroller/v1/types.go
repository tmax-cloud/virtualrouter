/*
Copyright 2017 The Kubernetes Authors.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NATRule is a specification for a NATRule resource
type NATRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NATRuleSpec   `json:"spec"`
	Status NATRuleStatus `json:"status"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FireWallRule is a specification for a FireWallRule resource
type FireWallRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FireWallRuleSpec   `json:"spec"`
	Status FireWallRuleStatus `json:"status"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalanceRule is a specification for a LoadBalanceRule resource
type LoadBalanceRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadBalanceRuleSpec   `json:"spec"`
	Status LoadBalanceRuleStatus `json:"status"`
}

// NATRuleSpec is the spec for a NATRule resource
type NATRuleSpec struct {
	Rules []Rules `json:"rules"`
}

// NATRuleStatus is the status for a NATRule resource
type NATRuleStatus struct {
	Deployed string `json:"deployed"`

	OldSrcIP string
	OldDstIP string
}

// FireWallRuleSpec is the spec for a FireWallRule resource
type FireWallRuleSpec struct {
	Rules []Rules `json:"rules"`
}

// FireWallRuleSpec is the status for a FireWallRule resource
type FireWallRuleStatus struct {
	Deployed string `json:"deployed"`

	OldSrcIP string
	OldDstIP string
}

// LoadBalanceRuleSpec is the spec for a LoadBalanceRule resource
type LoadBalanceRuleSpec struct {
	Rules []Rules `json:"rules"`
}

// LoadBalanceRuleSpec is the spec for a LoadBalanceRule resource
type LoadBalanceRuleStatus struct {
	Deployed string `json:"deployed"`

	OldSrcIP string
	OldDstIP string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NATRuleList is a list of NATRule resources
type NATRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NATRule `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FireWallRuleList is a list of FireWallRule resources
type FireWallRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []FireWallRule `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalanceRuleList is a list of LoadBalanceRule resources
type LoadBalanceRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []LoadBalanceRule `json:"items"`
}

type Rules struct {
	Match  Match  `json:"match"`
	Action Action `json:"action"`
}

type Match struct {
	SrcIP    string `json:"srcIP"`
	DstIP    string `json:"dstIP"`
	Protocol string `json:"protocol"`
}

type Action struct {
	SrcIP string `json:"srcIP"`
	DstIP string `json:"dstIP"`
}
