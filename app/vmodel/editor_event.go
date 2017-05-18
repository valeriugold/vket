package vmodel

import (
	"time"

	"github.com/valeriugold/vket/app/shared/database"
)

// *****************************************************************************
// User
// *****************************************************************************

// EditorEvent table contains the information for each user
type EditorEvent struct {
	// ObjectID  bson.ObjectId `bson:"_id"`
	ID        uint32    `db:"id" bson:"id,omitempty"` // Don't use Id, use UserID() instead for consistency with MongoDB
	EditorID  uint32    `db:"editor_id" bson:"editor_id,omitempty"`
	EventID   uint32    `db:"event_id" bson:"event_id,omitempty"`
	Status    string    `db:"status" bson:"status"`
	CreatedAt time.Time `db:"created_at" bson:"created_at"`
	UpdatedAt time.Time `db:"updated_at" bson:"updated_at"`
}

// EditorEventGetByID gets information from ID
func EditorEventGetByID(id uint32) (EditorEvent, error) {
	var err error

	result := EditorEvent{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM editor_event WHERE id = ? LIMIT 1", id)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EditorEventGetByEditorEventID gets information from Editor and Event ID
func EditorEventGetByEditorEventID(editorID, eventID uint32) (EditorEvent, error) {
	var err error

	result := EditorEvent{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM editor_event WHERE editor_id = ? AND event_id = ? LIMIT 1", editorID, eventID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EditorEventGetByEditorID gets information list for editor
func EditorEventGetByEditorID(editorID uint32) ([]EditorEvent, error) {
	var err error

	var result []EditorEvent

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&result,
			"SELECT * FROM editor_event WHERE editor_id = ?", editorID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EditorEventGetByEventID gets information list for event
func EditorEventGetByEventID(eventID uint32) ([]EditorEvent, error) {
	var err error

	var result []EditorEvent

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&result,
			"SELECT * FROM editor_event WHERE event_id = ?", eventID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EditorEventCreate creates editor-event association
func EditorEventCreate(editorID, eventID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO editor_event "+
			"(editor_id, event_id, status) VALUES (?,?,?)",
			editorID, eventID, "open")
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// EditorEventDelete remove by ID
func EditorEventDeleteByID(id uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM editor_event WHERE ID = ? LIMIT 1", id)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// EditorEventDelete remove by ID
func EditorEventDeleteByEditorEventID(editorID, eventID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM editor_event WHERE editor_id = ? and event_id = ? LIMIT 1", editorID, eventID)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}
