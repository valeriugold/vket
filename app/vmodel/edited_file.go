package vmodel

import (
	"time"

	"github.com/valeriugold/vket/app/shared/database"
)

// *****************************************************************************
// StoredFile
// *****************************************************************************

// EditedFile table contains the information for each stored file from table stored_file
type EditedFile struct {
	// ObjectID     bson.ObjectId `bson:"_id"`
	ID           uint32    `db:"id" bson:"id,omitempty"` // Don't use ID, use StoredFileID() instead for consistency with MongoDB
	EventID      uint32    `db:"event_id" bson:"event_id"`
	EditorID     uint32    `db:"editor_id" bson:"editor_id"`
	Name         string    `db:"name" bson:"name"` // The name of the file, as user sees it
	Status       string    `db:"status" bson:"status"`
	StoredFileID uint32    `db:"stored_file_id" bson:"stored_file_id"`
	CreatedAt    time.Time `db:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" bson:"updated_at"`
}

// Size         int64     `db:"size" bson:"size"`
// Md5          string    `db:"md5" bson:"md5"`

func EditedFileGetByEditedFileID(ID uint32) (EditedFile, error) {
	var err error

	result := EditedFile{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM edited_file WHERE ID = ? LIMIT 1", ID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func EditedFileGetByEventIDEditorIDName(eventID, editorID uint32, name string) (EditedFile, error) {
	var err error

	result := EditedFile{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM edited_file WHERE event_id = ? and editor_id = ? and name = ? LIMIT 1", eventID, editorID, name)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EditedFileGetAllForEventID gets all files for an event_id
func EditedFileGetAllForEventID(eventID uint32) ([]EditedFile, error) {
	var err error

	var result []EditedFile

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&result,
			"SELECT id,vevent_id, editor_id, name, stored_file_id, created_at, updated_at FROM edited_file "+
				"WHERE event_id = ?", eventID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EditedFileGetAllForEventID gets all files for an event_id and editorID
func EditedFileGetAllForEventIDEditorID(eventID, editorID uint32) ([]EditedFile, error) {
	var err error

	var result []EditedFile

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&result,
			"SELECT id, event_id, editor_id, name, stored_file_id, created_at, updated_at FROM edited_file "+
				"WHERE event_id = ? AND editor_id = ?", eventID, editorID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func EditedFileCreate(eventID, editorID uint32, name string, storedFileID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO edited_file (event_id, editor_id, name, status, stored_file_id) VALUES (?,?,?, 'processing',?)",
			eventID, editorID, name, storedFileID)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func EditedFileSetStoredFileID(eventID, editorID uint32, name string, storedFileID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("UPDATE edited_file SET stored_file_id = ? WHERE event_id = ? and editor_id = ? and name = ?",
			storedFileID, eventID, editorID, name)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func EditedFileDelete(eventID, editorID uint32, name string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM edited_file WHERE event_id = ? and editor_id = ? and name = ? LIMIT 1",
			eventID, editorID, name)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func EditedFileDeleteByID(ID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM edited_file WHERE ID = ? LIMIT 1", ID)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}
