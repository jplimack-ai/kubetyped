package kubetypes

// gvkInfo holds the typed Go struct info for a known Kubernetes GVK.
type gvkInfo struct {
	ShortName  string // e.g. "appsv1.Deployment"
	ImportPath string // e.g. "k8s.io/api/apps/v1"
}

// gvkKey returns the lookup key for a GVK: "apiVersion/kind".
func gvkKey(apiVersion, kind string) string { return apiVersion + "/" + kind }

// knownGVK maps "apiVersion/kind" to the typed Go struct info.
// This is the default table — never mutated after init.
var knownGVK = map[string]gvkInfo{
	// core/v1
	"v1/Pod":                   {ShortName: "corev1.Pod", ImportPath: "k8s.io/api/core/v1"},
	"v1/Service":               {ShortName: "corev1.Service", ImportPath: "k8s.io/api/core/v1"},
	"v1/ConfigMap":             {ShortName: "corev1.ConfigMap", ImportPath: "k8s.io/api/core/v1"},
	"v1/Secret":                {ShortName: "corev1.Secret", ImportPath: "k8s.io/api/core/v1"},
	"v1/ServiceAccount":        {ShortName: "corev1.ServiceAccount", ImportPath: "k8s.io/api/core/v1"},
	"v1/Namespace":             {ShortName: "corev1.Namespace", ImportPath: "k8s.io/api/core/v1"},
	"v1/PersistentVolume":      {ShortName: "corev1.PersistentVolume", ImportPath: "k8s.io/api/core/v1"},
	"v1/PersistentVolumeClaim": {ShortName: "corev1.PersistentVolumeClaim", ImportPath: "k8s.io/api/core/v1"},
	"v1/Node":                  {ShortName: "corev1.Node", ImportPath: "k8s.io/api/core/v1"},
	"v1/Endpoints":             {ShortName: "corev1.Endpoints", ImportPath: "k8s.io/api/core/v1"},
	"v1/ResourceQuota":         {ShortName: "corev1.ResourceQuota", ImportPath: "k8s.io/api/core/v1"},
	"v1/LimitRange":            {ShortName: "corev1.LimitRange", ImportPath: "k8s.io/api/core/v1"},

	// apps/v1
	"apps/v1/Deployment":  {ShortName: "appsv1.Deployment", ImportPath: "k8s.io/api/apps/v1"},
	"apps/v1/StatefulSet": {ShortName: "appsv1.StatefulSet", ImportPath: "k8s.io/api/apps/v1"},
	"apps/v1/DaemonSet":   {ShortName: "appsv1.DaemonSet", ImportPath: "k8s.io/api/apps/v1"},
	"apps/v1/ReplicaSet":  {ShortName: "appsv1.ReplicaSet", ImportPath: "k8s.io/api/apps/v1"},

	// batch/v1
	"batch/v1/Job":     {ShortName: "batchv1.Job", ImportPath: "k8s.io/api/batch/v1"},
	"batch/v1/CronJob": {ShortName: "batchv1.CronJob", ImportPath: "k8s.io/api/batch/v1"},

	// networking.k8s.io/v1
	"networking.k8s.io/v1/Ingress":       {ShortName: "networkingv1.Ingress", ImportPath: "k8s.io/api/networking/v1"},
	"networking.k8s.io/v1/NetworkPolicy": {ShortName: "networkingv1.NetworkPolicy", ImportPath: "k8s.io/api/networking/v1"},
	"networking.k8s.io/v1/IngressClass":  {ShortName: "networkingv1.IngressClass", ImportPath: "k8s.io/api/networking/v1"},

	// rbac.authorization.k8s.io/v1
	"rbac.authorization.k8s.io/v1/Role":               {ShortName: "rbacv1.Role", ImportPath: "k8s.io/api/rbac/v1"},
	"rbac.authorization.k8s.io/v1/ClusterRole":        {ShortName: "rbacv1.ClusterRole", ImportPath: "k8s.io/api/rbac/v1"},
	"rbac.authorization.k8s.io/v1/RoleBinding":        {ShortName: "rbacv1.RoleBinding", ImportPath: "k8s.io/api/rbac/v1"},
	"rbac.authorization.k8s.io/v1/ClusterRoleBinding": {ShortName: "rbacv1.ClusterRoleBinding", ImportPath: "k8s.io/api/rbac/v1"},

	// policy/v1
	"policy/v1/PodDisruptionBudget": {ShortName: "policyv1.PodDisruptionBudget", ImportPath: "k8s.io/api/policy/v1"},

	// storage.k8s.io/v1
	"storage.k8s.io/v1/StorageClass": {ShortName: "storagev1.StorageClass", ImportPath: "k8s.io/api/storage/v1"},

	// autoscaling/v2
	"autoscaling/v2/HorizontalPodAutoscaler": {ShortName: "autoscalingv2.HorizontalPodAutoscaler", ImportPath: "k8s.io/api/autoscaling/v2"},

	// admissionregistration.k8s.io/v1
	"admissionregistration.k8s.io/v1/ValidatingWebhookConfiguration": {ShortName: "admissionregistrationv1.ValidatingWebhookConfiguration", ImportPath: "k8s.io/api/admissionregistration/v1"},
	"admissionregistration.k8s.io/v1/MutatingWebhookConfiguration":   {ShortName: "admissionregistrationv1.MutatingWebhookConfiguration", ImportPath: "k8s.io/api/admissionregistration/v1"},

	// scheduling.k8s.io/v1
	"scheduling.k8s.io/v1/PriorityClass": {ShortName: "schedulingv1.PriorityClass", ImportPath: "k8s.io/api/scheduling/v1"},

	// discovery.k8s.io/v1
	"discovery.k8s.io/v1/EndpointSlice": {ShortName: "discoveryv1.EndpointSlice", ImportPath: "k8s.io/api/discovery/v1"},
}

// lookupGVK returns the typed struct info for a known apiVersion+kind combination.
func lookupGVK(table map[string]gvkInfo, apiVersion, kind string) (gvkInfo, bool) {
	info, ok := table[gvkKey(apiVersion, kind)]
	return info, ok
}
