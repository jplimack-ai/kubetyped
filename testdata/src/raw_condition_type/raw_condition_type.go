package raw_condition_type

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const condTypeReady = "Ready"

// --- Raw string Type: flagged ---

var rawType = metav1.Condition{
	Type:   "Ready", // want `raw string "Ready" for Condition\.Type; define a package-level const`
	Status: metav1.ConditionTrue,
}

var rawTypeAvailable = metav1.Condition{
	Type:   "Available", // want `raw string "Available" for Condition\.Type; define a package-level const`
	Status: metav1.ConditionFalse,
}

// --- Const Type: NOT flagged ---

var constType = metav1.Condition{
	Type:   condTypeReady,
	Status: metav1.ConditionTrue,
}

// --- Non-Condition type: NOT flagged ---

type myStruct struct {
	Type   string
	Status string
}

var notCondition = myStruct{
	Type:   "Ready",
	Status: "True",
}
