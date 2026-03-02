package deprecated_api_ignore

// Deprecated GVK in IgnoreGVKs: apps/v1beta1/Deployment is ignored, so no diagnostics should fire.
var manifest = map[string]any{
	"apiVersion": "apps/v1beta1",
	"kind":       "Deployment",
}

// A non-ignored deprecated GVK should still fire.
var cronJob = map[string]any{ // want `deprecated API version batch/v1beta1 for CronJob \(removed in Kubernetes 1\.25\); use batch/v1 instead` `map\[string\]any with apiVersion "batch/v1beta1" and kind "CronJob" constructs a Kubernetes manifest`
	"apiVersion": "batch/v1beta1",
	"kind":       "CronJob",
}
