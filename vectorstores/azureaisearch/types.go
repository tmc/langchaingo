package azureaisearch

import (
	"fmt"
)

type FieldType = string

var (
	FieldType_string         FieldType = "Edm.String"
	FieldType_single         FieldType = "Edm.Single"
	FieldType_int32          FieldType = "Edm.Int32"
	FieldType_int64          FieldType = "Edm.Int64"
	FieldType_double         FieldType = "Edm.Double"
	FieldType_boolean        FieldType = "Edm.Boolean"
	FieldType_datetimeOffset FieldType = "Edm.DateTimeOffset"
	FieldType_complexType    FieldType = "Edm.ComplexType"
)

func CollectionField(fieldType FieldType) FieldType {
	return fmt.Sprintf("Collection(%s)", fieldType)
}

type SimilaritySearchFilters struct {
	K int
}
