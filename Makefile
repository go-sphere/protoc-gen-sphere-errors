MODULE := $(shell go list -m)

TESTDATA := generate/errors/testdata

# Compile the test fixtures into committed descriptor sets. Only this target
# needs buf; `go test` runs against the committed *.pb files.
#
# buf resolves dependencies from $(TESTDATA)/buf.lock (no vendored protos),
# emits a FileDescriptorSet that bundles every import (so protogen can resolve
# the sphere.errors extensions), and keeps source info by default.
.PHONY: testdata
testdata:
	@mkdir -p $(TESTDATA)/pb
	@for p in $(TESTDATA)/proto/*.proto; do \
		name=$$(basename $$p .proto); \
		echo "building $$p -> $(TESTDATA)/pb/$$name.pb"; \
		buf build $(TESTDATA) --path $$p --as-file-descriptor-set \
			-o $(TESTDATA)/pb/$$name.pb || exit 1; \
	done

.PHONY: test
test: testdata
	go test ./...

# Regenerate golden files. Run after intentionally changing the template or
# generation logic, then review the diff before committing.
.PHONY: update-golden
update-golden: testdata
	go test ./generate/errors/ -run TestGolden -update-golden

.PHONY: lint
lint:
	go fix ./...
	go fmt ./...
	go vet ./...
	go get ./...
	go test ./...
	go mod tidy
	golangci-lint fmt --no-config --enable gofmt,goimports
	golangci-lint run --no-config --fix
	nilaway -include-pkgs="$(MODULE)" ./...
