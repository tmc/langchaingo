# Document Loader for Alloy DB 

Document loader is the utility for loading documents from Alloy DB.  

## Supported Go Versions

Go version >= go 1.22.0

## Document Loader Usage

`DocumentLoader` uses `PostgresEngine` for connecting with the database. [Learn more about the `PostgresEngine`](https://github.com/tmc/langchaingo/tree/main/internal/alloydbutil).

```go
package main

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/internal/alloydbutil"
	"github.com/tmc/langchaingo/documentloader/alloydb"
)

func main() {
	ctx := context.Background()
	pgEngine, err := alloydbutil.NewPostgresEngine(ctx,
		alloydbutil.WithUser("my-user"),
		alloydbutil.WithPassword("my-password"),
		alloydbutil.WithDatabase("my-database"),
		alloydbutil.WithAlloyDBInstance("my-project-id", "region", "my-cluster", "my-instance"),
	)
	if err != nil {
		panic(fmt.Errorf("error creating PostgresEngine: %s", err))
	}

	documentLoader, err := alloydb.NewDocumentLoader(ctx,
		pgEngine,
		alloydb.WithQuery("SELECT * FROM my_Table"),
		alloydb.WithCSVFormatter())
	if err != nil {
		panic(fmt.Errorf("error creating DocumentLoader: %s", err))
	}
	
	docs, err := documentLoader.Load(ctx)
	if err != nil {
		panic(fmt.Errorf("error loading documents: %s", err))
	}	
	
	for _, doc := range docs {
        	fmt.Printf("%v", doc)
	}
	
}
```
