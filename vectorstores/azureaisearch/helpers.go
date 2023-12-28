package azureaisearch

import (
	"encoding/json"
	"fmt"
)

func StructToMap(input any, output *map[string]interface{}) error {
	inrec, err := json.Marshal(input)
	if err != nil {
		fmt.Printf("error marshalling StructToMap input : %v\n", err)
		return err
	}

	return json.Unmarshal(inrec, output)
}
