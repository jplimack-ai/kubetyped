# kube-types

[![CI](https://github.com/togethercomputer/kube-types/actions/workflows/ci.yml/badge.svg)](https://github.com/togethercomputer/kube-types/actions/workflows/ci.yml)

A [golangci-lint](https://golangci-lint.run/) v2 module plugin that detects untyped Kubernetes manifest construction and suggests typed Go structs.

## Quick Start

Add to `.golangci.yml`:

```yaml
version: "2"

linters-settings:
  custom:
    kube-types:
      type: "module"
      settings:
        include_test_files: false

linters:
  enable:
    - kube-types
```

Example diagnostics:

```
deployment.go:12:5: use *appsv1.Deployment (import "k8s.io/api/apps/v1") instead of map[string]any for apps/v1/Deployment (kube-types)
deployment.go:20:2: deprecated API version extensions/v1beta1 for Ingress (removed in Kubernetes 1.22); use networking.k8s.io/v1 instead (kube-types)
deployment.go:28:2: os.ReadFile of YAML file "manifests/pod.yaml"; consider using typed Kubernetes structs from k8s.io/api (kube-types)
```

## Why

Kubernetes resources constructed via `map[string]any`, `fmt.Sprintf` YAML templates, or `unstructured.Unstructured` bypass Go's type system entirely. This means:

- No compile-time field validation
- No IDE autocompletion
- Typos in field names are runtime errors (or silent misconfigurations)
- Schema changes break silently

`kube-types` catches these patterns and points you to the typed struct you should use instead.

## Checks

### `map_literal`

Detects `map[string]any` (or `map[string]interface{}`) composite literals containing both `"apiVersion"` and `"kind"` keys, including named type aliases like `type Manifest map[string]any`.

```go
// Flagged:
m := map[string]any{
    "apiVersion": "apps/v1",
    "kind":       "Deployment",
    "metadata":   map[string]any{"name": "my-deploy"},
}
// Diagnostic: use *appsv1.Deployment (import "k8s.io/api/apps/v1") instead of map[string]any for apps/v1/Deployment

// Also supports const values:
const ver = "apps/v1"
m := map[string]any{"apiVersion": ver, "kind": "Deployment"}
```

### `sprintf_yaml`

Detects `fmt.Sprintf` and `fmt.Fprintf` calls where the format string contains YAML markers like `apiVersion:` or `kind:`, suggesting string-interpolated Kubernetes manifest construction. Supports both literal and `const` format strings.

```go
// Flagged:
yaml := fmt.Sprintf("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: %s", name)

// Also flagged (const format string):
const tmpl = "apiVersion: v1\nkind: Pod\nmetadata:\n  name: %s\n"
yaml := fmt.Sprintf(tmpl, name)
```

### `unstructured_gvk`

Detects `SetGroupVersionKind` calls on `*unstructured.Unstructured` with known GVK literals, and `SetAPIVersion` + `SetKind` call pairs on the same receiver.

```go
// Flagged:
u := &unstructured.Unstructured{}
u.SetGroupVersionKind(schema.GroupVersionKind{
    Group: "apps", Version: "v1", Kind: "Deployment",
})
// Diagnostic: SetGroupVersionKind(apiVersion="apps/v1", kind="Deployment") on unstructured.Unstructured:
//   use *appsv1.Deployment (import "k8s.io/api/apps/v1") instead

// Also flagged (SetAPIVersion + SetKind pair):
u.SetAPIVersion("apps/v1")
u.SetKind("Deployment")
```

### `raw_gvk_string`

Detects raw string literals (`"Deployment"`, `"apps/v1"`) in three contexts:

1. **TypeMeta field assignments** -- `metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"}`
2. **GroupVersionKind construction** -- `schema.GroupVersionKind{..., Kind: "Deployment"}`
3. **GVK equality comparisons** -- `obj.GetKind() == "Deployment"` or `gvk.Kind == "Deployment"`

Since `k8s.io/api` does not publish Kind constants, the check encourages defining package-level constants or using scheme-based approaches (`scheme.ObjectKinds()`, `apiutil.GVKForObject()`).

```go
// Flagged (each field independently):
metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"}

// Flagged (raw Kind in GVK literal):
schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}

// Flagged (comparison against raw string):
if obj.GetKind() == "Deployment" { ... }

// NOT flagged (const values are the desired pattern):
const kindDeployment = "Deployment"
metav1.TypeMeta{Kind: kindDeployment, APIVersion: constAPIVersion}
```

### `deprecated_api`

Detects usage of deprecated/removed Kubernetes beta API versions across all GVK construction patterns (map literals, TypeMeta, GroupVersionKind, unstructured). Covers ~25 removed beta APIs including:

- `extensions/v1beta1` (removed 1.22): Ingress, NetworkPolicy, DaemonSet, Deployment, ReplicaSet
- `apps/v1beta1`, `apps/v1beta2` (removed 1.16): Deployment, StatefulSet, DaemonSet, ReplicaSet
- `batch/v1beta1` (removed 1.25): CronJob
- `policy/v1beta1` (removed 1.25): PodDisruptionBudget, PodSecurityPolicy
- `rbac.authorization.k8s.io/v1beta1` (removed 1.22): Role, ClusterRole, RoleBinding, ClusterRoleBinding
- `admissionregistration.k8s.io/v1beta1` (removed 1.22): ValidatingWebhookConfiguration, MutatingWebhookConfiguration
- `autoscaling/v2beta1`, `v2beta2` (removed 1.26): HorizontalPodAutoscaler
- `flowcontrol.apiserver.k8s.io/v1beta1`, `v1beta2` (removed 1.29): FlowSchema, PriorityLevelConfiguration

```go
// Flagged:
m := map[string]any{
    "apiVersion": "apps/v1beta1",
    "kind":       "Deployment",
}
// Diagnostic: deprecated API version apps/v1beta1 for Deployment (removed in Kubernetes 1.16); use apps/v1 instead

// Also flagged in TypeMeta, GroupVersionKind, and unstructured patterns.
```

### `embedded_yaml`

Detects two patterns that load raw YAML Kubernetes manifests at runtime:

1. **`//go:embed` directives** referencing `.yaml` or `.yml` files
2. **`os.ReadFile` / `os.Open` calls** with YAML file path arguments (both literal and `const` strings)

```go
// Flagged (embed directive):
//go:embed deployment.yaml
var deploymentYAML string
// Diagnostic: //go:embed of YAML file "deployment.yaml"; consider using typed Kubernetes structs from k8s.io/api

// Flagged (ReadFile with literal path):
data, _ := os.ReadFile("manifests/pod.yaml")
// Diagnostic: os.ReadFile of YAML file "manifests/pod.yaml"; consider using typed Kubernetes structs from k8s.io/api

// Flagged (ReadFile with const path):
const yamlPath = "service.yaml"
data, _ := os.ReadFile(yamlPath)

// NOT flagged (dynamic path):
data, _ := os.ReadFile(pathVariable)

// NOT flagged (non-YAML files):
data, _ := os.ReadFile("config.json")
```

### `raw_condition_status`

Detects raw string literals `"True"`, `"False"`, `"Unknown"` used in `metav1.Condition` Status fields and comparisons, instead of the typed constants `metav1.ConditionTrue`, `metav1.ConditionFalse`, `metav1.ConditionUnknown`.

```go
// Flagged (composite literal):
cond := metav1.Condition{
    Type:   condTypeReady,
    Status: "True",
}
// Diagnostic: raw string "True" for Condition.Status; use metav1.ConditionTrue instead

// Flagged (comparison):
if cond.Status == "True" { ... }
// Diagnostic: raw string "True" in Condition.Status comparison; use metav1.ConditionTrue instead

// NOT flagged (typed constant):
cond := metav1.Condition{Status: metav1.ConditionTrue}
```

### `raw_condition_type`

Detects raw string literals in the `Type` field of `metav1.Condition` composite literals, encouraging package-level constants for condition type names.

```go
// Flagged:
cond := metav1.Condition{
    Type:   "Ready",
    Status: metav1.ConditionTrue,
}
// Diagnostic: raw string "Ready" for Condition.Type; define a package-level const

// NOT flagged (const value):
const ConditionReady = "Ready"
cond := metav1.Condition{Type: ConditionReady, Status: metav1.ConditionTrue}
```

### `condition_map_literal`

Detects `map[string]any` literals containing both `"type"` and `"status"` keys, which likely construct a Kubernetes status condition without type safety.

```go
// Flagged:
m := map[string]any{
    "type":   "Ready",
    "status": "True",
}
// Diagnostic: map[string]any with "type" and "status" keys constructs a Kubernetes status condition; use metav1.Condition instead

// NOT flagged (missing one key):
m := map[string]any{"type": "Ready", "name": "test"}
```

### `unstructured_status`

Detects untyped status manipulation via `unstructured.SetNestedField` / `SetNestedSlice` targeting `"status"` paths, and direct `u.Object["status"]` map access on `*unstructured.Unstructured`.

```go
// Flagged (SetNestedField):
unstructured.SetNestedField(obj, "Running", "status", "phase")
// Diagnostic: unstructured.SetNestedField targeting "status" path; use typed status subresource updates instead

// Flagged (direct map access):
status := u.Object["status"]
// Diagnostic: direct map access to Unstructured.Object["status"]; use typed status subresource updates instead

// NOT flagged (non-status paths):
unstructured.SetNestedField(obj, "test", "metadata", "name")
_ = u.Object["metadata"]
```

## Built-in GVKs

The plugin ships with ~35 known GVKs covering the most common Kubernetes resources:

- **core/v1**: Pod, Service, ConfigMap, Secret, ServiceAccount, Namespace, PersistentVolume, PersistentVolumeClaim, Node, Endpoints, ResourceQuota, LimitRange
- **apps/v1**: Deployment, StatefulSet, DaemonSet, ReplicaSet
- **batch/v1**: Job, CronJob
- **networking.k8s.io/v1**: Ingress, NetworkPolicy, IngressClass
- **rbac.authorization.k8s.io/v1**: Role, ClusterRole, RoleBinding, ClusterRoleBinding
- **policy/v1**: PodDisruptionBudget
- **storage.k8s.io/v1**: StorageClass
- **autoscaling/v2**: HorizontalPodAutoscaler
- **admissionregistration.k8s.io/v1**: ValidatingWebhookConfiguration, MutatingWebhookConfiguration
- **scheduling.k8s.io/v1**: PriorityClass
- **discovery.k8s.io/v1**: EndpointSlice

Unknown GVKs still produce a diagnostic suggesting you generate typed structs.

## Installation

Use golangci-lint's [custom-gcl module plugin builder](https://golangci-lint.run/plugins/module-plugins/) to include this plugin.

In your `.custom-gcl.yml`:

```yaml
version: v2.1.0
plugins:
  - module: github.com/togethercomputer/kube-types
    import: github.com/togethercomputer/kube-types
    version: latest
```

Then build your custom `golangci-lint`:

```sh
custom-gcl
```

## Configuration

Add to your `.golangci.yml`:

```yaml
version: "2"

linters-settings:
  custom:
    kube-types:
      type: "module"
      description: "Detects untyped Kubernetes manifest construction and suggests typed structs"
      settings:
        # Analyze test files (default: false).
        include_test_files: false

        # Per-check configuration. Omit for all checks enabled with defaults.
        # Valid checks: "map_literal", "sprintf_yaml", "unstructured_gvk",
        #               "raw_gvk_string", "deprecated_api", "embedded_yaml",
        #               "raw_condition_status", "raw_condition_type",
        #               "condition_map_literal", "unstructured_status"
        checks:
          map_literal:
            enabled: true
          sprintf_yaml:
            enabled: true
            additional_markers:
              - "metadata:"
          unstructured_gvk:
            enabled: true
          raw_gvk_string:
            enabled: true
          deprecated_api:
            enabled: true
          embedded_yaml:
            enabled: true
          raw_condition_status:
            enabled: true
          raw_condition_type:
            enabled: true
          condition_map_literal:
            enabled: true
          unstructured_status:
            enabled: true

        # Register additional GVKs beyond the built-in table.
        extra_known_gvks:
          - api_version: "example.io/v1"
            kind: "Widget"
            typed_package: "example.io/api/v1.Widget"

        # Skip diagnostics for specific GVKs (format: "apiVersion/kind").
        ignore_gvks:
          - "v1/ConfigMap"

        # Reject specific GVKs by project policy (format: "apiVersion/kind").
        # Produces an error-level diagnostic whenever the GVK is constructed.
        reject_gvks:
          - "v1/Pod"

linters:
  enable:
    - kube-types
```

### Configuration Reference

| Field | Type | Default | Description |
| ----- | ---- | ------- | ----------- |
| `include_test_files` | `bool` | `false` | Analyze `_test.go` files |
| `checks` | `map[string]CheckConfig` | all enabled | Per-check enable/disable and settings |
| `checks.<name>.enabled` | `*bool` | `true` | Enable or disable a specific check |
| `checks.sprintf_yaml.additional_markers` | `[]string` | `[]` | Extra YAML markers beyond `apiVersion:` and `kind:` |
| `extra_known_gvks` | `[]GVKEntry` | `[]` | Additional GVK-to-typed-struct mappings |
| `extra_known_gvks[].api_version` | `string` | required | API version (e.g. `"apps/v1"`) |
| `extra_known_gvks[].kind` | `string` | required | Kind (e.g. `"Deployment"`) |
| `extra_known_gvks[].typed_package` | `string` | required | Full typed package path (e.g. `"k8s.io/api/apps/v1.Deployment"`) |
| `ignore_gvks` | `[]string` | `[]` | GVK keys to suppress (`"apiVersion/kind"` format) |
| `reject_gvks` | `[]string` | `[]` | GVK keys to reject by policy (`"apiVersion/kind"` format) |

Valid check names: `map_literal`, `sprintf_yaml`, `unstructured_gvk`, `raw_gvk_string`, `deprecated_api`, `embedded_yaml`, `raw_condition_status`, `raw_condition_type`, `condition_map_literal`, `unstructured_status`.

### IgnoreGVKs Behavior

`ignore_gvks` suppresses diagnostics for the following checks:

- **`map_literal`** -- suppressed when both `apiVersion` and `kind` match
- **`unstructured_gvk`** -- suppressed for `SetGroupVersionKind` and `SetAPIVersion`/`SetKind` pairs
- **`raw_gvk_string` (composite literals)** -- suppressed for `TypeMeta{}` and `GroupVersionKind{}` when the full GVK can be resolved (both raw and const field values are considered)
- **`raw_gvk_string` (comparisons)** -- **not suppressed** (only one side of the comparison is visible, so a full GVK key cannot be constructed)
- **`deprecated_api`** -- suppressed (the ignore takes precedence over the deprecation warning)

### RejectGVKs Behavior

`reject_gvks` produces a policy-violation diagnostic whenever the specified GVK is constructed, regardless of whether it would otherwise be flagged. This is useful for enforcing organizational standards (e.g., "never construct bare Pods").

**Reject takes precedence over ignore.** If a GVK appears in both `reject_gvks` and `ignore_gvks`, the reject diagnostic fires and the ignore is not applied. This prevents accidental suppression of policy violations.

```yaml
# v1/Pod is in both lists -- reject wins, producing a diagnostic.
ignore_gvks:
  - "v1/Pod"
reject_gvks:
  - "v1/Pod"
```

### Suppressing Diagnostics

Use `//nolint:kube-types` to suppress a specific line:

```go
m := map[string]any{ //nolint:kube-types
    "apiVersion": "v1",
    "kind":       "ConfigMap",
}
```

## Development

```sh
# Run tests
make test

# Run linter
make lint

# Tidy deps
make tidy

# Build
make build

# Coverage report
make cover
```

## Known Limitations

- **Two-step map construction** is not detected. `m := make(map[string]any); m["apiVersion"] = "v1"; m["kind"] = "Pod"` won't fire because the keys are set via statements, not in the composite literal.
- **Cross-function SetAPIVersion/SetKind pairs** are not tracked. Both calls must be on the same receiver variable within the same function body.
- **Non-const variable format strings** in `fmt.Sprintf` are not analyzed. Only string literals and `const` strings are resolved.
- The `sprintf_yaml` check uses substring matching for markers. A string like `"log kind: info"` would be flagged if `kind:` is a marker.
- **IgnoreGVKs does not apply to comparisons** in the `raw_gvk_string` check. Expressions like `obj.GetKind() == "Deployment"` only expose one side of the GVK, making it impossible to construct the full `apiVersion/kind` key needed for suppression.
- **Positional GVK literals** like `schema.GroupVersionKind{"apps", "v1", "Deployment"}` are not detected. The analyzer requires named key-value pairs (e.g., `Group: "apps"`) to identify fields.
