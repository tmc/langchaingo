package aiplatformclient

import (
	"google.golang.org/protobuf/types/known/structpb"
)

func makeParams(params map[string]interface{}) (*structpb.Struct, error) {
	mergedParams := map[string]interface{}{}
	for paramKey, paramValue := range params {
		switch value := paramValue.(type) {
		case int, int32, int64, float32, float64:
			if value != 0 {
				mergedParams[paramKey] = value
			}
		default:
			mergedParams[paramKey] = value
		}
	}

	return structpb.NewStruct(mergedParams)
}

func convertArray(value []string) interface{} {
	newArray := make([]interface{}, len(value))
	for i, v := range value {
		newArray[i] = v
	}
	return newArray
}
