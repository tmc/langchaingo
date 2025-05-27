# Google CloudSQL Chat Message History Example

This example demonstrates how to use [CloudSQL for Postgres](https://cloud.google.com/sql) as a backend for the ChatMessageHistory for LangChain in Go.

## What This Example Does

1. **Creates a CloudSQL Chat Message History:**
    - Initializes the `cloudsql.PostgresEngine` object to establish a connection to the CloudSQL database.
    - Creates a new table to store the chat message history if it doesn't exist.
    - Initializes a `cloudsql.ChatMessageHistory` object, which provides methods to store, retrieve, and clear message contents with a specific session ID.

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

1. Set the following environment variables. Your CloudSQL values can be found in the [Google Cloud Console](https://console.cloud.google.com/sql/instances):
   ```
   export PROJECT_ID=<your project Id>
   export POSTGRES_USERNAME=<your user>
   export POSTGRES_PASSWORD=<your password>
   export POSTGRES_REGION=<your region>
   export POSTGRES_INSTANCE=<your instance>
   export POSTGRES_DATABASE=<your database>
   export POSTGRES_TABLE=<your tablename>
   export POSTGRES_SESSION_ID=<your sessionID>
   ```

2. Run the Go example:
   ```
   go run google_cloudsql_chat_message_history_example.go
   ```

## Key Features
- **CloudSQL Integration**: Connects to a CloudSQL instance for storing and managing chat messages.
- **Message Storage**: Provides methods for storing chat messages with different operations like add, overwrite, and clear.
- **Session-based Message Management**: Messages are stored and retrieved using a unique session ID, making it easy to manage separate chat sessions.
- **Clear and Overwrite Capabilities**: Allows overwriting and clearing messages to maintain the chat history as needed.