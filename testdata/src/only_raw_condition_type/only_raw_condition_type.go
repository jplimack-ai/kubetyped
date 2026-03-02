package only_raw_condition_type

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Triggers only raw_condition_type — Status uses const to avoid raw_condition_status.
var cond = metav1.Condition{
	Type:   "Ready",
	Status: metav1.ConditionTrue,
}
