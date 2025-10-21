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

  "github.com/tmc/langchaingo/internal/cloudsqlutil"
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

See the full [Vector Store example and tutorial](https://github.com/tmc/langchaingo/tree/main/examples/google-cloudsql-vectorstore-example).

## Engine Creation WithPool

Create a CloudSQLEngine with the `WithPool` method to connect to an instance of CloudSQL Omni or to customize your connection pool.


```go
package main

import (
  "context"
  "fmt"

  "github.com/jackc/pgx/v5/pgxpool"
  "github.com/tmc/langchaingo/internal/cloudsqlutil"
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

## Vector Store Usage

Use a vector store to store embedded data and perform vector search.

```go
package main

import (
  "context"
  "fmt"

  "github.com/tmc/langchaingo/embeddings"
  "github.com/tmc/langchaingo/internal/cloudsqlutil"
  "github.com/tmc/langchaingo/llms/googleai/vertex"
  "github.com/tmc/langchaingo/vectorstores/cloudsql"
)

func main() {
    ctx := context.Background()
    cloudSQLEngine, err := NewCloudSQLEngine(ctx)
    if err != nil {
        return nil, err
    }

    // Initialize table for the Vectorstore to use. You only need to do this the first time you use this table.
    vectorstoreTableoptions, err := &cloudsqlutil.VectorstoreTableOptions{
        TableName:  "table",
        VectorSize: 768,
    }
    if err != nil {
        log.Fatal(err)
    }

    err = pgEngine.InitVectorstoreTable(ctx, *vectorstoreTableoptions,
        []alloydbutil.Column{
            alloydbutil.Column{
                Name:     "area",
                DataType: "int",
                Nullable: false,
            },
            alloydbutil.Column{
                Name:     "population",
                DataType: "int",
                Nullable: false,
            },
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Initialize VertexAI LLM
    llm, err := vertex.New(ctx, googleai.WithCloudProject(projectID), googleai.WithCloudLocation(cloudLocation), googleai.WithDefaultModel("text-embedding-005"))
    if err != nil {
        log.Fatal(err)
    }

    myEmbedder, err := embeddings.NewEmbedder(llm)
    if err != nil {
        log.Fatal(err)
    }

    vectorStore := cloudsql.NewVectorStore(cloudSQLEngine, myEmbedder, "my-table", cloudsql.WithMetadataColumns([]string{"area", "population"}))
}
```