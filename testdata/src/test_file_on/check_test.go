package test_file_on

import "os"

// Map literal in test file: flagged when IncludeTestFiles=true.
var deployment = map[string]any{ // want `use \*appsv1\.Deployment \(import "k8s\.io/api/apps/v1"\) instead of map\[string\]any for apps/v1/Deployment`
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
}

func readInTest() {
	_, _ = os.ReadFile("in-test.yaml") // want `os\.ReadFile of YAML file "in-test\.yaml"; consider using typed Kubernetes structs from k8s\.io/api`
}
