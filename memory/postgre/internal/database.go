package internal

import (
	"errors"

	"github.com/tmc/langchaingo/schema"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ErrDBConnection     = errors.New("can't connect to database")
	ErrDBMigration      = errors.New("can't migrate database")
	ErrMissingSessionID = errors.New("session id can not be empty")
)

type Database struct {
	gorm      *gorm.DB
	history   *ChatHistory
	sessionID string
}

func NewDatabase(dsn string) (*Database, error) {
	database := &Database{
		history: &ChatHistory{},
	}

	gorm, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, ErrDBConnection
	}

	database.gorm = gorm

	err = database.gorm.AutoMigrate(ChatHistory{})
	if err != nil {
		return nil, ErrDBMigration
	}

	return database, nil
}

func (db *Database) SetSession(id string) {
	db.sessionID = id
}

func (db *Database) SessionID() string {
	return db.sessionID
}

func (db *Database) SaveHistory(msgs []schema.ChatMessage, bs string) error {
	if db.sessionID == "" {
		return ErrMissingSessionID
	}

	newMsgs := Messages{}
	for _, msg := range msgs {
		newMsgs = append(newMsgs, Message{
			Type: string(msg.GetType()),
			Text: msg.GetText(),
		})
	}

	db.history.SessionID = db.sessionID
	db.history.ChatHistory = &newMsgs
	db.history.BufferString = bs

	err := db.gorm.Save(&db.history).Error
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) GetHistroy() ([]schema.ChatMessage, error) {
	if db.sessionID == "" {
		return nil, ErrMissingSessionID
	}

	err := db.gorm.Where(ChatHistory{SessionID: db.sessionID}).Find(&db.history).Error
	if err != nil {
		return nil, err
	}

	msgs := []schema.ChatMessage{}
	if db.history.ChatHistory != nil {
		for i := range *db.history.ChatHistory {
			msg := (*db.history.ChatHistory)[i]

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

func (db *Database) ClearHistroy() error {
	return nil
}
