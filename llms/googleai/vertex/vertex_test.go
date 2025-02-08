package vertex

import (
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"testing"
)

/*
Description:
Author: xsl
Date: 2025/2/7

Modification History:
Date			Author		Description
-----------------------------------------
2025/2/7		    xsl			创建文件
*/

func TestConvertTools(t *testing.T) {
	availableTools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "BookCalendarSchedule",
				Description: "Useful when user want you to add event to calendar.\\nAdd event to USER'S CALENDAR by sending email.\\nOnly use this function when user EXPLICITLY ask you to add event to calendar.",
				Parameters: map[string]any{
					"type":     "object",
					"required": []string{"events"},
					"properties": map[string]any{
						"events": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type":     "object",
								"required": []string{"summary", "start", "end"},
								"properties": map[string]any{
									"end": map[string]any{
										"type":        "string",
										"description": "Event end time\\nFormat: 2023-10-03 23:10:03",
									},
									"start": map[string]any{
										"type":        "string",
										"description": "Event start time\\nFormat: 2023-10-03 23:10:03",
									},
									"summary": map[string]any{
										"type": "string",
									},
									"location": map[string]any{
										"type": "string",
									},
									"description": map[string]any{
										"type": "string",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	tools, err := convertTools(availableTools)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", tools)
}
