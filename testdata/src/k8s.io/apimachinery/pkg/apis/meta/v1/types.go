package v1

// TypeMeta is a minimal mock of k8s.io/apimachinery's TypeMeta for testing.
type TypeMeta struct {
	Kind       string
	APIVersion string
}
