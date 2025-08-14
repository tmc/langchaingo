package zep

import (
	"context"
	"testing"

	"github.com/0xDezzy/langchaingo/llms"
	"github.com/getzep/zep-go/v2"
	zepClient "github.com/getzep/zep-go/v2/client"
)

// MockZepClient implements a simple mock for testing
type MockZepClient struct {
	memory *zep.Memory
}

func (m *MockZepClient) Memory() MemoryService {
	return &MockMemoryService{memory: m.memory}
}

type MemoryService interface {
	Get(ctx context.Context, sessionID string, request *zep.MemoryGetRequest) (*zep.Memory, error)
	Add(ctx context.Context, sessionID string, request *zep.AddMemoryRequest) (*zep.Memory, error)
	Delete(ctx context.Context, sessionID string) (*zep.Memory, error)
}

type MockMemoryService struct {
	memory *zep.Memory
}

func (m *MockMemoryService) Get(_ context.Context, _ string, _ *zep.MemoryGetRequest) (*zep.Memory, error) {
	if m.memory == nil {
		return &zep.Memory{Messages: []*zep.Message{}}, nil
	}
	return m.memory, nil
}

func (m *MockMemoryService) Add(_ context.Context, _ string, request *zep.AddMemoryRequest) (*zep.Memory, error) {
	if m.memory == nil {
		m.memory = &zep.Memory{Messages: []*zep.Message{}}
	}
	m.memory.Messages = append(m.memory.Messages, request.Messages...)
	return m.memory, nil
}

func (m *MockMemoryService) Delete(_ context.Context, _ string) (*zep.Memory, error) {
	m.memory = &zep.Memory{Messages: []*zep.Message{}}
	return m.memory, nil
}

func createMockZepClient() *zepClient.Client {
	return &zepClient.Client{}
}

func TestNewMemory(t *testing.T) {
	t.Parallel()

	client := createMockZepClient()
	sessionID := "test-session"

	m := NewMemory(client, sessionID)

	if m.ZepClient != client {
		t.Errorf("Expected ZepClient to be set")
	}
	if m.SessionID != sessionID {
		t.Errorf("Expected SessionID to be %s, got %s", sessionID, m.SessionID)
	}
	if m.MemoryKey != "history" {
		t.Errorf("Expected default MemoryKey to be 'history', got %s", m.MemoryKey)
	}
	if m.HumanPrefix != "Human" {
		t.Errorf("Expected default HumanPrefix to be 'Human', got %s", m.HumanPrefix)
	}
	if m.AIPrefix != "AI" {
		t.Errorf("Expected default AIPrefix to be 'AI', got %s", m.AIPrefix)
	}
	if !m.ReturnMessages {
		t.Errorf("Expected default ReturnMessages to be true")
	}
}

func TestNewMemoryWithOptions(t *testing.T) {
	t.Parallel()

	client := createMockZepClient()
	sessionID := "test-session"

	m := NewMemory(
		client,
		sessionID,
		WithReturnMessages(false),
		WithMemoryKey("custom_history"),
		WithHumanPrefix("User"),
		WithAIPrefix("Assistant"),
		WithInputKey("input"),
		WithOutputKey("output"),
		WithMemoryType(zep.MemoryTypeSummaryRetriever),
	)

	if m.ReturnMessages {
		t.Errorf("Expected ReturnMessages to be false")
	}
	if m.MemoryKey != "custom_history" {
		t.Errorf("Expected MemoryKey to be 'custom_history', got %s", m.MemoryKey)
	}
	if m.HumanPrefix != "User" {
		t.Errorf("Expected HumanPrefix to be 'User', got %s", m.HumanPrefix)
	}
	if m.AIPrefix != "Assistant" {
		t.Errorf("Expected AIPrefix to be 'Assistant', got %s", m.AIPrefix)
	}
	if m.InputKey != "input" {
		t.Errorf("Expected InputKey to be 'input', got %s", m.InputKey)
	}
	if m.OutputKey != "output" {
		t.Errorf("Expected OutputKey to be 'output', got %s", m.OutputKey)
	}
	if m.MemoryType != zep.MemoryTypeSummaryRetriever {
		t.Errorf("Expected MemoryType to be SummaryRetriever")
	}
}

func TestMemoryVariables(t *testing.T) {
	t.Parallel()

	client := createMockZepClient()
	m := NewMemory(client, "test-session", WithMemoryKey("custom_key"))

	ctx := context.Background()
	variables := m.MemoryVariables(ctx)

	expected := []string{"custom_key"}
	if len(variables) != 1 || variables[0] != "custom_key" {
		t.Errorf("Expected variables %v, got %v", expected, variables)
	}
}

func TestGetMemoryKey(t *testing.T) {
	t.Parallel()

	client := createMockZepClient()
	m := NewMemory(client, "test-session", WithMemoryKey("test_key"))

	ctx := context.Background()
	key := m.GetMemoryKey(ctx)

	if key != "test_key" {
		t.Errorf("Expected memory key 'test_key', got %s", key)
	}
}

func TestNewZepChatMessageHistory(t *testing.T) {
	t.Parallel()

	client := createMockZepClient()
	sessionID := "test-session"

	h := NewZepChatMessageHistory(client, sessionID)

	if h.ZepClient != client {
		t.Errorf("Expected ZepClient to be set")
	}
	if h.SessionID != sessionID {
		t.Errorf("Expected SessionID to be %s, got %s", sessionID, h.SessionID)
	}
	if h.MemoryType != zep.MemoryTypePerpetual {
		t.Errorf("Expected default MemoryType to be Perpetual")
	}
}

func TestNewZepChatMessageHistoryWithOptions(t *testing.T) {
	t.Parallel()

	client := createMockZepClient()
	sessionID := "test-session"

	h := NewZepChatMessageHistory(
		client,
		sessionID,
		WithChatHistoryMemoryType(zep.MemoryTypeSummaryRetriever),
		WithChatHistoryHumanPrefix("User"),
		WithChatHistoryAIPrefix("Bot"),
	)

	if h.MemoryType != zep.MemoryTypeSummaryRetriever {
		t.Errorf("Expected MemoryType to be SummaryRetriever")
	}
	if h.HumanPrefix != "User" {
		t.Errorf("Expected HumanPrefix to be 'User', got %s", h.HumanPrefix)
	}
	if h.AIPrefix != "Bot" {
		t.Errorf("Expected AIPrefix to be 'Bot', got %s", h.AIPrefix)
	}
}

func TestMessagesFromZepMessages(t *testing.T) {
	t.Parallel()

	h := &ChatMessageHistory{}

	zepMessages := []*zep.Message{
		{
			Content:  "Hello",
			RoleType: zep.RoleTypeUserRole,
		},
		{
			Content:  "Hi there",
			RoleType: zep.RoleTypeAssistantRole,
		},
		{
			Content:  "Function result",
			RoleType: zep.RoleTypeFunctionRole,
		},
	}

	chatMessages := h.messagesFromZepMessages(zepMessages)

	if len(chatMessages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(chatMessages))
	}

	if human, ok := chatMessages[0].(llms.HumanChatMessage); !ok || human.Content != "Hello" {
		t.Errorf("Expected first message to be HumanChatMessage with content 'Hello'")
	}

	if ai, ok := chatMessages[1].(llms.AIChatMessage); !ok || ai.Content != "Hi there" {
		t.Errorf("Expected second message to be AIChatMessage with content 'Hi there'")
	}

	if tool, ok := chatMessages[2].(llms.ToolChatMessage); !ok || tool.Content != "Function result" {
		t.Errorf("Expected third message to be ToolChatMessage with content 'Function result'")
	}
}

func TestMessagesToZepMessages(t *testing.T) {
	t.Parallel()

	h := &ChatMessageHistory{
		HumanPrefix: "User",
		AIPrefix:    "Bot",
	}

	chatMessages := []llms.ChatMessage{
		llms.HumanChatMessage{Content: "Hello"},
		llms.AIChatMessage{Content: "Hi there"},
		llms.FunctionChatMessage{Content: "Function result"},
	}

	zepMessages := h.messagesToZepMessages(chatMessages)

	if len(zepMessages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(zepMessages))
	}

	if zepMessages[0].Content != "Hello" || zepMessages[0].RoleType != zep.RoleTypeUserRole {
		t.Errorf("Expected first message to be user role with content 'Hello'")
	}
	if *zepMessages[0].Role != "User" {
		t.Errorf("Expected first message role to be 'User', got %s", *zepMessages[0].Role)
	}

	if zepMessages[1].Content != "Hi there" || zepMessages[1].RoleType != zep.RoleTypeAssistantRole {
		t.Errorf("Expected second message to be assistant role with content 'Hi there'")
	}
	if *zepMessages[1].Role != "Bot" {
		t.Errorf("Expected second message role to be 'Bot', got %s", *zepMessages[1].Role)
	}

	if zepMessages[2].Content != "Function result" || zepMessages[2].RoleType != zep.RoleTypeFunctionRole {
		t.Errorf("Expected third message to be function role with content 'Function result'")
	}
}

// TestLoadMemoryVariablesReturnMessages tests LoadMemoryVariables with return messages = true
func TestLoadMemoryVariablesReturnMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := createMockZepClient()
	m := NewMemory(client, "test-session", WithReturnMessages(true))

	// Mock chat history with messages
	mockHistory := &mockChatHistory{
		messages: []llms.ChatMessage{
			llms.HumanChatMessage{Content: "Hello"},
			llms.AIChatMessage{Content: "Hi there"},
		},
	}
	m.ChatHistory = mockHistory

	result, err := m.LoadMemoryVariables(ctx, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	messages, ok := result["history"].([]llms.ChatMessage)
	if !ok {
		t.Fatalf("Expected result to contain messages slice")
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
}

// TestLoadMemoryVariablesBufferString tests LoadMemoryVariables with return messages = false
func TestLoadMemoryVariablesBufferString(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := createMockZepClient()
	m := NewMemory(client, "test-session", WithReturnMessages(false))

	mockHistory := &mockChatHistory{
		messages: []llms.ChatMessage{
			llms.HumanChatMessage{Content: "Hello"},
			llms.AIChatMessage{Content: "Hi there"},
		},
	}
	m.ChatHistory = mockHistory

	result, err := m.LoadMemoryVariables(ctx, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	bufferStr, ok := result["history"].(string)
	if !ok {
		t.Fatalf("Expected result to contain buffer string")
	}

	expected := "Human: Hello\nAI: Hi there"
	if bufferStr != expected {
		t.Errorf("Expected buffer string %q, got %q", expected, bufferStr)
	}
}

// TestLoadMemoryVariablesCustomKey tests LoadMemoryVariables with custom memory key
func TestLoadMemoryVariablesCustomKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := createMockZepClient()
	m := NewMemory(client, "test-session", WithMemoryKey("custom_history"))

	mockHistory := &mockChatHistory{
		messages: []llms.ChatMessage{},
	}
	m.ChatHistory = mockHistory

	result, err := m.LoadMemoryVariables(ctx, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, exists := result["custom_history"]; !exists {
		t.Errorf("Expected result to have key 'custom_history'")
	}
}

// TestLoadMemoryVariablesError tests LoadMemoryVariables error handling
func TestLoadMemoryVariablesError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := createMockZepClient()
	m := NewMemory(client, "test-session")

	mockHistory := &mockChatHistory{
		messagesErr: context.Canceled,
	}
	m.ChatHistory = mockHistory

	_, err := m.LoadMemoryVariables(ctx, nil)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

// TestSaveContext tests the SaveContext method
func TestSaveContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("BasicSave", func(t *testing.T) {
		client := createMockZepClient()
		m := NewMemory(client, "test-session")

		mockHistory := &mockChatHistory{}
		m.ChatHistory = mockHistory

		err := m.SaveContext(ctx,
			map[string]any{"input": "Hello"},
			map[string]any{"output": "Hi there"},
		)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if mockHistory.userMessageAdded != "Hello" {
			t.Errorf("Expected user message 'Hello', got %q", mockHistory.userMessageAdded)
		}
		if mockHistory.aiMessageAdded != "Hi there" {
			t.Errorf("Expected AI message 'Hi there', got %q", mockHistory.aiMessageAdded)
		}
	})

	t.Run("WithInputOutputKeys", func(t *testing.T) {
		client := createMockZepClient()
		m := NewMemory(client, "test-session",
			WithInputKey("user_input"),
			WithOutputKey("ai_output"),
		)

		mockHistory := &mockChatHistory{}
		m.ChatHistory = mockHistory

		err := m.SaveContext(ctx,
			map[string]any{"user_input": "Question", "other": "ignored"},
			map[string]any{"ai_output": "Answer", "other": "ignored"},
		)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if mockHistory.userMessageAdded != "Question" {
			t.Errorf("Expected user message 'Question', got %q", mockHistory.userMessageAdded)
		}
		if mockHistory.aiMessageAdded != "Answer" {
			t.Errorf("Expected AI message 'Answer', got %q", mockHistory.aiMessageAdded)
		}
	})

	t.Run("MissingInputKey", func(t *testing.T) {
		client := createMockZepClient()
		m := NewMemory(client, "test-session", WithInputKey("missing_key"))

		mockHistory := &mockChatHistory{}
		m.ChatHistory = mockHistory

		err := m.SaveContext(ctx,
			map[string]any{"input": "Hello"},
			map[string]any{"output": "Hi"},
		)
		if err == nil {
			t.Errorf("Expected error for missing input key")
		}
	})

	t.Run("ErrorFromAddUserMessage", func(t *testing.T) {
		client := createMockZepClient()
		m := NewMemory(client, "test-session")

		mockHistory := &mockChatHistory{
			addUserErr: context.DeadlineExceeded,
		}
		m.ChatHistory = mockHistory

		err := m.SaveContext(ctx,
			map[string]any{"input": "Hello"},
			map[string]any{"output": "Hi"},
		)
		if err == nil {
			t.Errorf("Expected error from AddUserMessage")
		}
	})
}

// TestClear tests the Clear method
func TestClear(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := createMockZepClient()
	m := NewMemory(client, "test-session")

	mockHistory := &mockChatHistory{}
	m.ChatHistory = mockHistory

	err := m.Clear(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !mockHistory.cleared {
		t.Errorf("Expected Clear to be called on chat history")
	}
}

// mockChatHistory implements schema.ChatMessageHistory for testing
type mockChatHistory struct {
	messages         []llms.ChatMessage
	messagesErr      error
	userMessageAdded string
	aiMessageAdded   string
	addUserErr       error
	addAIErr         error
	cleared          bool
}

func (m *mockChatHistory) Messages(_ context.Context) ([]llms.ChatMessage, error) {
	if m.messagesErr != nil {
		return nil, m.messagesErr
	}
	return m.messages, nil
}

func (m *mockChatHistory) AddUserMessage(_ context.Context, text string) error {
	m.userMessageAdded = text
	return m.addUserErr
}

func (m *mockChatHistory) AddAIMessage(_ context.Context, text string) error {
	m.aiMessageAdded = text
	return m.addAIErr
}

func (m *mockChatHistory) AddMessage(_ context.Context, _ llms.ChatMessage) error {
	return nil
}

func (m *mockChatHistory) SetMessages(_ context.Context, _ []llms.ChatMessage) error {
	return nil
}

func (m *mockChatHistory) Clear(_ context.Context) error {
	m.cleared = true
	return nil
}

// TestChatMessageHistoryMethods tests the ChatMessageHistory implementation methods
func TestChatMessageHistoryMethods(t *testing.T) {
	t.Parallel()

	// Since these methods require actual Zep client interaction,
	// we'll test the basic functionality and error paths

	t.Run("SetMessages", func(t *testing.T) {
		h := &ChatMessageHistory{}
		err := h.SetMessages(context.Background(), []llms.ChatMessage{})
		if err != nil {
			t.Errorf("SetMessages should return nil, got %v", err)
		}
	})
}
