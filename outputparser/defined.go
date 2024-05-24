package outputparser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// Defined parses JSON output from an LLM into Go structs. A properly tagged
// struct can generate TypeScript interfaces to help LLMs generate responses
// that follow the desired JSON structure.
type Defined[T any] struct {
	schema string
}

func NewDefined[T any](source T) (Defined[T], error) {
	var empty Defined[T]

	sourceType := reflect.TypeOf(source)
	if k := sourceType.Kind(); k != reflect.Struct {
		return empty, fmt.Errorf("expected a struct; got %s", k)
	}
	numFields := sourceType.NumField()
	if numFields == 0 {
		return empty, errors.New("schema source has no fields")
	}

	numTaggedFields := 0
	for i := 0; i < numFields; i++ {
		if _, ok := sourceType.Field(i).Tag.Lookup("describe"); ok {
			numTaggedFields++
		}
	}
	if numTaggedFields == 0 {
		return empty, errors.New("requires at least 1 field with the struct tag 'describe'")
	}

	var result bytes.Buffer
	switch sourceType.Kind() {
	case reflect.Struct:
		data, err := marshalStruct(sourceType, "_Root")
		if err != nil {
			return empty, err
		}
		result.Write(data)
	default:
		return empty, fmt.Errorf("unable to marshal '%s' field type", sourceType.Kind())
	}
	return Defined[T]{result.String()}, nil
}

var _ schema.OutputParser[any] = Defined[any]{}

func (p Defined[T]) GetFormatInstructions() string {
	const instructions = "Your output should be in JSON, structured according to this TypeScript:\n```typescript\n%s\n```"
	return fmt.Sprintf(instructions, p.schema)
}

func (p Defined[T]) Parse(text string) (T, error) {
	var target T

	// Removes '```json' and '```' from the start and end of the text.
	const opening = "```json"
	const closing = "```"
	if text[:len(opening)] != opening || text[len(text)-len(closing):] != closing {
		return target, fmt.Errorf("input text should start with %s and end with %s", opening, closing)
	}
	parseableJSON := text[len(opening) : len(text)-len(closing)]
	if err := json.Unmarshal([]byte(parseableJSON), &target); err != nil {
		return target, fmt.Errorf("could not parse generated JSON: %v", err)
	}
	return target, nil
}

// ParseWithPrompt is equivalent to Parse
func (p Defined[T]) ParseWithPrompt(text string, _ llms.PromptValue) (T, error) {
	return p.Parse(text)
}

func (p Defined[T]) Type() string {
	return "defined_parser"
}

func marshalStruct(vType reflect.Type, name string) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("interface ")
	b.WriteString(name)
	b.WriteString(" {\n")
	moreStructs := make([][]byte, 0, 5)
	for i := 0; i < vType.NumField(); i++ {
		field := vType.Field(i)
		b.WriteString("\t")
		name := field.Tag.Get("json")
		if name == "" {
			name = field.Name
		}
		b.WriteString(name)
		b.WriteString(": ")
		typeName := field.Type.Name()
		if typeName == "" {
			typeName = field.Name
		}
		switch field.Type.Kind() {
		case reflect.Struct:
			marshaled, err := marshalStruct(field.Type, typeName)
			if err != nil {
				return []byte{}, err
			}
			moreStructs = append(moreStructs, marshaled)
			b.WriteString(typeName)
		case reflect.Array, reflect.Slice:
			elemType := field.Type.Elem()
			switch elemType.Kind() {
			case reflect.Struct:
				marshaled, err := marshalStruct(elemType, typeName)
				if err != nil {
					return []byte{}, err
				}
				moreStructs = append(moreStructs, marshaled)
				b.WriteString(typeName)
			default:
				b.WriteString(field.Type.Elem().Kind().String())
			}
			b.WriteString("[]")
		default:
			b.WriteString(typeName)
		}
		b.WriteString(";")
		if describe := field.Tag.Get("describe"); describe != "" {
			b.WriteString(" // ")
			b.WriteString(describe)
		}
		b.WriteString("\n")
	}
	b.WriteString("}")
	if more := bytes.Join(moreStructs, []byte("\n")); len(more) > 0 {
		b.WriteString("\n")
		b.Write(more)
	}
	return b.Bytes(), nil
}
