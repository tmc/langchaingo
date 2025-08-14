package graphs

import (
	"encoding/json"

	"github.com/0xDezzy/langchaingo/schema"
)

// Node represents a node in a graph with associated properties.
type Node struct {
	// ID is a unique identifier for the node
	ID string `json:"id"`
	// Type is the type or label of the node
	Type string `json:"type"`
	// Properties contains additional properties and metadata associated with the node
	Properties map[string]interface{} `json:"properties"`
}

// NewNode creates a new Node with the given ID and type.
func NewNode(id, nodeType string) Node {
	return Node{
		ID:         id,
		Type:       nodeType,
		Properties: make(map[string]interface{}),
	}
}

// SetProperty sets a property on the node.
func (n *Node) SetProperty(key string, value interface{}) {
	if n.Properties == nil {
		n.Properties = make(map[string]interface{})
	}
	n.Properties[key] = value
}

// GetProperty gets a property from the node.
func (n *Node) GetProperty(key string) (interface{}, bool) {
	if n.Properties == nil {
		return nil, false
	}
	val, ok := n.Properties[key]
	return val, ok
}

// Relationship represents a directed relationship between two nodes in a graph.
type Relationship struct {
	// Source is the source node of the relationship
	Source Node `json:"source"`
	// Target is the target node of the relationship
	Target Node `json:"target"`
	// Type is the type of the relationship
	Type string `json:"type"`
	// Properties contains additional properties associated with the relationship
	Properties map[string]interface{} `json:"properties"`
}

// NewRelationship creates a new Relationship between source and target nodes.
func NewRelationship(source, target Node, relType string) Relationship {
	return Relationship{
		Source:     source,
		Target:     target,
		Type:       relType,
		Properties: make(map[string]interface{}),
	}
}

// SetProperty sets a property on the relationship.
func (r *Relationship) SetProperty(key string, value interface{}) {
	if r.Properties == nil {
		r.Properties = make(map[string]interface{})
	}
	r.Properties[key] = value
}

// GetProperty gets a property from the relationship.
func (r *Relationship) GetProperty(key string) (interface{}, bool) {
	if r.Properties == nil {
		return nil, false
	}
	val, ok := r.Properties[key]
	return val, ok
}

// GraphDocument represents a graph document consisting of nodes and relationships.
type GraphDocument struct {
	// Nodes is a list of nodes in the graph
	Nodes []Node `json:"nodes"`
	// Relationships is a list of relationships in the graph
	Relationships []Relationship `json:"relationships"`
	// Source is the document from which the graph information is derived
	Source schema.Document `json:"source"`
}

// NewGraphDocument creates a new GraphDocument with the given source document.
func NewGraphDocument(source schema.Document) GraphDocument {
	return GraphDocument{
		Nodes:         make([]Node, 0),
		Relationships: make([]Relationship, 0),
		Source:        source,
	}
}

// AddNode adds a node to the graph document.
func (gd *GraphDocument) AddNode(node Node) {
	gd.Nodes = append(gd.Nodes, node)
}

// AddRelationship adds a relationship to the graph document.
func (gd *GraphDocument) AddRelationship(rel Relationship) {
	gd.Relationships = append(gd.Relationships, rel)
}

// ToJSON converts the GraphDocument to JSON.
func (gd *GraphDocument) ToJSON() ([]byte, error) {
	return json.Marshal(gd)
}

// FromJSON creates a GraphDocument from JSON.
func FromJSON(data []byte) (*GraphDocument, error) {
	var gd GraphDocument
	err := json.Unmarshal(data, &gd)
	if err != nil {
		return nil, err
	}
	return &gd, nil
}
