# LLM Chain Multi Prompt Example

Welcome to this cheerful example of using LLM (Language Model) chains with LangChain in Go! 🎉

This example demonstrates how to create and use LLM chains for various natural language processing tasks. Let's dive in and see what exciting things we can do!

## What Does This Example Do?

1. **Company Name Generation** 🏢
   - We create an LLM chain that generates a company name based on a product.
   - It uses a simple prompt template: "What is a good name for a company that makes {{.product}}?"
   - We run this chain with "socks" as input and get a creative company name suggestion!

2. **Text Translation** 🌍
   - We set up another LLM chain for translating text between languages.
   - The prompt template asks to translate from one language to another.
   - We demonstrate translating "I love programming" from English to French.

## How It Works

1. We start by setting up an OpenAI LLM (Language Model).
2. For each task, we create a `PromptTemplate` with placeholders for inputs.
3. We then create `LLMChain` instances combining the LLM and the prompt templates.
4. For single-input chains, we use the `Run` function.
5. For multi-input chains, we use the `Call` function with a map of inputs.

## Running the Example

When you run this example, you'll see:
1. A suggested company name for a sock manufacturer.
2. The French translation of "I love programming".

It's a fun and practical demonstration of how LLM chains can be used for creative and linguistic tasks!

Happy coding, and enjoy exploring the world of LLM chains with Go! 🚀👨‍💻👩‍💻
