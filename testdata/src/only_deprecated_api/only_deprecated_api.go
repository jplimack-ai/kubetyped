package only_deprecated_api

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Uses const values to avoid triggering raw_gvk_string.
// This fixture isolates the deprecated_api check only.
// When deprecated_api is disabled, no diagnostics should fire.

const (
	apiVersion = "apps/v1beta1"
	kind       = "Deployment"
)

// TypeMeta with const values: raw_gvk_string does NOT fire, deprecated_api does when enabled.
var tm = metav1.TypeMeta{
	Kind:       kind,
	APIVersion: apiVersion,
}
