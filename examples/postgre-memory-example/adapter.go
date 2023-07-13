package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/tmc/langchaingo/schema"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ErrDBConnection     = errors.New("can't connect to database")
	ErrDBMigration      = errors.New("can't migrate database")
	ErrMissingSessionID = errors.New("session id can not be empty")
	ErrScan             = errors.New("can not scan type into Message")
)

type ChatHistory struct {
	ID           int       `gorm:"primary_key"`
	SessionID    string    `gorm:"type:varchar(256)"`
	BufferString string    `gorm:"type:text"`
	ChatHistory  *Messages `gorm:"type:jsonb;column:chat_history" json:"chat_history"`
}

type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Messages []Message

// Value implements the driver.Valuer interface, this method allows us to
// customize how we store the Message type in the database.
func (m Messages) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface, this method allows us to
// define how we convert the Message data from the database into our Message type.
func (m *Messages) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}
	return ErrScan
}

type PostgreAdapter struct {
	gorm      *gorm.DB
	history   *ChatHistory
	sessionID string
}

func NewPostgreAdapter() (*PostgreAdapter, error) {
	adapter := &PostgreAdapter{
		history: &ChatHistory{},
	}

	dsn := "host=localhost user=gpt-admin password=gpt-password dbname=gpt-db port=5432"
	gorm, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, ErrDBConnection
	}

	adapter.gorm = gorm

	err = adapter.gorm.AutoMigrate(ChatHistory{})
	if err != nil {
		return nil, ErrDBMigration
	}

	return adapter, nil
}

func (adapter *PostgreAdapter) SetSessionID(id string) {
	adapter.sessionID = id
}

func (adapter *PostgreAdapter) GetSessionID() string {
	return adapter.sessionID
}

func (adapter *PostgreAdapter) SaveDBContext(msgs []schema.ChatMessage, bufferString string) error {
	if adapter.sessionID == "" {
		return ErrMissingSessionID
	}

	newMsgs := Messages{}
	for _, msg := range msgs {
		newMsgs = append(newMsgs, Message{
			Type: string(msg.GetType()),
			Text: msg.GetText(),
		})
	}

	adapter.history.SessionID = adapter.sessionID
	adapter.history.ChatHistory = &newMsgs
	adapter.history.BufferString = bufferString

	err := adapter.gorm.Save(&adapter.history).Error
	if err != nil {
		return err
	}

	return nil
}

func (adapter *PostgreAdapter) LoadDBMemory() ([]schema.ChatMessage, error) {
	if adapter.sessionID == "" {
		return nil, ErrMissingSessionID
	}

	err := adapter.gorm.Where(ChatHistory{SessionID: adapter.sessionID}).Find(&adapter.history).Error
	if err != nil {
		return nil, err
	}

	msgs := []schema.ChatMessage{}
	if adapter.history.ChatHistory != nil {
		for i := range *adapter.history.ChatHistory {
			msg := (*adapter.history.ChatHistory)[i]

			if msg.Type == "human" {
				msgs = append(msgs, schema.HumanChatMessage{Text: msg.Text})
			}

			if msg.Type == "ai" {
				msgs = append(msgs, schema.AIChatMessage{Text: msg.Text})
			}
		}
	}

	return msgs, nil
}

func (adapter *PostgreAdapter) ClearDBContext() error {
	err := adapter.gorm.Where(ChatHistory{SessionID: adapter.sessionID}).Delete(&adapter.history).Error
	if err != nil {
		return err
	}
	return nil
}
