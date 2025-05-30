package googleai

import (
	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
	"testing"
)

func Test_convertTools(t *testing.T) {
	tests := []struct {
		name        string
		tools       []llms.Tool
		expected    []*genai.FunctionDeclaration
		expectedErr string
	}{
		{
			name:  "no tools",
			tools: nil,
		},
		{
			name: "unsupported tool type",
			tools: []llms.Tool{
				{Type: "unsupported"},
			},
			expectedErr: `unsupported type "unsupported", want 'function'`,
		},
		{
			name: "unsupported tool parameter type",
			tools: []llms.Tool{
				{Type: "function", Function: &llms.FunctionDefinition{Parameters: "unsupported"}},
			},
			expectedErr: `unsupported type string of Parameters`,
		},
		{
			name: "missing type",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{},
				}},
			},
			expectedErr: "type is missing",
		},
		{
			name: "type is not string",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type": 123,
					},
				}},
			},
			expectedErr: "type is not a string",
		},
		{
			name: "description is not string",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type":        "object",
						"description": 123,
					},
				}},
			},
			expectedErr: "description is not a string",
		},
		{
			name: "enum is not a slice",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type": "object",
						"enum": 123,
					},
				}},
			},
			expectedErr: "enum is not a slice",
		},
		{
			name: "required is not a slice",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type":     "object",
						"required": 123,
					},
				}},
			},
			expectedErr: "required field is not a slice",
		},
		{
			name: "required items are not strings",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type":     "object",
						"required": []any{"string", 123},
					},
				}},
			},
			expectedErr: "expected string for required",
		},
		{
			name: "items is not a map",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type":  "object",
						"items": 123,
					},
				}},
			},
			expectedErr: "items is not a map",
		},
		{
			name: "properties is not a map",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type":       "object",
						"properties": 123,
					},
				}},
			},
			expectedErr: "properties is not a map",
		},
		{
			name: "use given schema",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters:  &genai.Schema{},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters:  &genai.Schema{},
			}},
		},
		{
			name: "simple parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type":        "string",
						"description": "A simple string parameter",
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type:        genai.TypeString,
					Description: "A simple string parameter",
				},
			}},
		},
		{
			name: "object parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"string": map[string]any{
								"type":        "string",
								"description": "A string parameter",
							},
							"number": map[string]any{
								"type":        "number",
								"description": "A number parameter",
							},
							"integer": map[string]any{
								"type":        "integer",
								"description": "An integer parameter",
							},
							"boolean": map[string]any{
								"type":        "boolean",
								"description": "A boolean parameter",
							},
						},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"string": {
							Type:        genai.TypeString,
							Description: "A string parameter",
						},
						"number": {
							Type:        genai.TypeNumber,
							Description: "A number parameter",
						},
						"integer": {
							Type:        genai.TypeInteger,
							Description: "An integer parameter",
						},
						"boolean": {
							Type:        genai.TypeBoolean,
							Description: "A boolean parameter",
						},
					},
				},
			}},
		},
		{
			name: "required parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"string": map[string]any{
								"type":        "string",
								"description": "A string parameter",
							},
							"number": map[string]any{
								"type":        "number",
								"description": "A number parameter",
							},
						},
						"required": []string{"string"},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"string": {
							Type:        genai.TypeString,
							Description: "A string parameter",
						},
						"number": {
							Type:        genai.TypeNumber,
							Description: "A number parameter",
						},
					},
					Required: []string{"string"},
				},
			}},
		},
		{
			name: "enum parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"enum": map[string]any{
								"type":        "string",
								"description": "A enum parameter",
								"enum":        []string{"option1", "option2", "option3"},
							},
							"anyEnum": map[string]any{
								"type":        "string",
								"description": "A any enum parameter",
								"enum":        []any{1, 1.2, "option3"},
							},
						},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"enum": {
							Type:        genai.TypeString,
							Description: "A enum parameter",
							Enum:        []string{"option1", "option2", "option3"},
						},
						"anyEnum": {
							Type:        genai.TypeString,
							Description: "A any enum parameter",
							Enum:        []string{"1", "1.2", "option3"},
						},
					},
				},
			}},
		},
		{
			name: "simple array parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"string": map[string]any{
								"type":        "array",
								"description": "A string array parameter",
								"items": map[string]any{
									"type": "string",
								},
							},
							"number": map[string]any{
								"type":        "array",
								"description": "A number array parameter",
								"items": map[string]any{
									"type": "number",
								},
							},
							"integer": map[string]any{
								"type":        "array",
								"description": "An integer array parameter",
								"items": map[string]any{
									"type": "integer",
								},
							},
							"boolean": map[string]any{
								"type":        "array",
								"description": "A boolean array parameter",
								"items": map[string]any{
									"type": "boolean",
								},
							},
						},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"string": {
							Type:        genai.TypeArray,
							Description: "A string array parameter",
							Items: &genai.Schema{
								Type: genai.TypeString,
							},
						},
						"number": {
							Type:        genai.TypeArray,
							Description: "A number array parameter",
							Items: &genai.Schema{
								Type: genai.TypeNumber,
							},
						},
						"integer": {
							Type:        genai.TypeArray,
							Description: "An integer array parameter",
							Items: &genai.Schema{
								Type: genai.TypeInteger,
							},
						},
						"boolean": {
							Type:        genai.TypeArray,
							Description: "A boolean array parameter",
							Items: &genai.Schema{
								Type: genai.TypeBoolean,
							},
						},
					},
				},
			}},
		},
		{
			name: "complex array parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"array": map[string]any{
								"type":        "array",
								"description": "A array parameter",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"string": map[string]any{
											"type":        "string",
											"description": "A string parameter",
										},
										"number": map[string]any{
											"type":        "number",
											"description": "A number parameter",
										},
									},
									"required": []string{"string"},
								},
							},
						},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"array": {
							Type:        genai.TypeArray,
							Description: "A array parameter",
							Items: &genai.Schema{
								Type: genai.TypeObject,
								Properties: map[string]*genai.Schema{
									"string": {
										Type:        genai.TypeString,
										Description: "A string parameter",
									},
									"number": {
										Type:        genai.TypeNumber,
										Description: "A number parameter",
									},
								},
								Required: []string{"string"},
							},
						},
					},
				},
			}},
		},
		{
			name: "complex object parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"param1": map[string]any{
								"type":        "object",
								"description": "A object parameter",
								"properties": map[string]any{
									"array": map[string]any{
										"type":        "array",
										"description": "A array parameter",
										"items": map[string]any{
											"type": "object",
											"properties": map[string]any{
												"string": map[string]any{
													"type":        "string",
													"description": "A string parameter",
												},
												"number": map[string]any{
													"type":        "number",
													"description": "A number parameter",
												},
											},
											"required": []string{"string"},
										},
									},
								},
							},
							"param2": map[string]any{
								"type":        "string",
								"description": "A string parameter",
								"enum":        []string{"option1", "option2", "option3"},
							},
						},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"param1": {
							Type:        genai.TypeObject,
							Description: "A object parameter",
							Properties: map[string]*genai.Schema{
								"array": {
									Type:        genai.TypeArray,
									Description: "A array parameter",
									Items: &genai.Schema{
										Type: genai.TypeObject,
										Properties: map[string]*genai.Schema{
											"string": {
												Type:        genai.TypeString,
												Description: "A string parameter",
											},
											"number": {
												Type:        genai.TypeNumber,
												Description: "A number parameter",
											},
										},
										Required: []string{"string"},
									},
								},
							},
						},
						"param2": {
							Type:        genai.TypeString,
							Description: "A string parameter",
							Enum:        []string{"option1", "option2", "option3"},
						},
					},
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := convertTools(tt.tools)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			}

			if tt.expected == nil {
				assert.Empty(t, result)
			} else {
				assert.Len(t, result, 1)
				assert.Equal(t, tt.expected, result[0].FunctionDeclarations)
			}
		})
	}
}
