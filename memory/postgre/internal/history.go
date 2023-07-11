package internal

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

var ErrScanType = errors.New("could not scan type into Message")

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
	return ErrScanType
}
