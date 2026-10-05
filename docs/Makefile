# The site is built with md2html (https://github.com/tmc/md2html).
# The publish-docs workflow installs this version with "make install".
MD2HTML_VERSION ?= v0.0.0-20260923024514-2cf8d540c613
MD2HTML ?= md2html
MDVET ?= mdvet
SITE ?= _site
ADDR ?= :8080
SITE_URL ?= https://tmc.github.io/langchaingo/docs

MD2HTML_FLAGS = -nav -search -title LangChainGo

.PHONY: serve
serve:
	$(MD2HTML) $(MD2HTML_FLAGS) -C docs -http $(ADDR)

.PHONY: build
build:
	rm -rf $(SITE)
	mkdir -p $(SITE)
	cp -R static/. $(SITE)/
	$(MD2HTML) $(MD2HTML_FLAGS) -html-layout dir \
		-site-url $(SITE_URL) \
		-edit-url 'https://github.com/tmc/langchaingo/edit/main/docs/docs/{path}' \
		-html $(SITE)/docs docs

.PHONY: vet
vet:
	$(MDVET) -c links,anchors,case,assets,frontmatter,components docs

.PHONY: install
install:
	go install github.com/tmc/md2html/cmd/...@$(MD2HTML_VERSION)

.PHONY: clean
clean:
	rm -rf $(SITE)

.PHONY: lint-deps
lint-deps:
	@if ! command -v vale >/dev/null 2>&1; then \
		echo "Vale not found. Installing via brew..."; \
		if command -v brew >/dev/null 2>&1; then \
			brew install vale; \
		else \
			echo "Error: brew not found. Please install Vale manually: https://vale.sh/docs/vale-cli/installation/"; \
			exit 1; \
		fi; \
	else \
		echo "Vale is already installed"; \
	fi

.PHONY: lint-docs
lint-docs: lint-deps
	vale docs
