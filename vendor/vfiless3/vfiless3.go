package vfiless3

import (
	"io"

	"github.com/valeriugold/vket/vlog"
	model "github.com/valeriugold/vket/vmodel"
)

func S3Remove(id int32) error {
	//VG: todo
	return nil
}

type FileInfo struct {
	// Name string
	Size int64
	Md5  string
	Key  string
}

// func GetCloudFileInfo() (FileInfo, error)

type SaveLoader interface {
	Save(r io.Reader, nameHint string) (fileId string, size int64, savedMd5 string, err error)
	Load(w io.Writer, fileId string) error
	Remove(fileId string) error
	DoesExist(fileId string) bool
}

type VCloud interface {
	SaveUploadedFile(eventID uint32, fileName, s3Key, md5 string) error  // RecordUploadedFile
	DeleteFile(key interface{}) error
	// DeleteFileByEventIDName(eventID uint32, name string) error
	// DeleteFileByEventFileID(eventFileID uint32) error
	// DeleteFileByEventFile(ef model.EventFile) error
	GetDownloadLink(key interface{}) (string, error)
}

// var config Configuration
// type Configuration struct {
// 	Type   string               `json:"Type"`
// 	VLocal vlocal.Configuration `json:"VLocal"`
// }
// // InitConfiguration copy configuration to local config variable and init the system
// func InitConfiguration(c Configuration) {
// 	config = c
// 	if c.Type == "vlocal" {
// 		vlog.Trace.Printf("init vsl from vlocal")
// 		vsl = vlocal.InitConfiguration(config.VLocal)
// 	} else {
// 		log.Fatalf("wrong type for file store (%s), only vlocal is allowed for now", c.Type)
// 	}
// 	log.Printf("vfiles: %v\n", config)
// }

// Download data to user computer... TODO: use correct name and function
func LoadData(eventID uint32, name string, w io.Writer) error {
	uf, err := model.EventFileGetByEventIDName(eventID, name)
	if err != nil {
		return err
	}
	sf, err := model.StoredFileGetByID(uf.StoredFileID)
	if err != nil {
		return err
	}
	if err = ???.Load(w, sf.Name); err != nil {
		return err
	}
	return nil
}

func DeleteDataByEventIDName(eventID uint32, name string) error {
	ef, err := model.EventFileGetByEventIDName(eventID, name)
	if err != nil {
		return err
	}
	return DeleteDataByEventFile(ef)
}
func DeleteDataByEventFileID(eventFileID uint32) error {
	vlog.Trace.Printf("deleting file eventFileId=%d", eventFileID)
	ef, err := model.EventFileGetByEventFileID(eventFileID)
	if err != nil {
		return err
	}
	return DeleteDataByEventFile(ef)
}
func DeleteDataByEventFile(ef model.EventFile) error {
	sf, err := model.StoredFileGetByID(ef.StoredFileID)
	if err != nil {
		return err
	}
	if err = model.EventFileDeleteByID(ef.ID); err != nil {
		return err
	}
	if err = model.StoredFileDeleteByID(ef.StoredFileID); err != nil {
		return err
	}
	if sf.RefCount <= 1 {
		// nobody else has a reference to this file
		if err = S3Remove(ef.StoredFileID); err != nil {
			return err
		}
	}

	return nil
}

func SaveUploadedFile(eventID uint32, fileName, s3Key, md5 string) error {
	vlog.Trace.Printf("SaveUploadedFile eventID=%d, file=%s, s3key=%s", eventID, fileName, s3Key)
	var size int64 = 0

	sf, err := model.StoredFileCreate(s3Key, size, md5)
	if err != nil {
		return err
	}
	// VG: optimization for later
	// if sf.Name != fileId {
	// 	// the file already existed, its ref_counter was increased, but there is no need to store it again
	// 	vsl.Remove(fileId)
	// }

	// add to event_file
	if err = model.EventFileCreate(eventID, fileName, sf.ID); err != nil {
		model.StoredFileDeleteByID(sf.ID)
		return err
	}

	return nil
}
