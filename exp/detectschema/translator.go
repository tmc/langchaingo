package detectschema

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

type Translator interface {
	TranslateAttributeInfo(attributeInfo []schema.AttributeInfo) (any, error)
	GetAttributeInfoByNamespace(ctx context.Context, namespace string) ([]schema.AttributeInfo, error)
}
