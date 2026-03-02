package condition_map

// --- map[string]any with "type" + "status": flagged ---

var condMap = map[string]any{ // want `map\[string\]any with "type" and "status" keys constructs a Kubernetes status condition; use metav1\.Condition instead`
	"type":   "Ready",
	"status": "True",
}

// --- With extra keys: still flagged ---

var condMapExtra = map[string]any{ // want `map\[string\]any with "type" and "status" keys constructs a Kubernetes status condition; use metav1\.Condition instead`
	"type":    "Available",
	"status":  "False",
	"message": "not available",
}

// --- Missing "status": NOT flagged ---

var missingStatus = map[string]any{
	"type": "Ready",
	"name": "test",
}

// --- Missing "type": NOT flagged ---

var missingType = map[string]any{
	"status": "True",
	"name":   "test",
}

// --- Wrong map type (map[string]string): NOT flagged ---

var wrongMapType = map[string]string{
	"type":   "Ready",
	"status": "True",
}

// --- Has "apiVersion" + "kind" — this is a manifest, not a condition.
// The condition_map_literal check only looks for "type" + "status", so it won't fire.
// The map_literal check does fire on this (separate check).
var manifestNotCondition = map[string]any{ // want `use \*corev1\.Pod .* instead of map\[string\]any for v1/Pod`
	"apiVersion": "v1",
	"kind":       "Pod",
}
