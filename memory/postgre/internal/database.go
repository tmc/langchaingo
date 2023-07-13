package internal

import (
	"errors"

	"github.com/tmc/langchaingo/schema"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	// ErrDBConnection is returned when there is an issue connecting to the database. Usually incorrect DSN.
	ErrDBConnection = errors.New("can't connect to database")
	// ErrDBMigration is returned when there is an issue migrating the database.
	ErrDBMigration = errors.New("can't migrate database")
)

type Database struct {
	gorm      *gorm.DB
	history   *ChatHistory
	sessionID string
}

// NewDatabase creates a new Database instance.
//
// It takes a DSN (Data Source Name) string as a parameter.
// It returns a pointer to the created Database instance and an error.
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

// SetSession sets the session ID of the Database.
//
// id: the session ID to set.
func (db *Database) SetSession(id string) {
	db.sessionID = id
}

// SessionID returns the session ID of the Database.
//
// No parameters.
// Returns a string.
func (db *Database) SessionID() string {
	return db.sessionID
}

// SaveHistory saves the chat history to the database.
//
// It takes the following parameters:
// - id: the ID of the session
// - msgs: the chat messages to be saved
// - bs: the buffer string
//
// It returns an error if there was an issue saving the history.
func (db *Database) SaveHistory(id string, msgs []schema.ChatMessage, bs string) error {
	if db.sessionID == "" {
		db.sessionID = id
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

// GetHistory retrieves the chat history for a given session ID from the database.
//
// id: The ID of the session.
// []schema.ChatMessage: An array of chat messages representing the chat history.
// error: An error if there was a problem retrieving the chat history.
func (db *Database) GetHistroy(id string) ([]schema.ChatMessage, error) {
	if db.sessionID == "" {
		db.sessionID = id
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
	err := db.gorm.Where(ChatHistory{SessionID: db.sessionID}).Delete(&db.history).Error
	if err != nil {
		return err
	}
	return nil
}
