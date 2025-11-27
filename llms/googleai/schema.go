package googleai

import (
	"github.com/google/generative-ai-go/genai"
)

// Schema type aliases for convenient structured output usage.
// These are re-exported from github.com/google/generative-ai-go/genai.
type (
	// Schema represents the structure of generated content.
	// Use this with llms.WithResponseSchema() for structured output.
	Schema = genai.Schema

	// Type represents the data type of a Schema.
	Type = genai.Type
)

// Type constants for Schema definition.
const (
	// TypeUnspecified means not specified, should not be used.
	TypeUnspecified = genai.TypeUnspecified
	// TypeString means string type.
	TypeString = genai.TypeString
	// TypeNumber means number type (float).
	TypeNumber = genai.TypeNumber
	// TypeInteger means integer type.
	TypeInteger = genai.TypeInteger
	// TypeBoolean means boolean type.
	TypeBoolean = genai.TypeBoolean
	// TypeArray means array type.
	TypeArray = genai.TypeArray
	// TypeObject means object/struct type.
	TypeObject = genai.TypeObject
)
