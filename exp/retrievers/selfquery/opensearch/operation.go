package selfquery_opensearch

import (
	"fmt"
)

func (sqt SelfQueryOpensearchTranslator) Operation(operator string, args []interface{}) (interface{}, error) {
	arguments := []interface{}{}
	for _, arg := range args {
		result, err := sqt.HandleFilter(arg)
		if err != nil {
			return nil, err
		}

		arguments = append(arguments, result)
	}
	fmt.Printf("operator: %v\n", operator)
	fmt.Printf("OperatorMap: %v\n", OperatorMap)

	return map[string]interface{}{
		"bool": map[string]interface{}{
			operator: arguments,
		},
	}, nil
}
