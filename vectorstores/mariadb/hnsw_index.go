package mariadb

// DistanceFunction represents the type of distance function.
type DistanceFunction string

const (
	Cosine    DistanceFunction = "cosine"
	Euclidean DistanceFunction = "euclidean"
)

// HNSWIndex stores parameters for configuring the HNSW index.
type HNSWIndex struct {
	M            int
	DistanceFunc DistanceFunction
}
