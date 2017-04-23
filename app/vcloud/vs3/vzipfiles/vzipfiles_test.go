package vzipfiles

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/valeriugold/vket/app/vs3/vsecret"
)

type S3TestConfig struct {
	localFiles []string
	bucket     string
	prefix     string
	dirLoad    string
	dirSave    string
	eTag       string
}

var s3t = S3TestConfig{
	localFiles: []string{"aaa", "bbb", "ccc"},
	bucket:     "vket",
	prefix:     "test/",
	dirLoad:    "/tmp/vket-test-vs3/load/",
	dirSave:    "/tmp/vket-test-vs3/save/",
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

func hash_file_md5(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil

}

func TestZip(t *testing.T) {
	os.MkdirAll(s3t.dirLoad, 0777)
	os.MkdirAll(s3t.dirSave, 0777)
	// upload files to bucket prefix
	sess := session.New(&aws.Config{Region: aws.String("us-east-1")})
	svc := s3.New(sess)

	rand.Seed(time.Now().UnixNano())
	for _, name := range s3t.localFiles {
		// create file
		lName := s3t.dirLoad + name
		ioutil.WriteFile(lName, RandBytes(128), 0666)

		input, err := os.Open(lName)
		if err != nil {
			t.Errorf("open file %s, err: %v", lName, err)
		}

		params := &s3.PutObjectInput{
			Bucket: aws.String(s3t.bucket),        // Required
			Key:    aws.String(s3t.prefix + name), // Required
			Body:   input,
			// Body:               bytes.NewReader([]byte("PAYLOAD")),
		}
		resp, err := svc.PutObject(params)
		input.Close()

		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			t.Errorf("on PutObject: %v", err)
		}

		s3t.eTag = *resp.ETag
		// Pretty-print the response data.
		// t.Logf("PutObject %s: %v", s3t.prefix+name, resp)
	}

	// c := client{svc, &s3t.bucket}
	// // We let the service know that we want to do a multipart upload
	// output, err := c.s3Client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
	// 	Bucket: &bucket,
	// 	Key:    &key3,
	// })

	sf := make([]S3NameAndObjectID, len(s3t.localFiles))
	for i, name := range s3t.localFiles {
		sf[i].Name = name
		sf[i].ObjectID = s3t.prefix + name
	}
	// download the files in a zip
	handler := func(w http.ResponseWriter, r *http.Request) {
		z := NewZipMaker(vsecret.AWSAccessKey, vsecret.AWSSecretAccessKey, vsecret.AWSRegion, s3t.bucket)
		z.Zip(w, r, "vket-test.zip", sf)
		io.WriteString(w, "<html><body>Hello World!</body></html>")
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	destZipFileName := s3t.dirSave + "test.zip"
	ioutil.WriteFile(destZipFileName, body, 0666)
	// fmt.Println(string(body))

	// unzip file
	unzip(destZipFileName, s3t.dirSave)

	// compare the files with the original load
	for _, name := range s3t.localFiles {
		lName := s3t.dirLoad + name
		lm, _ := hash_file_md5(lName)
		dName := s3t.dirSave + name
		dm, _ := hash_file_md5(dName)
		if lm != dm {
			t.Errorf("diff md5 for %s: %s != %s", name, lm, dm)
		} else {
			t.Logf("same md5 for %s: %s", name, lm)
		}
	}

	// remove the files from S3 and local file system
	for _, name := range s3t.localFiles {
		params := &s3.DeleteObjectInput{
			Bucket: aws.String(s3t.bucket),        // Required
			Key:    aws.String(s3t.prefix + name), // Required
		}
		_, err := svc.DeleteObject(params)

		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			t.Errorf("on DeleteObject: %v", err)
		}
		// Pretty-print the response data.
		// t.Logf("DeleteObject %s: %v", s3t.prefix+name, resp)
	}
	os.RemoveAll(s3t.dirSave)
	os.RemoveAll(s3t.dirLoad)
}
