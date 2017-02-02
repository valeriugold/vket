package model

import (
	"testing"

	"github.com/valeriugold/vket/shared/database"
)

var configDb = database.Info{Type: database.TypeMySQL, MySQL: database.MySQLInfo{
	Username:  "valeriug",
	Password:  "tset",
	Name:      "vket",
	Hostname:  "127.0.0.1",
	Port:      3306,
	Parameter: "?parseTime=true"}}

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
	tuf := UserFile{UserID: x.ID, Name: "testUFname", Size: 12345, Md5: "12345678901234567890123456789012", StoredFileID: 1}
	if err = CreateUserFile(tuf.UserID, tuf.Name, tuf.Size, tuf.Md5, tuf.StoredFileID); err != nil {
		t.Errorf("err creating user_file %v, err: %v", tuf, err)
	}

	// retrieve user file
	uf, err := GetUserFileByUserIDName(tuf.UserID, tuf.Name)
	if err != nil {
		t.Errorf("err getting user_file (%d, %s), err: %v", tuf.UserID, tuf.Name, err)
	}

	// check tuf == uf
	if tuf.UserID != uf.UserID ||
		tuf.Name != uf.Name ||
		tuf.Size != uf.Size ||
		tuf.Md5 != uf.Md5 ||
		tuf.StoredFileID != uf.StoredFileID {
		t.Errorf("what was gotten %v is not what was set %v", uf, tuf)
	}
	//t.Logf("got back %v", uf)

	// try to add same user_file again
	if err = CreateUserFile(tuf.UserID, tuf.Name, tuf.Size, tuf.Md5, tuf.StoredFileID); err == nil {
		t.Errorf("no err creating duplicate user_file %v", tuf)
	}
	t.Logf("expected error creating duplicate user_file, err: %v", err)

	// delete user_file
	if err = DeleteUserFile(tuf.UserID, tuf.Name); err != nil {
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
	c, err := CreateStoredFile(f.Name, f.Size, f.Md5)
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
	c, err = CreateStoredFile(c.Name, c.Size, f.Md5)
	if err != nil {
		t.Errorf("creating another stored file %s, %d, %s, err: %v\n", c.Name, c.Size, c.Md5, err)
	}

	if f.Name != c.Name ||
		f.Size != c.Size ||
		f.Md5 != c.Md5 ||
		c.RefCount != 2 {
		t.Errorf("what was gotten on second create %v is not what was set %v", c, f)
	}

	if err = DeleteStoredFileByID(id); err != nil {
		t.Errorf("err on del stored_file ID=%d", id)
	}

	// get file should return ref_count =1
	if c, err = GetStoredFileByID(id); err != nil {
		t.Errorf("err on get stored_file ID=%d", id)
	}

	if f.Name != c.Name ||
		f.Size != c.Size ||
		f.Md5 != c.Md5 ||
		c.RefCount != 1 {
		t.Errorf("what was gotten on third get %v is not what was set %v", c, f)
	}

	if err = DeleteStoredFileByID(id); err != nil {
		t.Errorf("err on del stored_file ID=%d", id)
	}

	// get file should return error
	if c, err = GetStoredFileByID(id); err == nil {
		t.Errorf("err missing on get deleted stored_file ID=%d", id)
	}

	if err == ErrNoResult {
		t.Logf("expected error ErrNoResult getting deleted file")
	} else {
		t.Logf("expected error getting deleted file err: %v", err)
	}
}
