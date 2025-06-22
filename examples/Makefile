SYNC_REF?=latest

.PHONY: sync
sync: ## Sync all go.mod files
	@sh .update-all-to-latest.sh ${SYNC_REF}

.PHONY: install-all
install-all: ## Add go work files to all examples
	@find . -name go.mod -execdir go install \;

.PHONY: tidy-all
tidy-all: ## Tidy all go.mod files
	@find . -name go.mod -execdir go mod tidy \;

.PHONY: fmt-all
fmt-all: ## Run go fmt on all examples
	@find . -name go.mod -execdir go fmt \;
