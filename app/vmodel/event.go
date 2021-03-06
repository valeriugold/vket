package vmodel

import (
	"time"

	"github.com/valeriugold/vket/app/shared/database"
)

// *****************************************************************************
// Event
// *****************************************************************************

// Event table contains the information for each event
type Event struct {
	// ObjectID  bson.ObjectId `bson:"_id"`
	ID        uint32    `db:"id" bson:"id,omitempty"` // Don't use Id, use EventID() instead for consistency with MongoDB
	Name      string    `db:"name" bson:"name"`
	UserID    uint32    `db:"user_id" bson:"user_id"`
	Status    string    `db:"status" bson:"status"`
	CreatedAt time.Time `db:"created_at" bson:"created_at"`
	UpdatedAt time.Time `db:"updated_at" bson:"updated_at"`
}

// EventGetByUserIDName gets event information by UserID and Name
func EventGetByUserIDName(userID uint32, name string) (Event, error) {
	var err error

	result := Event{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM event WHERE user_id = ? AND name = ? LIMIT 1",
			userID, name)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EventGetByEventID gets event information by EventID
func EventGetByEventID(eventID uint32) (Event, error) {
	var err error

	result := Event{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM event WHERE id = ? LIMIT 1", eventID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EventGetAllForUserID gets all events for a user_id
func EventGetAllForUserID(userID uint32) ([]Event, error) {
	var err error

	var result []Event

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&result,
			"SELECT id, name, user_id, status, created_at, updated_at FROM event "+
				"WHERE user_id = ?", userID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EventCreate creates event
func EventCreate(userID uint32, name string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO event (name, user_id, status) VALUES (?,?,?)",
			name, userID, "open")
	default:
		err = ErrCode
	}

	// automatically add all events to editor id 1
	if err == nil {
		if ev, err := EventGetByUserIDName(userID, name); err == nil {
			EditorEventCreate(1, ev.ID, 0, "added automatically", []uint32{})
		}
	}

	return standardizeError(err)
}

func EventDelete(userID uint32, name string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM event WHERE user_id = ? AND name = ? LIMIT 1", userID, name)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}
