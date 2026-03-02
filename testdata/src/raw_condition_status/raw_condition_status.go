package raw_condition_status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	condStatusTrue = metav1.ConditionStatus("True")
	condTypeReady  = "Ready"
)

// --- Composite literal: raw string Status flagged ---

var rawStatus = metav1.Condition{
	Type:   condTypeReady,
	Status: "True", // want `raw string "True" for Condition\.Status; use metav1\.ConditionTrue instead`
}

var rawStatusFalse = metav1.Condition{
	Type:   condTypeReady,
	Status: "False", // want `raw string "False" for Condition\.Status; use metav1\.ConditionFalse instead`
}

var rawStatusUnknown = metav1.Condition{
	Type:   condTypeReady,
	Status: "Unknown", // want `raw string "Unknown" for Condition\.Status; use metav1\.ConditionUnknown instead`
}

// --- Composite literal: const Status NOT flagged ---

var constStatus = metav1.Condition{
	Type:   condTypeReady,
	Status: metav1.ConditionTrue,
}

// --- Binary comparison: cond.Status == "True" flagged ---

func checkStatusEq(cond metav1.Condition) bool {
	return cond.Status == "True" // want `raw string "True" in Condition\.Status comparison; use metav1\.ConditionTrue instead`
}

func checkStatusNeq(cond metav1.Condition) bool {
	return cond.Status != "False" // want `raw string "False" in Condition\.Status comparison; use metav1\.ConditionFalse instead`
}

// --- Reverse: "True" == cond.Status flagged ---

func checkStatusReverse(cond metav1.Condition) bool {
	return "Unknown" == cond.Status // want `raw string "Unknown" in Condition\.Status comparison; use metav1\.ConditionUnknown instead`
}

// --- Pointer receiver: flagged ---

func checkStatusPtr(cond *metav1.Condition) bool {
	return cond.Status == "True" // want `raw string "True" in Condition\.Status comparison; use metav1\.ConditionTrue instead`
}

// --- Const comparison: NOT flagged ---

func checkStatusConst(cond metav1.Condition) bool {
	return cond.Status == condStatusTrue
}

// --- Non-Condition type: NOT flagged ---

type myCondition struct {
	Status string
}

var notCondition = myCondition{
	Status: "True",
}

func checkNonConditionStatus(c myCondition) bool {
	return c.Status == "True"
}

// --- Non-status field: NOT flagged ---

func checkTypeField(cond metav1.Condition) bool {
	return cond.Type == "Ready"
}

// --- Unrelated raw string value (not True/False/Unknown): NOT flagged ---

var unrelatedStatus = metav1.Condition{
	Type:   condTypeReady,
	Status: "SomethingElse",
}
