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

// LoadBalancerRule is a specification for a LoadBalancerRule resource
type LoadBalancerRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadBalancerRuleSpec   `json:"spec"`
	Status LoadBalancerRuleStatus `json:"status"`
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
}

// LoadBalancerRuleSpec is the spec for a LoadBalancerRule resource
type LoadBalancerRuleSpec struct {
	Rules []LBRules `json:"rules"`
}

// LoadBalancerRuleSpec is the spec for a LoadBalancerRule resource
type LoadBalancerRuleStatus struct {
	Deployed string `json:"deployed"`
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

// LoadBalancerRuleList is a list of LoadBalancerRule resources
type LoadBalancerRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []LoadBalancerRule `json:"items"`
}

type Rules struct {
	Match  Match  `json:"match"`
	Action Action `json:"action"`
	Args   []string
}

type LBRules struct {
	LoadBalancerIP   string    `json:"loadBalancerIP"`
	LoadBalancerPort int       `json:"loadBalancerPort"`
	Protocol         string    `json:"Protocol"`
	Backends         []Backend `json:"backends"`
}

type Backend struct {
	BackendIP         string `json:"backendIP"`
	BackendPort       int    `json:"backendPort"`
	Weight            int    `json:"weight"`
	HealthCheckIP     string `json:"healthcheckIP"`
	HealthCheckPort   int    `json:"healthcheckPort"`
	HealthCheckMethod string `json:"healthcheckMethod"`
}

type Match struct {
	SrcIP    string `json:"srcIP"`
	DstIP    string `json:"dstIP"`
	SrcPort  int    `json:"srcPort"`
	DstPort  int    `json:"dstPort"`
	Protocol string `json:"protocol"`
}

type Action struct {
	SrcIP   string `json:"srcIP"`
	DstIP   string `json:"dstIP"`
	SrcPort int    `json:"srcPort"`
	DstPort int    `json:"dstPort"`
	Policy  string `json:"policy"`
}
