# LangChain Go CLI Tools Demonstration

This example demonstrates how to use the comprehensive CLI tools available in the LangChain Go repository. It showcases the built-in tools for managing examples, initializing projects, and validating LLM provider connections.

## What This Example Demonstrates

1. **Examples Management**: 
   - List all 70+ available examples
   - Filter examples by category (llm, vectorstore, chain, etc.)
   - Filter examples by tags (completion, chat, streaming, etc.)
   - Output in different formats (table, JSON)

2. **Project Initialization**:
   - Available templates (basic-llm, chat-bot, rag-system, gemini-llm)
   - Shows project structure that would be created
   - Demonstrates scaffolding capabilities

3. **Provider Validation**:
   - Test API key connectivity
   - Validate LLM provider configurations
   - Quick connection checks

## Prerequisites

The example automatically installs the LangChain CLI from the current repository. No manual setup required!

## Running the Example

```bash
go run langchain_cli_example.go
```

## Sample Output

The example will show:
- Complete list of available examples with categories
- Filtered results by category and tags  
- Project templates and their structures
- Validation command demonstrations

## CLI Tools Showcased

### Examples Management
```bash
langchain examples list                    # List all examples
langchain examples list --category llm     # Filter by category
langchain examples list --tag completion   # Filter by tags
langchain examples list --format json      # JSON output
langchain examples run openai-completion   # Run specific example
```

### Project Scaffolding  
```bash
langchain init my-project --template basic-llm   # Create new project
langchain init chat-app --template chat-bot      # Chat bot template
langchain init rag-app --template rag-system     # RAG system template
```

### Provider Validation
```bash
langchain validate --provider openai --quick     # Quick OpenAI test
langchain validate --provider anthropic          # Full Anthropic test
```

## Learn More

- Run `langchain --help` to see all commands
- Each command supports `--help` for detailed options
- Templates include working code, tests, and documentation