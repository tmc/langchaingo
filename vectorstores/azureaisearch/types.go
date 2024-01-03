package azureaisearch

import (
	"fmt"
)

// FieldType type for pseudo enum.
type FieldType = string

// Pseudo enum for all the different FieldType.
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

// CollectionField allows to define a fieldtype as a collection.
func CollectionField(fieldType FieldType) FieldType {
	return fmt.Sprintf("Collection(%s)", fieldType)
}
