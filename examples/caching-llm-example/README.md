# Caching LLM Example

This example demonstrates how to implement caching for a Language Model (LLM) using the LangChain Go library. The program showcases the benefits of caching by repeatedly querying an LLM and measuring the response time.

## What This Example Does

1. **Sets up an LLM**: 
   - Initializes an Ollama LLM using the "llama2" model.

2. **Implements Caching**:
   - Creates an in-memory cache that stores results for one minute.
   - Wraps the base LLM with the caching functionality.

3. **Performs Repeated Queries**:
   - Asks the same question ("Who was the first man to walk on the moon?") three times.
   - The first query will use the actual LLM, while subsequent queries will retrieve the cached response.

4. **Measures and Displays Performance**:
   - Records the time taken for each query.
   - Prints the response along with the time taken for each iteration.

5. **Formats Output**:
   - Uses word wrapping to ensure neat output within an 80-character width.
   - Separates each iteration with a line of "=" characters.

## Key Features

- **LLM Caching**: Demonstrates how to implement caching to improve response times for repeated queries.
- **Performance Measurement**: Shows the time difference between cached and non-cached responses.
- **Ollama Integration**: Uses the Ollama LLM with the "llama2" model.
- **Output Formatting**: Ensures readable output with proper word wrapping and separation between iterations.

## Running the Example

When you run this example, you'll see the LLM's response to the question about the first man on the moon, repeated three times. The first response will likely take longer as it queries the actual LLM, while the subsequent responses should be significantly faster due to caching.

This example is great for understanding how caching can dramatically improve response times in applications that use LLMs, especially when similar queries are likely to be repeated.
