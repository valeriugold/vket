package vfiles

import (
	"crypto/md5"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

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
		t.Error("error writing to file %s: %v", testLocalFile, err)
	} else {
		t.Logf("wrote %d fo file %s\n", n, testLocalFile)
	}
	in.Close()
}

const config = `{ "vfiles" : { "storageType" : "local", "params" : { "destDir" : "/tmp/vfiles_test_local" } } }`

func TestVfiles(t *testing.T) {
	c := Configuration{Type: "vlocal", VLocal: vlocal.Configuration{DestDir: "/tmp/vfiles_test_local"}}
	InitConfiguration(c)

}

func TestLocal(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	testDirLocal := "/tmp/vfiles_test_local"
	testLocalFile := testDirLocal + "/local1.txt"
	testRemoteName := "local1r.txt"
	// create local file
	if err := os.MkdirAll(testDirLocal, 0777); err != nil {
		t.Error("could not create local dir " + testDirLocal)
	}
	createRandomFile(t, testLocalFile)

	var rf io.Reader
	if rf, err = Open(testLocalFile); err != nil {
		t.Errorf("err on open file %s, err: %v", testLocalFile, err)
	}

	// add user
	tu := model.User{FirstName: "testFirst", LastName: "testLast", Email: "test@test", Password: "testPass", Role: "user"}
	err := model.UserCreate(tu.FirstName, tu.LastName, tu.Email, tu.Password, tu.Role)
	if err != nil {
		t.Errorf("adding user %v, err: %v\n", tu, err)
	}
	u, err := model.UserByEmail(tu.Email)
	if err != nil {
		t.Errorf("retriving user %s, err: %v\n", tu.Email, err)
	}

	uid := u.ID
	_, rcvName := filepath.Split(testLocalFile)
	if err = SaveFile(uid, rcvName, rf); err != nil {
		t.Errorf("saveFile %d, %s, err: %v", uid, rcvName, err)
	}

	// get saved file from DB

	// check file exists

	t.Logf("create/init vfiles")
	InitConfiguration([]byte(config))
	t.Logf("save %s to %s", testLocalFile, testRemoteName)
	if err = FileSave(testLocalFile, testRemoteName); err != nil {
		t.Error("FileSave(%s, %s) returned error %v", testLocalFile, testRemoteName, err)
	}

	// get the file
	testLocalGetFile := testDirLocal + "/gotfile1.txt"
	if err = FileGet(testLocalGetFile, testRemoteName); err != nil {
		t.Error("could not get testRemoteName %s to %s, err=%v", testRemoteName, testLocalGetFile, err)
	}

	// [md5.Size]byte
	md5f1, _ := getMd5FromFile(t, testLocalFile)
	md5f2, _ := getMd5FromFile(t, testLocalGetFile)
	t.Logf("original %v ==? %v got from remote", md5f1, md5f2)
	if md5f1 != md5f2 {
		t.Error("original file " + testLocalFile + "is not equal to gotten file " + testLocalGetFile)
	}

	// delete remote
	if err = FileRemove(testRemoteName); err != nil {
		t.Error("could not remove testRemoteName %s, err=%v", testRemoteName, err)
	}
}
