package vmodel

import (
	"time"

	"github.com/valeriugold/vket/app/shared/database"
)

// *****************************************************************************
// StoredFile
// *****************************************************************************

// EventFile table contains the information for each stored file from table stored_file
type EventFile struct {
	// ObjectID     bson.ObjectId `bson:"_id"`
	ID           uint32    `db:"id" bson:"id,omitempty"` // Don't use ID, use StoredFileID() instead for consistency with MongoDB
	EventID      uint32    `db:"event_id" bson:"event_id"`
	Name         string    `db:"name" bson:"name"` // The name of the file, as user sees it
	StoredFileID uint32    `db:"stored_file_id" bson:"stored_file_id"`
	CreatedAt    time.Time `db:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" bson:"updated_at"`
}

// Size         int64     `db:"size" bson:"size"`
// Md5          string    `db:"md5" bson:"md5"`

func EventFileGetByEventFileID(ID uint32) (EventFile, error) {
	var err error

	result := EventFile{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM event_file WHERE ID = ? LIMIT 1", ID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func EventFileGetByEventIDName(eventID uint32, name string) (EventFile, error) {
	var err error

	result := EventFile{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM event_file WHERE event_id = ? and name = ? LIMIT 1", eventID, name)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EventFileGetAllForEventID gets all files for an event_id
func EventFileGetAllForEventID(eventID uint32) ([]EventFile, error) {
	var err error

	var result []EventFile

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&result,
			"SELECT id,event_id, name, stored_file_id, created_at, updated_at FROM event_file "+
				"WHERE event_id = ?", eventID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func EventFileCreate(eventID uint32, name string, storedFileID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO event_file (event_id, name, stored_file_id) VALUES (?,?,?)",
			eventID, name, storedFileID)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func EventFileSetStoredFileID(eventID uint32, name string, storedFileID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("UPDATE event_file SET stored_file_id = ? WHERE event_id = ? and name = ?",
			storedFileID, eventID, name)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func EventFileDelete(eventID uint32, name string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM event_file WHERE event_id = ? and name = ? LIMIT 1", eventID, name)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func EventFileDeleteByID(ID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM event_file WHERE ID = ? LIMIT 1", ID)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}
