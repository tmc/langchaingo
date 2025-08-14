package main

import (
	"context"
	"fmt"
	"log"

	"github.com/0xDezzy/langchaingo/graphs"
	"github.com/0xDezzy/langchaingo/graphs/falkordb"
	"github.com/0xDezzy/langchaingo/schema"
)

func main() {
	ctx := context.Background()

	// Connect to FalkorDB (make sure it's running on localhost:6379)
	fmt.Println("🔌 Connecting to FalkorDB...")
	db, err := falkordb.New("social_network",
		falkordb.WithHost("localhost"),
		falkordb.WithPort(6379),
	)
	if err != nil {
		log.Fatalf("Failed to connect to FalkorDB: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Connected successfully!")

	// Create sample graph documents
	fmt.Println("\n📊 Creating social network graph...")
	graphDocs := createSocialNetworkGraphs()

	// Add graph documents to FalkorDB
	fmt.Println("📝 Adding nodes and relationships to graph...")
	err = db.AddGraphDocuments(ctx, graphDocs)
	if err != nil {
		log.Fatalf("Failed to add graph documents: %v", err)
	}

	fmt.Println("✅ Graph data added successfully!")

	// Refresh and display schema
	fmt.Println("\n🔍 Refreshing graph schema...")
	err = db.RefreshSchema(ctx)
	if err != nil {
		log.Fatalf("Failed to refresh schema: %v", err)
	}

	fmt.Println("\n📋 Current Graph Schema:")
	fmt.Println(db.GetSchema())

	// Execute various graph queries
	fmt.Println("\n🔎 Executing graph queries...")

	// Query 1: Find all people
	fmt.Println("\n👥 All people in the network:")
	results, err := db.Query(ctx, "MATCH (p:Person) RETURN p.name, p.age, p.city", nil)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		printResults(results)
	}

	// Query 2: Find all relationships
	fmt.Println("\n🤝 All relationships:")
	results, err = db.Query(ctx, "MATCH (a:Person)-[r]->(b:Person) RETURN a.name, type(r), b.name", nil)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		printResults(results)
	}

	// Query 3: Find people who know each other
	fmt.Println("\n👫 People who know each other:")
	results, err = db.Query(ctx, "MATCH (a:Person)-[:KNOWS]->(b:Person) RETURN a.name as person1, b.name as person2", nil)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		printResults(results)
	}

	// Query 4: Find people from the same city
	fmt.Println("\n🏙️ People grouped by city:")
	results, err = db.Query(ctx, "MATCH (p:Person) RETURN p.city as city, collect(p.name) as people", nil)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		printResults(results)
	}

	fmt.Println("\n🎉 Graph exploration complete!")
}

// createSocialNetworkGraphs creates sample graph documents for a social network
func createSocialNetworkGraphs() []graphs.GraphDocument {
	// Create a mock source document
	sourceDoc := schema.Document{
		PageContent: "Social network data for demonstration",
		Metadata: map[string]any{
			"source": "example",
			"type":   "social_network",
		},
	}

	// Create graph document
	graphDoc := graphs.NewGraphDocument(sourceDoc)

	// Create people (nodes)
	alice := graphs.NewNode("alice", "Person")
	alice.SetProperty("name", "Alice Johnson")
	alice.SetProperty("age", 28)
	alice.SetProperty("city", "New York")
	alice.SetProperty("occupation", "Software Engineer")

	bob := graphs.NewNode("bob", "Person")
	bob.SetProperty("name", "Bob Smith")
	bob.SetProperty("age", 32)
	bob.SetProperty("city", "New York")
	bob.SetProperty("occupation", "Data Scientist")

	carol := graphs.NewNode("carol", "Person")
	carol.SetProperty("name", "Carol Davis")
	carol.SetProperty("age", 25)
	carol.SetProperty("city", "San Francisco")
	carol.SetProperty("occupation", "Designer")

	david := graphs.NewNode("david", "Person")
	david.SetProperty("name", "David Wilson")
	david.SetProperty("age", 35)
	david.SetProperty("city", "San Francisco")
	david.SetProperty("occupation", "Product Manager")

	eve := graphs.NewNode("eve", "Person")
	eve.SetProperty("name", "Eve Brown")
	eve.SetProperty("age", 29)
	eve.SetProperty("city", "Boston")
	eve.SetProperty("occupation", "Marketing Manager")

	// Add nodes to graph
	graphDoc.AddNode(alice)
	graphDoc.AddNode(bob)
	graphDoc.AddNode(carol)
	graphDoc.AddNode(david)
	graphDoc.AddNode(eve)

	// Create relationships
	// Alice knows Bob (colleagues)
	rel1 := graphs.NewRelationship(alice, bob, "KNOWS")
	rel1.SetProperty("since", "2020")
	rel1.SetProperty("relationship_type", "colleague")
	graphDoc.AddRelationship(rel1)

	// Bob knows Carol (friends from college)
	rel2 := graphs.NewRelationship(bob, carol, "KNOWS")
	rel2.SetProperty("since", "2015")
	rel2.SetProperty("relationship_type", "friend")
	graphDoc.AddRelationship(rel2)

	// Carol knows David (worked together)
	rel3 := graphs.NewRelationship(carol, david, "KNOWS")
	rel3.SetProperty("since", "2019")
	rel3.SetProperty("relationship_type", "former_colleague")
	graphDoc.AddRelationship(rel3)

	// David knows Eve (siblings)
	rel4 := graphs.NewRelationship(david, eve, "FAMILY")
	rel4.SetProperty("relationship_type", "sibling")
	graphDoc.AddRelationship(rel4)

	// Alice knows Carol (met through Bob)
	rel5 := graphs.NewRelationship(alice, carol, "KNOWS")
	rel5.SetProperty("since", "2021")
	rel5.SetProperty("relationship_type", "friend")
	rel5.SetProperty("introduced_by", "Bob")
	graphDoc.AddRelationship(rel5)

	return []graphs.GraphDocument{graphDoc}
}

// printResults prints query results in a formatted way
func printResults(results []map[string]any) {
	if len(results) == 0 {
		fmt.Println("  No results found.")
		return
	}

	// Print results
	for i, result := range results {
		fmt.Printf("  Result %d: ", i+1)
		for key, value := range result {
			fmt.Printf("%s=%v ", key, value)
		}
		fmt.Println()
	}
}
