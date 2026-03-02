package only_raw_gvk

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Triggers only raw_gvk_string — no map literal, no sprintf, no unstructured.
var _ = metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"}
