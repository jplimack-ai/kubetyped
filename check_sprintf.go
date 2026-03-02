package kubetypes

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const sprintfYAMLURL = "https://github.com/togethercomputer/kube-types#sprintf_yaml"

// defaultYAMLMarkers are strings that suggest a YAML Kubernetes manifest template.
var defaultYAMLMarkers = []string{"apiVersion:", "kind:"}

// checkSprintfYAMLExpr detects fmt.Sprintf/Fprintf calls where the format string
// contains YAML markers like "apiVersion:" or "kind:", indicating string-interpolated
// Kubernetes manifest construction.
func checkSprintfYAMLExpr(pass *analysis.Pass, call *ast.CallExpr, markers []string) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok || !isPkgPath(pass, pkgIdent, "fmt") {
		return
	}

	fnName := sel.Sel.Name
	if fnName != "Sprintf" && fnName != "Fprintf" {
		return
	}

	// The format string is the first arg for Sprintf, second for Fprintf.
	fmtArgIdx := 0
	if fnName == "Fprintf" {
		fmtArgIdx = 1
	}

	if fmtArgIdx >= len(call.Args) {
		return
	}

	// Try literal first, then const string.
	fmtStr, ok := extractStringOrConstValue(pass, call.Args[fmtArgIdx])
	if !ok {
		return
	}

	for _, marker := range markers {
		if strings.Contains(fmtStr, marker) {
			pass.Report(analysis.Diagnostic{
				Pos:      call.Pos(),
				Category: checkSprintfYAML,
				URL:      sprintfYAMLURL,
				Message: fmt.Sprintf(
					"fmt.%s with YAML marker %q suggests string-interpolated Kubernetes manifest; use typed structs from k8s.io/api instead",
					fnName, marker,
				),
			})
			return
		}
	}
}
