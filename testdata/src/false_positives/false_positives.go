package false_positives

import "fmt"

// Labels map: NOT a manifest, should not be flagged.
var labels = map[string]string{
	"app":     "my-app",
	"version": "v1",
}

// Generic map: missing "kind" key, should not be flagged.
var partial = map[string]any{
	"apiVersion": "v1",
	"data":       "something",
}

// Map with only "kind": missing "apiVersion", should not be flagged.
var kindOnly = map[string]any{
	"kind": "Pod",
	"spec": "something",
}

// Map with non-literal values: should not be flagged.
var dynamic = map[string]any{
	"apiVersion": getVersion(),
	"kind":       getKind(),
}

func getVersion() string { return "v1" }
func getKind() string    { return "Pod" }

// Sprintf without YAML markers: should not be flagged.
var safeFmt = fmt.Sprintf("deploying %s to %s", "app", "cluster")

// Map[string]int: wrong value type, should not be flagged.
var intMap = map[string]int{
	"apiVersion": 1,
	"kind":       2,
}

// Sprintf with "kind:" in non-YAML context: we do flag this (documented as known behavior).
// If you have a legitimate use of "kind:" in a non-YAML string, use //nolint:kube-types.

// Empty map literal: NOT flagged.
var emptyMap = map[string]any{}

// Int const as map value: NOT flagged (const is not a string).
const deployPort = 8080

var intConstValue = map[string]any{
	"apiVersion": deployPort,
	"kind":       "Pod",
}

// Variable map key: NOT flagged (key can't be resolved at compile time).
var keyName = "apiVersion"

var varKeyMap = map[string]any{
	keyName: "v1",
	"kind":  "Pod",
}

// fmt.Errorf with YAML marker: NOT flagged (only Sprintf/Fprintf checked).
var _ = fmt.Errorf("apiVersion: v1\nkind: Pod\n")

// Two-step map construction: NOT flagged (known limitation).
var twoStep = make(map[string]any)

func buildManifest() {
	twoStep["apiVersion"] = "v1"
	twoStep["kind"] = "Pod"
}
