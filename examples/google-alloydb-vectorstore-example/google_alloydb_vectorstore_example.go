package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/averikitsch/langchaingo/embeddings"
	"github.com/averikitsch/langchaingo/llms/googleai"
	"github.com/averikitsch/langchaingo/llms/googleai/vertex"
	"github.com/averikitsch/langchaingo/schema"
	"github.com/averikitsch/langchaingo/util/alloydbutil"
	"github.com/averikitsch/langchaingo/vectorstores"
	"github.com/averikitsch/langchaingo/vectorstores/alloydb"
)

func getEnvVariables() (string, string, string, string, string, string, string, string, string) {
	// Requires environment variable ALLOYDB_USERNAME to be set.
	username := os.Getenv("ALLOYDB_USERNAME")
	if username == "" {
		log.Fatal("env variable ALLOYDB_USERNAME is empty")
	}
	// Requires environment variable ALLOYDB_PASSWORD to be set.
	password := os.Getenv("ALLOYDB_PASSWORD")
	if password == "" {
		log.Fatal("env variable ALLOYDB_PASSWORD is empty")
	}
	// Requires environment variable ALLOYDB_DATABASE to be set.
	database := os.Getenv("ALLOYDB_DATABASE")
	if database == "" {
		log.Fatal("env variable ALLOYDB_DATABASE is empty")
	}
	// Requires environment variable PROJECT_ID to be set.
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatal("env variable PROJECT_ID is empty")
	}
	// Requires environment variable ALLOYDB_REGION to be set.
	region := os.Getenv("ALLOYDB_REGION")
	if region == "" {
		log.Fatal("env variable ALLOYDB_REGION is empty")
	}
	// Requires environment variable ALLOYDB_INSTANCE to be set.
	instance := os.Getenv("ALLOYDB_INSTANCE")
	if instance == "" {
		log.Fatal("env variable ALLOYDB_INSTANCE is empty")
	}
	// Requires environment variable ALLOYDB_CLUSTER to be set.
	cluster := os.Getenv("ALLOYDB_CLUSTER")
	if cluster == "" {
		log.Fatal("env variable ALLOYDB_CLUSTER is empty")
	}
	// Requires environment variable ALLOYDB_TABLE to be set.
	table := os.Getenv("ALLOYDB_TABLE")
	if table == "" {
		log.Fatal("env variable ALLOYDB_TABLE is empty")
	}

	// Requires environment variable GOOGLE_CLOUD_LOCATION to be set.
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		log.Fatal("env variable GOOGLE_CLOUD_LOCATION is empty")
	}

	return username, password, database, projectID, region, instance, cluster, table, location
}

func main() {
	// Requires the Environment variables to be set as indicated in the getEnvVariables function.
	username, password, database, projectID, region, instance, cluster, table, cloudLocation := getEnvVariables()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgEngine, err := alloydbutil.NewPostgresEngine(ctx,
		alloydbutil.WithUser(username),
		alloydbutil.WithPassword(password),
		alloydbutil.WithDatabase(database),
		alloydbutil.WithAlloyDBInstance(projectID, region, cluster, instance),
		alloydbutil.WithIPType("PUBLIC"),
	)

	if err != nil {
		log.Fatal(err)
	}

	// Initialize table for the Vectorstore to use. You only need to do this the first time you use this table.
	vectorstoreTableoptions := alloydbutil.VectorstoreTableOptions{
		TableName:         table,
		VectorSize:        768,
		StoreMetadata:     true,
		OverwriteExisting: true,
		MetadataColumns: []alloydbutil.Column{
			{
				Name:     "area",
				DataType: "int",
			},
			{
				Name:     "population",
				DataType: "int",
			},
		},
	}

	err = pgEngine.InitVectorstoreTable(ctx, vectorstoreTableoptions)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize VertexAI LLM
	llm, err := vertex.New(ctx, googleai.WithCloudProject(projectID), googleai.WithCloudLocation(cloudLocation), googleai.WithDefaultModel("text-embedding-005"))
	if err != nil {
		log.Fatal(err)
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new AlloyDB Vectorstore
	vs, err := alloydb.NewVectorStore(pgEngine, e, table, alloydb.WithMetadataColumns([]string{"area", "population"}))
	if err != nil {
		log.Fatal(err)
	}

	_, err = vs.AddDocuments(ctx, []schema.Document{
		{
			PageContent: "Tokyo",
			Metadata: map[string]any{
				"population": 38,
				"area":       2190,
			},
		},
		{
			PageContent: "Paris",
			Metadata: map[string]any{
				"population": 11,
				"area":       105,
			},
		},
		{
			PageContent: "London",
			Metadata: map[string]any{
				"population": 9.5,
				"area":       1572,
			},
		},
		{
			PageContent: "Santiago",
			Metadata: map[string]any{
				"population": 6.9,
				"area":       641,
			},
		},
		{
			PageContent: "Buenos Aires",
			Metadata: map[string]any{
				"population": 15.5,
				"area":       203,
			},
		},
		{
			PageContent: "Rio de Janeiro",
			Metadata: map[string]any{
				"population": 13.7,
				"area":       1200,
			},
		},
		{
			PageContent: "Sao Paulo",
			Metadata: map[string]any{
				"population": 22.6,
				"area":       1523,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	docs, err := vs.SimilaritySearch(ctx, "Japan", 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Docs:", docs)
	filter := "\"area\" > 1500"
	filteredDocs, err := vs.SimilaritySearch(ctx, "Japan", 0, vectorstores.WithFilters(filter))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("FilteredDocs:", filteredDocs)
}
