/*
Copyright 2023.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Egressgwk3sSpec defines the desired state of Egressgwk3s
type Egressgwk3sSpec struct {
	GwNode     string               `json:"gwnode,omitempty"`
	SourcePods []SourcePodsSelector `json:"sourcepods,omitempty"`
}

type SourcePodsSelector struct {
	NamespaceSelector metav1.LabelSelector `json:"namespaceselector,omitempty"`
	PodSelector       metav1.LabelSelector `json:"podselector,omitempty"`
	IpBlock           *IPBlock             `json:"ipblock,omitempty"`
}

type IPBlock struct {
	CIDR   string   `json:"cidr,omitempty"`
	Except []string `json:"except,omitempty"`
}

// Egressgwk3sStatus defines the observed state of Egressgwk3s
type Egressgwk3sStatus struct {
	Pods []string `json:"pods,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.phase", name=Phase, type=string
// Egressgwk3s is the Schema for the egressgwk3s API
type Egressgwk3s struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Egressgwk3sSpec   `json:"spec,omitempty"`
	Status Egressgwk3sStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// Egressgwk3sList contains a list of Egressgwk3s
type Egressgwk3sList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Egressgwk3s `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Egressgwk3s{}, &Egressgwk3sList{})
}
