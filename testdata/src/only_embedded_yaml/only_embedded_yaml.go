package only_embedded_yaml

import "os"

// Triggers only embedded_yaml via os.ReadFile — no other checks.
// When embedded_yaml is disabled, no diagnostics should fire.
func readYAML() {
	_, _ = os.ReadFile("deployment.yaml")
}
