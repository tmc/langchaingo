package memory

import (
	"testing"
)

func TestRedisChatMessageHistory(t *testing.T) {
	t.Parallel()
	// h := NewRedisChatMessageHistory(
	//	WithRedisConfOptions(RedisConfOptions{
	//		Address:      "localhost:6379",
	//		Db:           5,
	//		Password:     "",
	//		ReadTimeout:  3000,
	//		WriteTimeout: 3000,
	//		IdleTimeout:  60000,
	//		PoolSize:     20,
	//		SessionId:    "aaaaaaaaabbbbbbbcccccc",
	//		KeyPrefix:    "langchaingo_redis",
	//	}),
	//	WithRedisPreviousMessages([]schema.ChatMessage{
	//		schema.AIChatMessage{Content: "foo"},
	//		schema.SystemChatMessage{Content: "bar"},
	//	}),
	//)
	// err := h.AddUserMessage(context.Background(), "zoo")
	// require.NoError(t, err)
	//
	//messages, err := h.Messages(context.Background())
	// require.NoError(t, err)
	//
	//assert.Equal(t, []schema.ChatMessage{
	//	schema.AIChatMessage{Content: "foo"},
	//	schema.SystemChatMessage{Content: "bar"},
	//	schema.HumanChatMessage{Content: "zoo"},
	//}, messages)
}

/*
=== RUN   TestRedisChatMessageHistory
=== PAUSE TestRedisChatMessageHistory
=== CONT  TestRedisChatMessageHistory
--- PASS: TestRedisChatMessageHistory (0.70s)
PASS
*/
