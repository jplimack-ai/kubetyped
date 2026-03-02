package reject_gvks

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// map_literal with rejected GVK: should fire rejection.
var manifest = map[string]any{ // want `construction of v1/Pod is rejected by project policy \(reject_gvks\)`
	"apiVersion": "v1",
	"kind":       "Pod",
}

// TypeMeta with rejected GVK.
var tm = metav1.TypeMeta{ // want `construction of v1/Pod is rejected by project policy \(reject_gvks\)`
	Kind:       "Pod",
	APIVersion: "v1",
}

// GroupVersionKind with rejected GVK.
var gvk = schema.GroupVersionKind{ // want `construction of v1/Pod is rejected by project policy \(reject_gvks\)`
	Version: "v1",
	Kind:    "Pod",
}

// SetGroupVersionKind with rejected GVK.
// checkUnstructuredGVKExpr fires at call pos; inner GVK literal is deduplicated.
func setGVK() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind( // want `construction of v1/Pod is rejected by project policy \(reject_gvks\)`
		schema.GroupVersionKind{
			Version: "v1",
			Kind:    "Pod",
		},
	)
}

// SetAPIVersion + SetKind with rejected GVK.
func setPair() {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("v1") // want `construction of v1/Pod is rejected by project policy \(reject_gvks\)`
	u.SetKind("Pod")
}

// Non-rejected GVK: should produce normal diagnostics.
var deployment = map[string]any{ // want `use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead of map\[string\]any for apps/v1/Deployment`
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
}
