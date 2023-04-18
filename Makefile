.PHONY: \
	tools \
	test \
	coverage \
	fmt \
	doc \

all: tools fmt test

print-%:
	@echo $* = $($*)

clean:
	rm -f coverage.txt

tools:
	go get github.com/axw/gocov/gocov
	go get github.com/matm/gocov-html
	go get github.com/golangci/golangci-lint/cmd/golangci-lint
	go get github.com/gordonklaus/ineffassign
	go get honnef.co/go/tools/cmd/staticcheck
	go get github.com/client9/misspell/cmd/misspell

test:  test_ineffassign    \
       test_misspell       \
       test_govet          \
       test_gotest_race    \
       test_gotest_cover

fmt:
	go fmt ./...

coverage:
	gocov test ./... > $(CURDIR)/coverage.out 2>/dev/null
	gocov report $(CURDIR)/coverage.out
	if test -z "$$CI"; then \
	  gocov-html $(CURDIR)/coverage.out > $(CURDIR)/coverage.html; \
	  if which open &>/dev/null; then \
	    open $(CURDIR)/coverage.html; \
	  fi; \
	fi

test_ineffassign:
	@echo "test: ineffassign"
	@find ./ -type f -name "*.go" -not -path "./vendor/*" -exec ineffassign {} \; || (echo "ineffassign failed"; exit 1)
	@echo "test: ok"

test_misspell:
	@echo "test: misspell"
	@find ./ -type f -name "*.go" -not -path "./vendor/*" -exec misspell {} \; || (echo "misspell failed"; exit 1)
	@echo "test: ok"

test_govet:
	@echo "test: go vet"
	@go vet ./... || (echo "go vet failed"; exit 1)
	@echo "test: ok"

test_gosec:
	@echo "test: gosec"
	@gosec -exclude=G107,G110 ./... || (echo "gosec failed"; exit 1)
	@echo "test: ok"

test_gotest_race:
	@echo "test: go test -race"
	@go test -race -coverprofile=coverage.txt -covermode=atomic ./... || (echo "go test -race failed"; exit 1)
	@echo "test: ok"

test_gotest_cover:
	@echo "test: go test -cover"
	@go test -cover ./... || (echo "go test -cover failed"; exit 1)
	@echo "test: ok"

testbadge:
	@echo "Running tests to update readme with badge coverage"
	@go tool cover -func=coverage.out -o=coverage.out
	@gobadge -filename=coverage.out -link https://github.com/tmc/langchaingo/actions/workflows/coverage.yml

doc:
	@echo "doc: http://localhost:8080/pkg/github.com/tmc/langchaingo"
	godoc -http=:8080 -index