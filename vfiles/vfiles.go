package vfiles

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strconv"

	"github.com/valeriugold/vket/vfiles/vlocal"
	"github.com/valeriugold/vket/vlog"
	model "github.com/valeriugold/vket/vmodel"
)

type SaveLoader interface {
	Save(r io.Reader, nameHint string) (fileId string, size int64, savedMd5 string, err error)
	Load(w io.Writer, fileId string) error
	Remove(fileId string) error
	DoesExist(fileId string) bool
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
		vlog.Trace.Printf("init vsl from vlocal")
		vsl = vlocal.InitConfiguration(config.VLocal)
	} else {
		log.Fatalf("wrong type for file store (%s), only vlocal is allowed for now", c.Type)
	}
	log.Printf("vfiles: %v\n", config)
}

// SaveMultipart handles a multipart files upload, by storing the actual files and
// filling all necessary DB information
func SaveMultipart(eventID uint32, mr *multipart.Reader) error {
	//copy each part to destination.
	vlog.Trace.Printf("SaveMultipart")
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			return nil
		}

		formName := part.FormName()
		vlog.Trace.Printf("formName=%v", formName)
		//if part.FileName() is empty, skip this iteration.
		if part.FileName() == "" {
			if formName == "eventID" {
				buf := new(bytes.Buffer)
				buf.ReadFrom(part)
				vlog.Trace.Println("eventID is: ", buf.String())
				if eid, err := strconv.ParseUint(buf.String(), 10, 32); err == nil {
					eventID = uint32(eid)
				} else {
					log.Fatalf("coudl not parseuint %s, err: %v", buf.String(), err)
				}
			}
			continue
		}

		if eventID == 0 {
			log.Fatalf("eventID=%d, it was not read from form", eventID)
		}
		if err = SaveData(eventID, part.FileName(), part); err != nil {
			return err
		}

	}
}

func LoadData(eventID uint32, name string, w io.Writer) error {
	uf, err := model.EventFileGetByEventIDName(eventID, name)
	if err != nil {
		return err
	}
	sf, err := model.StoredFileGetByID(uf.StoredFileID)
	if err != nil {
		return err
	}
	if err = vsl.Load(w, sf.Name); err != nil {
		return err
	}
	return nil
}

func DeleteData(eventID uint32, name string) error {
	uf, err := model.EventFileGetByEventIDName(eventID, name)
	if err != nil {
		return err
	}
	sf, err := model.StoredFileGetByID(uf.StoredFileID)
	if err != nil {
		return err
	}
	if err = vsl.Remove(sf.Name); err != nil {
		return err
	}
	if err = model.EventFileDelete(eventID, name); err != nil {
		return err
	}
	if err = model.StoredFileDeleteByID(uf.StoredFileID); err != nil {
		return err
	}
	return nil
}

// try limit size with io.LimitReader
func SaveData(eventID uint32, receivedName string, r io.Reader) error {
	vlog.Trace.Printf("multipart file: %s\n", receivedName)
	vlog.Trace.Printf("SaveData eventID=%d, rcvName=%s", eventID, receivedName)
	fileId, size, calculatedMd5, err := vsl.Save(r, fmt.Sprintf("%d-%s", eventID, receivedName))
	if err != nil {
		return err
	}

	sf, err := model.StoredFileCreate(fileId, size, calculatedMd5)
	if err != nil {
		vsl.Remove(fileId)
		return err
	}

	// add to event_file
	if err = model.EventFileCreate(eventID, receivedName, size, calculatedMd5, sf.ID); err != nil {
		model.StoredFileDeleteByID(sf.ID)
		vsl.Remove(fileId)
		return err
	}

	return nil
	// // get temp name
	// tmpName, _ := vfiles.getTmpName(receivedName)
	// dst, err = os.Create(tmpName)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// size, err := io.Copy(dst, f)
	// dst.Close()
	// if err != nil {
	// 	os.Remove(tmpName)
	// 	// http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return err
	// }

	// calculatedMd5 := fmt.Sprintf("%x", hash.Sum(nil))
	// // add to event_file (if possible)
	// if err = model.EventFileCreate(eventID, receivedName, size, calculatedMd5, 1); err != nil {
	// 	os.Remove(tmpName)
	// 	if model.IsDuplicateEntry(err) {
	// 		// file already exists
	// 		vlog.Trace("file %s for event %s already exists", receivedName, eventID)
	// 		return nil
	// 	}
	// 	return err
	// }

	// // transcode file or whatever...
	// tmpName, size, calculatedMd5, err = TransformFile(tmp)
	// if err != nil {
	// 	os.Remove(tmpName)
	// 	return err
	// }

	// // defer deleting the file from temp storage
	// defer os.Remove(tmpName)

	// sf, err := model.StoredFileCreate(tmpName, size, calculatedMd5)
	// if err != nil {
	// 	// delete from DB
	// 	model.EventFileDeleteByEventIDName(eventID, receivedName)
	// 	return err
	// }
	// // the file already exists if DB name is different than tmpName;
	// // otherwise the file needs to be stored permanently
	// if sf.Name == tmpName {
	// 	if err = storeActualFile(tmpName, calculatedMd5); err != nil {
	// 		// delete from DB
	// 		model.EventFileDeleteByEventIDName(eventID, receivedName)
	// 		model.StoredFileDeleteByID(sf.ID)
	// 		return err
	// 	}
	// }
	// if err = EventFileSetStoredFileID(eventID, receivedName, sf.ID); err != nil {
	// 	// delete from DB
	// 	model.EventFileDeleteByEventIDName(eventID, receivedName)
	// 	model.StoredFileDeleteByID(sf.ID)
	// 	return err
	// }
	// return nil
}

// func storeActualFile(crtName, md5 string) (string, error) {
// 	nameMd5 := crtName + "-" + calculatedMd5
// 	if err := vsl.Save(crtName, nameMd5); err != nil {
// 		return err
// 	}
// 	return nil
// 	// only copy the file if the remote file is not there already,
// 	// if Stat returns with not error, the file already exists
// 	if sn, err := os.Stat(name); err != nil {
// 		if os.IsNotExist(err) {
// 			err = os.Rename(crtName, name)
// 		}
// 	}
// 	return err
// }

// TransformFile applies transformation on original file, for example transcoding
// it doesn't do anything right now, maybe later
func TransformFile(tmpName string, size uint64, calculatedMd5 string) (string, uint64, string, error) {
	return tmpName, size, calculatedMd5, nil
}
