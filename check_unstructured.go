package kubetypes

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"maps"
	"slices"

	"golang.org/x/tools/go/analysis"
)

const (
	unstructuredPkgPath = "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	unstructuredGVKURL  = "https://github.com/togethercomputer/kube-types#unstructured_gvk"
)

// checkUnstructuredGVKExpr detects calls like:
//
//	u.SetGroupVersionKind(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"})
//
// where u is *unstructured.Unstructured and the GVK matches a known built-in type.
func checkUnstructuredGVKExpr(pass *analysis.Pass, call *ast.CallExpr, gvkTable map[string]gvkInfo, settings *Settings, enabled map[string]bool, policyChecked map[token.Pos]gvkAction) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "SetGroupVersionKind" {
		return
	}

	if pass.TypesInfo == nil {
		return
	}

	recvType := pass.TypesInfo.TypeOf(sel.X)
	if recvType == nil || !isUnstructuredType(recvType) {
		return
	}

	if len(call.Args) != 1 {
		return
	}

	lit, ok := call.Args[0].(*ast.CompositeLit)
	if !ok {
		return
	}

	apiVersion, kind := extractGVKFromCompositeLit(pass, lit)
	if apiVersion == "" || kind == "" {
		return
	}

	action := evalGVKPolicy(pass, call.Pos(), apiVersion, kind, checkUnstructuredGVK, settings, enabled)
	// Record the policy result so checkRawGVKStringCompositeLit won't emit
	// duplicate deprecated_api/reject diagnostics for the inner GVK literal.
	policyChecked[lit.Pos()] = action
	if action == gvkStop {
		return
	}

	if !enabled[checkUnstructuredGVK] {
		return
	}

	info, ok := lookupGVK(gvkTable, apiVersion, kind)
	if !ok {
		return
	}

	pass.Report(analysis.Diagnostic{
		Pos:      call.Pos(),
		Category: checkUnstructuredGVK,
		URL:      unstructuredGVKURL,
		Message: fmt.Sprintf(
			"SetGroupVersionKind(apiVersion=%q, kind=%q) on unstructured.Unstructured: use *%s (import %q) instead",
			apiVersion, kind, info.ShortName, info.ImportPath,
		),
	})
}

// extractGVKFromCompositeLit extracts apiVersion and kind from a schema.GroupVersionKind composite literal.
func extractGVKFromCompositeLit(pass *analysis.Pass, lit *ast.CompositeLit) (string, string) {
	var group, version, kind string
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		fieldIdent, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		val, ok := extractStringOrConstValue(pass, kv.Value)
		if !ok {
			continue
		}

		switch fieldIdent.Name {
		case "Group":
			group = val
		case "Version":
			version = val
		case "Kind":
			kind = val
		}
	}

	if version == "" || kind == "" {
		return "", ""
	}

	apiVersion := version
	if group != "" {
		apiVersion = group + "/" + version
	}
	return apiVersion, kind
}

// gvkParts tracks SetAPIVersion/SetKind calls on a receiver variable.
type gvkParts struct {
	apiVersion    string
	apiVersionPos token.Pos
	kind          string
	kindPos       token.Pos
}

// trackSetAPIVersionKind records SetAPIVersion/SetKind calls on unstructured receivers.
func trackSetAPIVersionKind(pass *analysis.Pass, call *ast.CallExpr, tracker map[token.Pos]*gvkParts) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	methodName := sel.Sel.Name
	if methodName != "SetAPIVersion" && methodName != "SetKind" {
		return
	}

	if pass.TypesInfo == nil {
		return
	}

	recvType := pass.TypesInfo.TypeOf(sel.X)
	if recvType == nil || !isUnstructuredType(recvType) {
		return
	}

	if len(call.Args) != 1 {
		return
	}

	val, ok := extractStringOrConstValue(pass, call.Args[0])
	if !ok {
		return
	}

	// Use the receiver variable's declaration position as a unique key.
	recvIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}
	obj := pass.TypesInfo.ObjectOf(recvIdent)
	if obj == nil {
		return
	}
	key := obj.Pos()

	parts, exists := tracker[key]
	if !exists {
		parts = &gvkParts{}
		tracker[key] = parts
	}

	switch methodName {
	case "SetAPIVersion":
		parts.apiVersion = val
		parts.apiVersionPos = call.Pos()
	case "SetKind":
		parts.kind = val
		parts.kindPos = call.Pos()
	}
}

// reportSetPairs reports diagnostics for receivers that had both SetAPIVersion and SetKind called.
func reportSetPairs(pass *analysis.Pass, tracker map[token.Pos]*gvkParts, gvkTable map[string]gvkInfo, settings *Settings, enabled map[string]bool) {
	keys := slices.Sorted(maps.Keys(tracker))
	for _, key := range keys {
		parts := tracker[key]
		if parts.apiVersion == "" || parts.kind == "" {
			continue
		}

		pos := min(parts.apiVersionPos, parts.kindPos)

		if evalGVKPolicy(pass, pos, parts.apiVersion, parts.kind, checkUnstructuredGVK, settings, enabled) == gvkStop {
			continue
		}

		if !enabled[checkUnstructuredGVK] {
			continue
		}

		info, ok := lookupGVK(gvkTable, parts.apiVersion, parts.kind)
		if !ok {
			continue
		}

		pass.Report(analysis.Diagnostic{
			Pos:      pos,
			Category: checkUnstructuredGVK,
			URL:      unstructuredGVKURL,
			Message: fmt.Sprintf(
				"SetAPIVersion(%q) + SetKind(%q) on unstructured.Unstructured: use *%s (import %q) instead",
				parts.apiVersion, parts.kind, info.ShortName, info.ImportPath,
			),
		})
	}
}

// isUnstructuredType checks whether a type is *unstructured.Unstructured or unstructured.Unstructured.
func isUnstructuredType(t types.Type) bool {
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

	return obj.Pkg().Path() == unstructuredPkgPath && obj.Name() == "Unstructured"
}
