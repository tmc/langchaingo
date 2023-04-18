# This file contains convenience targets for the project.
# It is not intended to be used as a build system.
# See the README for more information.

.PHONY: test
test:
	go run test ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test-race
test-race:
	go run test -race ./...

.PHONY: test-cover
test-cover:
	go run test -cover ./...

