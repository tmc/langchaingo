# This file contains convenience targets for the project.
# It is not intended to be used as a build system.
# See the README for more information.

.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Testing:"
	@echo "  test           - Run all tests with basic environment setup"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-cover     - Run tests with coverage reporting"
	@echo "  test-record    - Run tests with re-recording of httprr files"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint           - Run linter with auto-installation if needed"
	@echo "  lint-fix       - Run linter with automatic fixes"
	@echo ""
	@echo "Other:"
	@echo "  build-examples - Build all example projects to verify they compile"
	@echo "  docs           - Generate documentation"
	@echo "  clean          - Clean lint cache"
	@echo "  help           - Show this help message"

.PHONY: test
test:
	DOCKER_HOST=$$(docker context inspect -f='{{.Endpoints.docker.Host}}' 2>/dev/null || echo "unix:///var/run/docker.sock") \
	TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE="/var/run/docker.sock" \
	go test ./...

.PHONY: lint
lint: lint-deps
	golangci-lint run --color=always ./...

.PHONY: lint-exp
lint-exp:
	golangci-lint run --fix --config .golangci-exp.yaml ./...

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix --skip-dirs=./exp ./...

.PHONY: lint-all
lint-all:
	golangci-lint run --color=always ./...

.PHONY: lint-deps
lint-deps:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo >&2 "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6; \
		command -v golangci-lint >/dev/null 2>&1 || { \
			echo >&2 "Failed to detect golangci-lint after installation. Please check your Go installation and PATH."; \
			exit 1; \
		} \
	}
	@golangci-lint version | grep -q "version v2" || { echo "Error: golangci-lint v2.x.x required, found:" && golangci-lint version && exit 1; }

.PHONY: docs
docs:
	@echo "Generating documentation..."
	$(MAKE) -C docs build

.PHONY: test-race
test-race:
	DOCKER_HOST=$$(docker context inspect -f='{{.Endpoints.docker.Host}}' 2>/dev/null || echo "unix:///var/run/docker.sock") \
	TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE="/var/run/docker.sock" \
	go test -race ./...

.PHONY: test-cover
test-cover:
	DOCKER_HOST=$$(docker context inspect -f='{{.Endpoints.docker.Host}}' 2>/dev/null || echo "unix:///var/run/docker.sock") \
	TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE="/var/run/docker.sock" \
	go test -cover ./...

.PHONY: test-record
test-record:
	@echo "Re-recording HTTP interactions for all packages using httprr..."
	PACKAGES=$$(go run ./internal/devtools/rrtool list-packages -format=paths) && \
	echo "Recording HTTP interactions for packages:" && \
	echo "$$PACKAGES" | tr ' ' '\n' | sed 's/^/  /' && \
	echo "" && \
	env DOCKER_HOST=$$(docker context inspect -f='{{.Endpoints.docker.Host}}' 2>/dev/null || echo "unix:///var/run/docker.sock") \
	TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE="/var/run/docker.sock" \
	go test $$PACKAGES -httprecord=. -httprecord-delay=1s -timeout=300s


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

.PHONY: lint-devtools
lint-devtools:
	go run ./internal/devtools/lint -v

.PHONY: lint-devtools-fix
lint-devtools-fix:
	go run ./internal/devtools/lint -fix -v

.PHONY: lint-architecture
lint-architecture:
	go run ./internal/devtools/lint -architecture -v

.PHONY: lint-prepush
lint-prepush:
	go run ./internal/devtools/lint -prepush -v

.PHONY: lint-prepush-fix
lint-prepush-fix:
	go run ./internal/devtools/lint -prepush -fix -v
