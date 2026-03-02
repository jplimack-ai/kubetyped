package kubetypes

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const (
	conditionURL    = "https://github.com/togethercomputer/kube-types#condition_checks"
	conditionMapURL = "https://github.com/togethercomputer/kube-types#condition_map_literal"
)

// conditionStatusConstants maps raw string values to their typed constant names.
var conditionStatusConstants = map[string]string{
	"True":    "metav1.ConditionTrue",
	"False":   "metav1.ConditionFalse",
	"Unknown": "metav1.ConditionUnknown",
}

// checkConditionCompositeLit flags raw string literals in metav1.Condition composite literals
// for both Status (raw_condition_status) and Type (raw_condition_type) fields.
func checkConditionCompositeLit(pass *analysis.Pass, lit *ast.CompositeLit, enabled map[string]bool) {
	if pass.TypesInfo == nil {
		return
	}

	t := pass.TypesInfo.TypeOf(lit)
	if t == nil {
		return
	}

	named, ok := t.(*types.Named)
	if !ok {
		return
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return
	}

	if obj.Pkg().Path() != typeMetaPkgPath || obj.Name() != "Condition" {
		return
	}

	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		fieldIdent, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		switch fieldIdent.Name {
		case "Status":
			if !enabled[checkRawConditionStatus] {
				continue
			}
			rawVal, ok := extractStringValue(kv.Value)
			if !ok {
				continue
			}
			constName, known := conditionStatusConstants[rawVal]
			if !known {
				continue
			}
			pass.Report(analysis.Diagnostic{
				Pos:      kv.Value.Pos(),
				Category: checkRawConditionStatus,
				URL:      conditionURL,
				Message:  fmt.Sprintf("raw string %q for Condition.Status; use %s instead", rawVal, constName),
			})
		case "Type":
			if !enabled[checkRawConditionType] {
				continue
			}
			rawVal, ok := extractStringValue(kv.Value)
			if !ok {
				continue
			}
			pass.Report(analysis.Diagnostic{
				Pos:      kv.Value.Pos(),
				Category: checkRawConditionType,
				URL:      conditionURL,
				Message:  fmt.Sprintf("raw string %q for Condition.Type; define a package-level const", rawVal),
			})
		}
	}
}

// checkConditionStatusBinaryExpr flags raw string comparisons against Condition.Status fields.
func checkConditionStatusBinaryExpr(pass *analysis.Pass, expr *ast.BinaryExpr) {
	if !isConditionStatusAccess(pass, expr.X) && !isConditionStatusAccess(pass, expr.Y) {
		return
	}

	// Check both sides — one should be the field access, the other a raw string.
	if isConditionStatusAccess(pass, expr.X) {
		if rawVal, ok := extractStringValue(expr.Y); ok {
			if constName, known := conditionStatusConstants[rawVal]; known {
				pass.Report(analysis.Diagnostic{
					Pos:      expr.Y.Pos(),
					Category: checkRawConditionStatus,
					URL:      conditionURL,
					Message:  fmt.Sprintf("raw string %q in Condition.Status comparison; use %s instead", rawVal, constName),
				})
			}
		}
	} else if isConditionStatusAccess(pass, expr.Y) {
		if rawVal, ok := extractStringValue(expr.X); ok {
			if constName, known := conditionStatusConstants[rawVal]; known {
				pass.Report(analysis.Diagnostic{
					Pos:      expr.X.Pos(),
					Category: checkRawConditionStatus,
					URL:      conditionURL,
					Message:  fmt.Sprintf("raw string %q in Condition.Status comparison; use %s instead", rawVal, constName),
				})
			}
		}
	}
}

// isConditionStatusAccess checks for .Status access on a metav1.Condition type.
func isConditionStatusAccess(pass *analysis.Pass, expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "Status" {
		return false
	}

	if pass.TypesInfo == nil {
		return false
	}

	t := pass.TypesInfo.TypeOf(sel.X)
	if t == nil {
		return false
	}

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	return obj.Pkg().Path() == typeMetaPkgPath && obj.Name() == "Condition"
}

// checkConditionMapLiteralExpr flags map[string]any literals with both "type" and "status" keys,
// which likely construct a Kubernetes status condition.
func checkConditionMapLiteralExpr(pass *analysis.Pass, lit *ast.CompositeLit) {
	if !isMapStringAnyLit(pass, lit) {
		return
	}

	var hasType, hasStatus bool
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := extractStringOrConstValue(pass, kv.Key)
		if !ok {
			continue
		}

		switch key {
		case "type":
			hasType = true
		case "status":
			hasStatus = true
		}
	}

	if !hasType || !hasStatus {
		return
	}

	pass.Report(analysis.Diagnostic{
		Pos:      lit.Pos(),
		Category: checkConditionMapLiteral,
		URL:      conditionMapURL,
		Message:  "map[string]any with \"type\" and \"status\" keys constructs a Kubernetes status condition; use metav1.Condition instead",
	})
}
