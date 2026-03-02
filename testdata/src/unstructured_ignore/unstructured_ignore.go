package unstructured_ignore

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SetGroupVersionKind with ignored GVK: should NOT be flagged.
func ignoredGVK() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
}

// SetAPIVersion + SetKind with ignored GVK: should NOT be flagged.
func ignoredPair() {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("apps/v1")
	u.SetKind("Deployment")
}

// Non-ignored GVK: should be flagged.
func notIgnored() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{ // want `SetGroupVersionKind\(apiVersion="v1", kind="Pod"\) on unstructured.Unstructured: use \*corev1\.Pod \(import "k8s\.io/api/core/v1"\) instead`
		Version: "v1",
		Kind:    "Pod", // want `raw string "Pod" for GroupVersionKind\.Kind; use \*corev1\.Pod \(import "k8s\.io/api/core/v1"\) or define a const`
	})
}
