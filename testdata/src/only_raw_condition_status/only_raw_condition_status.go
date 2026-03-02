package only_raw_condition_status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Triggers only raw_condition_status — no other checks.
var cond = metav1.Condition{
	Type:   constType,
	Status: "True",
}

const constType = "Ready"
