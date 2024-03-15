/*
Copyright 2024 nineinfra.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceConfig struct {
	// The replicas of the cluster workload.Default value is 1
	// +optional
	Replicas int32 `json:"replicas"`
	// num of the disks. default value is 1
	// +optional
	Disks int32 `json:"disks"`
	// the storage class. default value is nineinfra-default
	// +optional
	StorageClass string `json:"storageClass"`
	// The resource requirements of the cluster workload.
	// +optional
	ResourceRequirements corev1.ResourceRequirements `json:"resourceRequirements"`
}

type ImageConfig struct {
	Repository string `json:"repository"`
	// Image tag. Usually the vesion of the cluster, default: `latest`.
	// +optional
	Tag string `json:"tag,omitempty"`
	// Image pull policy. One of `Always, Never, IfNotPresent`, default: `Always`.
	// +kubebuilder:default:=Always
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	// +optional
	PullPolicy string `json:"pullPolicy,omitempty"`
	// Secrets for image pull.
	// +optional
	PullSecrets string `json:"pullSecret,omitempty"`
}

// KafkaClusterSpec defines the desired state of KafkaCluster
type KafkaClusterSpec struct {
	// Version. version of the cluster.
	Version string `json:"version"`
	// Image. image config of the cluster.
	Image ImageConfig `json:"image"`
	// Resource. resouce config of the cluster.
	// +optional
	Resource ResourceConfig `json:"resource,omitempty"`
	// Conf. k/v configs for the server.properties.
	// +optional
	Conf map[string]string `json:"conf,omitempty"`
	// K8sConf. k/v configs for the cluster in k8s.such as the cluster domain
	// +optional
	K8sConf map[string]string `json:"k8sConf,omitempty"`
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KafkaCluster is the Schema for the kafkaclusters API
type KafkaCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaClusterSpec   `json:"spec,omitempty"`
	Status KafkaClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaClusterList contains a list of KafkaCluster
type KafkaClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KafkaCluster{}, &KafkaClusterList{})
}
