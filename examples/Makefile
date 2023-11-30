SYNC_REF?=main

.PHONY: sync
sync: ## Sync all go.mod files
	@sh .update-all-to-latest.sh ${SYNC_REF}

.PHONY: install-all
install-all: ## Add go work files to all examples
	@find . -name go.mod -execdir go install \;
