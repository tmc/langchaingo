# 🕸️ FalkorDB Graph Database Example 📊

Welcome to the FalkorDB Graph Database Example! This program demonstrates how to use FalkorDB with langchaingo to create and query knowledge graphs. FalkorDB is a super-fast graph database that's perfect for building GraphRAG applications!

## 🚀 What Does This Program Do?

This example showcases the power of graph databases:

1. **Connects to FalkorDB**: Sets up a connection to a local FalkorDB instance.
2. **Creates Graph Documents**: Builds a social network with people and their relationships.
3. **Adds Nodes and Relationships**: Inserts people (nodes) and their connections (relationships) into the graph.
4. **Queries the Graph**: Demonstrates various Cypher queries to explore the data.
5. **Schema Introspection**: Shows how to examine the graph structure and properties.

## 🌐 The Social Network

Our example creates a simple social network with:

- **People**: Nodes representing individuals with properties like name, age, and city
- **Relationships**: Connections showing who knows whom, who works where, and family relationships
- **Complex Queries**: Finding mutual friends, shortest paths, and community detection

## 🛠️ Prerequisites

Before running this example, you'll need:

1. **FalkorDB Running**: Start a FalkorDB instance (usually via Docker):
   ```bash
   docker run -p 6379:6379 falkordb/falkordb:edge
   ```

2. **Go Environment**: Make sure you have Go installed and configured.

## 🎯 Key Features Demonstrated

- **Graph Document Creation**: Building structured graph data
- **Batch Operations**: Efficiently adding multiple nodes and relationships  
- **Cypher Queries**: Executing graph traversals and pattern matching
- **Schema Management**: Introspecting node types and relationship patterns
- **Property Handling**: Working with typed properties on nodes and edges

## 🔍 Sample Queries

The program demonstrates several types of graph queries:

- **Find all people**: `MATCH (p:Person) RETURN p.name, p.age`
- **Explore relationships**: `MATCH (a:Person)-[r:KNOWS]->(b:Person) RETURN a.name, b.name`
- **Path finding**: `MATCH path = (a:Person)-[:KNOWS*1..3]-(b:Person) RETURN path`
- **Aggregations**: `MATCH (p:Person) RETURN p.city, count(*) as population`

## 🚀 Ready to Run?

Just make sure FalkorDB is running on localhost:6379, then execute the program. You'll see the social network being built and various queries exploring the connections!

This is perfect for learning about:
- Graph database concepts
- Knowledge graph construction  
- GraphRAG foundations
- Social network analysis
- Recommendation systems

Happy graph exploring! 🕸️✨