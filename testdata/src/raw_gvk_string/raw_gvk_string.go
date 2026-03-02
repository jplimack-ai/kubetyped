package raw_gvk_string

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	constKind       = "Deployment"
	constAPIVersion = "apps/v1"
)

// --- TypeMeta with raw strings: both fields flagged independently ---

var rawTypeMeta = metav1.TypeMeta{
	Kind:       "Deployment", // want `raw string "Deployment" for TypeMeta\.Kind; define a package-level const`
	APIVersion: "apps/v1",    // want `raw string "apps/v1" for TypeMeta\.APIVersion; define a package-level const`
}

// --- TypeMeta with const values: NOT flagged ---

var constTypeMeta = metav1.TypeMeta{
	Kind:       constKind,
	APIVersion: constAPIVersion,
}

// --- TypeMeta with mixed: only raw field flagged ---

var mixedTypeMeta = metav1.TypeMeta{
	Kind:       constKind,
	APIVersion: "apps/v1", // want `raw string "apps/v1" for TypeMeta\.APIVersion; define a package-level const`
}

// --- Non-TypeMeta struct with Kind/APIVersion fields: NOT flagged ---

type myStruct struct {
	Kind       string
	APIVersion string
}

var notTypeMeta = myStruct{
	Kind:       "Deployment",
	APIVersion: "apps/v1",
}

// --- schema.GroupVersionKind with raw Kind, known GVK: flagged with typed suggestion ---

var knownGVK = schema.GroupVersionKind{
	Group:   "apps",
	Version: "v1",
	Kind:    "Deployment", // want `raw string "Deployment" for GroupVersionKind\.Kind; use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) or define a const`
}

// --- schema.GroupVersionKind with raw Kind, unknown GVK: flagged with generic message ---

var unknownGVK = schema.GroupVersionKind{
	Group:   "example.io",
	Version: "v1",
	Kind:    "Widget", // want `raw string "Widget" for GroupVersionKind\.Kind; define a package-level const`
}

// --- schema.GroupVersionKind with const Kind: NOT flagged ---

var constGVK = schema.GroupVersionKind{
	Group:   "apps",
	Version: "v1",
	Kind:    constKind,
}

// --- GVK method call comparisons ---

type kubeObj interface {
	GetKind() string
	GetAPIVersion() string
}

// GetKind() == raw string literal: flagged.
func checkKindMethod(obj kubeObj) bool {
	return obj.GetKind() == "Deployment" // want `raw string "Deployment" in GVK comparison; define a package-level const`
}

// GetAPIVersion() == raw string literal: flagged.
func checkAPIVersionMethod(obj kubeObj) bool {
	return obj.GetAPIVersion() == "apps/v1" // want `raw string "apps/v1" in GVK comparison; define a package-level const`
}

// GetKind() != raw string literal: also flagged.
func checkKindNotEqual(obj kubeObj) bool {
	return obj.GetKind() != "Pod" // want `raw string "Pod" in GVK comparison; define a package-level const`
}

// Reverse order: raw string == GetKind(): flagged.
func checkKindReverse(obj kubeObj) bool {
	return "StatefulSet" == obj.GetKind() // want `raw string "StatefulSet" in GVK comparison; define a package-level const`
}

// --- GVK field access comparisons ---

// gvk.Kind == raw string literal: flagged.
func checkKindField(gvk schema.GroupVersionKind) bool {
	return gvk.Kind == "Deployment" // want `raw string "Deployment" in GVK comparison; define a package-level const`
}

// --- Reverse field access: raw string == gvk.Kind ---

func checkKindFieldReverse(gvk schema.GroupVersionKind) bool {
	return "Deployment" == gvk.Kind // want `raw string "Deployment" in GVK comparison; define a package-level const`
}

// --- Binary non-comparison: concatenation is NOT flagged ---

func checkKindConcat(gvk schema.GroupVersionKind) string {
	return gvk.Kind + "suffix"
}

// --- Negative cases for comparisons ---

// GetKind() == const value: NOT flagged.
func checkKindConst(obj kubeObj) bool {
	return obj.GetKind() == constKind
}

// Unrelated string comparison: NOT flagged.
func checkUnrelated(s string) bool {
	return s == "Deployment"
}

// Field access on non-k8s type: NOT flagged.
func checkLocalKind(s myStruct) bool {
	return s.Kind == "Deployment"
}

// --- Embedded TypeMeta: inner TypeMeta flagged ---

type MyResource struct {
	metav1.TypeMeta
}

var embeddedTypeMeta = MyResource{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Pod", // want `raw string "Pod" for TypeMeta\.Kind; define a package-level const`
		APIVersion: "v1",  // want `raw string "v1" for TypeMeta\.APIVersion; define a package-level const`
	},
}

// --- Empty TypeMeta: NOT flagged ---

var emptyTypeMeta = metav1.TypeMeta{}

// --- Pointer-receiver field access: flagged ---

func checkKindFieldPtr(gvk *schema.GroupVersionKind) bool {
	return gvk.Kind == "Deployment" // want `raw string "Deployment" in GVK comparison; define a package-level const`
}

// --- GetKind with args: NOT flagged ---

type argKind struct{}

func (argKind) GetKind(s string) string { return s }

func checkGetKindWithArgs(a argKind) bool {
	return a.GetKind("x") == "Foo"
}

// --- Version field comparison: NOT flagged ---

func checkVersionField(gvk schema.GroupVersionKind) bool {
	return gvk.Version == "v1"
}

// --- Group field comparison: NOT flagged ---

func checkGroupField(gvk schema.GroupVersionKind) bool {
	return gvk.Group == "apps"
}
