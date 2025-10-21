.PHONY: start
start:
	pnpm run start

.PHONY: build
build:
	pnpm run build

.PHONY: lint
lint:
	pnpm run lint

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

.PHONY: docker-build
docker-build: docker-setup
	docker build -t langchaingo-docs .

.PHONY: docker-run
docker-run: docker-build
	@echo "Starting documentation server at http://localhost:3000"
	docker run --rm -p 3000:3000 -v $(PWD):/app langchaingo-docs

.PHONY: docker-dev
docker-dev: docker-setup
	@echo "Starting development environment with live reload..."
	docker run --rm -it -p 3000:3000 -v $(PWD):/app -w /app node:18-alpine sh -c "\
		npm install -g pnpm && \
		pnpm install && \
		pnpm run start --host 0.0.0.0"

.PHONY: docker-clean
docker-clean:
	@echo "Cleaning up Docker resources..."
	-docker rmi langchaingo-docs
	-rm -f Dockerfile
