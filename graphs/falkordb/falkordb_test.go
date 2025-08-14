package falkordb

import (
	"context"
	"testing"

	"github.com/0xDezzy/langchaingo/graphs"
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/stretchr/testify/assert"
)

func TestBuildPropertiesString(t *testing.T) {
	f := &FalkorDB{}

	tests := []struct {
		name     string
		props    map[string]any
		expected string
	}{
		{
			name:     "empty properties",
			props:    map[string]any{},
			expected: "",
		},
		{
			name: "string property",
			props: map[string]any{
				"name": "Alice",
			},
			expected: ", name:'Alice'",
		},
		{
			name: "mixed properties",
			props: map[string]any{
				"name":   "Alice",
				"age":    30,
				"active": true,
			},
			expected: ", active:true, age:30, name:'Alice'", // Note: map iteration order varies
		},
		{
			name: "string with quotes",
			props: map[string]any{
				"description": "Alice's profile",
			},
			expected: ", description:'Alice\\'s profile'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.buildPropertiesString(tt.props)
			if tt.name == "mixed properties" {
				// For mixed properties, just check that it contains all expected parts
				assert.Contains(t, result, "name:'Alice'")
				assert.Contains(t, result, "age:30")
				assert.Contains(t, result, "active:true")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGraphDocumentOperations(t *testing.T) {
	// Test graph document creation and manipulation
	sourceDoc := schema.Document{
		PageContent: "Test document",
		Metadata:    map[string]any{"source": "test"},
	}

	graphDoc := graphs.NewGraphDocument(sourceDoc)

	// Test node creation
	node1 := graphs.NewNode("1", "Person")
	node1.SetProperty("name", "Alice")
	node1.SetProperty("age", 30)

	node2 := graphs.NewNode("2", "Person")
	node2.SetProperty("name", "Bob")
	node2.SetProperty("age", 25)

	// Add nodes
	graphDoc.AddNode(node1)
	graphDoc.AddNode(node2)

	assert.Len(t, graphDoc.Nodes, 2)
	assert.Equal(t, "Alice", graphDoc.Nodes[0].Properties["name"])
	assert.Equal(t, 30, graphDoc.Nodes[0].Properties["age"])

	// Test relationship creation
	rel := graphs.NewRelationship(node1, node2, "KNOWS")
	rel.SetProperty("since", "2020")

	graphDoc.AddRelationship(rel)

	assert.Len(t, graphDoc.Relationships, 1)
	assert.Equal(t, "KNOWS", graphDoc.Relationships[0].Type)
	assert.Equal(t, "2020", graphDoc.Relationships[0].Properties["since"])
}

func TestNodeOperations(t *testing.T) {
	// Test node property operations
	node := graphs.NewNode("test", "TestType")

	// Test setting properties
	node.SetProperty("key1", "value1")
	node.SetProperty("key2", 42)

	// Test getting properties
	val1, exists1 := node.GetProperty("key1")
	assert.True(t, exists1)
	assert.Equal(t, "value1", val1)

	val2, exists2 := node.GetProperty("key2")
	assert.True(t, exists2)
	assert.Equal(t, 42, val2)

	// Test non-existent property
	_, exists3 := node.GetProperty("nonexistent")
	assert.False(t, exists3)
}

func TestRelationshipOperations(t *testing.T) {
	// Test relationship property operations
	source := graphs.NewNode("1", "Person")
	target := graphs.NewNode("2", "Person")

	rel := graphs.NewRelationship(source, target, "KNOWS")

	// Test setting properties
	rel.SetProperty("since", "2020")
	rel.SetProperty("strength", 0.8)

	// Test getting properties
	since, exists1 := rel.GetProperty("since")
	assert.True(t, exists1)
	assert.Equal(t, "2020", since)

	strength, exists2 := rel.GetProperty("strength")
	assert.True(t, exists2)
	assert.Equal(t, 0.8, strength)

	// Test non-existent property
	_, exists3 := rel.GetProperty("nonexistent")
	assert.False(t, exists3)
}

func TestOptions(t *testing.T) {
	// Test functional options
	options := graphs.NewOptions()

	// Test default values
	assert.False(t, options.IncludeSource)
	assert.Equal(t, 100, options.BatchSize)
	assert.Equal(t, 0, options.Timeout)

	// Test option functions
	withSource := graphs.WithIncludeSource(true)
	withBatchSize := graphs.WithBatchSize(50)
	withTimeout := graphs.WithTimeout(5000)

	withSource(options)
	withBatchSize(options)
	withTimeout(options)

	assert.True(t, options.IncludeSource)
	assert.Equal(t, 50, options.BatchSize)
	assert.Equal(t, 5000, options.Timeout)
}

// Integration test that requires a running FalkorDB instance
// This test is skipped by default - run with FalkorDB running to test
func TestFalkorDBIntegration(t *testing.T) {
	t.Skip("Integration test - requires running FalkorDB instance")

	ctx := context.Background()

	// Connect to FalkorDB
	db, err := New("test_graph",
		WithHost("localhost"),
		WithPort(6379),
	)
	if err != nil {
		t.Fatalf("Failed to connect to FalkorDB: %v", err)
	}
	defer db.Close()

	// Create test graph document
	sourceDoc := schema.Document{
		PageContent: "Integration test",
		Metadata:    map[string]any{"test": true},
	}

	graphDoc := graphs.NewGraphDocument(sourceDoc)

	node1 := graphs.NewNode("test1", "TestPerson")
	node1.SetProperty("name", "TestAlice")
	graphDoc.AddNode(node1)

	node2 := graphs.NewNode("test2", "TestPerson")
	node2.SetProperty("name", "TestBob")
	graphDoc.AddNode(node2)

	rel := graphs.NewRelationship(node1, node2, "TEST_KNOWS")
	graphDoc.AddRelationship(rel)

	// Add to database
	err = db.AddGraphDocuments(ctx, []graphs.GraphDocument{graphDoc})
	assert.NoError(t, err)

	// Query the data
	results, err := db.Query(ctx, "MATCH (p:TestPerson) RETURN p.name", nil)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Clean up
	_, err = db.Query(ctx, "MATCH (n:TestPerson) DETACH DELETE n", nil)
	assert.NoError(t, err)
}
