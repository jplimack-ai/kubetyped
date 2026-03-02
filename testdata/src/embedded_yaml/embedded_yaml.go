package embedded_yaml

import "os"

//go:embed deployment.yaml
var deploymentYAML string // want `//go:embed of YAML file "deployment\.yaml"; consider using typed Kubernetes structs from k8s\.io/api`

//go:embed config.yml
var configYML string // want `//go:embed of YAML file "config\.yml"; consider using typed Kubernetes structs from k8s\.io/api`

//go:embed schema.json
var schemaJSON string // no diagnostic for non-YAML

//go:embed templates/service.yaml other.txt
var multiEmbed string // want `//go:embed of YAML file "templates/service\.yaml"; consider using typed Kubernetes structs from k8s\.io/api`

func readYAML() {
	_, _ = os.ReadFile("manifests/pod.yaml") // want `os\.ReadFile of YAML file "manifests/pod\.yaml"; consider using typed Kubernetes structs from k8s\.io/api`
	_, _ = os.ReadFile("config.yml")         // want `os\.ReadFile of YAML file "config\.yml"; consider using typed Kubernetes structs from k8s\.io/api`
	_, _ = os.ReadFile("data.json")          // no diagnostic for non-YAML
}

func openYAML() {
	_, _ = os.Open("service.yaml") // want `os\.Open of YAML file "service\.yaml"; consider using typed Kubernetes structs from k8s\.io/api`
	_, _ = os.Open("main.go")     // no diagnostic for non-YAML
}

func dynamicPath(path string) {
	_, _ = os.ReadFile(path) // no diagnostic for dynamic path
}

func createYAML() {
	_, _ = os.Create("foo.yaml") // no diagnostic — os.Create is not reading a manifest
}
