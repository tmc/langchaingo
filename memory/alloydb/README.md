# AlloyDB for PostgreSQL for LangChain Go

- [Product Documentation](https://cloud.google.com/alloydb)

The **AlloyDB for PostgreSQL for LangChain** package provides a first class experience for connecting to
AlloyDB instances from the LangChain ecosystem while providing the following benefits:

- **Simplified & Secure Connections**: easily and securely create shared connection pools to connect to Google Cloud databases utilizing IAM for authorization and database authentication without needing to manage SSL certificates, configure firewall rules, or enable authorized networks.
- **Improved performance & Simplified management**: use a single-table schema can lead to faster query execution, especially for large collections.
- **Improved metadata handling**: store metadata in columns instead of JSON, resulting in significant performance improvements.
- **Clear separation**: clearly separate table and extension creation, allowing for distinct permissions and streamlined workflows.
- **Better integration with AlloyDB**: built-in methods to take advantage of AlloyDB's advanced indexing and scalability capabilities.

## Quick Start

In order to use this package, you first need to go through the following
steps:

1. [Select or create a Cloud Platform project.](https://console.cloud.google.com/project)
2. [Enable billing for your project.](https://cloud.google.com/billing/docs/how-to/modify-project#enable_billing_for_a_project)
3. [Enable the AlloyDB API.](https://console.cloud.google.com/flows/enableapi?apiid=alloydb.googleapis.com)
4. [Authentication with CloudSDK.](https://cloud.google.com/sdk/gcloud/reference/auth/application-default/login)

## Supported Go Versions

Go version >= go 1.22.0

## Engine Creation

The `AlloyDBEngine` configures a connection pool to your AlloyDB database. 

```go
package main

import (
  "context"
  "fmt"

  "github.com/tmc/langchaingo/internal/alloydbutil"
)

func NewAlloyDBEngine(ctx context.Context) (*alloydbutil.PostgresEngine, error) {
	// Call NewPostgresEngine to initialize the database connection
    pgEngine, err := alloydbutil.NewPostgresEngine(ctx,
        alloydbutil.WithUser("my-user"),
        alloydbutil.WithPassword("my-password"),
        alloydbutil.WithDatabase("my-database"),
        alloydbutil.WithAlloyDBInstance("my-project-id", "region", "my-cluster", "my-instance"),
    )
    if err != nil {
        return nil, fmt.Errorf("Error creating PostgresEngine: %s", err)
    }
    return pgEngine, nil
}

func main() {
    ctx := context.Background()
    alloyDBEngine, err := NewAlloyDBEngine(ctx)
    if err != nil {
         return nil, err
    }
}
```

See the full [Chat Message History example and tutorial](https://github.com/tmc/langchaingo/tree/main/examples/google-alloydb-chat-message-history-example).

## Engine Creation WithPool

Create an AlloyDBEngine with the `WithPool` method to connect to an instance of AlloyDB Omni or to customize your connection pool.


```go
package main

import (
  "context"
  "fmt"

  "github.com/jackc/pgx/v5/pgxpool"
  "github.com/tmc/langchaingo/internal/alloydbutil"
)

func NewAlloyDBWithPoolEngine(ctx context.Context) (*alloydbutil.PostgresEngine, error) {
    myPool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
    if err != nil {
        return err
    }
	// Call NewPostgresEngine to initialize the database connection
    pgEngineWithPool, err := alloydbutil.NewPostgresEngine(ctx, alloydbutil.WithPool(myPool))
    if err != nil {
        return nil, fmt.Errorf("Error creating PostgresEngine with pool: %s", err)
    }
    return pgEngineWithPool, nil
}

func main() {
    ctx := context.Background()
    alloyDBEngine, err := NewAlloyDBWithPoolEngine(ctx)
    if err != nil {
        return nil, err
    }
}
```

## Chat Message History Usage

Use a table to store the history of chat messages.

```go
package main

import (
  "context"
  "fmt"

  "github.com/tmc/langchaingo/internal/alloydbutil"
  "github.com/tmc/langchaingo/llms"
  "github.com/tmc/langchaingo/memory/alloydb"
)

func main() {
    ctx := context.Background()
    alloyDBEngine, err := NewAlloyDBEngine(ctx)
    if err != nil {
        return nil, err
    }
    
	// Creates a new table in the Postgres database, which will be used for storing Chat History.
	err = alloyDBEngine.InitChatHistoryTable(ctx, "tableName")
	if err != nil {
		log.Fatal(err)
	}

    // Creates a new Chat Message History
    cmh, err := alloydb.NewChatMessageHistory(ctx, *alloyDBEngine, "tableName", "sessionID")
    if err != nil {
        log.Fatal(err)
    }

    // Creates individual messages and adds them to the chat message history.
    aiMessage := llms.AIChatMessage{Content: "test AI message"}
    humanMessage := llms.HumanChatMessage{Content: "test HUMAN message"}
    // Adds a user message to the chat message history.
    err = cmh.AddUserMessage(ctx, string(aiMessage.GetContent()))
    if err != nil {
        log.Fatal(err)
    }
    // Adds a user message to the chat message history.
    err = cmh.AddUserMessage(ctx, string(humanMessage.GetContent()))
    if err != nil {
        log.Fatal(err)
    }
    msgs, err := cmh.Messages(ctx)
    if err != nil {
        log.Fatal(err)
    }
    for _, msg := range msgs {
        fmt.Println("Message:", msg)
    }
}
```