# Cloud SQL for PostgreSQL for LangChain Go

- [Product Documentation](https://cloud.google.com/sql/docs)

The **Cloud SQL for PostgreSQL for LangChain** package provides a first class experience for connecting to
Cloud SQL instances from the LangChain ecosystem while providing the following benefits:

- **Simplified & Secure Connections**: easily and securely create shared connection pools to connect to Google Cloud databases utilizing IAM for authorization and database authentication without needing to manage SSL certificates, configure firewall rules, or enable authorized networks.
- **Improved performance & Simplified management**: use a single-table schema can lead to faster query execution, especially for large collections.
- **Improved metadata handling**: store metadata in columns instead of JSON, resulting in significant performance improvements.
- **Clear separation**: clearly separate table and extension creation, allowing for distinct permissions and streamlined workflows.

## Quick Start

In order to use this package, you first need to go through the following
steps:

1. [Select or create a Cloud Platform project.](https://console.cloud.google.com/project)
2. [Enable billing for your project.](https://cloud.google.com/billing/docs/how-to/modify-project#enable_billing_for_a_project)
3. [Enable the Cloud SQL API.](https://console.cloud.google.com/apis/enableflow?apiid=sql.googleapis.com)
4. [Authentication with CloudSDK.](https://cloud.google.com/sdk/gcloud/reference/auth/application-default/login)

## Supported Go Versions

Go version >= go 1.22.0

## Engine Creation

The `CloudSQLEngine` configures a connection pool to your CloudSQL database. 

```go
package main

import (
  "context"
  "fmt"

  "github.com/vendasta/langchaingo/internal/cloudsqlutil"
)

func NewCloudSQLEngine(ctx context.Context) (*cloudsqlutil.PostgresEngine, error) {
	// Call NewPostgresEngine to initialize the database connection
    pgEngine, err := cloudsqlutil.NewPostgresEngine(ctx,
        cloudsqlutil.WithUser("my-user"),
        cloudsqlutil.WithPassword("my-password"),
        cloudsqlutil.WithDatabase("my-database"),
        cloudsqlutil.WithCloudSQLInstance("my-project-id", "region", "my-instance"),
    )
    if err != nil {
        return nil, fmt.Errorf("Error creating PostgresEngine: %s", err)
    }
    return pgEngine, nil
}

func main() {
    ctx := context.Background()
    cloudSQLEngine, err := NewCloudSQLEngine(ctx)
    if err != nil {
         return nil, err
    }
}
```

See the full [Chat Message History example and tutorial](https://github.com/vendasta/langchaingo/tree/main/examples/google-cloudsql-chat-message-history-example).

## Engine Creation WithPool

Create a CloudSQLEngine with the `WithPool` method to connect to an instance of CloudSQL Omni or to customize your connection pool.


```go
package main

import (
  "context"
  "fmt"

  "github.com/jackc/pgx/v5/pgxpool"
  "github.com/vendasta/langchaingo/internal/cloudsqlutil"
)

func NewCloudSQLWithPoolEngine(ctx context.Context) (*cloudsqlutil.PostgresEngine, error) {
    myPool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
    if err != nil {
        return err
    }
	// Call NewPostgresEngine to initialize the database connection
    pgEngineWithPool, err := cloudsqlutil.NewPostgresEngine(ctx, cloudsqlutil.WithPool(myPool))
    if err != nil {
        return nil, fmt.Errorf("Error creating PostgresEngine with pool: %s", err)
    }
    return pgEngineWithPool, nil
}

func main() {
    ctx := context.Background()
    cloudSQLEngine, err := NewCloudSQLWithPoolEngine(ctx)
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

  "github.com/vendasta/langchaingo/internal/cloudsqlutil"
  "github.com/vendasta/langchaingo/llms"
  "github.com/vendasta/langchaingo/memory/cloudsql"
)

func main() {
    ctx := context.Background()
    cloudSQLEngine, err := NewCloudSQLEngine(ctx)
    if err != nil {
        return nil, err
    }

	// Creates a new table in the Postgres database, which will be used for storing Chat History.
	err = cloudSQLEngine.InitChatHistoryTable(ctx, "tableName")
	if err != nil {
		log.Fatal(err)
	}

    // Creates a new Chat Message History
    cmh, err := cloudsql.NewChatMessageHistory(ctx, *cloudSQLEngine, "tableName", "sessionID")
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