# Contributing to LangChainGo

Thank you for your interest in contributing to LangChainGo! This guide helps you get started.

## Ways to contribute

### 1. Code contributions

- **Bug fixes**: Help us squash bugs and improve stability
- **New features**: Implement new LLM providers, tools, or chains
- **Performance improvements**: Optimize existing code
- **Tests**: Improve test coverage and add missing tests

### 2. Documentation contributions

- **Tutorials**: Write step-by-step guides for common use cases
- **How-to guides**: Create practical solutions for specific problems
- **API documentation**: Improve code comments and examples
- **Conceptual guides**: Explain architectural decisions and patterns

### 3. Community support

- **Answer questions**: Help others in GitHub Discussions
- **Report issues**: File detailed bug reports
- **Review PRs**: Provide feedback on pull requests
- **Share examples**: Showcase your LangChainGo projects

## Getting started

### Development setup

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/langchaingo.git
   cd langchaingo
   ```

3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/tmc/langchaingo.git
   ```

4. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Code style

- Follow standard [Go conventions and idioms](https://go.dev/doc/effective_go)
- Run `go fmt` before committing
- Ensure all tests pass with `go test ./...`
- Add tests for new functionality
- Use package-prefixed commit messages (see PR Guidelines below)
- Keep commits focused to a single topic

### Testing

When contributing code that interacts with external APIs:

1. Use the internal `httprr` tool for recording HTTP interactions
2. Never commit real API keys or secrets
3. Ensure tests can run without external dependencies
4. See the [Architecture Guide](/docs/concepts/architecture#http-testing-with-httprr) for details

## Contribution process

1. **Check existing issues**: Look for existing issues or discussions about your idea
2. **Open an issue**: For significant changes, open an issue to discuss first
3. **Make changes**: Implement your changes in a feature branch
4. **Follow commit style**: Use Go-style package-prefixed commit messages
5. **Test thoroughly**: Ensure all tests pass and add new ones as needed
6. **Submit PR**: Open a pull request with a clear description following our guidelines
7. **Address feedback**: Respond to review comments promptly

## Pull request guidelines

### PR title format

**Use Go-style package-prefixed commit messages** following the [Go Contribute Guidelines](https://go.dev/doc/contribute#commit_messages):

- `memory: add interfaces for custom storage backends`
- `llms/openai: fix streaming response handling`
- `chains: implement conversation chain with memory`
- `vectorstores/chroma: add support for metadata filtering`
- `docs: update getting started guide for new API`
- `agents: add tool calling support for GPT-4`
- `examples: add RAG implementation tutorial`

**Format**: `package: description in lowercase without period`

Examples of good commit messages:
- `llms/anthropic: implement function calling support`
- `memory: fix buffer overflow in conversation memory`
- `tools: add calculator tool with error handling`
- `all: update dependencies and organize go.mod file`

### PR description
Include:
- Summary of changes
- Related issue numbers  
- Testing performed
- Breaking changes (if any)
- Reference to similar features in Python/TypeScript LangChain (when applicable)

## Documentation contributions

See our dedicated [Documentation Contribution Guide](./documentation) for details on:
- Writing tutorials
- Creating how-to guides
- Documentation style guide
- Building and testing docs locally

## Code of conduct

Please note that this project follows a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Recognition

Contributors are recognized in:
- The project's contributor list
- Release notes for significant contributions
- Documentation credits for written content

## Questions?

- Open a [GitHub Discussion](https://github.com/tmc/langchaingo/discussions)
- Check existing issues and PRs
- Review the documentation

Thank you for helping make LangChainGo better!