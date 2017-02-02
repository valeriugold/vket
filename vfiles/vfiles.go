package vfiles

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/valeriugold/vket/vfiles"
	"github.com/valeriugold/vket/vfiles/vlocal"
	"github.com/valeriugold/vket/vlog"
	model "github.com/valeriugold/vket/vmodel"
)

type SaveLoader interface {
	Save(nameHere, nameThere string) error
	Load(nameHere, nameThere string) error
	Remove(nameThere string) error
	DoesExist(nameThere string) bool
}

// vsl is the object that will perform the actual save and load operation on files
// it will handle only the part related to moving the actual file around, not DB
var vsl SaveLoader

var config Configuration

type Configuration struct {
	Type   string               `json:"Type"`
	VLocal vlocal.Configuration `json:"VLocal"`
}

// InitConfiguration copy configuration to local config variable and init the system
func InitConfiguration(c Configuration) {
	config = c
	if c.Type == "vlocal" {
		vsl = vlocal.InitConfiguration(config.VLocal)
	} else {
		log.Fatalf("wrong type for file store (%s), only vlocal is allowed for now", c.Type)
	}
	log.Printf("vfiles: %v\n", config)
}

// SaveMultipart handles a multipart files upload, by storing the actual files and
// filling all necessary DB information
func SaveMultipart(userID uint32, mr multipart.Reader) error {
	//copy each part to destination.
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			return nil
		}

		//if part.FileName() is empty, skip this iteration.
		if part.FileName() == "" {
			continue
		}

		if err = SaveFile(userID, part.FileName(), part); err != nil {
			return err
		}

	}
}

func getTmpName(name string) string {
	return "/tmp/" + name, nil
}

func SaveFile(userID uint32, receivedName string, r io.Reader) error {
	vlog.Trace.Printf("multipart file: %s\n", receivedName)

	// prepare hash
	hash := md5.New()
	f := io.TeeReader(r, hash)

	// get temp name
	tmpName, _ := vfiles.getTmpName(receivedName)
	dst, err = os.Create(tmpName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	size, err := io.Copy(dst, f)
	dst.Close()
	if err != nil {
		os.Remove(tmpName)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	calculatedMd5 := fmt.Sprintf("%x", hash.Sum(nil))
	// add to user_file (if possible)
	if err = model.CreateUserFile(userID, receivedName, size, calculatedMd5, 1); err != nil {
		os.Remove(tmpName)
		if model.IsDuplicateEntry(err) {
			// file already exists
			vlog.Trace("file %s for user %s already exists", receivedName, userID)
			return nil
		}
		return err
	}

	// transcode file or whatever...
	tmpName, size, calculatedMd5, err = TransformFile(tmp)
	if err != nil {
		os.Remove(tmpName)
		return err
	}

	// defer deleting the file from temp storage
	defer os.Remove(tmpName)

	sf, err := model.CreateStoredFile(tmpName, size, calculatedMd5)
	if err != nil {
		// delete from DB
		model.DeleteUserFileByUserIDName(userID, receivedName)
		return err
	}
	// the file already exists if DB name is different than tmpName;
	// otherwise the file needs to be stored permanently
	if sf.Name == tmpName {
		if err = storeActualFile(tmpName, calculatedMd5); err != nil {
			// delete from DB
			model.DeleteUserFileByUserIDName(userID, receivedName)
			model.DeleteStoredFileByID(sf.ID)
			return err
		}
	}
	if err = SetStoredFileID(userID, receivedName, sf.ID); err != nil {
		// delete from DB
		model.DeleteUserFileByUserIDName(userID, receivedName)
		model.DeleteStoredFileByID(sf.ID)
		return err
	}
	return nil
}

func storeActualFile(crtName, md5 string) (string, error) {
	nameMd5 := crtName + "-" + calculatedMd5
	if err := vsl.Save(crtName, nameMd5); err != nil {
		return err
	}
	return nil
	// only copy the file if the remote file is not there already,
	// if Stat returns with not error, the file already exists
	if sn, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			err = os.Rename(crtName, name)
		}
	}
	return err
}

func retrieveActualFile(w io.Writer, name string, md5 string) {
	nameMd5 := name + "-" + calculatedMd5

}

// TransformFile applies transformation on original file, for example transcoding
// it doesn't do anything right now, maybe later
func TransformFile(tmpName string, size uint64, calculatedMd5 string) (string, uint64, string, error) {
	return tmpName, size, calculatedMd5, nil
}

func LoadFile(userID uint32, receivedName string, w io.Writer) error {
	// get file data from DB
	// user_file
	uf, err := GetUserFileByUserIDName(userID, receivedName)
	if err != nil {
		return err
	}
	// stored_file
	sf, err := GetStoredFileByID(uf.StoredFileID)
	if err != nil {
		return err
	}

}
