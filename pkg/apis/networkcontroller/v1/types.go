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

// NATRuleSpec is the spec for a NATRule resource
type NATRuleSpec struct {
	Rules []Rules
}

type Rules struct {
	Match  Match  `json:"match"`
	Action Action `json:"action"`
}

type Match struct {
	SrcIP   string `json:"srcIP"`
	DstIP   string `json:"dstIP"`
	Protocl string `json:"protocol"`
}

type Action struct {
	SrcIP string `json:"srcIP"`
	DstIP string `json:"dstIP"`
}

// NATRuleStatus is the status for a NATRule resource
type NATRuleStatus struct {
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
