package selfquery_opensearch

import (
	"errors"
	"fmt"
	"slices"

	"github.com/tmc/langchaingo/exp/tools/queryconstructor"
)

func (sqt SelfQueryOpensearchTranslator) Comparison(comparator string, args []interface{}) (interface{}, error) {

	if len(args) != 2 {
		return nil, errors.New("there should be exactly 2 arguments for a comparison")
	}

	attribute, isFirstArgString := args[0].(string)
	if !isFirstArgString {
		return nil, errors.New("first argument of comparison should be a string")
	}

	value := args[1]

	field := fmt.Sprintf("metadata.%s", attribute)

	switch {
	case isRange(comparator):
		return map[string]interface{}{
			"range": map[string]interface{}{
				field: map[string]interface{}{
					comparator: value,
				},
			},
		}, nil
	case comparator == queryconstructor.ComparatorLIKE:
		return map[string]interface{}{
			comparator: map[string]interface{}{
				field: map[string]interface{}{
					"value": value,
				},
			},
		}, nil
	case isString(value):
		field = fmt.Sprintf("%s.keyword", attribute)
		fallthrough
	default:
		return map[string]interface{}{
			comparator: map[string]interface{}{
				field: value,
			},
		}, nil
	}

}

func isRange(comparator queryconstructor.Comparator) bool {
	return slices.Contains[[]queryconstructor.Comparator, queryconstructor.Comparator]([]queryconstructor.Comparator{
		queryconstructor.ComparatorLT,
		queryconstructor.ComparatorLTE,
		queryconstructor.ComparatorGT,
		queryconstructor.ComparatorGTE}, comparator)
}

func isString(value interface{}) bool {
	_, ok := value.(string)
	return ok
}
