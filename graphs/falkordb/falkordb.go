package falkordb

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/0xDezzy/langchaingo/graphs"
	redisgraph "github.com/FalkorDB/falkordb-go"
	"github.com/gomodule/redigo/redis"
)

const (
	nodePropertiesQuery = `
		MATCH (n)
		WITH keys(n) as keys, labels(n) AS labels
		WITH CASE WHEN keys = [] THEN [NULL] ELSE keys END AS keys, labels
		UNWIND labels AS label
		UNWIND keys AS key
		WITH label, collect(DISTINCT key) AS keys
		RETURN {label:label, keys:keys} AS output
	`

	relPropertiesQuery = `
		MATCH ()-[r]->()
		WITH keys(r) as keys, type(r) AS types
		WITH CASE WHEN keys = [] THEN [NULL] ELSE keys END AS keys, types 
		UNWIND types AS type
		UNWIND keys AS key WITH type,
		collect(DISTINCT key) AS keys 
		RETURN {types:type, keys:keys} AS output
	`

	relationshipsQuery = `
		MATCH (n)-[r]->(m)
		UNWIND labels(n) as src_label
		UNWIND labels(m) as dst_label
		UNWIND type(r) as rel_type
		RETURN DISTINCT {start: src_label, type: rel_type, end: dst_label} AS output
	`
)

// FalkorDB implements the graphs.GraphStore interface for FalkorDB.
type FalkorDB struct {
	// Connection configuration
	host         string
	port         int
	username     string
	password     string
	ssl          bool
	httpClient   *http.Client
	timeout      time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration

	// FalkorDB client and graph
	conn  redis.Conn
	graph *redisgraph.Graph

	// Schema caching
	schema           string
	structuredSchema map[string]any

	// Graph database name
	database string
}

// New creates a new FalkorDB graph store instance.
func New(database string, opts ...Option) (*FalkorDB, error) {
	f := &FalkorDB{
		host:             "localhost",
		port:             6379,
		ssl:              false,
		timeout:          30 * time.Second,
		readTimeout:      30 * time.Second,
		writeTimeout:     30 * time.Second,
		database:         database,
		structuredSchema: make(map[string]any),
	}

	// Apply options
	for _, opt := range opts {
		opt(f)
	}

	// Create connection
	if err := f.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to FalkorDB: %w", err)
	}

	// Initialize schema
	if err := f.RefreshSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return f, nil
}

// connect establishes connection to FalkorDB.
func (f *FalkorDB) connect() error {
	// Create Redis connection
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", f.host, f.port))
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Authenticate if credentials provided
	if f.username != "" && f.password != "" {
		if _, err := conn.Do("AUTH", f.username, f.password); err != nil {
			conn.Close()
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	} else if f.password != "" {
		if _, err := conn.Do("AUTH", f.password); err != nil {
			conn.Close()
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	f.conn = conn
	f.graph = &redisgraph.Graph{}
	*f.graph = f.graph.New(f.database, conn)

	return nil
}

// AddGraphDocuments adds graph documents to the FalkorDB store.
func (f *FalkorDB) AddGraphDocuments(ctx context.Context, docs []graphs.GraphDocument, options ...graphs.Option) error {
	opts := graphs.NewOptions()
	for _, option := range options {
		option(opts)
	}

	for _, doc := range docs {
		// Add nodes
		for _, node := range doc.Nodes {
			if err := f.addNode(ctx, node); err != nil {
				return fmt.Errorf("failed to add node %s: %w", node.ID, err)
			}
		}

		// Add relationships
		for _, rel := range doc.Relationships {
			if err := f.addRelationship(ctx, rel); err != nil {
				return fmt.Errorf("failed to add relationship %s->%s: %w", rel.Source.ID, rel.Target.ID, err)
			}
		}
	}

	return nil
}

// addNode adds a single node to the graph.
func (f *FalkorDB) addNode(ctx context.Context, node graphs.Node) error {
	// Build properties string
	propsStr := f.buildPropertiesString(node.Properties)

	query := fmt.Sprintf("MERGE (n:%s {id:'%s'%s}) RETURN distinct 'done' AS result",
		node.Type, node.ID, propsStr)

	_, err := f.graph.Query(query)
	if err != nil {
		return fmt.Errorf("failed to execute node query: %w", err)
	}

	return nil
}

// addRelationship adds a single relationship to the graph.
func (f *FalkorDB) addRelationship(ctx context.Context, rel graphs.Relationship) error {
	// Sanitize relationship type (replace spaces with underscores and uppercase)
	relType := strings.ToUpper(strings.ReplaceAll(rel.Type, " ", "_"))

	// Build properties string for relationship
	propsStr := f.buildPropertiesString(rel.Properties)

	query := fmt.Sprintf(
		"MATCH (a:%s {id:'%s'}), (b:%s {id:'%s'}) MERGE (a)-[r:%s%s]->(b) RETURN distinct 'done' AS result",
		rel.Source.Type, rel.Source.ID,
		rel.Target.Type, rel.Target.ID,
		relType, propsStr)

	_, err := f.graph.Query(query)
	if err != nil {
		return fmt.Errorf("failed to execute relationship query: %w", err)
	}

	return nil
}

// Query executes a Cypher query against the FalkorDB instance.
func (f *FalkorDB) Query(ctx context.Context, query string, params map[string]any) ([]map[string]any, error) {
	// Note: FalkorDB's redisgraph client doesn't support parameterized queries the same way
	// For now, we'll execute the query as-is and handle parameters through string formatting
	result, err := f.graph.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Convert result to map format
	var results []map[string]any
	if len(result.Results) > 1 { // First row is headers
		headers := result.Results[0]
		for i := 1; i < len(result.Results); i++ {
			row := make(map[string]any)
			for j, header := range headers {
				if j < len(result.Results[i]) {
					row[header] = result.Results[i][j]
				}
			}
			results = append(results, row)
		}
	}

	return results, nil
}

// RefreshSchema refreshes the schema information from FalkorDB.
func (f *FalkorDB) RefreshSchema(ctx context.Context) error {
	// Query node properties
	nodeProperties, err := f.Query(ctx, nodePropertiesQuery, nil)
	if err != nil {
		return fmt.Errorf("failed to query node properties: %w", err)
	}

	// Query relationship properties
	relProperties, err := f.Query(ctx, relPropertiesQuery, nil)
	if err != nil {
		return fmt.Errorf("failed to query relationship properties: %w", err)
	}

	// Query relationships
	relationships, err := f.Query(ctx, relationshipsQuery, nil)
	if err != nil {
		return fmt.Errorf("failed to query relationships: %w", err)
	}

	// Build structured schema
	f.structuredSchema = map[string]any{
		"node_props":    extractNodeProperties(nodeProperties),
		"rel_props":     extractRelProperties(relProperties),
		"relationships": extractRelationships(relationships),
	}

	// Build schema string
	f.schema = fmt.Sprintf("Node properties: %v\nRelationships properties: %v\nRelationships: %v\n",
		nodeProperties, relProperties, relationships)

	return nil
}

// extractNodeProperties extracts node properties from query results.
func extractNodeProperties(results []map[string]any) map[string][]string {
	nodeProps := make(map[string][]string)
	for _, result := range results {
		if output, ok := result["output"].(map[string]any); ok {
			if label, ok := output["label"].(string); ok {
				if keys, ok := output["keys"].([]any); ok {
					var keyStrings []string
					for _, key := range keys {
						if keyStr, ok := key.(string); ok {
							keyStrings = append(keyStrings, keyStr)
						}
					}
					nodeProps[label] = keyStrings
				}
			}
		}
	}
	return nodeProps
}

// extractRelProperties extracts relationship properties from query results.
func extractRelProperties(results []map[string]any) map[string][]string {
	relProps := make(map[string][]string)
	for _, result := range results {
		if output, ok := result["output"].(map[string]any); ok {
			if types, ok := output["types"].(string); ok {
				if keys, ok := output["keys"].([]any); ok {
					var keyStrings []string
					for _, key := range keys {
						if keyStr, ok := key.(string); ok {
							keyStrings = append(keyStrings, keyStr)
						}
					}
					relProps[types] = keyStrings
				}
			}
		}
	}
	return relProps
}

// extractRelationships extracts relationship patterns from query results.
func extractRelationships(results []map[string]any) []map[string]string {
	var relationships []map[string]string
	for _, result := range results {
		if output, ok := result["output"].(map[string]any); ok {
			rel := make(map[string]string)
			if start, ok := output["start"].(string); ok {
				rel["start"] = start
			}
			if relType, ok := output["type"].(string); ok {
				rel["type"] = relType
			}
			if end, ok := output["end"].(string); ok {
				rel["end"] = end
			}
			relationships = append(relationships, rel)
		}
	}
	return relationships
}

// GetSchema returns the current schema as a string.
func (f *FalkorDB) GetSchema() string {
	return f.schema
}

// GetStructuredSchema returns the structured schema information.
func (f *FalkorDB) GetStructuredSchema() map[string]any {
	return f.structuredSchema
}

// buildPropertiesString converts a properties map to a Cypher properties string.
func (f *FalkorDB) buildPropertiesString(props map[string]any) string {
	if len(props) == 0 {
		return ""
	}

	var parts []string
	for key, value := range props {
		switch v := value.(type) {
		case string:
			parts = append(parts, fmt.Sprintf("%s:'%s'", key, strings.ReplaceAll(v, "'", "\\'")))
		case int, int64, float64:
			parts = append(parts, fmt.Sprintf("%s:%v", key, v))
		case bool:
			parts = append(parts, fmt.Sprintf("%s:%t", key, v))
		default:
			// Convert to string as fallback
			parts = append(parts, fmt.Sprintf("%s:'%v'", key, v))
		}
	}

	if len(parts) > 0 {
		return ", " + strings.Join(parts, ", ")
	}
	return ""
}

// Close closes the connection to FalkorDB.
func (f *FalkorDB) Close() error {
	if f.conn != nil {
		err := f.conn.Close()
		f.conn = nil
		f.graph = nil
		return err
	}
	return nil
}
