package only_map_literal

// Triggers only map_literal — no sprintf, no TypeMeta/GVK structs, no unstructured.
var m = map[string]any{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
}
