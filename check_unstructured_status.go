package kubetypes

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const unstructuredStatusURL = "https://github.com/togethercomputer/kube-types#unstructured_status"

// checkUnstructuredStatusCall flags unstructured.SetNestedField/SetNestedSlice calls
// targeting the "status" path.
func checkUnstructuredStatusCall(pass *analysis.Pass, call *ast.CallExpr) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	methodName := sel.Sel.Name
	if methodName != "SetNestedField" && methodName != "SetNestedSlice" {
		return
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}

	if !isPkgPath(pass, pkgIdent, unstructuredPkgPath) {
		return
	}

	// SetNestedField(obj, value, fields...) — first variadic field is at index 2.
	// SetNestedSlice(obj, value, fields...) — first variadic field is at index 2.
	if len(call.Args) < 3 {
		return
	}

	firstField, ok := extractStringValue(call.Args[2])
	if !ok {
		return
	}

	if firstField != "status" {
		return
	}

	pass.Report(analysis.Diagnostic{
		Pos:      call.Pos(),
		Category: checkUnstructuredStatus,
		URL:      unstructuredStatusURL,
		Message:  "unstructured.SetNestedField targeting \"status\" path; use typed status subresource updates instead",
	})
}

// checkUnstructuredStatusIndex flags u.Object["status"] map access on Unstructured types.
func checkUnstructuredStatusIndex(pass *analysis.Pass, expr *ast.IndexExpr) {
	// Match: sel.Object["status"] where sel.X is *unstructured.Unstructured.
	sel, ok := expr.X.(*ast.SelectorExpr)
	if !ok {
		return
	}

	if sel.Sel.Name != "Object" {
		return
	}

	if pass.TypesInfo == nil {
		return
	}

	recvType := pass.TypesInfo.TypeOf(sel.X)
	if recvType == nil || !isUnstructuredType(recvType) {
		return
	}

	indexVal, ok := extractStringValue(expr.Index)
	if !ok {
		return
	}

	if indexVal != "status" {
		return
	}

	pass.Report(analysis.Diagnostic{
		Pos:      expr.Pos(),
		Category: checkUnstructuredStatus,
		URL:      unstructuredStatusURL,
		Message:  "direct map access to Unstructured.Object[\"status\"]; use typed status subresource updates instead",
	})
}
