package serpapi

import (
	"strings"

	"github.com/tmc/langchaingo/exp/tools"
	"github.com/tmc/langchaingo/exp/tools/serpapi/internal"
)

//Create a new tool for serpapi to search on internet
func New() (*tools.Tool, error) {
	client, err := internal.New()
	if err != nil {
		return nil, err
	}
	return &tools.Tool{
		Name: "Google Search",
		Description: `"A wrapper around Google Search. "
        "Useful for when you need to answer questions about current events. "
		"Always one of the first options when you need to find information on internet"
        "Input should be a search query."`,
		Run: func(input string) (string, error) {
			result, err := client.Search(input)
			if err != nil {
				return "", err
			}
			if len(result) == 0 {
				return "No good Google Search Result was found", nil
			}
			return strings.Join(strings.Fields(result), " "), nil
		},
	}, nil

}
