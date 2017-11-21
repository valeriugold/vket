package vmodel

import (
	"fmt"
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
	x, err := EventGetByUserIDName(tur.ID, teName)
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

func prepareDBUserEvent(t *testing.T) (User, Event) {
	// Connect to database
	database.Connect(configDb)

	// add user
	tu := User{FirstName: "testFirst", LastName: "testLast", Email: "test@test", Password: "testPass", Role: "user"}
	if err := UserCreate(tu.FirstName, tu.LastName, tu.Email, tu.Password, tu.Role); err != nil {
		t.Errorf("create user err: %v", err)
	}
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
	// retrive event
	ev, err := EventGetByUserIDName(tur.ID, teName)
	if err != nil {
		t.Errorf("retriving events %d, err: %v\n", tur.ID, err)
	}

	return tur, ev
}

func cleanupDBUserEvent(us User, ev Event) {
	EventGetByEventID(ev.ID)
	UserDeleteByID(us.ID)
}

func createEventFile(t *testing.T, ev Event, usID uint32, status string, sfid uint32) EventFile {
	// add event file
	tuf := EventFile{EventID: ev.ID, OwnerID: usID, Name: fmt.Sprintf("testUFname-%d", sfid), StoredFileID: sfid}
	if err := EventFileCreate(tuf.EventID, usID, status, tuf.Name, tuf.StoredFileID); err != nil {
		t.Errorf("err creating event_file %v, err: %v", tuf, err)
	}

	// retrieve event file
	ef, err := EventFileGetByEventIDOwnerIDName(tuf.EventID, usID, tuf.Name)
	if err != nil {
		t.Errorf("err getting event_file (%d, %s), err: %v", tuf.EventID, tuf.Name, err)
	}

	// check tuf == uf
	if tuf.EventID != ef.EventID || tuf.Name != ef.Name {
		t.Errorf("what was gotten %v is not what was set %v", ef, tuf)
	}
	return ef
}

type storedFiles struct {
	sfs []StoredFile
}

func (s *storedFiles) clean() {
	for _, sf := range s.sfs {
		StoredFileDeleteByID(sf.ID)
	}
}

func createStoredFiles(count int) (storedFiles, error) {
	s := storedFiles{}
	for i := 0; i < count; i++ {
		f := StoredFile{Name: fmt.Sprintf("testStoredNamex%d", i),
			Size: int64(12345000 + i),
			Md5:  fmt.Sprintf("AA345678901234567890123456789%.3d", i%1000)}
		sf, err := StoredFileCreate(f.Name, f.Size, f.Md5)
		if err != nil {
			s.clean()
			// t.Errorf("creating stored file %s, %d, %s, err: %v\n", f.Name, f.Size, f.Md5, err)
			return s, err
		}
		s.sfs = append(s.sfs, sf)
	}
	return s, nil
}

func TestEventFile(t *testing.T) {
	s, err := createStoredFiles(1)
	if err != nil {
		t.Errorf("createStoredFiles error: %v", err)
	}
	defer s.clean()
	sf := s.sfs[0]
	// f := StoredFile{Name: "testStoredNamex",
	// 	Size: 12345678,
	// 	Md5:  "AA345678901234567890123456789011"}
	// // create and retrieve stored file
	// sf, err := StoredFileCreate(f.Name, f.Size, f.Md5)
	// if err != nil {
	// 	t.Errorf("creating stored file %s, %d, %s, err: %v\n", f.Name, f.Size, f.Md5, err)
	// }
	// defer StoredFileDeleteByID(sf.ID)

	us, ev := prepareDBUserEvent(t)
	defer cleanupDBUserEvent(us, ev)

	t.Logf("event = %v\n", ev)
	// add event file
	deleteEventFile := true
	ef := createEventFile(t, ev, us.ID, "original", sf.ID)
	defer func() {
		if deleteEventFile {
			EventFileDelete(ef.EventID, us.ID, ef.Name)
		}
	}()

	// try to add same user_file again
	err = EventFileCreate(ef.EventID, us.ID, "original", ef.Name, ef.StoredFileID)
	if err == nil {
		t.Errorf("no err creating duplicate event_file %v", ef)
		//os.Exit(1)
	}
	t.Logf("expected error creating duplicate event_file, err: %v", err)

	// delete event_file
	if err = EventFileDelete(ef.EventID, us.ID, ef.Name); err != nil {
		t.Errorf("deleting event_file (%d, %s), err: %v\n", ef.EventID, ef.Name, err)
	}
	deleteEventFile = false

	// try retrieve deleted event file
	_, err = EventFileGetByEventIDOwnerIDName(ef.EventID, us.ID, ef.Name)
	if err == nil {
		t.Errorf("no error on getting deleted event_file")
	}
	t.Logf("expected error retriving deleted event_file err: %v\n", err)
}

func TestEventFilePreview(t *testing.T) {
	us, ev := prepareDBUserEvent(t)
	defer cleanupDBUserEvent(us, ev)

	f := StoredFile{Name: "testStoredNameO",
		Size: 12345678,
		Md5:  "AA345678901234567890123456789014"}
	// create and retrieve stored file
	sf, err := StoredFileCreate(f.Name, f.Size, f.Md5)
	if err != nil {
		t.Errorf("creating stored file %s, %d, %s, err: %v\n", f.Name, f.Size, f.Md5, err)
	}
	defer StoredFileDeleteByID(sf.ID)

	var editorID uint32 = 1

	deleteEventFile := true
	ef := createEventFile(t, ev, editorID, "proposal", sf.ID)
	defer func() {
		if deleteEventFile {
			EventFileDelete(ef.EventID, us.ID, ef.Name)
		}
	}()

	fp := StoredFile{Name: "testStoredNameP",
		Size: 12345678,
		Md5:  "AA345678901234567890123456789015"}
	// create and retrieve stored file
	sfp, err := StoredFileCreate(fp.Name, fp.Size, fp.Md5)
	if err != nil {
		t.Errorf("creating stored file %s, %d, %s, err: %v\n", fp.Name, fp.Size, fp.Md5, err)
	}
	defer StoredFileDeleteByID(sfp.ID)

	// retrieve preview file
	if err := EventFileCreatePreview(ev.ID, editorID, ef.Name, sfp.ID); err != nil {
		t.Errorf("Err on EventFileCreatePreview(%d, %d, %s, %d): %v", ev.ID, editorID, ef.Name, sf.ID, err)
	}

	// get preview
	pr, err := EventFileGetPreview(ef)
	if err != nil {
		t.Errorf("on EventFileGetPreview(%v), err=%v", ef, err)
	}
	defer EventFileDeleteByID(pr.ID)

	// delete event file, should delete preview as well
	if err = EventFileAcceptPreviewID(pr.ID); err != nil {
		t.Errorf("EventFileAcceptPreviewID(%d), err: %v\n", pr.ID, err)
	}

	// check that preview file is gone
	_, err = EventFileGetByEventFileID(pr.ID)
	if err == nil {
		t.Errorf("no error on getting deleted preview event_file")
	}
	t.Logf("expected error retriving deleted preview event_file err: %v\n", err)
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

func TestEditorEvent(t *testing.T) {
	// create users, event
	us, ev := prepareDBUserEvent(t)
	defer cleanupDBUserEvent(us, ev)

	// create files
	s, err := createStoredFiles(5)
	if err != nil {
		t.Errorf("createStoredFiles error: %v", err)
	}
	defer s.clean()

	// create event files
	deleteEventFile := true
	efs := make([]EventFile, len(s.sfs))
	for i, sf := range s.sfs {
		efs[i] = createEventFile(t, ev, us.ID, "original", sf.ID)
	}
	defer func() {
		if deleteEventFile {
			for _, ef := range efs {
				EventFileDelete(ef.EventID, us.ID, ef.Name)
			}
		}
	}()

	// confirm there is editor 1 added to the event (by default)
	ee, err := EditorEventGetByEditorEventID(1, ev.ID)
	if err != nil {
		t.Error("err from EditorEventGetByEditorEventID:", err)
	}
	t.Logf("ee = %v", ee)

	// add files to EditorEvent
	efids := make([]uint32, len(efs))
	for i, v := range efs {
		efids[i] = v.ID
	}
	price := CalculatePrice(ev, efids)
	if err = EditorEventCreate(1, ev.ID, price, "instructions,,,", efids); err != nil {
		t.Error("err from EditorEventCreate:", err)
	}
	defer func() {
		EditorEventFileDeleteAllByEditorEventID(ee.ID)
		EditorEventDeleteByID(ee.ID)
	}()

	// display ee and eef
	eeGet, err := EditorEventGetByEditorEventID(1, ev.ID)
	if err != nil {
		t.Error("err from EditorEventGetByEditorEventID:", err)
	}
	t.Logf("eeGet = %v", eeGet)
	efidsGet, err := EditorEventFileGetEFIDs(eeGet.ID)
	if err != nil {
		t.Error("err from EditorEventGetByEditorEventID:", err)
	}
	for _, efid := range efidsGet {
		ef, err := EventFileGetByEventFileID(efid)
		if err != nil {
			t.Error("err from EditorEventGetByEditorEventID:", err)
		}
		t.Logf("efGet = %v", ef)
	}
	t.Logf("efidsGet = %v", efidsGet)
}
