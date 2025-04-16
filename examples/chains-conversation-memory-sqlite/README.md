# Conversational Memory with SQLite in LangChain

Hello there! 👋 This example demonstrates how to create a conversational AI system with memory persistence using SQLite in Go with the LangChain library. Let's break down what this exciting code does!

## What Does This Example Do?

1. **Sets up an OpenAI Language Model**: It initializes an OpenAI language model to power our conversational AI.

2. **Creates a SQLite Database**: The code sets up a SQLite database to store conversation history.

3. **Implements Conversation Memory**: It uses SQLite to maintain a persistent memory of the conversation, allowing the AI to remember previous interactions.

4. **Prepares Sample Data**: If the database is empty, it inserts a sample message to kickstart the conversation.

5. **Runs a Conversation**: The example runs a conversation chain, asking the AI a question that requires memory of previous interactions.

## Key Components

- **SQLite Chat Message History**: Uses `sqlite3.NewSqliteChatMessageHistory` to create a chat history stored in SQLite.
- **Conversation Buffer**: Implements `memory.NewConversationBuffer` to manage the conversation memory.
- **Conversation Chain**: Creates a `chains.NewConversation` to handle the flow of the conversation.

## How It Works

1. The code first checks if there's any existing data in the SQLite database.
2. If empty, it inserts a sample message: "Hi there, my name is Murilo!"
3. It then asks the AI: "What's my name? How many times did I ask this?"
4. The AI responds based on the conversation history stored in the SQLite database.

This example showcases how to create a conversational AI system with persistent memory, allowing for more context-aware and personalized interactions over time!

Feel free to run this example and experiment with different questions to see how the AI remembers and uses previous conversation context! 🚀🤖
