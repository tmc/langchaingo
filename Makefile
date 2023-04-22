# This file contains convenience targets for the project.
# It is not intended to be used as a build system.
# See the README for more information.

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run --color=always --sort-results --skip-dirs=exp ./...

.PHONY: lint-exp
lint-exp:
	golangci-lint run --fix --config .golangci-exp.yaml ./...

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix --skip-dirs=exp ./...

.PHONY: lint-all
lint-all:
	golangci-lint run --color=always --sort-results ./...

.PHONY: test-race
test-race:
	go run test -race ./...

.PHONY: test-cover
test-cover:
	go run test -cover ./...

.PHONY: run-pkgsite
run-pkgsite:
	go run golang.org/x/pkgsite/cmd/pkgsite@latest
