# Google AlloyDB Chat Message History Example

This example demonstrates how to use [AlloyDB for Postgres](https://cloud.google.com/products/alloydb) as a backend for the ChatMessageHistory for LangChain in Go.

## What This Example Does

1. **Creates an AlloyDB Chat Message History:**
    - Initializes the `alloydb.PostgresEngine` object to establish a connection to the AlloyDB database.
    - Creates a new table to store the chat message history if it doesn't exist.
    - Initializes an `alloydb.ChatMessageHistory` object, which provides methods to store, retrieve, and clear message contents with a specific session ID.

2. **Add Single Messages:**
    - Creates individual AI and Human messages and stores them in the chat message history.
    - Prints the stored messages to the console.

3. **Add Multiple Messages:**
    - Creates multiple messages of different types (AI and Human) and stores them all at once in the chat message history.
    - Prints the stored messages to the console.

4. **Overwrite Messages:**
    - Clears the existing messages and then stores a new set of messages.
    - Prints the updated list of stored messages.

5. **Clear All Messages:**
    - Clears all messages stored for the current session.

## How to Run the Example

1. Set the following environment variables. Your AlloyDB values can be found in the [Google Cloud Console](https://console.cloud.google.com/alloydb/clusters):
   ```
   export PROJECT_ID=<your project Id>
   export ALLOYDB_USERNAME=<your user>
   export ALLOYDB_PASSWORD=<your password>
   export ALLOYDB_REGION=<your region>
   export ALLOYDB_CLUSTER=<your cluster>
   export ALLOYDB_INSTANCE=<your instance>
   export ALLOYDB_DATABASE=<your database>
   export ALLOYDB_TABLE=<your tablename>
   export ALLOYDB_SESSION_ID=<your sessionID>
   ```

2. Run the Go example:
   ```
   go run google_alloydb_chat_message_history_example.go
   ```

## Key Features
- **AlloyDB Integration**: Connects to an AlloyDB instance for storing and managing chat messages.
- **Message Storage**: Provides methods for storing chat messages with different operations like add, overwrite, and clear.
- **Session-based Message Management**: Messages are stored and retrieved using a unique session ID, making it easy to manage separate chat sessions.
- **Clear and Overwrite Capabilities**: Allows overwriting and clearing messages to maintain the chat history as needed.
