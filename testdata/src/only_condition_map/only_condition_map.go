package only_condition_map

// Triggers only condition_map_literal — no apiVersion/kind, no TypeMeta/GVK structs.
var m = map[string]any{
	"type":   "Ready",
	"status": "True",
}
