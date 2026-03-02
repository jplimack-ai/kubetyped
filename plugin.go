package kubetypes

import (
	"fmt"
	"maps"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("kube-types", New)
}

// New creates a new kube-types linter plugin instance.
func New(settings any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, fmt.Errorf("decoding kube-types settings: %w", err)
	}

	if err := s.validateChecks(); err != nil {
		return nil, err
	}

	if err := s.validateExtraGVKs(); err != nil {
		return nil, err
	}

	if err := s.validateGVKKeys(); err != nil {
		return nil, err
	}

	table := buildGVKTable(s.ExtraKnownGVKs)
	return &plugin{settings: s, gvkTable: table}, nil
}

// buildGVKTable creates a GVK lookup table by copying knownGVK and adding extra entries.
func buildGVKTable(extras []GVKEntry) map[string]gvkInfo {
	table := make(map[string]gvkInfo, len(knownGVK)+len(extras))
	maps.Copy(table, knownGVK)
	for _, entry := range extras {
		table[gvkKey(entry.APIVersion, entry.Kind)] = parseGVKEntry(entry)
	}
	return table
}

// parseGVKEntry converts a GVKEntry into a gvkInfo by splitting the TypedPackage.
// Example: "k8s.io/api/apps/v1.Deployment" splits into ImportPath="k8s.io/api/apps/v1", ShortName="v1.Deployment".
func parseGVKEntry(entry GVKEntry) gvkInfo {
	tp := entry.TypedPackage
	if idx := strings.LastIndex(tp, "."); idx >= 0 {
		importPath := tp[:idx]
		typeName := tp[idx+1:]
		alias := importPath
		if slashIdx := strings.LastIndex(importPath, "/"); slashIdx >= 0 {
			alias = importPath[slashIdx+1:]
		}
		return gvkInfo{
			ShortName:  alias + "." + typeName,
			ImportPath: importPath,
		}
	}
	return gvkInfo{ShortName: tp, ImportPath: tp}
}

type plugin struct {
	settings Settings
	gvkTable map[string]gvkInfo
}

// GetLoadMode returns the load mode required by the plugin.
func (p *plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

// BuildAnalyzers returns the analyzers for the plugin.
func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		newAnalyzer(p),
	}, nil
}
