package unstructured_gvk

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SetGroupVersionKind with known GVK literal: should be flagged.
func knownGVK() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{ // want `SetGroupVersionKind\(apiVersion="apps/v1", kind="Deployment"\) on unstructured.Unstructured: use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead`
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment", // want `raw string "Deployment" for GroupVersionKind\.Kind; use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) or define a const`
	})
}

// SetGroupVersionKind with core API (no group): should be flagged.
func coreGVK() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{ // want `SetGroupVersionKind\(apiVersion="v1", kind="Pod"\) on unstructured.Unstructured: use \*corev1\.Pod \(import "k8s\.io/api/core/v1"\) instead`
		Version: "v1",
		Kind:    "Pod", // want `raw string "Pod" for GroupVersionKind\.Kind; use \*corev1\.Pod \(import "k8s\.io/api/core/v1"\) or define a const`
	})
}

// SetGroupVersionKind with unknown GVK: should NOT be flagged.
func unknownGVK() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "example.io",
		Version: "v1",
		Kind:    "Widget", // want `raw string "Widget" for GroupVersionKind\.Kind; define a package-level const`
	})
}

// SetGroupVersionKind with variable arg: should NOT be flagged.
func variableGVK() {
	u := &unstructured.Unstructured{}
	gvk := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment", // want `raw string "Deployment" for GroupVersionKind\.Kind; use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) or define a const`
	}
	u.SetGroupVersionKind(gvk)
}

// SetAPIVersion + SetKind pair with known GVK: should be flagged.
func setPair() {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("apps/v1") // want `SetAPIVersion\("apps/v1"\) \+ SetKind\("Deployment"\) on unstructured\.Unstructured: use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead`
	u.SetKind("Deployment")
}

// Only SetAPIVersion, no SetKind: should NOT be flagged.
func setAPIVersionOnly() {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("apps/v1")
}

// Only SetKind, no SetAPIVersion: should NOT be flagged.
func setKindOnly() {
	u := &unstructured.Unstructured{}
	u.SetKind("Deployment")
}

// SetAPIVersion + SetKind with unknown GVK: should NOT be flagged.
func setPairUnknown() {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("example.io/v1")
	u.SetKind("Widget")
}

// Value receiver (non-pointer): should still be flagged.
func valueRecv() {
	var u unstructured.Unstructured
	u.SetGroupVersionKind(schema.GroupVersionKind{ // want `SetGroupVersionKind\(apiVersion="apps/v1", kind="Deployment"\) on unstructured.Unstructured: use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead`
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment", // want `raw string "Deployment" for GroupVersionKind\.Kind; use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) or define a const`
	})
}

// Reversed pair order: SetKind before SetAPIVersion, diagnostic at min pos.
func setPairReversed() {
	u := &unstructured.Unstructured{}
	u.SetKind("Deployment")      // want `SetAPIVersion\("apps/v1"\) \+ SetKind\("Deployment"\) on unstructured\.Unstructured: use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead`
	u.SetAPIVersion("apps/v1")
}

// Dynamic variable as argument: should NOT be flagged (can't resolve at compile time).
func setPairDynamic(apiVersion string) {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion(apiVersion)
	u.SetKind("Widget")
}
