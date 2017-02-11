package vfiles

import (
	"crypto/md5"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/valeriugold/vket/shared/database"
	"github.com/valeriugold/vket/vfiles/vlocal"
	model "github.com/valeriugold/vket/vmodel"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

func getMd5FromFile(t *testing.T, s string) ([md5.Size]byte, error) {
	// t.Logf("md5 for %s\n", s)
	hash := md5.New()

	file, err := os.Open(s)
	if err != nil {
		t.Error("err " + err.Error())
		// return returnMD5String, err
	}
	defer file.Close()
	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return [md5.Size]byte{}, err
	}
	h := hash.Sum(nil)
	var a [md5.Size]byte
	copy(a[:], h)
	return a, nil
}

func createRandomFile(t *testing.T, testLocalFile string) {
	if err := os.Remove(testLocalFile); err != nil {
		t.Log("ignore error " + err.Error())
	}
	// in, err := os.Open(testLocalFile)
	in, err := os.Create(testLocalFile)
	if err != nil {
		t.Error("could not create local file " + testLocalFile + " err: " + err.Error())
	}
	// t.Log("write file now!!!!!!!")
	if n, err := in.Write(RandBytes(1024)); err != nil {
		t.Error("error writing to file", testLocalFile, ": ", err)
	} else {
		t.Logf("wrote %d fo file %s\n", n, testLocalFile)
	}
	in.Close()
}

func TestVfiles(t *testing.T) {
	c := Configuration{Type: "vlocal", VLocal: vlocal.Configuration{DestDir: "/tmp/vfiles_test_local"}}
	InitConfiguration(c)

}

var configDb = database.Info{Type: database.TypeMySQL, MySQL: database.MySQLInfo{
	Username:  "valeriug",
	Password:  "tset",
	Name:      "vket",
	Hostname:  "127.0.0.1",
	Port:      3306,
	Parameter: "?parseTime=true"}}

func TestLocal(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	testDirLocal := "/tmp/vfiles_test_local"
	testLocalFile := testDirLocal + "/local1.txt"
	testLocalLoadedFile := testDirLocal + "/loaded.txt"
	// create local file
	if err := os.MkdirAll(testDirLocal, 0777); err != nil {
		t.Error("could not create local dir " + testDirLocal)
	}
	createRandomFile(t, testLocalFile)

	// add user
	// Connect to database
	database.Connect(configDb)
	tu := model.User{FirstName: "testFirstVfiles", LastName: "testLastVfiles", Email: "Vfiles@test", Password: "testPass", Role: "user"}
	err := model.UserCreate(tu.FirstName, tu.LastName, tu.Email, tu.Password, tu.Role)
	if err != nil {
		if !model.IsDuplicateEntry(err) {
			t.Errorf("adding user %v, err: %v\n", tu, err)
		}
	}
	defer model.UserDelete(tu.Email)
	u, err := model.UserByEmail(tu.Email)
	if err != nil {
		t.Errorf("retriving user %s, err: %v\n", tu.Email, err)
	}

	// add event
	teName := "test name"
	if err := model.EventCreate(u.ID, teName); err != nil {
		t.Errorf("create event err: %v", err)
	}
	defer model.EventDelete(u.ID, teName)
	// retrive event
	x, err := model.EventByUserIDName(u.ID, teName)
	if err != nil {
		t.Errorf("retriving events %d, err: %v\n", u.ID, err)
	}

	rf, err := os.Open(testLocalFile)
	if err != nil {
		t.Errorf("err on open file %s, err: %v", testLocalFile, err)
	}
	eid := x.ID
	rcvName := filepath.Base(testLocalFile)
	if err = SaveData(eid, rcvName, rf); err != nil {
		t.Errorf("saveFile %d, %s, err: %v", eid, rcvName, err)
	}
	rf.Close()
	defer DeleteData(eid, rcvName)

	// get saved file from DB
	nf, err := os.Create(testLocalLoadedFile)
	if err != nil {
		t.Error("create file testLocalLoadedFile ", err)
	}

	if err = LoadData(eid, rcvName, nf); err != nil {
		t.Error("LoadData ", err)
	}
	nf.Close()

	// check file exists
	// t.Logf("create/init vfiles")
	// testConfig := Configuration{Type: "vlocal",
	// 	VLocal: vlocal.Configuration{DestDir: "/tmp/vfiles_test_local"}}
	// InitConfiguration(testConfig)
	// t.Logf("save %s to %s", testLocalFile, testRemoteName)
	// if err = SaveData(eid, rcvName, rf); err != nil {
	// 	t.Error("FileSave(%s, %s) returned error %v", testLocalFile, testRemoteName, err)
	// }

	// // get the file
	// testLocalGetFile := testDirLocal + "/gotfile1.txt"
	// if err = FileGet(testLocalGetFile, testRemoteName); err != nil {
	// 	t.Error("could not get testRemoteName %s to %s, err=%v", testRemoteName, testLocalGetFile, err)
	// }

	// [md5.Size]byte
	md5f1, _ := getMd5FromFile(t, testLocalFile)
	md5f2, _ := getMd5FromFile(t, testLocalLoadedFile)
	t.Logf("original %v ==? %v got from remote", md5f1, md5f2)
	if md5f1 != md5f2 {
		t.Error("original file " + testLocalFile + "is not equal to gotten file " + testLocalLoadedFile)
	}

	// // delete remote
	// if err = FileRemove(testRemoteName); err != nil {
	// 	t.Error("could not remove testRemoteName %s, err=%v", testRemoteName, err)
	// }
}
