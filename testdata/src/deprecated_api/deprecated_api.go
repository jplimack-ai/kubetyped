package deprecated_api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// map_literal with deprecated API: fires both deprecated_api and map_literal (unknown GVK).
var manifest = map[string]any{ // want `deprecated API version apps/v1beta1 for Deployment \(removed in Kubernetes 1\.16\); use apps/v1 instead` `map\[string\]any with apiVersion "apps/v1beta1" and kind "Deployment" constructs a Kubernetes manifest`
	"apiVersion": "apps/v1beta1",
	"kind":       "Deployment",
}

// TypeMeta with deprecated API: fires deprecated_api (at TypeMeta pos) and raw_gvk_string (at field pos).
var tm = metav1.TypeMeta{ // want `deprecated API version apps/v1beta1 for Deployment \(removed in Kubernetes 1\.16\); use apps/v1 instead`
	Kind:       "Deployment",   // want `raw string "Deployment" for TypeMeta\.Kind; define a package-level const`
	APIVersion: "apps/v1beta1", // want `raw string "apps/v1beta1" for TypeMeta\.APIVersion; define a package-level const`
}

// GroupVersionKind with deprecated API: fires deprecated_api (at GVK pos) and raw_gvk_string (at Kind pos).
var gvk = schema.GroupVersionKind{ // want `deprecated API version extensions/v1beta1 for Ingress \(removed in Kubernetes 1\.22\); use networking\.k8s\.io/v1 instead`
	Group:   "extensions",
	Version: "v1beta1",
	Kind:    "Ingress", // want `raw string "Ingress" for GroupVersionKind\.Kind; define a package-level const`
}

// SetGroupVersionKind with deprecated API on unstructured.
// checkUnstructuredGVKExpr fires deprecated_api at call pos.
func setGVK() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind( // want `deprecated API version apps/v1beta2 for StatefulSet \(removed in Kubernetes 1\.16\); use apps/v1 instead`
		schema.GroupVersionKind{
			Group:   "apps",
			Version: "v1beta2",
			Kind:    "StatefulSet", // want `raw string "StatefulSet" for GroupVersionKind\.Kind; define a package-level const`
		},
	)
}

// SetAPIVersion + SetKind pair with deprecated API.
func setPairDeprecated() {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("batch/v1beta1") // want `deprecated API version batch/v1beta1 for CronJob \(removed in Kubernetes 1\.25\); use batch/v1 instead`
	u.SetKind("CronJob")
}

// PodSecurityPolicy: removed entirely.
var psp = map[string]any{ // want `deprecated API version policy/v1beta1 for PodSecurityPolicy \(removed in Kubernetes 1\.25\); this API has been removed entirely` `map\[string\]any with apiVersion "policy/v1beta1" and kind "PodSecurityPolicy" constructs a Kubernetes manifest`
	"apiVersion": "policy/v1beta1",
	"kind":       "PodSecurityPolicy",
}

// Non-deprecated API: should NOT fire deprecated_api but DOES fire map_literal.
var stable = map[string]any{ // want `use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead of map\[string\]any for apps/v1/Deployment`
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
}
