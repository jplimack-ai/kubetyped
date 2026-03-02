package test_file_off

// Map literal in test file: NOT flagged when IncludeTestFiles=false.
var deployment = map[string]any{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
}
