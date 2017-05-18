package vmodel

import (
	"fmt"
	"net/http"

	"github.com/valeriugold/vket/app/shared/vlog"
	"github.com/valeriugold/vket/app/vcloud"
	"github.com/valeriugold/vket/app/vcloud/vs3/vzipfiles"
	"github.com/valeriugold/vket/app/vmodel/vmodelcallbacks"
)

// type VModelRecorder interface {
// 	RecordUploadedFile(eventID uint32, fileName, s3Key, md5 string) error
// 	RecordDeleteFile(eventID uint32, name string) error
// }

// type VModelRecorderDownloaderDeleter interface {
// 	RecordUploadedFile(eventID uint32, fileName, s3Key, md5 string) error
// 	RecordDeleteFile(eventID uint32, name string) error
// 	DownloadFiles(w http.ResponseWriter, r *http.Request, userID, eventID uint32, fileEventIDs []uint32, zpr vzipfiles.Zipper) error
//	DeleteDataByEventFileID(eventFileID uint32) error
// }

type vLogic struct {
	vc vcloud.VCloud
}

func New(vc vcloud.VCloud) vmodelcallbacks.VModelRecorderDownloaderDeleter {
	vr := new(vLogic)
	vr.vc = vc
	return vr
}

func (v *vLogic) RecordUploadedFile(eventID, editorID uint32, fileName, s3Key, md5 string) error {
	vlog.Trace.Printf("RecordUploadedFile eventID=%d, editorID=%d, file=%s, s3key=%s", eventID, editorID, fileName, s3Key)
	var size int64 = 0

	sf, err := StoredFileCreate(s3Key, size, md5)
	if err != nil {
		return err
	}
	// VG: optimization for later
	// if sf.Name != fileId {
	// 	// the file already existed, its ref_counter was increased, but there is no need to store it again
	// 	vsl.Remove(fileId)
	// }

	if editorID == 0 {
		// this is an original file, add it to event_file
		if err = EventFileCreate(eventID, fileName, sf.ID); err != nil {
			StoredFileDeleteByID(sf.ID)
			return err
		}
	} else {
		// this is an edited file, add it to edited_file
		if err = EditedFileCreate(eventID, editorID, fileName, sf.ID); err != nil {
			StoredFileDeleteByID(sf.ID)
			return err
		}
	}
	return nil
}

func (v *vLogic) RecordDeleteFile(eventID uint32, name string) error {
	ef, err := EventFileGetByEventIDName(eventID, name)
	if err != nil {
		return err
	}
	return v.deleteDataByEventFile(ef)
}

func (v *vLogic) DeleteDataByEventFileID(eventFileID uint32) error {
	vlog.Trace.Printf("deleting file eventFileId=%d", eventFileID)
	ef, err := EventFileGetByEventFileID(eventFileID)
	if err != nil {
		return err
	}
	// vlog.Trace.Printf("delete evFile=%v", ef)
	return v.deleteDataByEventFile(ef)
}

func (v *vLogic) deleteDataByEventFile(ef EventFile) error {
	sf, err := StoredFileGetByID(ef.StoredFileID)
	if err != nil {
		return err
	}
	if err = EventFileDeleteByID(ef.ID); err != nil {
		return err
	}
	if err = StoredFileDeleteByID(ef.StoredFileID); err != nil {
		return err
	}
	if sf.RefCount <= 1 {
		vlog.Info.Printf("deleting file name: %s", sf.Name)
		err = v.vc.DeleteFile(sf.Name)
		// or
		// nobody else has a reference to this file
		// VG: todo ---> see what happens here
		// if err = S3Remove(ef.StoredFileID); err != nil {
		// 	return err
		// }
	}
	return err
}

func (v *vLogic) DeleteDataByEditedFileID(editedFileID uint32) error {
	vlog.Trace.Printf("deleting file editedFileId=%d", editedFileID)
	df, err := EditedFileGetByEditedFileID(editedFileID)
	if err != nil {
		return err
	}
	// vlog.Trace.Printf("delete evFile=%v", ef)
	return v.deleteDataByEditedFile(df)
}

func (v *vLogic) deleteDataByEditedFile(df EditedFile) error {
	sf, err := StoredFileGetByID(df.StoredFileID)
	if err != nil {
		return err
	}
	if err = EditedFileDeleteByID(df.ID); err != nil {
		return err
	}
	if err = StoredFileDeleteByID(df.StoredFileID); err != nil {
		return err
	}
	if sf.RefCount <= 1 {
		vlog.Info.Printf("deleting file name: %s", sf.Name)
		err = v.vc.DeleteFile(sf.Name)
		// or
		// nobody else has a reference to this file
		// VG: todo ---> see what happens here
		// if err = S3Remove(ef.StoredFileID); err != nil {
		// 	return err
		// }
	}
	return err
}

func (v *vLogic) DownloadFiles(w http.ResponseWriter, r *http.Request, eventID uint32, areEditedFiles bool, fileEventIDs []uint32, zpr vzipfiles.Zipper) error {
	// check if the event belongs to this authenticated user
	ev, err := EventGetByEventID(eventID)
	if err != nil {
		vlog.Warning.Printf("Could not find event id %d, err:%v", eventID, err)
		return err
	}
	vlog.Info.Printf("Download files id %v", fileEventIDs)
	// get all files: user_names and file_names
	zp := make([]vzipfiles.S3NameAndObjectID, len(fileEventIDs))
	for i, fid := range fileEventIDs {
		// fid, err := stringToUint32(f)
		// if err != nil {
		// 	vlog.Warning.Printf("event file ID %s is not integer, err=%v", f, err)
		// 	continue
		// }
		var sfID uint32
		var name string
		if areEditedFiles {
			// these are edited files
			df, err := EditedFileGetByEditedFileID(fid)
			if err != nil {
				vlog.Warning.Printf("df for id=%d err: %v", fid, err)
				continue
			}
			sfID = df.StoredFileID
			name = df.Name
			vlog.Info.Printf("df===%v", df)
		} else {
			// these are original files
			ef, err := EventFileGetByEventFileID(fid)
			if err != nil {
				vlog.Warning.Printf("ef for id=%d err: %v", fid, err)
				continue
			}
			sfID = ef.StoredFileID
			name = ef.Name
			vlog.Info.Printf("ef===%v", ef)
		}
		sf, err := StoredFileGetByID(sfID)
		if err != nil {
			vlog.Warning.Printf("sf for id=%d (eventID=%d) err: %v", sfID, fid, err)
			continue
		}
		zp[i].Name = name
		zp[i].ObjectID = sf.Name
		vlog.Info.Printf("sf===%v", sf)
		vlog.Info.Printf("file EventFile.Name=%s, StoredFile.Name=%s", zp[i].Name, zp[i].ObjectID)
	}
	zipName := fmt.Sprintf("evfiles-%d-%s.zip", ev.UserID, ev.Name)
	err = zpr.Zip(w, r, zipName, zp)
	if err != nil {
		vlog.Warning.Printf("on Zip err: %v", err)
	}
	return err
}
