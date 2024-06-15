package redisvector

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"reflect"
	"strconv"

	"sigs.k8s.io/yaml"
)

var (
	ErrInvalidSchemaFormat = errors.New("invalid schema format")
	ErrEmptySchemaContent  = errors.New("empty schema content")
)

type (
	IndexType           string
	DistanceMetric      string
	VectorAlgorithm     string
	VectorDataType      string
	PhoneticMatcherType string
)

const (
	// IndexType enum values.
	JSONIndexType IndexType = "JSON"
	HASHIndexType IndexType = "HASH"

	// Distance Metric enums values.
	L2DistanceMetric     DistanceMetric = "L2"
	CosineDistanceMetric DistanceMetric = "COSINE"
	IPDistanceMetric     DistanceMetric = "IP"

	// Vector Algorithm enum values.
	FlatVectorAlgorithm VectorAlgorithm = "FLAT"
	HNSWVectorAlgorithm VectorAlgorithm = "HNSW"

	// Vector DataType enum values.
	FLOAT32VectorDataType VectorDataType = "FLOAT32"
	FLOAT64VectorDataType VectorDataType = "FLOAT64"

	// Phonetic Matchers enum values.
	PhoneticDoubleMetaphoneEnglish    PhoneticMatcherType = "dm:en"
	PhoneticDoubleMetaphoneFrench     PhoneticMatcherType = "dm:fr"
	PhoneticDoubleMetaphonePortuguese PhoneticMatcherType = "dm:pt"
	PhoneticDoubleMetaphoneSpanish    PhoneticMatcherType = "dm:es"
)

type RedisIndexSchemaField interface {
	// AsCommand convert schema into redis command string
	AsCommand() []string
}

type TagField struct {
	Name          string
	As            string `json:"as,omitempty"             yaml:"as,omitempty"`
	Separator     string `json:"separator"                yaml:"separator"` // default=","
	NoIndex       bool   `json:"no_index,omitempty"       yaml:"no_index,omitempty"`
	Sortable      bool   `json:"sortable,omitempty"       yaml:"sortable,omitempty"`
	CaseSensitive bool   `json:"case_sensitive,omitempty" yaml:"case_sensitive,omitempty"`
}

func (f TagField) AsCommand() []string {
	argsOut := []string{f.Name}
	if f.As != "" {
		argsOut = append(argsOut, "AS", f.As, "TAG")
	} else {
		argsOut = append(argsOut, "TAG")
	}
	if f.Separator == "" {
		// default ,
		argsOut = append(argsOut, ",")
	} else {
		argsOut = append(argsOut, f.Separator)
	}
	if f.CaseSensitive {
		argsOut = append(argsOut, "CASESENSITIVE")
	}
	if f.NoIndex {
		argsOut = append(argsOut, "NOINDEX")
	}
	if f.Sortable {
		argsOut = append(argsOut, "SORTABLE")
	}
	return argsOut
}

type TextField struct {
	Name            string
	As              string              `json:"as,omitempty" yaml:"as,omitempty"`
	Weight          float32             // default=1
	NoStem          bool                `json:"no_stem,omitempty"          yaml:"no_stem,omitempty"`
	WithSuffixtrie  bool                `json:"withsuffixtrie,omitempty"   yaml:"withsuffixtrie,omitempty"`
	PhoneticMatcher PhoneticMatcherType `json:"phonetic_matcher,omitempty" yaml:"phonetic_matcher,omitempty"`
	NoIndex         bool                `json:"no_index,omitempty"         yaml:"no_index,omitempty"`
	Sortable        bool                `json:"sortable,omitempty"         yaml:"sortable,omitempty"`
}

func (f TextField) AsCommand() []string {
	argsOut := []string{f.Name}
	if f.As != "" {
		argsOut = append(argsOut, "AS", f.As, "TEXT")
	} else {
		argsOut = append(argsOut, "TEXT")
	}
	if f.Weight != 0 && f.Weight != 1 {
		argsOut = append(argsOut, "WEIGHT", strconv.FormatFloat(float64(f.Weight), 'f', -1, 32))
	}
	if f.PhoneticMatcher == PhoneticDoubleMetaphoneEnglish || f.PhoneticMatcher == PhoneticDoubleMetaphoneFrench || f.PhoneticMatcher == PhoneticDoubleMetaphonePortuguese || f.PhoneticMatcher == PhoneticDoubleMetaphoneSpanish {
		argsOut = append(argsOut, "PHONETIC", string(f.PhoneticMatcher))
	}
	if f.WithSuffixtrie {
		argsOut = append(argsOut, "WITHSUFFIXTRIE")
	}
	if f.NoStem {
		argsOut = append(argsOut, "NOSTEM")
	}
	if f.NoIndex {
		argsOut = append(argsOut, "NOINDEX")
	}
	if f.Sortable {
		argsOut = append(argsOut, "SORTABLE")
	}
	return argsOut
}

type NumericField struct {
	Name     string
	As       string `json:"as,omitempty"       yaml:"as,omitempty"`
	NoIndex  bool   `json:"no_index,omitempty" yaml:"no_index,omitempty"`
	Sortable bool   `json:"sortable,omitempty" yaml:"sortable,omitempty"`
}

func (f NumericField) AsCommand() []string {
	argsOut := []string{f.Name}
	if f.As != "" {
		argsOut = append(argsOut, "AS", f.As, "NUMERIC")
	} else {
		argsOut = append(argsOut, "NUMERIC")
	}
	if f.NoIndex {
		argsOut = append(argsOut, "NOINDEX")
	}
	if f.Sortable {
		argsOut = append(argsOut, "SORTABLE")
	}
	return argsOut
}

type VectorField struct {
	Name      string
	As        string          `json:"as,omitempty" yaml:"as,omitempty"`
	Algorithm VectorAlgorithm // default="FLAT"

	// mandatory attributes
	Dims           int            // int = Field(...)
	Datatype       VectorDataType // default="FLOAT32"
	DistanceMetric DistanceMetric `json:"distance_metric" yaml:"distance_metric"` // default="COSINE"

	// optional attributes
	InitialCap int `json:"initial_cap,omitempty" yaml:"initial_cap,omitempty"` // Optional[int] = None

	// only FLAT attributes
	BlockSize int `json:"block_size,omitempty" yaml:"block_size,omitempty"`

	// only HSNW attributes
	M              int     `json:"m,omitempty"               yaml:"m,omitempty"`               // m: int = Field(default=16)
	EfConstruction int     `json:"ef_construction,omitempty" yaml:"ef_construction,omitempty"` // ef_construction: int = Field(default=200)
	EfRuntime      int     `json:"ef_runtime,omitempty"      yaml:"ef_runtime,omitempty"`      // ef_runtime: int = Field(default=10)
	Epsilon        float32 `json:"epsilon,omitempty"         yaml:"epsilon,omitempty"`         // epsilon: float = Field(default=0.01)
}

// nolint: cyclop
func (f VectorField) AsCommand() []string {
	argsOut := []string{}

	if f.Datatype == FLOAT32VectorDataType || f.Datatype == FLOAT64VectorDataType {
		argsOut = append(argsOut, "TYPE", string(f.Datatype))
	} else {
		argsOut = append(argsOut, "TYPE", string(FLOAT32VectorDataType))
	}

	if f.Dims > 0 {
		argsOut = append(argsOut, "DIM", strconv.Itoa(f.Dims))
	} else {
		argsOut = append(argsOut, "DIM", "128")
	}

	if f.DistanceMetric == CosineDistanceMetric || f.DistanceMetric == L2DistanceMetric || f.DistanceMetric == IPDistanceMetric {
		argsOut = append(argsOut, "DISTANCE_METRIC", string(f.DistanceMetric))
	} else {
		argsOut = append(argsOut, "DISTANCE_METRIC", string(CosineDistanceMetric))
	}

	// vector has 3 mandatory attributes at least (TYPE, DIM, DISTANCE_METRIC)
	count := 3

	//nolint: nestif
	if f.Algorithm == HNSWVectorAlgorithm {
		if f.M > 0 {
			argsOut = append(argsOut, "M", strconv.Itoa(f.M))
			count++
		}
		if f.EfConstruction > 0 {
			argsOut = append(argsOut, "EF_CONSTRUCTION", strconv.Itoa(f.EfConstruction))
			count++
		}
		if f.EfRuntime > 0 {
			argsOut = append(argsOut, "EF_RUNTIME", strconv.Itoa(f.EfRuntime))
			count++
		}
		if f.Epsilon > 0 {
			argsOut = append(argsOut, "EPSILON", strconv.FormatFloat(float64(f.Epsilon), 'f', -1, 32))
			count++
		}
	} else {
		f.Algorithm = FlatVectorAlgorithm
		if f.BlockSize > 0 {
			argsOut = append(argsOut, "BLOCK_SIZE", strconv.Itoa(f.BlockSize))
			count++
		}
	}

	if f.As != "" {
		argsOut = append([]string{f.Name, "AS", f.As, "VECTOR", string(f.Algorithm), strconv.Itoa(count * 2)}, argsOut...)
	} else {
		argsOut = append([]string{f.Name, "VECTOR", string(f.Algorithm), strconv.Itoa(count * 2)}, argsOut...)
	}
	return argsOut
}

type IndexSchema struct {
	Tag     []TagField     `json:"tag"     yaml:"tag"`
	Text    []TextField    `json:"text"    yaml:"text"`
	Numeric []NumericField `json:"numeric" yaml:"numeric"`
	Vector  []VectorField  `json:"vector"  yaml:"vector"`
	// TODO GEO
}

// return names of field exclude vector key.
func (s *IndexSchema) MetadataKeys() map[string]any {
	keys := map[string]any{}
	for _, tag := range s.Tag {
		keys[tag.Name] = struct{}{}
	}
	for _, tag := range s.Text {
		keys[tag.Name] = struct{}{}
	}
	for _, tag := range s.Numeric {
		keys[tag.Name] = struct{}{}
	}
	for _, tag := range s.Vector {
		if tag.Name != defaultContentVectorFieldKey {
			keys[tag.Name] = struct{}{}
		}
	}
	return keys
}

func (s *IndexSchema) AsCommand() []string {
	argsOut := []string{}
	for _, tag := range s.Tag {
		argsOut = append(argsOut, tag.AsCommand()...)
	}
	for _, tag := range s.Text {
		argsOut = append(argsOut, tag.AsCommand()...)
	}
	for _, tag := range s.Numeric {
		argsOut = append(argsOut, tag.AsCommand()...)
	}
	for _, tag := range s.Vector {
		argsOut = append(argsOut, tag.AsCommand()...)
	}
	return argsOut
}

type schemaGenerator struct {
	format   SchemaFormat
	filePath string
	buf      []byte
}

// generate index schema with yaml config file or bytes.
func (s *schemaGenerator) generate() (*IndexSchema, error) {
	if s.filePath == "" && len(s.buf) == 0 {
		return nil, ErrEmptySchemaContent
	}

	if s.filePath != "" && len(s.buf) == 0 {
		file, err := os.Open(s.filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, file)
		if err != nil {
			return nil, err
		}
		s.buf = buf.Bytes()
	}

	schema := &IndexSchema{}
	switch s.format {
	case JSONSchemaFormat:
		if err := json.Unmarshal(s.buf, schema); err != nil {
			return nil, err
		}
	case YAMLSchemaFormat:
		if err := yaml.Unmarshal(s.buf, schema); err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidSchemaFormat
	}
	// check content & content_vector field exists
	// for _, text := range schema.Text {

	// }

	return schema, nil
}

// generate index schema with metadata & default values
// metadata has be appended with content & content_vector.
func generateSchemaWithMetadata(data map[string]any) (*IndexSchema, error) {
	defaultVectorField := VectorField{
		Name:           defaultContentVectorFieldKey,
		Algorithm:      FlatVectorAlgorithm,
		Dims:           1536,
		Datatype:       FLOAT32VectorDataType,
		DistanceMetric: CosineDistanceMetric,
	}
	schema := IndexSchema{}
	for key, value := range data {
		// nolint:nestif
		// content_vector
		if key == defaultContentVectorFieldKey {
			field := defaultVectorField
			if _value, ok := value.([]float32); ok {
				field.Dims = len(_value)
				schema.Vector = append(schema.Vector, field)
			} else if _value, ok := value.([]float64); ok {
				field.Dims = len(_value)
				schema.Vector = append(schema.Vector, field)
			} else {
				return nil, errors.New("the vector type is not []float32 or []float64")
			}
		} else {
			if value == nil {
				slog.Warn("Ignore nil value", "key", key)
				continue
			}
			//nolint: exhaustive
			switch reflect.TypeOf(value).Kind() {
			case reflect.String:
				field := TextField{Weight: 1, Name: key}
				schema.Text = append(schema.Text, field)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
				field := NumericField{Name: key}
				schema.Numeric = append(schema.Numeric, field)
			case reflect.Slice:
				field := TagField{Name: key, Separator: ","}
				schema.Tag = append(schema.Tag, field)
			default:
				slog.Warn("ignore invalid metadata value", "key", key, "value", value, "type", reflect.TypeOf(value).String())
			}
		}
	}

	return &schema, nil
}

type RedisIndex struct {
	name      string
	prefix    []string
	indexType IndexType
	schema    IndexSchema
}

func NewIndex(name string, prefix []string, indexType IndexType, schema IndexSchema) *RedisIndex {
	return &RedisIndex{
		name:      name,
		prefix:    prefix,
		indexType: indexType,
		schema:    schema,
	}
}

func (i *RedisIndex) AsCommand() ([]string, error) {
	cmd := []string{"FT.CREATE", i.name}
	if i.indexType != HASHIndexType && i.indexType != JSONIndexType {
		return nil, errors.New("invalid index type")
	}
	cmd = append(cmd, "ON", string(i.indexType))

	if len(i.prefix) > 0 {
		cmd = append(cmd, "PREFIX", strconv.Itoa(len(i.prefix)))
		cmd = append(cmd, i.prefix...)
	}
	cmd = append(cmd, "SCORE", "1.0", "SCHEMA")
	cmd = append(cmd, i.schema.AsCommand()...)
	return cmd, nil
}
