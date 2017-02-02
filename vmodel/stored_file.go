package model

import (
	"time"

	"github.com/valeriugold/vket/shared/database"
)

// *****************************************************************************
// StoredFile
// *****************************************************************************

// StoredFile table contains the information for each stored file from table stored_file
type StoredFile struct {
	// ObjectID  bson.ObjectId `bson:"_id"`
	ID        uint32    `db:"id" bson:"id,omitempty"` // Don't use ID, use StoredFileID() instead for consistency with MongoDB
	Name      string    `db:"name" bson:"name"`
	Size      uint64    `db:"size" bson:"size"`
	Md5       string    `db:"md5" bson:"md5"`
	RefCount  uint32    `db:"ref_count" bson:"ref_count"`
	CreatedAt time.Time `db:"created_at" bson:"created_at"`
	UpdatedAt time.Time `db:"updated_at" bson:"updated_at"`
}

func GetStoredFileByID(ID uint32) (StoredFile, error) {
	var err error

	result := StoredFile{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM stored_file WHERE id = ? LIMIT 1", ID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func GetStoredFileByMd5(md5 string) (StoredFile, error) {
	var err error

	result := StoredFile{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM stored_file WHERE md5 = ? LIMIT 1", md5)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func CreateStoredFile(name string, size uint64, md5 string) (StoredFile, error) {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO stored_file (name, size, md5, ref_count) VALUES (?,?,?,1)"+
			" ON DUPLICATE KEY UPDATE ref_count=ref_count+1",
			name, size, md5)
	default:
		err = ErrCode
	}

	if err != nil {
		return StoredFile{}, standardizeError(err)
	}
	return GetStoredFileByMd5(md5)
}

func DeleteStoredFileByMd5(md5 string) error {
	var err error
	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		if _, err = database.SQL.Exec("UPDATE stored_file SET ref_count = ref_count - 1 WHERE md5 = ? and ref_count > 0", md5); err == nil {
			_, err = database.SQL.Exec("DELETE FROM stored_file WHERE md5 = ? and ref_count = 0", md5)
		}
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func DeleteStoredFileByID(ID uint32) error {
	var err error
	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		if _, err = database.SQL.Exec("UPDATE stored_file SET ref_count = ref_count - 1 WHERE id = ? and ref_count > 0", ID); err == nil {
			_, err = database.SQL.Exec("DELETE FROM stored_file WHERE id = ? and ref_count = 0", ID)
		}
	default:
		err = ErrCode
	}

	return standardizeError(err)
}
