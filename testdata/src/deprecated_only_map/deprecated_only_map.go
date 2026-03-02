package deprecated_only_map

// map_literal is disabled but deprecated_api is enabled.
// Only the deprecated_api diagnostic should fire, not map_literal.
var manifest = map[string]any{ // want `deprecated API version apps/v1beta1 for Deployment \(removed in Kubernetes 1\.16\); use apps/v1 instead`
	"apiVersion": "apps/v1beta1",
	"kind":       "Deployment",
}

// Non-deprecated: no diagnostics at all (map_literal disabled).
var stable = map[string]any{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
}
