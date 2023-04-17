package prompts

import "fmt"

type interpolator func(template string, values map[string]any) (string, error)

var defaultFormatterMapping = map[string]interpolator{
	"f-string": interpolateFString,
}

func getDefaultInterpolator(templateFormat string) (interpolator, error) {
	formatter, ok := defaultFormatterMapping[templateFormat]
	if !ok {
		validFormats := ""
		for format := range defaultFormatterMapping {
			validFormats += format + " "
		}

		return formatter, fmt.Errorf("Invalid template format. Got %s; should be one of %s", templateFormat, validFormats)
	}

	return formatter, nil
}

func renderTemplate(template string, templateFormat string, inputValues map[string]any) (string, error) {
	formatter, err := getDefaultInterpolator(templateFormat)
	if err != nil {
		return "", err
	}

	return formatter(template, inputValues)
}

type parsedFStringNode interface{}
type fStringLiteral struct{ text string }
type fStringVariable struct{ name string }

func paresFString(_template string) ([]parsedFStringNode, error) {
	template := []rune(_template)
	nodes := make([]parsedFStringNode, 0)

	for i := 0; i < len(template); {

		if template[i] == '{' && i+1 < len(template) && template[i+1] == '{' {
			nodes = append(nodes, fStringLiteral{text: "{"})
			i += 2
			continue
		}

		if template[i] == '}' && i+1 < len(template) && template[i+1] == '}' {
			nodes = append(nodes, fStringLiteral{text: "}"})
			i += 2
			continue
		}

		if template[i] == '{' {
			next := getNextBracket(template, []rune{'}'}, i)
			if next < 0 {
				return nodes, fmt.Errorf("Unclosed '{' in template.")
			}

			nodes = append(nodes, fStringVariable{name: string(template[i+1 : next])})

			i = next + 1

			continue
		}

		if template[i] == '}' {
			return nodes, fmt.Errorf("Single '}' in template.")
		}

		next := getNextBracket(template, []rune{'{', '}'}, i)
		if next < 0 { // If no more brackets in template
			nodes = append(nodes, fStringLiteral{text: string(template[i:])})
			break
		}

		nodes = append(nodes, fStringLiteral{text: string(template[i:next])})
		i = next
	}

	return nodes, nil
}

func getNextBracket(template []rune, bracketTypes []rune, start int) int {
	for i := start; i < len(template); i++ {
		for j := 0; j < len(bracketTypes); j++ {
			if template[i] == bracketTypes[j] {
				return i
			}
		}
	}

	return -1
}

func interpolateFString(template string, values map[string]any) (string, error) {
	parsed, err := paresFString(template)
	if err != nil {
		return "", err
	}

	output := ""

	for _, node := range parsed {
		switch n := node.(type) {
		case fStringVariable:
			variableValue, ok := values[n.name]

			if !ok {
				return "", fmt.Errorf("Missing value for input %s", n.name)
			}

			output += fmt.Sprintf("%v", variableValue)
		case fStringLiteral:
			output += n.text

		}
	}

	return output, nil
}

func checkValidTemplate(template string, templateFormat string, inputVariables []string) error {
	dummyValues := make(map[string]any, 0)
	for i := 0; i < len(inputVariables); i++ {
		dummyValues[inputVariables[i]] = "foo"
	}

	_, err := renderTemplate(template, templateFormat, dummyValues)
	if err != nil {
		return fmt.Errorf("Invalid prompt schema. %e", err)
	}

	return nil
}
