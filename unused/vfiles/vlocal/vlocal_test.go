package vlocal

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"
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

func TestLocal(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	testDirLocal := "/tmp/vfiles_test_local"
	testLocalFile := testDirLocal + "/local1.txt"
	testDirRemote := "/tmp/vfiles_test_remote"
	testRemoteName := "local1r.txt"
	// create local file
	if err := os.MkdirAll(testDirLocal, 0777); err != nil {
		t.Error("could not create local dir " + testDirLocal)
	}
	if err := os.MkdirAll(testDirRemote, 0777); err != nil {
		t.Error("could not create remote dir " + testDirRemote)
	}
	createRandomFile(t, testLocalFile)

	t.Logf("save %s to %s, in dir %s", testLocalFile, testRemoteName, testDirRemote)
	r, err := os.Open(testLocalFile)
	if err != nil {
		t.Error("open testLocalFile", testLocalFile, "err=", err)
	}
	x := VFilesLocal{testDirRemote}
	storedId, size, md5Saved, err := x.Save(r, testRemoteName)
	if err != nil {
		t.Error("Save(%s, %s) returned error %v", testLocalFile, testRemoteName, err)
	}
	t.Logf("created x with %s / %s, size=%d, md5=%s", x.dir, storedId, size, md5Saved)
	r.Close()

	// check that remote file exists
	testRemoteFile := testDirRemote + "/" + storedId
	_, err = os.Stat(testRemoteFile)
	if err != nil {
		t.Error("there is no testRemoteFile %s, err=%v", testRemoteFile, err)
	}

	// get the file
	testLocalGetFile := testDirLocal + "/gotfile1.txt"
	w, err := os.Create(testLocalGetFile)
	if err != nil {
		t.Error(err)
	}
	if err = x.Load(w, storedId); err != nil {
		t.Error("could not get testRemoteName %s to %s, err=%v", testRemoteName, testLocalGetFile, err)
	}

	var md5fs [md5.Size]byte
	md5f1, _ := getMd5FromFile(t, testLocalFile)
	md5f2, _ := getMd5FromFile(t, testLocalGetFile)
	tmp, _ := hex.DecodeString(md5Saved)
	copy(md5fs[:], tmp)
	t.Logf("original %v ==? %v got from remote, saved=%s, %v", md5f1, md5f2, md5Saved, md5fs)
	if md5f1 != md5f2 || md5f1 != md5fs {
		t.Error("original file " + testLocalFile + "is not equal to gotten file " + testLocalGetFile)
	}

	// delete remote
	if err = x.Remove(testRemoteName); err != nil {
		t.Error("could not remove testRemoteName %s, err=%v", testRemoteName, err)
	}
}

func TestMain(m *testing.M) {
	//TestLocal(m)
	os.Exit(m.Run())
}
