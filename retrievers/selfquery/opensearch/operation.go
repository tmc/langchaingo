package selfqueryopensearch

func (sqt Translator) operation(operator string, args []interface{}) (interface{}, error) {
	arguments := []interface{}{}
	for _, arg := range args {
		result, err := sqt.handleFilter(arg)
		if err != nil {
			return nil, err
		}

		arguments = append(arguments, result)
	}

	return map[string]interface{}{
		"bool": map[string]interface{}{
			operator: arguments,
		},
	}, nil
}
