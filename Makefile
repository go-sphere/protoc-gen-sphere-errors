MODULE := $(shell go list -m)

ERRORS_PROTO_DIR := $(shell go list -m -f '{{.Dir}}' github.com/go-sphere/errors)/proto
TESTDATA_DIR := generate/errors/testdata
TESTDATA_PROTO_DIR := $(TESTDATA_DIR)/proto
TESTDATA_PB_DIR := $(TESTDATA_DIR)/pb

# Compile test .proto files into FileDescriptorSet (.pb) fixtures used by the
# integration and golden tests. Run after editing any testdata proto.
.PHONY: testdata
testdata:
	@mkdir -p $(TESTDATA_PB_DIR)
	@for proto in $(wildcard $(TESTDATA_PROTO_DIR)/*.proto); do \
		name=$$(basename $$proto .proto); \
		echo "Compiling $$proto -> $(TESTDATA_PB_DIR)/$$name.pb"; \
		protoc --proto_path=$(TESTDATA_PROTO_DIR) \
			--proto_path=$(ERRORS_PROTO_DIR) \
			--descriptor_set_out=$(TESTDATA_PB_DIR)/$$name.pb \
			--include_imports \
			$$proto || exit 1; \
	done

.PHONY: test
test:
	go test ./...

# Regenerate golden files. Run after intentionally changing the template or
# generation logic, then review the diff before committing.
.PHONY: update-golden
update-golden:
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
