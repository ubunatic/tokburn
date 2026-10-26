.PHONY: ⚙️ 🤖
_prim := \033[36m
_rst  := \033[0m

⚙️:
🤖:

help: 🤖  # show this help
	@grep -E '^[a-zA-Z_-]+:.*[⚙🤖].*#+' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*#+ "}; {printf "    $(_prim)%-15s$(_rst) %s\n", $$1, $$2}'

build: ⚙️  # build all packages
	go build ./...

install: ⚙️  # install the tokburn binary
	go install .

test: ⚙️  # run vet and unit tests
	go vet ./...
	go test ./...

check: test ⚙️  # all checks (alias for test)

dist: ⚙️  # cross-build release binaries into dist/ (used by harnez release)
	rm -rf dist
	for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		os=$${target%/*}; arch=$${target#*/}; ext=""; \
		if test "$$os" = windows; then ext=".exe"; fi; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -o dist/tokburn-$$os-$$arch$$ext . || exit 1; \
	done
