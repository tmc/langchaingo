package redisvector

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"unsafe"
)

type IndexVectorSearch struct {
	index          string
	vector         []float32
	scoreThreshold float32
	preFilters     string
	returns        []string
	offset         int
	limit          int
	sortBy         []string
}

type SearchOption func(s *IndexVectorSearch)

func NewIndexVectorSearch(index string, vector []float32, opts ...SearchOption) (*IndexVectorSearch, error) {
	if index == "" {
		return nil, errors.New("invalid index")
	}
	if len(vector) == 0 {
		return nil, errors.New("invalid vector")
	}
	s := &IndexVectorSearch{
		index:   index,
		vector:  vector,
		returns: []string{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

func WithScoreThreshold(scoreThreshold float32) SearchOption {
	return func(s *IndexVectorSearch) {
		if scoreThreshold > 0 && scoreThreshold < 1 {
			s.scoreThreshold = scoreThreshold
		}
	}
}

func WithPreFilters(preFilters string) SearchOption {
	return func(s *IndexVectorSearch) {
		if len(preFilters) != 0 {
			s.preFilters = preFilters
		}
	}
}

func WithReturns(returns []string) SearchOption {
	return func(s *IndexVectorSearch) {
		if returns != nil {
			s.returns = returns
		}
	}
}

func WithOffsetLimit(offset, limit int) SearchOption {
	return func(s *IndexVectorSearch) {
		if limit == 0 {
			limit = 1
		}
		s.offset = offset
		s.limit = limit
	}
}

func (s IndexVectorSearch) AsCommand() []string {
	// "FT.SEARCH" "users"
	// "({prefilters})=>[KNN 5 @content_vector $vector AS distance]"
	// "RETURN" "4" "content" "distance" "user" "age"
	// "SORTBY" "distance" "ASC"
	// "DIALECT" "2"
	// "LIMIT" "0" "5"
	// "params" "2" "vector"

	// "FT.SEARCH" "users"
	// "@job:("engineer") @content_vector:[VECTOR_RANGE $distance_threshold $vector]=>{$yield_distance_as: distance}"
	// "RETURN" "4" "content" "distance" "user" "age"
	// "SORTBY" "distance" "ASC"
	// "DIALECT" "2"
	// "LIMIT" "0" "3"
	// "params" "n" "vector" "xxx"  "distance_threshold" "0.1"
	cmd := []string{"FT.SEARCH", s.index}

	if s.limit == 0 {
		s.limit = 1
	}

	const vectorField = "vector"
	const vectorFieldAs = defaultDistanceFieldKey
	const disThresholdFiled = "distance_threshold"
	const vectorKey = defaultContentVectorFieldKey
	params := []string{vectorField, VectorString32(s.vector)}

	if s.scoreThreshold > 0 && s.scoreThreshold < 1 {
		// Range search
		// "@content_vector:[VECTOR_RANGE $distance_threshold $vector]=>{$yield_distance_as: distance}"
		filter := fmt.Sprintf("@%s:[VECTOR_RANGE $%s $%s]=>{$yield_distance_as: %s}", vectorKey, disThresholdFiled, vectorField, vectorFieldAs)
		if len(s.preFilters) > 0 {
			filter = fmt.Sprintf("(%s) %s", s.preFilters, filter)
		}
		cmd = append(cmd, filter)
		params = append(params, disThresholdFiled, strconv.FormatFloat(float64(s.scoreThreshold), 'f', -1, 32))
	} else {
		// KNN search
		// "(*)=>[KNN n @content_vector $vector AS distance]"
		filter := "*"
		if len(s.preFilters) > 0 {
			filter = s.preFilters
		}
		cmd = append(cmd, fmt.Sprintf("(%s)=>[KNN %d @%s $%s AS %s]", filter, s.limit, vectorKey, vectorField, vectorFieldAs))
	}

	if l := len(s.returns); l > 0 {
		s.returns = append(s.returns, defaultDistanceFieldKey)
		cmd = append(cmd, "RETURN", strconv.Itoa(len(s.returns)))
		cmd = append(cmd, s.returns...)
	}

	cmd = append(cmd, "SORTBY")
	if len(s.sortBy) == 0 {
		s.sortBy = []string{vectorFieldAs, "ASC"}
	}
	cmd = append(cmd, s.sortBy...)

	cmd = append(cmd, "DIALECT", "2")
	cmd = append(cmd, "LIMIT", strconv.Itoa(s.offset), strconv.Itoa(s.limit))

	cmd = append(cmd, "PARAMS", strconv.Itoa(len(params)))
	cmd = append(cmd, params...)

	return cmd
}

// convert []float32 into string.
func VectorString32(v []float32) string {
	b := make([]byte, len(v)*4)
	for i, e := range v {
		i := i * 4
		binary.LittleEndian.PutUint32(b[i:i+4], math.Float32bits(e))
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// convert []float64 into string.
func VectorString64(v []float64) string {
	b := make([]byte, len(v)*8)
	for i, e := range v {
		i := i * 8
		binary.LittleEndian.PutUint64(b[i:i+8], math.Float64bits(e))
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}
