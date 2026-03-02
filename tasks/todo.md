# kube-types golangci-lint Plugin

## Implementation

- [x] 1. Scaffold: go.mod, Makefile, .golangci.yml
- [x] 2. settings.go — configuration
- [x] 3. known_gvks.go — GVK lookup table (~35 entries)
- [x] 4. plugin.go — plugin registration
- [x] 5. analyzer.go — analyzer wiring + helpers
- [x] 6. check_map_literal.go — map literal detection
- [x] 7. check_sprintf.go — sprintf YAML detection
- [x] 8. check_unstructured.go — unstructured GVK detection
- [x] 9. Test fixtures + plugin_test.go
- [x] 10. .golangci.example.yml

## Verification

- [x] `go build ./...` compiles cleanly
- [x] `make test` — 3/3 tests pass
- [x] `make lint` — 0 issues
