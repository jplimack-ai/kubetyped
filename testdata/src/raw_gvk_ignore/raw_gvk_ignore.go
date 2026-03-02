package raw_gvk_ignore

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TypeMeta with ignored GVK (apps/v1/Deployment): NOT flagged.
var ignoredTypeMeta = metav1.TypeMeta{
	Kind:       "Deployment",
	APIVersion: "apps/v1",
}

// TypeMeta with non-ignored GVK (v1/Pod): flagged.
var nonIgnoredTypeMeta = metav1.TypeMeta{
	Kind:       "Pod", // want `raw string "Pod" for TypeMeta\.Kind; define a package-level const`
	APIVersion: "v1",  // want `raw string "v1" for TypeMeta\.APIVersion; define a package-level const`
}

// GVK with ignored GVK: NOT flagged.
var ignoredGVK = schema.GroupVersionKind{
	Group:   "apps",
	Version: "v1",
	Kind:    "Deployment",
}

// GVK with non-ignored GVK: flagged.
var nonIgnoredGVK = schema.GroupVersionKind{
	Version: "v1",
	Kind:    "Pod", // want `raw string "Pod" for GroupVersionKind\.Kind; use \*corev1\.Pod \(import "k8s\.io/api/core/v1"\) or define a const`
}
