package v1

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	CRDKind    = "WebServerCluster"
	CRDPlural  = strings.ToLower(CRDKind) + "s"
	CRDGroup   = "demo.io"
	CRDVersion = "v1"
	CRDName    = CRDPlural + "." + CRDGroup
)

var SchemeGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}

func AddKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&WebServerCluster{},
		&WebServerClusterList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)

	return nil
}
