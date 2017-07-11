package vmodel

import (
	"errors"
	"strings"
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
	OwnerID      uint32    `db:"owner_id" bson:"owner_id"`
	Status       string    `db:"status" bson:"status"`
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

func EventFileGetByEventIDOwnerIDName(eventID, ownerID uint32, name string) (EventFile, error) {
	var err error

	result := EventFile{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT * FROM event_file WHERE event_id = ? and owner_id = ? and name = ? LIMIT 1",
			eventID, ownerID, name)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// EventFileGetAllForEventIDOwnerID gets all files for an event_id
func EventFileGetAllForEventIDOwnerID(eventID, ownerID uint32) ([]EventFile, error) {
	var err error

	var result []EventFile

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&result,
			"SELECT id,event_id, name, stored_file_id, created_at, updated_at FROM event_file "+
				"WHERE event_id = ? and owner_id = ?", eventID, ownerID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func EventFileCreate(eventID, ownerID uint32, status, name string, storedFileID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO event_file (event_id, owner_id, status, name, stored_file_id) VALUES (?,?,?,?,?)",
			eventID, ownerID, status, name, storedFileID)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func EventFileSetStoredFileID(eventID, ownerID, uint32, name string, storedFileID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("UPDATE event_file SET stored_file_id = ? WHERE event_id = ? and owner_id = > and name = ?",
			storedFileID, eventID, ownerID, name)
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

func EventFileDelete(eventID, ownerID uint32, name string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM event_file WHERE event_id = ? and owner_id = ? and name = ? LIMIT 1", eventID, ownerID, name)
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

func EventFileCreatePreview(eventID, ownerID uint32, name string, storedFileID uint32) error {
	// if status != "proposal" {
	// 	return errors.New("The file status is not proposal")
	// }
	name = name + ".preview"
	return EventFileCreate(eventID, ownerID, "preview", name, storedFileID)
}

func EventFileDeletePreview(ef EventFile) error {
	pef, err := EventFileGetPreview(ef)
	if err == nil {
		err = EventFileDeleteByID(pef.ID)
	}
	return err
}

func EventFileGetPreview(ef EventFile) (EventFile, error) {
	if ef.Status != "proposal" {
		return EventFile{}, errors.New("The file status is not proposal")
	}
	name := ef.Name + ".preview"
	return EventFileGetByEventIDOwnerIDName(ef.EventID, ef.OwnerID, name)
}

func EventFileGetProposal(preview EventFile) (EventFile, error) {
	if preview.Status != "preview" {
		return EventFile{}, errors.New("The file status is not preview")
	}
	if !strings.HasSuffix(preview.Name, ".preview") {
		return EventFile{}, errors.New("The file name does not end with .preview")
	}
	name := strings.TrimSuffix(preview.Name, ".preview")
	return EventFileGetByEventIDOwnerIDName(preview.EventID, preview.OwnerID, name)
}

// func EventFileAccept(af, pf EventFile, ev Event, ownerID uint32) error {
func EventFileAcceptProposalID(efid uint32) error {
	aef, err := EventFileGetByEventFileID(efid)
	if err != nil {
		return err
	}
	ev, err := EventGetByEventID(aef.EventID)
	if err != nil {
		return err
	}
	// tranfer ownership
	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("UPDATE event_file SET owner_id = ?, status = 'accepted' WHERE ID = ?",
			ev.UserID, aef.ID)
	default:
		err = ErrCode
	}

	if err != nil {
		return standardizeError(err)
	}

	// delete preview file
	return EventFileDeletePreview(aef)
}

func EventFileAcceptPreviewID(efid uint32) error {
	pref, err := EventFileGetByEventFileID(efid)
	if err != nil {
		return err
	}
	aef, err := EventFileGetProposal(pref)
	if err != nil {
		return err
	}
	ev, err := EventGetByEventID(aef.EventID)
	if err != nil {
		return err
	}
	// tranfer ownership
	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("UPDATE event_file SET owner_id = ?, status = 'accepted' WHERE ID = ?",
			ev.UserID, aef.ID)
	default:
		err = ErrCode
	}
	if err != nil {
		return standardizeError(err)
	}
	// delete preview file
	return EventFileDeleteByID(pref.ID)
}

func EventFileRejectPreviewID(efid uint32) error {
	pref, err := EventFileGetByEventFileID(efid)
	if err != nil {
		return err
	}
	aef, err := EventFileGetProposal(pref)
	if err != nil {
		return err
	}
	// change status to 'rejected'
	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("UPDATE event_file SET status = 'rejected' WHERE ID = ?", aef.ID)
	default:
		err = ErrCode
	}
	if err != nil {
		return standardizeError(err)
	}
	// delete preview file
	return EventFileDeleteByID(pref.ID)
}
