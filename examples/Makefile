.PHONY: sync
sync: ## Sync all go.mod files
	@sh .update-all-to-latest.sh

.PHONY: add-go-work-files
add-go-work-files: ## Add go work files to all examples
	find . -name go.mod -execdir sh -c 'go work init && go work use . && go work use ../..' \;

.PHONY: rm-go-work-files
rm-go-work-files: ## Add go work files to all examples
	rm -f ./*/go.work*

.PHONY: install-all
install-all: ## Add go work files to all examples
	@find . -name go.mod -execdir go install \;
