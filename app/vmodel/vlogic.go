package vmodel

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/valeriugold/vket/app/shared/vlog"
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

type vLogic struct{}

func New() vmodelcallbacks.VModelRecorderDownloaderDeleter {
	vr := new(vLogic)
	return vr
}

func (v *vLogic) RecordUploadedFile(eventID uint32, fileName, s3Key, md5 string) error {
	vlog.Trace.Printf("RecordUploadedFile eventID=%d, file=%s, s3key=%s", eventID, fileName, s3Key)
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

	// add to event_file
	if err = EventFileCreate(eventID, fileName, sf.ID); err != nil {
		StoredFileDeleteByID(sf.ID)
		return err
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
		// vs3.DeleteFile(key string) error
		// or
		// nobody else has a reference to this file
		// VG: todo ---> see what happens here
		// if err = S3Remove(ef.StoredFileID); err != nil {
		// 	return err
		// }
	}
	return nil
}

func (v *vLogic) DownloadFiles(w http.ResponseWriter, r *http.Request, userID, eventID uint32, fileEventIDs []uint32, zpr vzipfiles.Zipper) error {
	// check if the event belongs to this authenticated user
	ev, err := EventByEventID(eventID)
	if err != nil {
		vlog.Warning.Printf("Could not find event id %d, err:%v", eventID, err)
		return err
	}
	if ev.UserID != userID {
		vlog.Warning.Printf("event id %d does not belong to user %d", eventID, userID)
		return errors.New("event does not belong to user")
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
		ef, err := EventFileGetByEventFileID(fid)
		if err != nil {
			vlog.Warning.Printf("ef for id=%d err: %v", fid, err)
			continue
		}
		sf, err := StoredFileGetByID(ef.StoredFileID)
		if err != nil {
			vlog.Warning.Printf("sf for id=%d (eventID=%d) err: %v", ef.StoredFileID, fid, err)
			continue
		}
		zp[i].Name = ef.Name
		zp[i].ObjectID = sf.Name
		vlog.Info.Printf("ef===%v, sf===%v", ef, sf)
		vlog.Info.Printf("file EventFile.Name=%s, StoredFile.Name=%s", zp[i].Name, zp[i].ObjectID)
	}
	zipName := fmt.Sprintf("evfiles-%d-%s.zip", userID, ev.Name)
	err = zpr.Zip(w, r, zipName, zp)
	if err != nil {
		vlog.Warning.Printf("on Zip err: %v", err)
	}
	return err
}
