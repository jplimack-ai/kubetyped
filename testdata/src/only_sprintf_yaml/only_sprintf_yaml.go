package only_sprintf_yaml

import "fmt"

// Triggers only sprintf_yaml — no map literal, no TypeMeta/GVK structs, no unstructured.
var _ = fmt.Sprintf("apiVersion: apps/v1\nkind: Deployment\n")
