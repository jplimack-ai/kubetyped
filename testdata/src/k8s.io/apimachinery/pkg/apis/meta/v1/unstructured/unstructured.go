package unstructured

import "k8s.io/apimachinery/pkg/runtime/schema"

// Unstructured is a minimal mock of k8s.io/apimachinery's Unstructured type for testing.
type Unstructured struct {
	Object map[string]any
}

// SetGroupVersionKind is a mock of the real Unstructured.SetGroupVersionKind.
func (u *Unstructured) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	u.Object = map[string]any{
		"apiVersion": gvk.Group + "/" + gvk.Version,
		"kind":       gvk.Kind,
	}
}

// SetAPIVersion is a mock of the real Unstructured.SetAPIVersion.
func (u *Unstructured) SetAPIVersion(apiVersion string) {
	if u.Object == nil {
		u.Object = make(map[string]any)
	}
	u.Object["apiVersion"] = apiVersion
}

// SetKind is a mock of the real Unstructured.SetKind.
func (u *Unstructured) SetKind(kind string) {
	if u.Object == nil {
		u.Object = make(map[string]any)
	}
	u.Object["kind"] = kind
}

// SetNestedField is a mock of the real unstructured.SetNestedField.
func SetNestedField(obj map[string]any, value any, fields ...string) error { return nil }

// SetNestedSlice is a mock of the real unstructured.SetNestedSlice.
func SetNestedSlice(obj map[string]any, value []any, fields ...string) error { return nil }
