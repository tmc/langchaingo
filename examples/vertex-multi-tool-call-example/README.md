# Vertex AI Multi Tool Call Example

Welcome to this cheerful example of using Google's Vertex AI with LangChain Go! 🎉

This example demonstrates how to use the Vertex AI API to generate text completions using LangChain Go. It's a simple and fun way to interact with Google's powerful language models!

## What This Example Does

1. **Sets Up Vertex AI**: The code configures a connection to Google's Vertex AI service using your GCP project details and credentials.

2. **Creates a Language Model**: It initializes a language model (LLM) using the Vertex AI service.

3. **Uses tools to get a response**: The example then uses this LLM to generate an answer using tool calls to the question ""What is the weather like in Chicago? And what's the elevation in Chicago?"

4. **Prints the Result**: Finally, it prints the generated answer to the console.

## How to Run

Before running this example, make sure you've set the following environment variables:

- `VERTEX_PROJECT`: Your Google Cloud Project ID
- `VERTEX_LOCATION`: The GCP location (region) for Vertex AI (e.g., "us-central1")
- `VERTEX_CREDENTIALS`: Path to your GCP service account credentials JSON file

Once you've set these up, just run the Go file, and watch the magic happen! 🚀

## Why It's Cool

This example showcases how easy it is to tap into advanced AI capabilities using LangChain Go and Google's Vertex AI. Whether you're building a chatbot, a question-answering system, or any other AI-powered application, this code provides a great starting point!

Happy coding, and may your AI adventures be ever exciting! 🌟
