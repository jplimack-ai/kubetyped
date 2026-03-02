package only_unstructured_status

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Triggers only unstructured_status — no GVK patterns.
func f() {
	obj := map[string]any{}
	unstructured.SetNestedField(obj, "Running", "status", "phase")
}
