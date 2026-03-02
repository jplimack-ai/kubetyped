package map_literal_const

const (
	deployAPIVersion = "apps/v1"
	deployKind       = "Deployment"
)

// Map literal with const values: should be flagged.
var deployment = map[string]any{ // want `use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead of map\[string\]any for apps/v1/Deployment`
	"apiVersion": deployAPIVersion,
	"kind":       deployKind,
}

// Mixed: one const, one literal.
var mixed = map[string]any{ // want `use \*corev1\.Pod \(import "k8s\.io/api/core/v1"\) instead of map\[string\]any for v1/Pod`
	"apiVersion": "v1",
	"kind":       podKind,
}

const podKind = "Pod"

// Non-const variable: should NOT be flagged (can't resolve at compile time).
var dynamicVersion = "apps/v1"

var notFlagged = map[string]any{
	"apiVersion": dynamicVersion,
	"kind":       "Deployment",
}

// Const map key: should be flagged after C0 fix resolves const keys.
const apiVersionKey = "apiVersion"

var constKeyMap = map[string]any{ // want `use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead of map\[string\]any for apps/v1/Deployment`
	apiVersionKey: "apps/v1",
	"kind":        "Deployment",
}
