package only_unstructured

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Uses consts so raw_gvk_string does NOT fire (it only fires on raw string literals).
const (
	group   = "apps"
	version = "v1"
	kind    = "Deployment"
)

// Triggers only unstructured_gvk — const fields avoid raw_gvk_string.
func f() {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{Group: group, Version: version, Kind: kind})
}
