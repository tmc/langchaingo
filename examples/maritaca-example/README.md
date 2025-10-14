# 🦜 Maritaca Chat Example

Hello there! 👋 This fun little Go program demonstrates how to use the Maritaca AI language model to answer questions. Let's break it down!

## 🎯 What does this example do?

This example:
1. Sets up a connection to the Maritaca AI service
2. Asks a simple question about Brazil's population
3. Prints out the AI's response

## 🔑 Key Components

- We're using the `github.com/vendasta/langchaingo/llms/maritaca` package to interact with Maritaca AI.
- The program reads your Maritaca API key from an environment variable called `MARITACA_KEY`.
- We're using the "sabia-2-medium" model, which is pretty smart!

## 🚀 How it works

1. First, we set up our Maritaca client with the API key and model choice.
2. Then, we prepare a simple question: "How many people live in Brazil?"
3. We send this question to the AI using the `Call` method.
4. Finally, we print out whatever answer the AI gives us!

## 🎉 Running the example

To run this example, make sure you:
1. Have your Maritaca API key set as an environment variable.
2. Have all the necessary Go packages installed.

Then just run the program and see what the AI says about Brazil's population!

## 🤔 Why is this cool?

This example shows how easy it is to integrate AI into your Go programs. With just a few lines of code, you can ask an AI model complex questions and get intelligent responses. How awesome is that? 🎈

Happy coding, and have fun chatting with AI! 🤖💬
