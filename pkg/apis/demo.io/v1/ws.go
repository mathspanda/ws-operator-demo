package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WebServerCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   WebServerClusterSpec   `json:"spec"`
	Status WebServerClusterStatus `json:"status"`
}

type WebServerClusterSpec struct {
	Replicas    *int32 `json:"replicas"`
	Image       string `json:"image"`
	ServicePort int32  `json:"port"`
}

type WebServerClusterStatus struct {
	Replicas int32 `json:"replicas"`
}

type WebServerClusterList struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Items []WebServerCluster `json:"items"`
}
