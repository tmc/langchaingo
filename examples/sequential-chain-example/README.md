# Sequential Chain Example

Welcome to the Sequential Chain Example! This Go program demonstrates how to use the LangChain library to create and run sequential chains of language models. It's a fun and practical way to see how multiple AI models can work together to generate creative content.

## What Does This Example Do?

This example showcases two types of sequential chains:

1. **Simple Sequential Chain**: A chain with a single input that passes through multiple models.
2. **Sequential Chain**: A more complex chain that can handle multiple inputs and outputs.

Let's break down each example:

### Simple Sequential Chain

In this part, we create a chain that:
1. Takes a play title as input.
2. Uses an AI playwright to generate a synopsis based on the title.
3. Passes the synopsis to an AI play critic, who writes a review.

The process goes like this:
```
Play Title -> AI Playwright -> Synopsis -> AI Critic -> Review
```

### Sequential Chain

This more advanced chain:
1. Takes a play title and an era as inputs.
2. Uses an AI playwright to create a synopsis based on both the title and era.
3. Passes the synopsis to an AI critic for a review.

The process looks like this:
```
Play Title ─┐
            └─> AI Playwright -> Synopsis -> AI Critic -> Review
Era ────────┘
```

## How It Works

The example uses OpenAI's language model and the LangChain library to create these chains. It demonstrates how to:

- Set up prompts for each step in the chain
- Create LLM chains with specific inputs and outputs
- Combine chains into sequential structures
- Run the chains with given inputs

## Running the Example

When you run this program, it will execute both the simple and complex sequential chain examples. You'll see the generated reviews printed to the console.

This example is a great way to understand how you can chain together multiple AI models to create more complex and interesting outputs. Have fun exploring the possibilities of sequential AI chains!
