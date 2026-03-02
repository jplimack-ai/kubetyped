package disabled_check

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// All checks are disabled — nothing here should produce diagnostics.

// map_literal trigger.
var manifest = map[string]any{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
}

// sprintf_yaml trigger.
var yaml = fmt.Sprintf("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: %s", "test")

// raw_gvk_string trigger: TypeMeta with raw strings.
var tm = metav1.TypeMeta{
	Kind:       "Deployment",
	APIVersion: "apps/v1",
}

// raw_gvk_string trigger: GVK with raw Kind.
var gvk = schema.GroupVersionKind{
	Group:   "apps",
	Version: "v1",
	Kind:    "Deployment",
}

// unstructured_gvk trigger: SetGroupVersionKind on Unstructured.
func setGVK() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
}
