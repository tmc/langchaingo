package azureaisearch

import (
	"encoding/json"
	"fmt"
)

func structToMap(input any, output *map[string]interface{}) error {
	inrec, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("error marshalling StructToMap input : %w", err)
	}

	return json.Unmarshal(inrec, output)
}
