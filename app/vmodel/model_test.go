package vmodel

import (
	"testing"

	"github.com/valeriugold/vket/app/shared/database"
)

var configDb = database.Info{Type: database.TypeMySQL, MySQL: database.MySQLInfo{
	Username:  "valeriug",
	Password:  "tset",
	Name:      "vket",
	Hostname:  "127.0.0.1",
	Port:      3306,
	Parameter: "?parseTime=true"}}

func TestEvent(t *testing.T) {
	// Connect to database
	database.Connect(configDb)

	// add user
	tu := User{FirstName: "testFirst", LastName: "testLast", Email: "test@test", Password: "testPass", Role: "user"}
	if err := UserCreate(tu.FirstName, tu.LastName, tu.Email, tu.Password, tu.Role); err != nil {
		t.Errorf("create user err: %v", err)
	}
	defer UserDelete(tu.Email)
	// retrive user
	tur, err := UserByEmail(tu.Email)
	if err != nil {
		t.Errorf("retriving user %s, err: %v\n", tur.Email, err)
	}

	// add event
	teName := "test name"
	if err := EventCreate(tur.ID, teName); err != nil {
		t.Errorf("create event err: %v", err)
	}
	defer EventDelete(tur.ID, teName)
	// retrive user
	x, err := EventByUserIDName(tur.ID, teName)
	if err != nil {
		t.Errorf("retriving events %d, err: %v\n", tur.ID, err)
	}

	// check the event is the same
	if tur.ID != x.UserID || teName != x.Name {
		t.Errorf("returned wrong event\n")
	}

	// try to add the same event again
	if err = EventCreate(tur.ID, teName); err == nil {
		t.Errorf("adding same event did not return ern")
	}
	t.Logf("expected error adding same event %v\n", err)
}

func TestUser(t *testing.T) {
	// Connect to database
	database.Connect(configDb)

	// add user
	tu := User{FirstName: "testFirst", LastName: "testLast", Email: "test@test", Password: "testPass", Role: "user"}
	err := UserCreate(tu.FirstName, tu.LastName, tu.Email, tu.Password, tu.Role)
	if err != nil {
		t.Errorf("adding user %v, err: %v\n", tu, err)
	}
	// try to add the same user again
	err = UserCreate(tu.FirstName, tu.LastName, tu.Email, tu.Password, tu.Role)
	if err == nil {
		t.Errorf("adding same user %v returned no err", tu)
	}
	t.Logf("expected error adding same user %v\n", err)
	// retrieve user
	x, err := UserByEmail(tu.Email)
	if err != nil {
		t.Errorf("retriving user %s, err: %v\n", tu.Email, err)
	}

	// check the user is the same
	if tu.FirstName != x.FirstName ||
		tu.LastName != x.LastName ||
		tu.Email != x.Email ||
		tu.Password != x.Password ||
		tu.Role != x.Role {
		t.Errorf("what was gotten %v is not what was set %v", x, tu)
	}

	if err = UserDelete(tu.Email); err != nil {
		t.Errorf("deleting user %s, err: %v\n", tu.Email, err)
	}

	// retrieve deleted user
	_, err = UserByEmail(tu.Email)
	if err == nil {
		t.Errorf("no error on retriving deleted user %s\n", tu.Email)
	}

	t.Logf("expected error retriving deleted user %s, err: %v\n", tu.Email, err)
}

func TestEventFile(t *testing.T) {
	// Connect to database
	database.Connect(configDb)

	// add user
	tu := User{FirstName: "testFirst", LastName: "testLast", Email: "test@test", Password: "testPass", Role: "user"}
	if err := UserCreate(tu.FirstName, tu.LastName, tu.Email, tu.Password, tu.Role); err != nil {
		t.Errorf("create user err: %v", err)
	}
	defer UserDelete(tu.Email)
	// retrive user
	tur, err := UserByEmail(tu.Email)
	if err != nil {
		t.Errorf("retriving user %s, err: %v\n", tur.Email, err)
	}

	// add event
	teName := "test name"
	if err := EventCreate(tur.ID, teName); err != nil {
		t.Errorf("create event err: %v", err)
	}
	defer EventDelete(tur.ID, teName)
	// retrive event
	x, err := EventByUserIDName(tur.ID, teName)
	if err != nil {
		t.Errorf("retriving events %d, err: %v\n", tur.ID, err)
	}

	// add user file
	tuf := EventFile{EventID: x.ID, Name: "testUFname", StoredFileID: 1}
	if err = EventFileCreate(tuf.EventID, tuf.Name, tuf.StoredFileID); err != nil {
		t.Errorf("err creating event_file %v, err: %v", tuf, err)
	}
	defer EventFileDelete(tuf.EventID, tuf.Name)

	// retrieve user file
	uf, err := EventFileGetByEventIDName(tuf.EventID, tuf.Name)
	if err != nil {
		t.Errorf("err getting event_file (%d, %s), err: %v", tuf.EventID, tuf.Name, err)
	}

	// check tuf == uf
	if tuf.EventID != uf.EventID || tuf.Name != uf.Name {
		t.Errorf("what was gotten %v is not what was set %v", uf, tuf)
	}

	// try to add same user_file again
	if err = UserFileCreate(tuf.EventID, tuf.Name, tuf.StoredFileID); err == nil {
		t.Errorf("no err creating duplicate event_file %v", tuf)
	}
	t.Logf("expected error creating duplicate event_file, err: %v", err)

	// delete event_file
	if err = EventFileDelete(tuf.EventID, tuf.Name); err != nil {
		t.Errorf("deleting event_file (%d, %s), err: %v\n", tuf.EventID, tuf.Name, err)
	}

	// try retrieve deleted event file
	_, err = EventFileGetByEventIDName(tuf.EventID, tuf.Name)
	if err == nil {
		t.Errorf("no error on getting deleted event_file")
	}
	t.Logf("expected error retriving deleted event_file err: %v\n", err)
}

func TestUserFile(t *testing.T) {
	// Connect to database
	database.Connect(configDb)

	// add user
	tu := User{FirstName: "testFirst", LastName: "testLast", Email: "test@test", Password: "testPass", Role: "user"}
	err := UserCreate(tu.FirstName, tu.LastName, tu.Email, tu.Password, tu.Role)
	// retrive user
	x, err := UserByEmail(tu.Email)
	if err != nil {
		t.Errorf("retriving user %s, err: %v\n", tu.Email, err)
	}

	// add user file
	tuf := UserFile{UserID: x.ID, Name: "testUFname", StoredFileID: 1}
	if err = UserFileCreate(tuf.UserID, tuf.Name, tuf.StoredFileID); err != nil {
		t.Errorf("err creating user_file %v, err: %v", tuf, err)
	}

	// retrieve user file
	uf, err := UserFileGetByUserIDName(tuf.UserID, tuf.Name)
	if err != nil {
		t.Errorf("err getting user_file (%d, %s), err: %v", tuf.UserID, tuf.Name, err)
	}

	// check tuf == uf
	if tuf.UserID != uf.UserID ||
		tuf.Name != uf.Name ||
		tuf.StoredFileID != uf.StoredFileID {
		t.Errorf("what was gotten %v is not what was set %v", uf, tuf)
	}
	//t.Logf("got back %v", uf)

	// try to add same user_file again
	if err = UserFileCreate(tuf.UserID, tuf.Name, tuf.StoredFileID); err == nil {
		t.Errorf("no err creating duplicate user_file %v", tuf)
	}
	t.Logf("expected error creating duplicate user_file, err: %v", err)

	// delete user_file
	if err = UserFileDelete(tuf.UserID, tuf.Name); err != nil {
		t.Errorf("deleting user_file (%d, %s), err: %v\n", tuf.UserID, tuf.Name, err)
	}

	// delete user
	if err = UserDelete(tu.Email); err != nil {
		t.Errorf("deleting user %s, err: %v\n", tu.Email, err)
	}
}

func TestStoredFile(t *testing.T) {
	// Connect to database
	database.Connect(configDb)

	f := StoredFile{Name: "testStoredName",
		Size: 12345678,
		Md5:  "AA345678901234567890123456789012"}

	// create and retrieve stored file
	c, err := StoredFileCreate(f.Name, f.Size, f.Md5)
	if err != nil {
		t.Errorf("creating stored file %s, %d, %s, err: %v\n", f.Name, f.Size, f.Md5, err)
	}
	if f.Name != c.Name ||
		f.Size != c.Size ||
		f.Md5 != c.Md5 ||
		c.RefCount != 1 {
		t.Errorf("what was gotten %v is not what was set %v", c, f)
	}
	id := c.ID

	// add the same file, all we expect is a bigger ref count
	c.Name = "otherName"
	c, err = StoredFileCreate(c.Name, c.Size, f.Md5)
	if err != nil {
		t.Errorf("creating another stored file %s, %d, %s, err: %v\n", c.Name, c.Size, c.Md5, err)
	}

	if f.Name != c.Name ||
		f.Size != c.Size ||
		f.Md5 != c.Md5 ||
		c.RefCount != 2 {
		t.Errorf("what was gotten on second create %v is not what was set %v", c, f)
	}

	if err = StoredFileDeleteByID(id); err != nil {
		t.Errorf("err on del stored_file ID=%d", id)
	}

	// get file should return ref_count =1
	if c, err = StoredFileGetByID(id); err != nil {
		t.Errorf("err on get stored_file ID=%d", id)
	}

	if f.Name != c.Name ||
		f.Size != c.Size ||
		f.Md5 != c.Md5 ||
		c.RefCount != 1 {
		t.Errorf("what was gotten on third get %v is not what was set %v", c, f)
	}

	if err = StoredFileDeleteByID(id); err != nil {
		t.Errorf("err on del stored_file ID=%d", id)
	}

	// get file should return error
	if c, err = StoredFileGetByID(id); err == nil {
		t.Errorf("err missing on get deleted stored_file ID=%d", id)
	}

	if err == ErrNoResult {
		t.Logf("expected error ErrNoResult getting deleted file")
	} else {
		t.Logf("expected error getting deleted file err: %v", err)
	}
}
