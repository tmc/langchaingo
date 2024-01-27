package detectschemaopensearch

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/exp/detectschema"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/opensearch"
)

type Translator struct {
	vectorstore opensearch.Store
}

var _ detectschema.Translator = Translator{}

func New(vectorstore opensearch.Store) Translator {
	return Translator{
		vectorstore: vectorstore,
	}
}

func (t Translator) TranslateAttributeInfo(attributeInfo []schema.AttributeInfo) (any, error) {
	output := map[string]interface{}{}
	fmt.Printf("attributeInfo: %v\n", attributeInfo)
	for _, attribute := range attributeInfo {
		output[attribute.Name] = attribute.Type
		switch attribute.Type {
		case detectschema.AllowedTypeString:
			output[attribute.Name] = map[string]interface{}{
				"type": "keyword",
			}
		case detectschema.AllowedTypeBool:
			output[attribute.Name] = map[string]interface{}{
				"type": "boolean",
			}
		case detectschema.AllowedTypeFloat:
			output[attribute.Name] = map[string]interface{}{
				"type": "float",
			}
		case detectschema.AllowedTypeInt:
			output[attribute.Name] = map[string]interface{}{
				"type": "integer",
			}
		default:
			return nil, fmt.Errorf("unknown type: %s", attribute.Type)
		}
	}
	return output, nil
}

func (t Translator) GetAttributeInfoByNamespace(ctx context.Context, namespace string) ([]schema.AttributeInfo, error) {
	output := map[string]interface{}{}
	if err := t.vectorstore.GetIndex(ctx, namespace, &output); err != nil {
		return nil, err
	}
	fmt.Printf("output: %+v\n", output)

	attributeInfo := []schema.AttributeInfo{}

	return attributeInfo, nil
}
