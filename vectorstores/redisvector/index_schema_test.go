package redisvector

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/schema"
)

func TestGenerateSchema(t *testing.T) {
	t.Parallel()

	testYamlFile := "./testdata/schema.yml"
	generatorYAML := schemaGenerator{
		format:   YAMLSchemaFormat,
		filePath: testYamlFile,
	}
	schemaWithYAML, err := generatorYAML.generate()
	require.NoError(t, err)
	assert.Len(t, schemaWithYAML.Vector, 1)
	assert.Equal(t, FlatVectorAlgorithm, schemaWithYAML.Vector[0].Algorithm)
	assert.Equal(t, 1536, schemaWithYAML.Vector[0].Dims)
	assert.Equal(t, CosineDistanceMetric, schemaWithYAML.Vector[0].DistanceMetric)
	assert.Empty(t, schemaWithYAML.Tag)
	assert.Len(t, schemaWithYAML.Text, 4)
	assert.Len(t, schemaWithYAML.Numeric, 1)

	testJSONFile := "./testdata/schema.json"
	generatorJSON := schemaGenerator{
		format:   JSONSchemaFormat,
		filePath: testJSONFile,
	}
	schemaWithJSON, err := generatorJSON.generate()
	require.NoError(t, err)
	assert.Len(t, schemaWithJSON.Vector, 1)
	assert.Equal(t, FlatVectorAlgorithm, schemaWithJSON.Vector[0].Algorithm)
	assert.Equal(t, 1536, schemaWithJSON.Vector[0].Dims)
	assert.Equal(t, CosineDistanceMetric, schemaWithJSON.Vector[0].DistanceMetric)
	assert.Empty(t, schemaWithJSON.Tag)
	assert.Len(t, schemaWithJSON.Text, 4)
	assert.Len(t, schemaWithJSON.Numeric, 1)

	data := []schema.Document{
		{
			PageContent: "Tokyo",
			Metadata: map[string]any{
				"population":        38.2,
				"area":              2190,
				"content":           "foo",
				"content_vector":    []float32{0.1},
				"tag":               []int{1, 2},
				"ignore":            nil,
				"ignore_other_type": map[string]string{},
			},
		},
	}
	schema, err := generateSchemaWithMetadata(data[0].Metadata)
	require.NoError(t, err)
	assert.Len(t, schema.Vector, 1)
	assert.Equal(t, FlatVectorAlgorithm, schema.Vector[0].Algorithm)
	assert.Equal(t, 1, schema.Vector[0].Dims)
	assert.Len(t, schema.Tag, 1)
	assert.Equal(t, ",", schema.Tag[0].Separator)
	assert.Len(t, schema.Text, 1)
	assert.Len(t, schema.Numeric, 2)
}

func TestSchemaAsCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args RedisIndexSchemaField
		want string
	}{
		{
			"TagField: with all keys",
			TagField{Name: "tag", As: "_tag", Separator: "|", NoIndex: true, Sortable: true, CaseSensitive: true},
			"tag AS _tag TAG | CASESENSITIVE NOINDEX SORTABLE",
		},
		{
			"TagField: with separator",
			TagField{Name: "tag", Separator: "|"},
			"tag TAG |",
		},
		{
			"TagField: only name",
			TagField{Name: "tag"},
			"tag TAG ,",
		},
		{
			"TextField: with all keys",
			TextField{Name: "text", As: "_text", Weight: 0.6, NoIndex: true, Sortable: true, NoStem: true, WithSuffixtrie: true, PhoneticMatcher: "dm:en"},
			"text AS _text TEXT WEIGHT 0.6 PHONETIC dm:en WITHSUFFIXTRIE NOSTEM NOINDEX SORTABLE",
		},
		{
			"TextField: with invalid enum value",
			TextField{Name: "text", NoIndex: true, Sortable: true, NoStem: true, WithSuffixtrie: true, PhoneticMatcher: "invalid enum"},
			"text TEXT WITHSUFFIXTRIE NOSTEM NOINDEX SORTABLE",
		},
		{
			"TextField: weight=1",
			TextField{Name: "text", As: "_text", Weight: 1},
			"text AS _text TEXT",
		},
		{
			"NumericField: with all keys",
			NumericField{Name: "number", As: "_number", NoIndex: true, Sortable: true},
			"number AS _number NUMERIC NOINDEX SORTABLE",
		},
		{
			"VectorField: FLAT vector with all keys",
			VectorField{Name: "vector", As: "_vector", Algorithm: FlatVectorAlgorithm, Dims: 1024, DistanceMetric: L2DistanceMetric, Datatype: FLOAT64VectorDataType, BlockSize: 100},
			"vector AS _vector VECTOR FLAT 8 TYPE FLOAT64 DIM 1024 DISTANCE_METRIC L2 BLOCK_SIZE 100",
		},
		{
			"VectorField: FLAT vector with default value",
			VectorField{Name: "vector", As: "_vector", BlockSize: 100},
			"vector AS _vector VECTOR FLAT 8 TYPE FLOAT32 DIM 128 DISTANCE_METRIC COSINE BLOCK_SIZE 100",
		},
		{
			"VectorField: HNSW vector with all keys",
			VectorField{Name: "vector", As: "_vector", Algorithm: HNSWVectorAlgorithm, Dims: 1024, DistanceMetric: CosineDistanceMetric, Datatype: FLOAT64VectorDataType, M: 10, EfConstruction: 100, EfRuntime: 1000, Epsilon: 0.5},
			"vector AS _vector VECTOR HNSW 14 TYPE FLOAT64 DIM 1024 DISTANCE_METRIC COSINE M 10 EF_CONSTRUCTION 100 EF_RUNTIME 1000 EPSILON 0.5",
		},
		{
			"VectorField: HNSW vector with basic keys",
			VectorField{Name: "vector", Algorithm: HNSWVectorAlgorithm, Dims: 1024, DistanceMetric: CosineDistanceMetric, Datatype: FLOAT32VectorDataType},
			"vector VECTOR HNSW 6 TYPE FLOAT32 DIM 1024 DISTANCE_METRIC COSINE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, strings.Join(tt.args.AsCommand(), " "))
		})
	}
}

func TestIndexSearchAsCommand(t *testing.T) {
	t.Parallel()

	type Args struct {
		name   string
		vector []float32
		opts   []SearchOption
	}

	tests := []struct {
		name string
		args Args
		want string
	}{
		{
			"basic search",
			Args{"demo", []float32{0.111}, []SearchOption{}},
			"FT.SEARCH demo (*)=>[KNN 1 @content_vector $vector AS distance] SORTBY distance ASC DIALECT 2 LIMIT 0 1 PARAMS 2 vector \xf8S\xe3=",
		},
		{
			"search with limit",
			Args{"demo", []float32{0.111}, []SearchOption{WithOffsetLimit(0, 10)}},
			"FT.SEARCH demo (*)=>[KNN 10 @content_vector $vector AS distance] SORTBY distance ASC DIALECT 2 LIMIT 0 10 PARAMS 2 vector \xf8S\xe3=",
		},
		{
			"search with score threshold",
			Args{"demo", []float32{0.111}, []SearchOption{WithScoreThreshold(0.5)}},
			"FT.SEARCH demo @content_vector:[VECTOR_RANGE $distance_threshold $vector]=>{$yield_distance_as: distance} SORTBY distance ASC DIALECT 2 LIMIT 0 1 PARAMS 4 vector \xf8S\xe3= distance_threshold 0.5",
		},
		{
			"search with filter",
			Args{"demo", []float32{0.111}, []SearchOption{WithPreFilters("@job{engineer}")}},
			"FT.SEARCH demo (@job{engineer})=>[KNN 1 @content_vector $vector AS distance] SORTBY distance ASC DIALECT 2 LIMIT 0 1 PARAMS 2 vector \xf8S\xe3=",
		},
		{
			"search with score threshold and filter",
			Args{"demo", []float32{0.111}, []SearchOption{WithScoreThreshold(0.5), WithPreFilters("@job{engineer}")}},
			"FT.SEARCH demo (@job{engineer}) @content_vector:[VECTOR_RANGE $distance_threshold $vector]=>{$yield_distance_as: distance} SORTBY distance ASC DIALECT 2 LIMIT 0 1 PARAMS 4 vector \xf8S\xe3= distance_threshold 0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			search, err := NewIndexVectorSearch(tt.args.name, tt.args.vector, tt.args.opts...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, strings.Join(search.AsCommand(), " "))
		})
	}
}
