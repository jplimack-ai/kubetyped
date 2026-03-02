package kubetypes

import (
	"go/ast"
	"go/token"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// testAnalyzer creates an analyzer from settings using the same path as New().
func testAnalyzer(t *testing.T, s Settings) *plugin {
	t.Helper()

	if err := s.validateChecks(); err != nil {
		t.Fatalf("invalid settings: %v", err)
	}
	if err := s.validateExtraGVKs(); err != nil {
		t.Fatalf("invalid settings: %v", err)
	}

	return &plugin{settings: s, gvkTable: buildGVKTable(s.ExtraKnownGVKs)}
}

func TestMapLiteral(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "map_literal")
}

func TestSprintfYAML(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "sprintf_yaml")
}

func TestFalsePositives(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "false_positives")
}

func TestUnstructuredGVK(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "unstructured_gvk")
}

func TestInvalidCheckName(t *testing.T) {
	s := Settings{
		Checks: map[string]CheckConfig{
			"map_literals": {}, // typo
		},
	}
	if err := s.validateChecks(); err == nil {
		t.Fatal("expected error for invalid check name, got nil")
	}
}

func TestInvalidExtraGVK_EmptyAPIVersion(t *testing.T) {
	s := Settings{
		ExtraKnownGVKs: []GVKEntry{
			{APIVersion: "", Kind: "Widget", TypedPackage: "example.io/v1.Widget"},
		},
	}
	if err := s.validateExtraGVKs(); err == nil {
		t.Fatal("expected error for empty api_version, got nil")
	}
}

func TestInvalidExtraGVK_EmptyKind(t *testing.T) {
	s := Settings{
		ExtraKnownGVKs: []GVKEntry{
			{APIVersion: "example.io/v1", Kind: "", TypedPackage: "example.io/v1.Widget"},
		},
	}
	if err := s.validateExtraGVKs(); err == nil {
		t.Fatal("expected error for empty kind, got nil")
	}
}

func TestInvalidExtraGVK_EmptyTypedPackage(t *testing.T) {
	s := Settings{
		ExtraKnownGVKs: []GVKEntry{
			{APIVersion: "example.io/v1", Kind: "Widget", TypedPackage: ""},
		},
	}
	if err := s.validateExtraGVKs(); err == nil {
		t.Fatal("expected error for empty typed_package, got nil")
	}
}

func TestCheckEnabled_DefaultNil(t *testing.T) {
	// nil Enabled means default-on.
	s := Settings{
		Checks: map[string]CheckConfig{
			checkMapLiteral: {Enabled: nil},
		},
	}
	enabled := s.enabledChecks()
	if !enabled[checkMapLiteral] {
		t.Fatal("expected map_literal to be enabled with nil Enabled")
	}
	if !enabled[checkSprintfYAML] {
		t.Fatal("expected sprintf_yaml to be enabled (not mentioned in config)")
	}
}

func TestExtraKnownGVKs(t *testing.T) {
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		ExtraKnownGVKs: []GVKEntry{
			{APIVersion: "example.io/v1", Kind: "Widget", TypedPackage: "example.io/api/v1.Widget"},
		},
	})
	// Run against map_literal_extra which has a Widget GVK.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "map_literal_extra")
}

func TestIgnoreGVKs(t *testing.T) {
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		IgnoreGVKs:       []string{"apps/v1/Deployment"},
	})
	// Run against a fixture where Deployment is ignored.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "ignore_gvk")
}

func TestAllChecksEnabledByDefault(t *testing.T) {
	s := Settings{}
	enabled := s.enabledChecks()
	for _, c := range allChecks {
		if !enabled[c] {
			t.Fatalf("expected check %q to be enabled by default", c)
		}
	}
}

func TestEmptyChecksMap_AllEnabled(t *testing.T) {
	s := Settings{Checks: map[string]CheckConfig{}}
	enabled := s.enabledChecks()
	for _, c := range allChecks {
		if !enabled[c] {
			t.Fatalf("expected check %q to be enabled with empty checks map", c)
		}
	}
}

func TestParseGVKEntry(t *testing.T) {
	tests := []struct {
		name       string
		entry      GVKEntry
		wantShort  string
		wantImport string
	}{
		{
			name:       "standard k8s path",
			entry:      GVKEntry{TypedPackage: "k8s.io/api/apps/v1.Deployment"},
			wantShort:  "v1.Deployment",
			wantImport: "k8s.io/api/apps/v1",
		},
		{
			name:       "custom CRD path",
			entry:      GVKEntry{TypedPackage: "example.io/api/v1.Widget"},
			wantShort:  "v1.Widget",
			wantImport: "example.io/api/v1",
		},
		{
			name:       "no dot separator",
			entry:      GVKEntry{TypedPackage: "something"},
			wantShort:  "something",
			wantImport: "something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parseGVKEntry(tt.entry)
			if info.ShortName != tt.wantShort {
				t.Errorf("ShortName = %q, want %q", info.ShortName, tt.wantShort)
			}
			if info.ImportPath != tt.wantImport {
				t.Errorf("ImportPath = %q, want %q", info.ImportPath, tt.wantImport)
			}
		})
	}
}

func TestMarkersForSprintfYAML(t *testing.T) {
	t.Run("default markers", func(t *testing.T) {
		s := Settings{}
		markers := s.markersForSprintfYAML()
		if len(markers) != 2 {
			t.Fatalf("expected 2 markers, got %d", len(markers))
		}
		if markers[0] != "apiVersion:" {
			t.Fatalf("markers[0] = %q, want %q", markers[0], "apiVersion:")
		}
		if markers[1] != "kind:" {
			t.Fatalf("markers[1] = %q, want %q", markers[1], "kind:")
		}
	})

	t.Run("additional markers", func(t *testing.T) {
		s := Settings{
			Checks: map[string]CheckConfig{
				checkSprintfYAML: {
					AdditionalMarkers: []string{"metadata:"},
				},
			},
		}
		markers := s.markersForSprintfYAML()
		if len(markers) != 3 {
			t.Fatalf("expected 3 markers, got %d", len(markers))
		}
		if markers[0] != "apiVersion:" {
			t.Fatalf("markers[0] = %q, want %q", markers[0], "apiVersion:")
		}
		if markers[1] != "kind:" {
			t.Fatalf("markers[1] = %q, want %q", markers[1], "kind:")
		}
		if markers[2] != "metadata:" {
			t.Fatalf("markers[2] = %q, want %q", markers[2], "metadata:")
		}
	})
}

func TestIsGVKIgnored(t *testing.T) {
	s := Settings{
		IgnoreGVKs: []string{"apps/v1/Deployment", "v1/Pod"},
	}

	if !s.isGVKIgnored("apps/v1", "Deployment") {
		t.Fatal("expected apps/v1 Deployment to be ignored")
	}
	if !s.isGVKIgnored("v1", "Pod") {
		t.Fatal("expected v1 Pod to be ignored")
	}
	if s.isGVKIgnored("v1", "Service") {
		t.Fatal("expected v1 Service to NOT be ignored")
	}
}

func TestMapLiteralConst(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "map_literal_const")
}

func TestMapLiteralNamed(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "map_literal_named")
}

func TestCheckDisabled_MapLiteral(t *testing.T) {
	disabled := false
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		Checks: map[string]CheckConfig{
			checkMapLiteral: {Enabled: &disabled},
		},
	})
	// only_map_literal triggers ONLY map_literal. Disabling it → zero diagnostics.
	// If disable is broken, unexpected diagnostics fail the test.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "only_map_literal")
}

func TestCheckDisabled_SprintfYAML(t *testing.T) {
	disabled := false
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		Checks: map[string]CheckConfig{
			checkSprintfYAML: {Enabled: &disabled},
		},
	})
	// only_sprintf_yaml triggers ONLY sprintf_yaml. Disabling it → zero diagnostics.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "only_sprintf_yaml")
}

func TestSprintfAdditionalMarkers(t *testing.T) {
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		Checks: map[string]CheckConfig{
			checkSprintfYAML: {
				AdditionalMarkers: []string{"metadata:"},
			},
		},
	})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "sprintf_markers")
}

func TestRawGVKString(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "raw_gvk_string")
}

func TestRawGVKStringDisabled(t *testing.T) {
	disabled := false
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		Checks: map[string]CheckConfig{
			checkRawGVKString: {Enabled: &disabled},
		},
	})
	// only_raw_gvk triggers ONLY raw_gvk_string. Disabling it → zero diagnostics.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "only_raw_gvk")
}

func TestUnstructuredIgnoreGVK(t *testing.T) {
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		IgnoreGVKs:       []string{"apps/v1/Deployment"},
	})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "unstructured_ignore")
}

func TestPluginInterface(t *testing.T) {
	p, err := New(map[string]any{})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	lp := p.(*plugin)
	if mode := lp.GetLoadMode(); mode != "typesinfo" {
		t.Fatalf("GetLoadMode() = %q, want %q", mode, "typesinfo")
	}

	analyzers, err := lp.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() failed: %v", err)
	}
	if len(analyzers) != 1 {
		t.Fatalf("BuildAnalyzers() returned %d analyzers, want 1", len(analyzers))
	}
	if analyzers[0].Name != "kubetypes" {
		t.Fatalf("analyzer Name = %q, want %q", analyzers[0].Name, "kubetypes")
	}
}

func TestNewWithInvalidSettings(t *testing.T) {
	// Invalid settings type should fail decoding.
	_, err := New("invalid")
	if err == nil {
		t.Fatal("expected error for invalid settings, got nil")
	}
}

func TestGVKTableNotMutated(t *testing.T) {
	// Verify that creating a plugin does not mutate the global knownGVK map.
	origLen := len(knownGVK)

	_, err := New(map[string]any{
		"extra_known_gvks": []map[string]any{
			{
				"api_version":   "test.io/v1",
				"kind":          "TestResource",
				"typed_package": "test.io/api/v1.TestResource",
			},
		},
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if len(knownGVK) != origLen {
		t.Fatalf("global knownGVK was mutated: had %d entries, now has %d", origLen, len(knownGVK))
	}
}

func TestAllChecksDisabled(t *testing.T) {
	disabled := false
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		Checks: map[string]CheckConfig{
			checkMapLiteral:      {Enabled: &disabled},
			checkSprintfYAML:     {Enabled: &disabled},
			checkUnstructuredGVK: {Enabled: &disabled},
			checkRawGVKString:    {Enabled: &disabled},
		},
	})
	// All checks disabled — disabled_check triggers all 4, expects zero diagnostics.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "disabled_check")
}

func TestCheckDisabled_UnstructuredGVK(t *testing.T) {
	disabled := false
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		Checks: map[string]CheckConfig{
			checkUnstructuredGVK: {Enabled: &disabled},
		},
	})
	// only_unstructured triggers ONLY unstructured_gvk. Disabling it → zero diagnostics.
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "only_unstructured")
}

func TestIncludeTestFiles_On(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: true})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "test_file_on")
}

func TestIncludeTestFiles_Off(t *testing.T) {
	p := testAnalyzer(t, Settings{IncludeTestFiles: false})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "test_file_off")
}

func TestRawGVKStringIgnoreGVK(t *testing.T) {
	p := testAnalyzer(t, Settings{
		IncludeTestFiles: true,
		IgnoreGVKs:       []string{"apps/v1/Deployment"},
	})
	analysistest.Run(t, analysistest.TestData(), newAnalyzer(p), "raw_gvk_ignore")
}

func TestGVKKey(t *testing.T) {
	tests := []struct {
		apiVersion, kind, want string
	}{
		{"apps/v1", "Deployment", "apps/v1/Deployment"},
		{"v1", "Pod", "v1/Pod"},
		{"networking.k8s.io/v1", "Ingress", "networking.k8s.io/v1/Ingress"},
		{"", "Pod", "/Pod"},
		{"v1", "", "v1/"},
		{"", "", "/"},
	}
	for _, tt := range tests {
		if got := gvkKey(tt.apiVersion, tt.kind); got != tt.want {
			t.Errorf("gvkKey(%q, %q) = %q, want %q", tt.apiVersion, tt.kind, got, tt.want)
		}
	}
}

func TestExtractStringValue(t *testing.T) {
	tests := []struct {
		name   string
		expr   ast.Expr
		want   string
		wantOK bool
	}{
		{
			name:   "double-quoted string",
			expr:   &ast.BasicLit{Kind: token.STRING, Value: `"hello"`},
			want:   "hello",
			wantOK: true,
		},
		{
			name:   "backtick string",
			expr:   &ast.BasicLit{Kind: token.STRING, Value: "`world`"},
			want:   "world",
			wantOK: true,
		},
		{
			name:   "int literal",
			expr:   &ast.BasicLit{Kind: token.INT, Value: "42"},
			wantOK: false,
		},
		{
			name:   "ident (not a literal)",
			expr:   &ast.Ident{Name: "x"},
			wantOK: false,
		},
		{
			name:   "malformed string literal",
			expr:   &ast.BasicLit{Kind: token.STRING, Value: `"unterminated`},
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := extractStringValue(tt.expr)
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("extractStringValue() = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestIsGVKMethodCall(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expr
		want bool
	}{
		{
			name: "GetKind() zero args",
			expr: &ast.CallExpr{
				Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "obj"}, Sel: &ast.Ident{Name: "GetKind"}},
			},
			want: true,
		},
		{
			name: "GetAPIVersion() zero args",
			expr: &ast.CallExpr{
				Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "obj"}, Sel: &ast.Ident{Name: "GetAPIVersion"}},
			},
			want: true,
		},
		{
			name: "GetKind with one arg",
			expr: &ast.CallExpr{
				Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "obj"}, Sel: &ast.Ident{Name: "GetKind"}},
				Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"x"`}},
			},
			want: false,
		},
		{
			name: "OtherMethod()",
			expr: &ast.CallExpr{
				Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "obj"}, Sel: &ast.Ident{Name: "OtherMethod"}},
			},
			want: false,
		},
		{
			name: "non-call expression",
			expr: &ast.Ident{Name: "x"},
			want: false,
		},
		{
			name: "direct function call (not selector)",
			expr: &ast.CallExpr{
				Fun: &ast.Ident{Name: "GetKind"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isGVKMethodCall(tt.expr); got != tt.want {
				t.Errorf("isGVKMethodCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMapStringAnyDirect(t *testing.T) {
	tests := []struct {
		name string
		typ  *ast.MapType
		want bool
	}{
		{
			name: "map[string]any",
			typ: &ast.MapType{
				Key:   &ast.Ident{Name: "string"},
				Value: &ast.Ident{Name: "any"},
			},
			want: true,
		},
		{
			name: "map[string]interface{}",
			typ: &ast.MapType{
				Key:   &ast.Ident{Name: "string"},
				Value: &ast.InterfaceType{Methods: &ast.FieldList{}},
			},
			want: true,
		},
		{
			name: "key is SelectorExpr (not Ident)",
			typ: &ast.MapType{
				Key:   &ast.SelectorExpr{X: &ast.Ident{Name: "pkg"}, Sel: &ast.Ident{Name: "String"}},
				Value: &ast.Ident{Name: "any"},
			},
			want: false,
		},
		{
			name: "key is int (not string)",
			typ: &ast.MapType{
				Key:   &ast.Ident{Name: "int"},
				Value: &ast.Ident{Name: "any"},
			},
			want: false,
		},
		{
			name: "value is string (not any)",
			typ: &ast.MapType{
				Key:   &ast.Ident{Name: "string"},
				Value: &ast.Ident{Name: "string"},
			},
			want: false,
		},
		{
			name: "interface with methods",
			typ: &ast.MapType{
				Key: &ast.Ident{Name: "string"},
				Value: &ast.InterfaceType{Methods: &ast.FieldList{
					List: []*ast.Field{{Names: []*ast.Ident{{Name: "Foo"}}}},
				}},
			},
			want: false,
		},
		{
			name: "value is StarExpr (neither Ident nor InterfaceType)",
			typ: &ast.MapType{
				Key:   &ast.Ident{Name: "string"},
				Value: &ast.StarExpr{X: &ast.Ident{Name: "int"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMapStringAnyDirect(tt.typ); got != tt.want {
				t.Errorf("isMapStringAnyDirect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckEnabled_ExplicitTrue(t *testing.T) {
	enabled := true
	s := Settings{
		Checks: map[string]CheckConfig{
			checkMapLiteral: {Enabled: &enabled},
		},
	}
	checks := s.enabledChecks()
	if !checks[checkMapLiteral] {
		t.Fatal("expected map_literal to be enabled with explicit true")
	}
	if !checks[checkSprintfYAML] {
		t.Fatal("expected sprintf_yaml to be enabled (not mentioned in config)")
	}
}

func TestNewWithInvalidCheckName(t *testing.T) {
	_, err := New(map[string]any{
		"checks": map[string]any{
			"map_literals": map[string]any{},
		},
	})
	if err == nil {
		t.Fatal("expected New() to fail for invalid check name, got nil")
	}
}

func TestNewWithInvalidExtraGVK(t *testing.T) {
	_, err := New(map[string]any{
		"extra_known_gvks": []map[string]any{
			{
				"api_version":   "",
				"kind":          "Widget",
				"typed_package": "example.io/v1.Widget",
			},
		},
	})
	if err == nil {
		t.Fatal("expected New() to fail for empty api_version, got nil")
	}
}
