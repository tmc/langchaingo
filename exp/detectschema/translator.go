package detectschema

import (
	"github.com/tmc/langchaingo/schema"
)

type Translator interface {
	TranslateAttributeInfo(attributeInfo []schema.AttributeInfo) (any, error)
}
