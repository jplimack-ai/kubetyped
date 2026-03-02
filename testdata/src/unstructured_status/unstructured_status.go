package unstructured_status

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// --- SetNestedField with "status" path: flagged ---

func setStatusField() {
	obj := map[string]any{}
	unstructured.SetNestedField(obj, "Running", "status", "phase") // want `unstructured\.SetNestedField targeting "status" path; use typed status subresource updates instead`
}

// --- SetNestedSlice with "status" path: flagged ---

func setStatusSlice() {
	obj := map[string]any{}
	unstructured.SetNestedSlice(obj, []any{"cond1"}, "status", "conditions") // want `unstructured\.SetNestedField targeting "status" path; use typed status subresource updates instead`
}

// --- u.Object["status"] access: flagged ---

func readStatus() {
	u := &unstructured.Unstructured{}
	_ = u.Object["status"] // want `direct map access to Unstructured\.Object\["status"\]; use typed status subresource updates instead`
}

// --- SetNestedField with non-status path: NOT flagged ---

func setMetadataField() {
	obj := map[string]any{}
	unstructured.SetNestedField(obj, "test", "metadata", "name")
}

// --- SetNestedField with "spec" path: NOT flagged ---

func setSpecField() {
	obj := map[string]any{}
	unstructured.SetNestedField(obj, 3, "spec", "replicas")
}

// --- u.Object["metadata"] access: NOT flagged ---

func readMetadata() {
	u := &unstructured.Unstructured{}
	_ = u.Object["metadata"]
}

// --- Non-Unstructured Object["status"]: NOT flagged ---

type myObj struct {
	Object map[string]any
}

func readNonUnstructured() {
	o := myObj{Object: map[string]any{}}
	_ = o.Object["status"]
}
