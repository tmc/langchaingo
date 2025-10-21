# Documentation contribution guide

This guide helps you contribute documentation to LangChainGo. We especially need help with tutorials and how-to guides!

## Documentation structure

Our documentation is organized into four main categories:

### 1. Concepts
**Purpose**: Explain ideas and provide background
- Architecture overviews
- Design decisions
- Theoretical foundations

### 2. Tutorials  
**Purpose**: Step-by-step learning experiences
- Complete, runnable projects
- Progressive complexity
- Real-world applications

### 3. How-to guides
**Purpose**: Solve specific problems
- Focused on single tasks
- Assume some knowledge
- Practical solutions

### 4. API reference
**Purpose**: Technical specifications
- Generated from code comments
- Complete parameter documentation
- Usage examples

## Writing tutorials

Tutorials are complete learning experiences. Here's how to write a great tutorial:

### Tutorial template

```markdown
# Building [What You're Building]

[One sentence description of what the reader will build]

## What you'll build

A [type of application] that:
- [Feature 1]
- [Feature 2]
- [Feature 3]

## Prerequisites

- Go 1.21+
- [Required API keys]
- [Other requirements]

## Step 1: [First Task]

[Brief explanation of what this step accomplishes]

```go
// Complete, runnable code
```

## Step 2: [next task]

[Continue with progressive steps...]

## Running the application

```bash
# Clear commands to run
```

## Next steps

- [Potential improvements]
- [Related tutorials]
```

### Tutorial guidelines

1. **Start Simple**: Begin with minimal code that works
2. **Build Progressively**: Add complexity step by step
3. **Explain Why**: Don't just show how, explain why
4. **Complete Code**: Every code block should be runnable
5. **Test Everything**: Ensure all code examples work

## Writing how-to guides

How-to guides solve specific problems. They differ from tutorials:

### How-to template

```markdown
# How to [Specific Task]

## Problem

[Clear description of the problem being solved]

## Solution

[Brief overview of the approach]

## Implementation

```go
// Focused code example
```

## Considerations

- [Performance implications]
- [Security considerations]
- [Alternative approaches]

## Related guides

- [Link to related how-tos]
```

### How-to guidelines

1. **One Problem**: Focus on solving one specific issue
2. **Clear Title**: "How to X" format
3. **Minimal Setup**: Don't repeat basic setup
4. **Multiple Solutions**: Show alternatives when relevant
5. **Practical Focus**: Real problems developers face

## Documentation style guide

### Language and tone

- **Direct and Clear**: Avoid flowery language
- **Active Voice**: "Configure the client" not "The client should be configured"
- **Present Tense**: "This function returns" not "This function will return"
- **You/Your**: Address the reader directly

### Code examples

```go
// DO: Complete, runnable examples
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/tmc/langchaingo/llms/openai"
)

func main() {
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }
    // ... rest of example
}
```

```go
// DON'T: Incomplete fragments
llm := openai.New() // Missing error handling
// ... magic happens here
```

### Formatting conventions

- **Headers**: Use sentence case, not title case
- **Code Blocks**: Always specify language (` ```go`)
- **Emphasis**: Use **bold** for important concepts
- **Lists**: Use `-` for unordered lists
- **Links**: Use descriptive link text, not "click here"

### Things to avoid

- No emojis in documentation
- No marketing language or hype
- No incomplete examples
- No hardcoded API keys
- No external service dependencies

## Contributing missing documentation

We have several tutorials and guides marked as "Coming Soon". Here's how to contribute:

### 1. Choose a topic

- Check our [Tutorials](/docs/tutorials) and [How-To Guides](/docs/how-to) for topics marked as coming soon. 
- Review open issues for topics that have already been claimed (avoid duplicate work).

### 2. Open an issue

Before writing, open an issue to:
- Claim the topic (avoid duplicate work)
- Discuss the approach
- Get feedback on the outline

### 3. Write the content

Follow the templates and guidelines above.

### 4. Test everything

- Ensure all code examples run
- Test on a clean environment
- Verify API keys are handled properly

### 5. Submit PR

Create a pull request with:
- Clear title: `docs: add tutorial for [topic]`
- Link to the tracking issue
- Summary of what's covered

## Local development

### Building documentation

#### Local development

```bash
cd docs
npm install
npm run start
```

This starts a local server at `http://localhost:3000`

#### Docker development

For a containerized environment:

```bash
cd docs

# Quick development server with live reload
make docker-dev

# Or build and run a persistent container
make docker-run

# Clean up when done
make docker-clean
```

The Docker approach ensures consistent Node.js environment and dependencies.

### Testing documentation

Before submitting:

1. **Check Links**: Ensure all links work
2. **Run Code**: Test all code examples
3. **Review Formatting**: Check rendering in browser
4. **Lint Documentation**: Run Vale to check style consistency
5. **Spell Check**: Use your editor's spell checker

#### Running Vale linting

Vale automatically checks documentation style and consistency:

```bash
# Install Vale (on macOS)
make lint-deps

# Lint all documentation
make lint-docs

# Or run Vale directly
vale docs
```

Vale checks for:
- Sentence case headers
- Consistent terminology
- Spelling of technical terms
- Writing style guidelines

## Examples of good documentation

### Good tutorial example
- [Building an AI Code Reviewer](/docs/tutorials/code-reviewer)
- Complete, practical application
- Progressive complexity
- Real-world use case

### Good how-to guide example
- [How to configure different LLM providers](/docs/how-to/configure-llm-providers)
- Focused on specific task
- Multiple provider examples
- Clear configuration steps

## Need help?

- Check existing documentation for style examples
- Open a [GitHub Discussion](https://github.com/tmc/langchaingo/discussions) for questions
- Tag your PR with `documentation` for faster review

## Recognition

Documentation contributors are credited in:
- The documentation itself (author attribution)
- Release notes
- Contributors list

Thank you for helping improve LangChainGo documentation!
