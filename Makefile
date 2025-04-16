# This file contains convenience targets for the project.
# It is not intended to be used as a build system.
# See the README for more information.

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint: lint-deps
	golangci-lint run --color=always --sort-results ./...

.PHONY: lint-exp
lint-exp:
	golangci-lint run --fix --config .golangci-exp.yaml ./...

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix --skip-dirs=./exp ./...

.PHONY: lint-all
lint-all:
	golangci-lint run --color=always --sort-results ./...

.PHONY: lint-deps
lint-deps:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo >&2 "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0; \
	}

.PHONY: docs
docs:
	@echo "Generating documentation..."
	$(MAKE) -C docs build

.PHONY: test-race
test-race:
	go run test -race ./...

.PHONY: test-cover
test-cover:
	go run test -cover ./...

.PHONY: run-pkgsite
run-pkgsite:
	go run golang.org/x/pkgsite/cmd/pkgsite@latest

.PHONY: clean
clean: clean-lint-cache

.PHONY: clean-lint-cache
clean-lint-cache:
	golangci-lint cache clean

.PHONY: build-examples
build-examples:
	for example in $(shell find ./examples -mindepth 1 -maxdepth 1 -type d); do \
		(cd $$example; echo Build $$example; go mod tidy; go build -o /dev/null) || exit 1; done

.PHONY: add-go-work
add-go-work:
	go work init .
	go work use -r .
