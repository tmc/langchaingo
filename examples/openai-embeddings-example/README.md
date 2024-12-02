# OpenAI Embeddings Example

Hello there! 👋 Welcome to this fun little Go program that demonstrates how to create embeddings using OpenAI's API. Let's break down what this exciting example does!

## What does this program do?

This program showcases how to:

1. Set up an OpenAI client with specific model options
2. Create embeddings for given text inputs
3. Print the resulting embeddings

## How it works

1. First, we configure our OpenAI client:
   - We use the "gpt-3.5-turbo-0125" model for general language tasks
   - We specify "text-embedding-3-large" as our embedding model

2. We create a new OpenAI client with these options

3. We prepare two simple words for embedding: "ola" and "mundo" (Hello and World in Portuguese)

4. We call the `CreateEmbedding` function to generate embeddings for these words

5. Finally, we print out the resulting embeddings

## Why is this cool?

Embeddings are super useful! They convert words or phrases into numerical vectors, which can be used for all sorts of neat tricks like:

- Finding similar texts
- Clustering related concepts
- Improving search functionality
- And much more!

This example gives you a starting point to experiment with embeddings in your own projects. Have fun exploring the world of vector representations! 🚀🧠

## Running the program

Make sure you have your OpenAI API key set up properly, and then just run the program. You'll see the embeddings printed out in all their numerical glory!

Happy coding! 😊
