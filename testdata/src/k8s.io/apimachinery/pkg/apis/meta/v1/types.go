package v1

// TypeMeta is a minimal mock of k8s.io/apimachinery's TypeMeta for testing.
type TypeMeta struct {
	Kind       string
	APIVersion string
}

// ConditionStatus represents the status of a Condition.
type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// Condition is a minimal mock of k8s.io/apimachinery's Condition for testing.
type Condition struct {
	Type               string
	Status             ConditionStatus
	ObservedGeneration int64
	LastTransitionTime string
	Reason             string
	Message            string
}
