package kubetypes

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const embeddedYAMLURL = "https://github.com/togethercomputer/kube-types#embedded_yaml"

// checkEmbeddedYAMLFiles scans file comments for //go:embed directives
// that reference .yaml or .yml files.
func checkEmbeddedYAMLFiles(pass *analysis.Pass, settings *Settings) {
	for _, f := range pass.Files {
		if !settings.IncludeTestFiles {
			pos := pass.Fset.Position(f.Pos())
			if strings.HasSuffix(pos.Filename, "_test.go") {
				continue
			}
		}

		for _, cg := range f.Comments {
			for _, c := range cg.List {
				text := c.Text
				if !strings.HasPrefix(text, "//go:embed ") {
					continue
				}
				for p := range strings.FieldsSeq(text[len("//go:embed "):]) {
					if isYAMLPath(p) {
						// Find the next declaration after this comment to report at its position.
						reportPos := findNextDeclPos(f, c.End())
						pass.Report(analysis.Diagnostic{
							Pos:      reportPos,
							Category: checkEmbeddedYAML,
							URL:      embeddedYAMLURL,
							Message: fmt.Sprintf(
								"//go:embed of YAML file %q; consider using typed Kubernetes structs from k8s.io/api",
								p,
							),
						})
					}
				}
			}
		}
	}
}

// findNextDeclPos finds the position of the first declaration after the given position.
func findNextDeclPos(f *ast.File, after token.Pos) token.Pos {
	for _, decl := range f.Decls {
		if decl.Pos() > after {
			return decl.Pos()
		}
	}
	// Fallback: return the position after the comment.
	return after
}

// checkReadFileYAML detects os.ReadFile or os.Open calls with a YAML file path argument.
// Both literal and const string arguments are resolved.
func checkReadFileYAML(pass *analysis.Pass, call *ast.CallExpr) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok || !isPkgPath(pass, pkgIdent, "os") {
		return
	}

	fnName := sel.Sel.Name
	if fnName != "ReadFile" && fnName != "Open" {
		return
	}

	if len(call.Args) == 0 {
		return
	}

	path, ok := extractStringOrConstValue(pass, call.Args[0])
	if !ok {
		return
	}

	if !isYAMLPath(path) {
		return
	}

	pass.Report(analysis.Diagnostic{
		Pos:      call.Pos(),
		Category: checkEmbeddedYAML,
		URL:      embeddedYAMLURL,
		Message: fmt.Sprintf(
			"os.%s of YAML file %q; consider using typed Kubernetes structs from k8s.io/api",
			fnName, path,
		),
	})
}

// isYAMLPath returns true if the path ends in .yaml or .yml.
func isYAMLPath(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}
