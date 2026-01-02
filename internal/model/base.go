package model

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ID uuid.UUID

func NewID() ID {
	return ID(uuid.Must(uuid.NewV7()))
}

func NewTimestamp() time.Time {
	return time.Now().UTC()
}

func ParseID(s string) (ID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return ID{}, fmt.Errorf("invalid uuid string: %w", err)
	}
	return ID(id), nil
}

func FromBytes(b []byte) (ID, error) {
	id, err := uuid.FromBytes(b)
	if err != nil {
		return ID{}, fmt.Errorf("invalid uuid bytes: %w", err)
	}
	return ID(id), nil
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

func (id ID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

func (id *ID) Scan(value interface{}) error {
	var u uuid.UUID
	if err := u.Scan(value); err != nil {
		return fmt.Errorf("scan id: %w", err)
	}
	*id = ID(u)
	return nil
}
