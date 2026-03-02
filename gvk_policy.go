package kubetypes

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

const (
	deprecatedAPIURL = "https://github.com/togethercomputer/kube-types#deprecated_api"
	rejectGVKsURL    = "https://github.com/togethercomputer/kube-types#reject_gvks"
)

// gvkAction indicates what the caller should do after evaluating GVK policy.
type gvkAction int

const (
	gvkContinue gvkAction = iota // proceed with the caller's normal diagnostic
	gvkStop                      // reject or ignore handled it; caller should return
)

// evalGVKPolicy runs the reject → ignore → deprecated chain for a GVK.
// It reports reject or deprecated diagnostics as needed and returns what
// the caller should do next. The category parameter is used as the Category
// for reject diagnostics so that //nolint works correctly.
func evalGVKPolicy(pass *analysis.Pass, pos token.Pos, apiVersion, kind string, category string, settings *Settings, enabled map[string]bool) gvkAction {
	if settings.isGVKRejected(apiVersion, kind) {
		pass.Report(analysis.Diagnostic{
			Pos:      pos,
			Category: category,
			URL:      rejectGVKsURL,
			Message: fmt.Sprintf(
				"construction of %s/%s is rejected by project policy (reject_gvks)",
				apiVersion, kind,
			),
		})
		return gvkStop
	}

	if settings.isGVKIgnored(apiVersion, kind) {
		return gvkStop
	}

	if enabled[checkDeprecatedAPI] {
		if info, ok := lookupDeprecatedGVK(apiVersion, kind); ok {
			pass.Report(analysis.Diagnostic{
				Pos:      pos,
				Category: checkDeprecatedAPI,
				URL:      deprecatedAPIURL,
				Message:  deprecatedAPIMsg(apiVersion, kind, info),
			})
		}
	}

	return gvkContinue
}

// deprecatedAPIInfo describes a deprecated Kubernetes API version.
type deprecatedAPIInfo struct {
	NewAPIVersion string // replacement API version, empty if removed entirely
	RemovedIn     string // Kubernetes version that removed the API
}

// deprecatedGVKs maps "apiVersion/kind" to deprecation info for removed beta APIs.
var deprecatedGVKs = map[string]deprecatedAPIInfo{
	// extensions/v1beta1 — removed in 1.22
	"extensions/v1beta1/Ingress":       {NewAPIVersion: "networking.k8s.io/v1", RemovedIn: "1.22"},
	"extensions/v1beta1/NetworkPolicy": {NewAPIVersion: "networking.k8s.io/v1", RemovedIn: "1.22"},
	"extensions/v1beta1/DaemonSet":     {NewAPIVersion: "apps/v1", RemovedIn: "1.22"},
	"extensions/v1beta1/Deployment":    {NewAPIVersion: "apps/v1", RemovedIn: "1.22"},
	"extensions/v1beta1/ReplicaSet":    {NewAPIVersion: "apps/v1", RemovedIn: "1.22"},

	// apps/v1beta1 — removed in 1.16
	"apps/v1beta1/Deployment":  {NewAPIVersion: "apps/v1", RemovedIn: "1.16"},
	"apps/v1beta1/StatefulSet": {NewAPIVersion: "apps/v1", RemovedIn: "1.16"},

	// apps/v1beta2 — removed in 1.16
	"apps/v1beta2/Deployment":  {NewAPIVersion: "apps/v1", RemovedIn: "1.16"},
	"apps/v1beta2/StatefulSet": {NewAPIVersion: "apps/v1", RemovedIn: "1.16"},
	"apps/v1beta2/DaemonSet":   {NewAPIVersion: "apps/v1", RemovedIn: "1.16"},
	"apps/v1beta2/ReplicaSet":  {NewAPIVersion: "apps/v1", RemovedIn: "1.16"},

	// batch/v1beta1 — removed in 1.25
	"batch/v1beta1/CronJob": {NewAPIVersion: "batch/v1", RemovedIn: "1.25"},

	// policy/v1beta1 — removed in 1.25
	"policy/v1beta1/PodDisruptionBudget": {NewAPIVersion: "policy/v1", RemovedIn: "1.25"},
	"policy/v1beta1/PodSecurityPolicy":   {RemovedIn: "1.25"}, // removed entirely

	// rbac.authorization.k8s.io/v1beta1 — removed in 1.22
	"rbac.authorization.k8s.io/v1beta1/Role":               {NewAPIVersion: "rbac.authorization.k8s.io/v1", RemovedIn: "1.22"},
	"rbac.authorization.k8s.io/v1beta1/ClusterRole":        {NewAPIVersion: "rbac.authorization.k8s.io/v1", RemovedIn: "1.22"},
	"rbac.authorization.k8s.io/v1beta1/RoleBinding":        {NewAPIVersion: "rbac.authorization.k8s.io/v1", RemovedIn: "1.22"},
	"rbac.authorization.k8s.io/v1beta1/ClusterRoleBinding": {NewAPIVersion: "rbac.authorization.k8s.io/v1", RemovedIn: "1.22"},

	// admissionregistration.k8s.io/v1beta1 — removed in 1.22
	"admissionregistration.k8s.io/v1beta1/ValidatingWebhookConfiguration": {NewAPIVersion: "admissionregistration.k8s.io/v1", RemovedIn: "1.22"},
	"admissionregistration.k8s.io/v1beta1/MutatingWebhookConfiguration":   {NewAPIVersion: "admissionregistration.k8s.io/v1", RemovedIn: "1.22"},

	// autoscaling/v2beta1 — removed in 1.26
	"autoscaling/v2beta1/HorizontalPodAutoscaler": {NewAPIVersion: "autoscaling/v2", RemovedIn: "1.26"},

	// autoscaling/v2beta2 — removed in 1.26
	"autoscaling/v2beta2/HorizontalPodAutoscaler": {NewAPIVersion: "autoscaling/v2", RemovedIn: "1.26"},

	// flowcontrol.apiserver.k8s.io/v1beta1 — removed in 1.29
	"flowcontrol.apiserver.k8s.io/v1beta1/FlowSchema":                 {NewAPIVersion: "flowcontrol.apiserver.k8s.io/v1", RemovedIn: "1.29"},
	"flowcontrol.apiserver.k8s.io/v1beta1/PriorityLevelConfiguration": {NewAPIVersion: "flowcontrol.apiserver.k8s.io/v1", RemovedIn: "1.29"},

	// flowcontrol.apiserver.k8s.io/v1beta2 — removed in 1.29
	"flowcontrol.apiserver.k8s.io/v1beta2/FlowSchema":                 {NewAPIVersion: "flowcontrol.apiserver.k8s.io/v1", RemovedIn: "1.29"},
	"flowcontrol.apiserver.k8s.io/v1beta2/PriorityLevelConfiguration": {NewAPIVersion: "flowcontrol.apiserver.k8s.io/v1", RemovedIn: "1.29"},
}

// lookupDeprecatedGVK checks whether a given apiVersion+kind is a deprecated Kubernetes API.
func lookupDeprecatedGVK(apiVersion, kind string) (deprecatedAPIInfo, bool) {
	info, ok := deprecatedGVKs[gvkKey(apiVersion, kind)]
	return info, ok
}

// deprecatedAPIMsg builds a diagnostic message for a deprecated API.
func deprecatedAPIMsg(apiVersion, kind string, info deprecatedAPIInfo) string {
	if info.NewAPIVersion == "" {
		return fmt.Sprintf(
			"deprecated API version %s for %s (removed in Kubernetes %s); this API has been removed entirely",
			apiVersion, kind, info.RemovedIn,
		)
	}
	return fmt.Sprintf(
		"deprecated API version %s for %s (removed in Kubernetes %s); use %s instead",
		apiVersion, kind, info.RemovedIn, info.NewAPIVersion,
	)
}
