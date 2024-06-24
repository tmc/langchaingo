package bedrockclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/mitchellh/mapstructure"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/tmc/langchaingo/internal/util"
	"github.com/tmc/langchaingo/llms"
)

func getBedrockSystemContentBlock(parts []llms.ContentPart) ([]types.SystemContentBlock, error) {
	systemContent := make([]types.SystemContentBlock, 0)
	for _, part := range parts {
		switch p := part.(type) {
		case llms.TextContent:
			systemContent = append(systemContent, &types.SystemContentBlockMemberText{Value: p.Text})
		default:
			return nil, errors.New("system content part must be text")
		}
	}
	return systemContent, nil
}

func getConverseMessageImageType(typ string) (types.ImageFormat, error) {
	switch typ {
	case "png":
		return types.ImageFormatPng, nil
	case "jpeg":
		return types.ImageFormatJpeg, nil
	case "gif":
		return types.ImageFormatGif, nil
	case "webp":
		return types.ImageFormatWebp, nil
	default:
		return "", errors.New("unsupported image type")
	}
}

func getBedrockContentBlock(parts []llms.ContentPart) ([]types.ContentBlock, error) {
	convertedParts := make([]types.ContentBlock, 0, len(parts))
	for _, part := range parts {
		var out types.ContentBlock
		switch p := part.(type) {
		case llms.TextContent:
			out = &types.ContentBlockMemberText{Value: p.Text}
		case llms.BinaryContent:
			return nil, errors.New("binary content not supported")
		case llms.ImageURLContent:
			typ, data, err := util.DownloadImageData(p.URL)
			if err != nil {
				return nil, err
			}
			t, err := getConverseMessageImageType(typ)
			if err != nil {
				return nil, err
			}
			out = &types.ContentBlockMemberImage{
				Value: types.ImageBlock{
					Format: t,
					Source: &types.ImageSourceMemberBytes{
						Value: data,
					},
				},
			}
		case llms.ToolCall:
			input := make(map[string]any)
			err := json.Unmarshal([]byte(p.FunctionCall.Arguments), &input)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool call input: %w", err)
			}
			out = &types.ContentBlockMemberToolUse{
				Value: types.ToolUseBlock{
					Input:     document.NewLazyDocument(input),
					Name:      &p.FunctionCall.Name,
					ToolUseId: &p.ID,
				},
			}
		case llms.ToolCallResponse:
			// TODO implement file response
			out = &types.ContentBlockMemberToolResult{
				Value: types.ToolResultBlock{
					Content:   []types.ToolResultContentBlock{&types.ToolResultContentBlockMemberText{Value: p.Content}},
					ToolUseId: &p.ToolCallID,
					Status:    types.ToolResultStatusSuccess,
				},
			}
		default:
			return nil, errors.New("unsupported content part")
		}

		convertedParts = append(convertedParts, out)
	}
	return convertedParts, nil
}

// process the input messages to anthropic supported input
// returns the input content and system prompt.
func processInputMessagesBedrock(messages []llms.MessageContent) ([]types.SystemContentBlock, []types.Message, error) {
	// merge messages with the same role
	chunkedMessages := make([][]llms.MessageContent, 0, len(messages))
	currentChunk := make([]llms.MessageContent, 0, len(messages))
	var lastRole string
	for _, message := range messages {
		role, err := getAnthropicRole(message.Role)
		if err != nil {
			return nil, nil, err
		}
		if role != lastRole {
			if len(currentChunk) > 0 {
				chunkedMessages = append(chunkedMessages, currentChunk)
			}
			currentChunk = make([]llms.MessageContent, 0)
		}
		currentChunk = append(currentChunk, message)
		lastRole = role
	}

	if len(currentChunk) > 0 {
		chunkedMessages = append(chunkedMessages, currentChunk)
	}

	inputContents := make([]types.Message, 0, len(messages))
	systemContents := make([]types.SystemContentBlock, 0)
	for _, chunk := range chunkedMessages {
		role, err := getAnthropicRole(chunk[0].Role)
		if err != nil {
			return nil, nil, err
		}
		if role == AnthropicSystem {
			for _, message := range chunk {
				content, err := getBedrockSystemContentBlock(message.Parts)
				if err != nil {
					return nil, nil, err
				}
				systemContents = append(systemContents, content...)
			}
			continue
		}

		var messageRole types.ConversationRole
		switch role {
		case AnthropicRoleAssistant:
			messageRole = types.ConversationRoleAssistant
		case AnthropicRoleUser:
			messageRole = types.ConversationRoleUser
		default:
			return nil, nil, errors.New("role not supported")
		}

		contentBlocks := make([]types.ContentBlock, 0, len(chunk))
		for _, message := range chunk {
			blocks, err := getBedrockContentBlock(message.Parts)
			if err != nil {
				return nil, nil, err
			}
			contentBlocks = append(contentBlocks, blocks...)
		}
		inputContents = append(inputContents, types.Message{
			Content: contentBlocks,
			Role:    messageRole,
		})
	}
	return systemContents, inputContents, nil
}

func updateToolUse(
	tools []llms.ToolCall,
	delta *types.ToolUseBlockDelta,
	start *types.ContentBlockStartMemberToolUse,
) ([]byte, []llms.ToolCall, error) {
	// TODO handle bytes
	if len(tools) == 0 {
		tools = append(tools, llms.ToolCall{})
	}
	if start != nil {
		if tools[len(tools)-1].ID != "" && tools[len(tools)-1].ID != *start.Value.ToolUseId {
			tools = append(tools, llms.ToolCall{
				ID: *start.Value.ToolUseId,
				FunctionCall: &llms.FunctionCall{
					Name: *start.Value.Name,
				},
			})
		} else {
			tools[len(tools)-1].ID = *start.Value.ToolUseId
			tools[len(tools)-1].FunctionCall = &llms.FunctionCall{
				Name: *start.Value.Name,
			}
		}
	}
	if delta != nil && len(*delta.Input) > 0 {
		tools[len(tools)-1].FunctionCall.Arguments += *delta.Input
	}

	return nil, tools, nil
}

// handleConverseStreamEvents handles the stream events and returns the content response.
// TODO: support multiple content choices.
func handleConverseStreamEvents(
	ctx context.Context,
	streamOutput *bedrockruntime.ConverseStreamOutput,
	options *llms.CallOptions,
) (*llms.ContentResponse, error) {
	stream := streamOutput.GetStream()
	if stream == nil {
		return nil, errors.New("no stream")
	}
	defer func() {
		if err := stream.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close stream", "err", err)
		}
	}()

	contentChoices := []*llms.ContentChoice{{GenerationInfo: map[string]interface{}{}}}
	tools := make([]llms.ToolCall, 0)
	var err error
	for e := range stream.Events() {
		if err = stream.Err(); err != nil {
			return nil, err
		}

		switch event := e.(type) {
		case *types.ConverseStreamOutputMemberContentBlockStart:
			switch start := event.Value.Start.(type) {
			case *types.ContentBlockStartMemberToolUse:
				// TODO handle bytes stream
				_, tools, err = updateToolUse(tools, nil, start)
			default:
				slog.WarnContext(ctx, "content block start not supported", "value", event.Value)
			}
		case *types.ConverseStreamOutputMemberContentBlockStop:
			slog.DebugContext(ctx, "content block stop not supported", "value", event.Value)
		case *types.ConverseStreamOutputMemberMessageStart:
			slog.DebugContext(ctx, "message start not supported", "value", event.Value)
		case *types.ConverseStreamOutputMemberMessageStop:
			contentChoices[0].StopReason = string(event.Value.StopReason)
		case *types.ConverseStreamOutputMemberContentBlockDelta:
			switch delta := event.Value.Delta.(type) {
			case *types.ContentBlockDeltaMemberText:
				if err = options.StreamingFunc(ctx, []byte(delta.Value)); err != nil {
					return nil, err
				}
				contentChoices[0].Content += delta.Value
			case *types.ContentBlockDeltaMemberToolUse:
				var err error
				_, tools, err = updateToolUse(tools, &delta.Value, nil)
				if err != nil {
					return nil, err
				}
			default:
				slog.WarnContext(ctx, "tool use not supported", "value", delta)
			}
		case *types.ConverseStreamOutputMemberMetadata:
			contentChoices[0].GenerationInfo["input_tokens"] = event.Value.Usage.InputTokens
			contentChoices[0].GenerationInfo["output_tokens"] = event.Value.Usage.OutputTokens
			contentChoices[0].GenerationInfo["total_tokens"] = event.Value.Usage.TotalTokens
			contentChoices[0].GenerationInfo["latency_ms"] = event.Value.Metrics.LatencyMs
		}
	}
	contentChoices[0].ToolCalls = tools
	return &llms.ContentResponse{Choices: contentChoices}, nil
}

func convertTools(tools []llms.Tool) []types.Tool {
	convertedTools := make([]types.Tool, 0, len(tools))
	for _, tool := range tools {
		convertedTools = append(convertedTools, &types.ToolMemberToolSpec{
			Value: types.ToolSpecification{
				InputSchema: &types.ToolInputSchemaMemberJson{
					Value: document.NewLazyDocument(tool.Function.Parameters),
				},
				Name:        &tool.Function.Name,
				Description: &tool.Function.Description,
			},
		})
	}
	return convertedTools
}

func convertToolChoice(toolChoice any) (types.ToolChoice, error) {
	if toolChoice == nil {
		return &types.ToolChoiceMemberAuto{}, nil
	}
	if toolChoice, ok := toolChoice.(string); ok {
		switch toolChoice {
		case "none":
			return nil, errors.New("tool choice none not supported")
		case "any":
			return &types.ToolChoiceMemberAny{}, nil
		case "auto":
			return &types.ToolChoiceMemberAuto{}, nil
		default:
			return nil, errors.New("unsupported tool choice")
		}
	}
	if toolChoice, ok := toolChoice.(llms.ToolChoice); ok {
		if toolChoice.Function == nil {
			return &types.ToolChoiceMemberAuto{}, nil
		}
		return &types.ToolChoiceMemberTool{
			Value: types.SpecificToolChoice{
				Name: &toolChoice.Function.Name,
			},
		}, nil
	}

	var llmsToolChoice llms.ToolChoice
	if err := mapstructure.Decode(toolChoice, &toolChoice); err != nil {
		return nil, fmt.Errorf("failed to decode tool choice: %w", err)
	}
	return &types.ToolChoiceMemberTool{
		Value: types.SpecificToolChoice{
			Name: &llmsToolChoice.Function.Name,
		},
	}, nil
}
