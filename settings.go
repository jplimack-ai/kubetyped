package kubetypes

import (
	"fmt"
	"slices"
)

// Settings holds configuration for the kube-types linter.
type Settings struct {
	// IncludeTestFiles controls whether test files are analyzed.
	// Default false — test files are skipped.
	IncludeTestFiles bool `json:"include_test_files"`

	// ExtraKnownGVKs allows users to register additional GVKs beyond the built-in table.
	ExtraKnownGVKs []GVKEntry `json:"extra_known_gvks"`

	// IgnoreGVKs lists GVK keys ("apiVersion/kind") that should never produce diagnostics.
	IgnoreGVKs []string `json:"ignore_gvks"`

	// Checks configures which checks are enabled and their per-check settings.
	// nil or empty map means all checks enabled with defaults.
	Checks map[string]CheckConfig `json:"checks"`
}

// CheckConfig holds per-check configuration.
type CheckConfig struct {
	// Enabled controls whether this check runs. nil = enabled (default on).
	Enabled *bool `json:"enabled"`

	// AdditionalMarkers (sprintf_yaml only) extends the default YAML markers list.
	AdditionalMarkers []string `json:"additional_markers"`
}

// GVKEntry represents a known Kubernetes GroupVersionKind and its typed struct info.
type GVKEntry struct {
	APIVersion   string `json:"api_version"`
	Kind         string `json:"kind"`
	TypedPackage string `json:"typed_package"` // e.g. "k8s.io/api/apps/v1.Deployment"
}

const (
	checkMapLiteral      = "map_literal"
	checkSprintfYAML     = "sprintf_yaml"
	checkUnstructuredGVK = "unstructured_gvk"
	checkRawGVKString    = "raw_gvk_string"
)

var allChecks = []string{checkMapLiteral, checkSprintfYAML, checkUnstructuredGVK, checkRawGVKString}

// enabledChecks returns the set of enabled check names based on settings.
func (s *Settings) enabledChecks() map[string]bool {
	// nil/empty Checks map → all enabled.
	if len(s.Checks) == 0 {
		m := make(map[string]bool, len(allChecks))
		for _, c := range allChecks {
			m[c] = true
		}
		return m
	}

	m := make(map[string]bool, len(allChecks))
	for _, c := range allChecks {
		cfg, configured := s.Checks[c]
		if !configured {
			// Not mentioned in config → enabled by default.
			m[c] = true
			continue
		}
		if cfg.Enabled == nil || *cfg.Enabled {
			m[c] = true
		}
	}
	return m
}

// validateChecks returns an error if any configured check name is invalid.
func (s *Settings) validateChecks() error {
	for name := range s.Checks {
		if !slices.Contains(allChecks, name) {
			return fmt.Errorf("unknown check %q; valid checks: %v", name, allChecks)
		}
	}
	return nil
}

// validateExtraGVKs returns an error if any ExtraKnownGVKs entry has empty fields.
func (s *Settings) validateExtraGVKs() error {
	for i, entry := range s.ExtraKnownGVKs {
		if entry.APIVersion == "" {
			return fmt.Errorf("extra_known_gvks[%d]: api_version must not be empty", i)
		}
		if entry.Kind == "" {
			return fmt.Errorf("extra_known_gvks[%d]: kind must not be empty", i)
		}
		if entry.TypedPackage == "" {
			return fmt.Errorf("extra_known_gvks[%d]: typed_package must not be empty", i)
		}
	}
	return nil
}

// markersForSprintfYAML returns the combined list of default + additional markers.
func (s *Settings) markersForSprintfYAML() []string {
	cfg, ok := s.Checks[checkSprintfYAML]
	if !ok || len(cfg.AdditionalMarkers) == 0 {
		return defaultYAMLMarkers
	}
	markers := make([]string, len(defaultYAMLMarkers)+len(cfg.AdditionalMarkers))
	copy(markers, defaultYAMLMarkers)
	copy(markers[len(defaultYAMLMarkers):], cfg.AdditionalMarkers)
	return markers
}

// isGVKIgnored returns true if the given apiVersion/kind is in the IgnoreGVKs list.
func (s *Settings) isGVKIgnored(apiVersion, kind string) bool {
	return slices.Contains(s.IgnoreGVKs, gvkKey(apiVersion, kind))
}
