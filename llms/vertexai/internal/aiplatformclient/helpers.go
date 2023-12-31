package aiplatformclient

import (
	"github.com/tmc/langchaingo/llms/vertexai/internal/schema"
	"google.golang.org/protobuf/types/known/structpb"
)

func mergeParams(defaultParams, params map[string]interface{}) *structpb.Struct {
	mergedParams := cloneDefaultParameters()
	for paramKey, paramValue := range params {
		switch value := paramValue.(type) {
		case float64:
			if value != 0 {
				mergedParams[paramKey] = value
			}
		case int:
		case int32:
		case int64:
			if value != 0 {
				mergedParams[paramKey] = value
			}
		case []interface{}:
			mergedParams[paramKey] = value
		}
	}
	return convertToOutputStruct(defaultParams, mergedParams)
}

func convertToOutputStruct(defaultParams map[string]interface{}, mergedParams map[string]interface{}) *structpb.Struct {
	smergedParams, err := structpb.NewStruct(mergedParams)
	if err != nil {
		smergedParams, _ = structpb.NewStruct(defaultParams)
		return smergedParams
	}
	return smergedParams
}

func cloneDefaultParameters() map[string]interface{} {
	mergedParams := map[string]interface{}{}
	for paramKey, paramValue := range schema.DefaultParameters {
		mergedParams[paramKey] = paramValue
	}
	return mergedParams
}

func convertArray(value []string) interface{} {
	newArray := make([]interface{}, len(value))
	for i, v := range value {
		newArray[i] = v
	}
	return newArray
}
