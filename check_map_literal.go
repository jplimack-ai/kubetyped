package kubetypes

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const mapLiteralURL = "https://github.com/togethercomputer/kube-types#map_literal"

// checkMapLiteralExpr detects map[string]any{} or map[string]interface{}{} composite
// literals (including named type aliases) that contain both "apiVersion" and "kind"
// string keys, indicating a hand-constructed Kubernetes manifest.
func checkMapLiteralExpr(pass *analysis.Pass, lit *ast.CompositeLit, gvkTable map[string]gvkInfo, settings *Settings, enabled map[string]bool) {
	if !isMapStringAnyLit(pass, lit) {
		return
	}

	var apiVersion, kind string
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
		case "apiVersion":
			apiVersion, _ = extractStringOrConstValue(pass, kv.Value)
		case "kind":
			kind, _ = extractStringOrConstValue(pass, kv.Value)
		}
	}

	if apiVersion == "" || kind == "" {
		return
	}

	if evalGVKPolicy(pass, lit.Pos(), apiVersion, kind, checkMapLiteral, settings, enabled) == gvkStop {
		return
	}

	if !enabled[checkMapLiteral] {
		return
	}

	if info, ok := lookupGVK(gvkTable, apiVersion, kind); ok {
		pass.Report(analysis.Diagnostic{
			Pos:      lit.Pos(),
			Category: checkMapLiteral,
			URL:      mapLiteralURL,
			Message: fmt.Sprintf(
				"use *%s (import %q) instead of map[string]any for %s/%s",
				info.ShortName, info.ImportPath, apiVersion, kind,
			),
		})
	} else {
		pass.Report(analysis.Diagnostic{
			Pos:      lit.Pos(),
			Category: checkMapLiteral,
			URL:      mapLiteralURL,
			Message: fmt.Sprintf(
				"map[string]any with apiVersion %q and kind %q constructs a Kubernetes manifest without type safety; consider generating typed structs with controller-gen",
				apiVersion, kind,
			),
		})
	}
}
