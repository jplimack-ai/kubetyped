package test_file_off

import "os"

// Map literal in test file: NOT flagged when IncludeTestFiles=false.
var deployment = map[string]any{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
}

func readInTest() {
	_, _ = os.ReadFile("in-test.yaml") // no diagnostic — test files excluded
}
