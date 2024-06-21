# Zep Memory with LangChain and OpenAI Example

Hello there! 👋 This example demonstrates how to use Zep memory with LangChain and OpenAI to create a conversational AI system. Let's break down what this exciting code does!

## What This Example Does

1. **Sets up the environment**: The code uses environment variables for the Zep API key and session ID.

2. **Initializes clients**:
   - Creates a Zep client using the API key.
   - Sets up an OpenAI language model.

3. **Creates a conversation chain**:
   - Uses LangChain's conversation chain.
   - Integrates Zep memory for maintaining context.

4. **Runs two conversation turns**:
   - First turn: Introduces "John Doe".
   - Second turn: Asks about the name to test memory retention.

5. **Prints responses**: Displays the AI's responses to each input.

## Key Features

- **Perpetual Memory**: Uses Zep's perpetual memory type to maintain long-term context.
- **Custom Prefixes**: Sets custom prefixes for AI ("Robot") and human ("Joe") messages.
- **Error Handling**: Includes error checking for robust operation.

## How to Run

1. Set the required environment variables:
   - `ZEP_API_KEY`: Your Zep API key
   - `ZEP_SESSION_ID`: A unique session identifier
   - `OPENAI_API_KEY`: Your OpenAI API key (required by the OpenAI client)

2. Run the Go program:
   ```
   go run main.go
   ```

## Expected Output

You should see responses from the AI to both inputs. The second response should demonstrate that the AI remembers the name "John Doe" from the first interaction.

Enjoy exploring this powerful combination of Zep, LangChain, and OpenAI! 🚀🤖
