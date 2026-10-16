# Example with Document Loader and Ollama

## Prerequisite

This example uses Ollama as the LLM engine.
So, you should install it. You can find all informations [here](https://ollama.com/).

Example of an installation with Docker : 

```bash
docker run -d -v ollama:/root/.ollama -p 11434:11434 --name ollama ollama/ollama
```

And after, you can load a model :

```bash
docker exec ollama ollama run tinydolphin
```

## Example explanations

This example present the integration of Langchain with Ollama while showcasing the use of a PDF document to enrich the context.

Actually, the PDF of the example contains 1 page and contains information from Langchain documentation.

Steps of the example : 

1. Load documents (for the context) from the PDF file
2. Initialize the llm and the chain
3. Prompt the model with context loads from the PDF

The prompt targets a question where the answer is contained in the context.
