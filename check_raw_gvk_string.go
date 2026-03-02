package kubetypes

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	typeMetaPkgPath = "k8s.io/apimachinery/pkg/apis/meta/v1"
	schemaPkgPath   = "k8s.io/apimachinery/pkg/runtime/schema"
	rawGVKStringURL = "https://github.com/togethercomputer/kube-types#raw_gvk_string"
)

// checkRawGVKStringCompositeLit flags raw string literals in TypeMeta and GroupVersionKind composite literals.
func checkRawGVKStringCompositeLit(pass *analysis.Pass, lit *ast.CompositeLit, gvkTable map[string]gvkInfo, settings *Settings, enabled map[string]bool, policyChecked map[token.Pos]gvkAction) {
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

	pkgPath := obj.Pkg().Path()
	typeName := obj.Name()

	switch {
	case pkgPath == typeMetaPkgPath && typeName == "TypeMeta":
		checkTypeMetaRawStrings(pass, lit, settings, enabled, policyChecked)
	case pkgPath == schemaPkgPath && typeName == "GroupVersionKind":
		checkGVKRawStrings(pass, lit, gvkTable, settings, enabled, policyChecked)
	}
}

// checkTypeMetaRawStrings flags raw string literals in TypeMeta.Kind and TypeMeta.APIVersion.
// Both field values (raw or const) are extracted first so IgnoreGVKs can suppress the full GVK.
func checkTypeMetaRawStrings(pass *analysis.Pass, lit *ast.CompositeLit, settings *Settings, enabled map[string]bool, policyChecked map[token.Pos]gvkAction) {
	type rawField struct {
		name string
		val  string
		pos  token.Pos
	}

	var apiVersion, kind string
	var raws []rawField

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
		case "Kind":
			if val, ok := extractStringValue(kv.Value); ok {
				kind = val
				raws = append(raws, rawField{"Kind", val, kv.Value.Pos()})
			} else if val, ok := extractStringOrConstValue(pass, kv.Value); ok {
				kind = val
			}
		case "APIVersion":
			if val, ok := extractStringValue(kv.Value); ok {
				apiVersion = val
				raws = append(raws, rawField{"APIVersion", val, kv.Value.Pos()})
			} else if val, ok := extractStringOrConstValue(pass, kv.Value); ok {
				apiVersion = val
			}
		}
	}

	if apiVersion != "" && kind != "" {
		if action, checked := policyChecked[lit.Pos()]; checked {
			// checkUnstructuredGVKExpr already ran the policy for this literal.
			// If it returned gvkStop (reject or ignore), suppress raw diagnostics too.
			if action == gvkStop {
				return
			}
		} else if evalGVKPolicy(pass, lit.Pos(), apiVersion, kind, checkRawGVKString, settings, enabled) == gvkStop {
			return
		}
	}

	if len(raws) == 0 || !enabled[checkRawGVKString] {
		return
	}

	for _, f := range raws {
		pass.Report(analysis.Diagnostic{
			Pos:      f.pos,
			Category: checkRawGVKString,
			URL:      rawGVKStringURL,
			Message: fmt.Sprintf(
				"raw string %q for TypeMeta.%s; define a package-level const or use scheme-based type resolution",
				f.val, f.name,
			),
		})
	}
}

// checkGVKRawStrings flags raw string literals in GroupVersionKind.Kind and looks up the GVK table for suggestions.
func checkGVKRawStrings(pass *analysis.Pass, lit *ast.CompositeLit, gvkTable map[string]gvkInfo, settings *Settings, enabled map[string]bool, policyChecked map[token.Pos]gvkAction) {
	var group, version, kind string
	var kindValue ast.Expr
	var kindIsRaw bool

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
		case "Group":
			group, _ = extractStringOrConstValue(pass, kv.Value)
		case "Version":
			version, _ = extractStringOrConstValue(pass, kv.Value)
		case "Kind":
			if val, ok := extractStringValue(kv.Value); ok {
				kind = val
				kindValue = kv.Value
				kindIsRaw = true
			} else if val, ok := extractStringOrConstValue(pass, kv.Value); ok {
				kind = val
			}
		}
	}

	apiVersion := version
	if group != "" {
		apiVersion = group + "/" + version
	}

	if apiVersion != "" && kind != "" {
		if action, checked := policyChecked[lit.Pos()]; checked {
			if action == gvkStop {
				return
			}
		} else if evalGVKPolicy(pass, lit.Pos(), apiVersion, kind, checkRawGVKString, settings, enabled) == gvkStop {
			return
		}
	}

	if !kindIsRaw || kind == "" || !enabled[checkRawGVKString] {
		return
	}

	if info, ok := lookupGVK(gvkTable, apiVersion, kind); ok {
		pass.Report(analysis.Diagnostic{
			Pos:      kindValue.Pos(),
			Category: checkRawGVKString,
			URL:      rawGVKStringURL,
			Message: fmt.Sprintf(
				"raw string %q for GroupVersionKind.Kind; use *%s (import %q) or define a const",
				kind, info.ShortName, info.ImportPath,
			),
		})
	} else {
		pass.Report(analysis.Diagnostic{
			Pos:      kindValue.Pos(),
			Category: checkRawGVKString,
			URL:      rawGVKStringURL,
			Message: fmt.Sprintf(
				"raw string %q for GroupVersionKind.Kind; define a package-level const or use scheme-based type resolution",
				kind,
			),
		})
	}
}

// checkRawGVKStringBinaryExpr flags raw string comparisons against GVK accessors.
func checkRawGVKStringBinaryExpr(pass *analysis.Pass, expr *ast.BinaryExpr) {
	if expr.Op != token.EQL && expr.Op != token.NEQ {
		return
	}

	if isGVKAccess(pass, expr.X) {
		if val, ok := extractStringValue(expr.Y); ok {
			reportRawGVKComparison(pass, expr.Y, val)
		}
	} else if isGVKAccess(pass, expr.Y) {
		if val, ok := extractStringValue(expr.X); ok {
			reportRawGVKComparison(pass, expr.X, val)
		}
	}
}

func reportRawGVKComparison(pass *analysis.Pass, lit ast.Expr, val string) {
	pass.Report(analysis.Diagnostic{
		Pos:      lit.Pos(),
		Category: checkRawGVKString,
		URL:      rawGVKStringURL,
		Message:  fmt.Sprintf("raw string %q in GVK comparison; define a package-level const", val),
	})
}

// isGVKAccess returns true if the expression is a GVK method call or field access.
func isGVKAccess(pass *analysis.Pass, expr ast.Expr) bool {
	return isGVKMethodCall(expr) || isGVKFieldAccess(pass, expr)
}

// isGVKMethodCall checks for GetKind() or GetAPIVersion() calls with zero arguments.
func isGVKMethodCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	name := sel.Sel.Name
	return (name == "GetKind" || name == "GetAPIVersion") && len(call.Args) == 0
}

// isGVKFieldAccess checks for .Kind or .APIVersion field access on types from k8s.io/apimachinery/.
func isGVKFieldAccess(pass *analysis.Pass, expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	name := sel.Sel.Name
	if name != "Kind" && name != "APIVersion" {
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

	return strings.HasPrefix(obj.Pkg().Path(), "k8s.io/apimachinery/")
}
