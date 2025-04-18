package alloydb

import "fmt"

type distanceStrategy interface {
	String() string
	operator() string
	searchFunction() string
	similaritySearchFunction() string
}

type Index interface {
	Options() string
}

type Euclidean struct{}

func (Euclidean) String() string {
	return "euclidean"
}

func (Euclidean) operator() string {
	return "<->"
}

func (Euclidean) searchFunction() string {
	return "vector_l2_ops"
}

func (Euclidean) similaritySearchFunction() string {
	return "l2_distance"
}

type CosineDistance struct{}

func (CosineDistance) String() string {
	return "cosineDistance"
}

func (CosineDistance) operator() string {
	return "<=>"
}

func (CosineDistance) searchFunction() string {
	return "vector_cosine_ops"
}

func (CosineDistance) similaritySearchFunction() string {
	return "cosine_distance"
}

type InnerProduct struct{}

func (InnerProduct) String() string {
	return "innerProduct"
}

func (InnerProduct) operator() string {
	return "<#>"
}

func (InnerProduct) searchFunction() string {
	return "vector_ip_ops"
}

func (InnerProduct) similaritySearchFunction() string {
	return "inner_product"
}

// HNSWOptions holds the configuration for the hnsw index.
type HNSWOptions struct {
	M              int
	EfConstruction int
}

func (h HNSWOptions) Options() string {
	return fmt.Sprintf("(m = %d, ef_construction = %d)", h.M, h.EfConstruction)
}

// IVFFlatOptions holds the configuration for the ivfflat index.
type IVFFlatOptions struct {
	Lists int
}

func (i IVFFlatOptions) Options() string {
	return fmt.Sprintf("(lists = %d)", i.Lists)
}

// IVFOptions holds the configuration for the ivf index.
type IVFOptions struct {
	Lists     int
	Quantizer string
}

func (i IVFOptions) Options() string {
	return fmt.Sprintf("(lists = %d, quantizer = %s)", i.Lists, i.Quantizer)
}

// SCANNOptions holds the configuration for the ScaNN index.
type SCANNOptions struct {
	NumLeaves int
	Quantizer string
}

func (s SCANNOptions) Options() string {
	return fmt.Sprintf("(num_leaves = %d, quantizer = %s)", s.NumLeaves, s.Quantizer)
}

// indexOptions returns the specific options for the index based on the index type.
func (index *BaseIndex) indexOptions() string {
	return index.options.Options()
}
