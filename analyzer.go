package kubetypes

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func newAnalyzer(p *plugin) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "kubetypes",
		Doc:      "detects untyped Kubernetes manifest construction (map literals, sprintf YAML, unstructured) and suggests typed structs",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, &p.settings, p.gvkTable)
		},
	}
}

func run(pass *analysis.Pass, settings *Settings, gvkTable map[string]gvkInfo) (any, error) {
	enabled := settings.enabledChecks()
	markers := settings.markersForSprintfYAML()
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Track SetAPIVersion/SetKind pairs per receiver for the unstructured check.
	pairTracker := make(map[token.Pos]*gvkParts)

	nodeFilter := []ast.Node{
		(*ast.CompositeLit)(nil),
		(*ast.CallExpr)(nil),
		(*ast.BinaryExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		if !settings.IncludeTestFiles {
			pos := pass.Fset.Position(n.Pos())
			if strings.HasSuffix(pos.Filename, "_test.go") {
				return
			}
		}

		switch node := n.(type) {
		case *ast.CompositeLit:
			if enabled[checkMapLiteral] {
				checkMapLiteralExpr(pass, node, gvkTable, settings)
			}
			if enabled[checkRawGVKString] {
				checkRawGVKStringCompositeLit(pass, node, gvkTable, settings)
			}
		case *ast.CallExpr:
			if enabled[checkSprintfYAML] {
				checkSprintfYAMLExpr(pass, node, markers)
			}
			if enabled[checkUnstructuredGVK] {
				checkUnstructuredGVKExpr(pass, node, gvkTable, settings)
				trackSetAPIVersionKind(pass, node, pairTracker)
			}
		case *ast.BinaryExpr:
			if enabled[checkRawGVKString] {
				checkRawGVKStringBinaryExpr(pass, node)
			}
		}
	})

	// Report SetAPIVersion/SetKind pairs with known GVKs.
	if enabled[checkUnstructuredGVK] {
		reportSetPairs(pass, pairTracker, gvkTable, settings)
	}

	return nil, nil
}

// extractStringValue extracts a plain string value from a *ast.BasicLit of kind STRING.
func extractStringValue(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	val, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return val, true
}

// extractStringOrConstValue extracts a string value from a literal or a const string identifier.
func extractStringOrConstValue(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	if val, ok := extractStringValue(expr); ok {
		return val, true
	}

	ident, ok := expr.(*ast.Ident)
	if !ok {
		return "", false
	}

	if pass.TypesInfo == nil {
		return "", false
	}

	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return "", false
	}

	c, ok := obj.(*types.Const)
	if !ok {
		return "", false
	}

	if basic, ok := c.Type().(*types.Basic); !ok || (basic.Kind() != types.String && basic.Kind() != types.UntypedString) {
		return "", false
	}

	return constant.StringVal(c.Val()), true
}

// isMapStringAnyLit checks whether a composite literal represents a map[string]any
// (or map[string]interface{}, or a named type alias thereof).
func isMapStringAnyLit(pass *analysis.Pass, lit *ast.CompositeLit) bool {
	// Fast path: direct map[string]any syntax.
	if lit.Type != nil {
		if mapType, ok := lit.Type.(*ast.MapType); ok {
			return isMapStringAnyDirect(mapType)
		}
	}

	// Named type — use the type checker to resolve the underlying type.
	t := pass.TypesInfo.TypeOf(lit)
	if t == nil {
		return false
	}
	m, ok := t.Underlying().(*types.Map)
	if !ok {
		return false
	}
	keyBasic, ok := m.Key().(*types.Basic)
	if !ok || keyBasic.Kind() != types.String {
		return false
	}
	// Unalias to handle `any` which is a type alias for `interface{}`.
	elem := types.Unalias(m.Elem())
	valIface, ok := elem.(*types.Interface)
	if !ok {
		return false
	}
	return valIface.NumMethods() == 0
}

// isMapStringAnyDirect checks if an ast.MapType is map[string]any or map[string]interface{}.
func isMapStringAnyDirect(mapType *ast.MapType) bool {
	keyIdent, ok := mapType.Key.(*ast.Ident)
	if !ok || keyIdent.Name != "string" {
		return false
	}

	switch val := mapType.Value.(type) {
	case *ast.Ident:
		return val.Name == "any"
	case *ast.InterfaceType:
		return val.Methods != nil && len(val.Methods.List) == 0
	}
	return false
}
