package disabled_check

import (
	"fmt"
	"os"

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

// deprecated_api trigger: deprecated API version.
var deprecated = map[string]any{
	"apiVersion": "apps/v1beta1",
	"kind":       "Deployment",
}

// embedded_yaml trigger: os.ReadFile with YAML path.
func readYAML() {
	_, _ = os.ReadFile("deployment.yaml")
}

// raw_condition_status trigger: raw string in Condition.Status.
var cond = metav1.Condition{
	Type:   "Ready",
	Status: "True",
}

// condition_map_literal trigger: map[string]any with "type" + "status".
var condMap = map[string]any{
	"type":   "Ready",
	"status": "True",
}

// unstructured_status trigger: SetNestedField targeting "status".
func setStatus() {
	obj := map[string]any{}
	unstructured.SetNestedField(obj, "Running", "status", "phase")
}

// raw_condition_type trigger: raw string in Condition.Type.
// (cond above also triggers this — both raw_condition_status and raw_condition_type fire on it.)
