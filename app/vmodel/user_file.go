package vmodel

import (
	"time"

	"github.com/valeriugold/vket/app/shared/database"
)

// *****************************************************************************
// StoredFile
// *****************************************************************************

// StoredFile table contains the information for each stored file from table stored_file
type UserFile struct {
	// ObjectID     bson.ObjectId `bson:"_id"`
	ID     uint32 `db:"id" bson:"id,omitempty"` // Don't use ID, use StoredFileID() instead for consistency with MongoDB
	UserID uint32 `db:"user_id" bson:"user_id"`
	Name   string `db:"name" bson:"name"` // The name of the file, as user sees it
	// Size         int64     `db:"size" bson:"size"`
	// Md5          string    `db:"md5" bson:"md5"`
	StoredFileID uint32    `db:"stored_file_id" bson:"stored_file_id"`
	CreatedAt    time.Time `db:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" bson:"updated_at"`
}

func UserFileGetByUserIDName(userID uint32, name string) (UserFile, error) {
	var err error

	result := UserFile{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM user_file WHERE user_id = ? and name = ? LIMIT 1", userID, name)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func UserFileCreate(userID uint32, name string, storedFileID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO user_file (user_id, name, stored_file_id) VALUES (?,?,?)",
			userID, name, storedFileID)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func UserFileSetStoredFileID(userID uint32, name string, storedFileID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("UPDATE user_file SET stored_file_id = ? WHERE user_id = ? and name = ?",
			storedFileID, userID, name)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func UserFileDelete(userID uint32, name string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM user_file WHERE user_id = ? and name = ? LIMIT 1", userID, name)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}
