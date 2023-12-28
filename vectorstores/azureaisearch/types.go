package azureaisearch

import (
	"fmt"
)

type FieldType = string

const (
	FieldTypeString         FieldType = "Edm.String"
	FieldTypeSingle         FieldType = "Edm.Single"
	FieldTypeInt32          FieldType = "Edm.Int32"
	FieldTypeInt64          FieldType = "Edm.Int64"
	FieldTypeDouble         FieldType = "Edm.Double"
	FieldTypeBoolean        FieldType = "Edm.Boolean"
	FieldTypeDatetimeOffset FieldType = "Edm.DateTimeOffset"
	FieldTypeComplexType    FieldType = "Edm.ComplexType"
)

func CollectionField(fieldType FieldType) FieldType {
	return fmt.Sprintf("Collection(%s)", fieldType)
}

type SimilaritySearchFilters struct {
	K int
}
