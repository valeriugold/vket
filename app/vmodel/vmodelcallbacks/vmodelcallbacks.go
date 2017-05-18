package vmodelcallbacks

import (
	"net/http"

	"github.com/valeriugold/vket/app/vcloud/vs3/vzipfiles"
)

type VModelRecorder interface {
	RecordUploadedFile(eventID, editorID uint32, fileName, s3Key, md5 string) error
	RecordDeleteFile(eventID uint32, name string) error
}

type VModelRecorderDownloaderDeleter interface {
	RecordUploadedFile(eventID, editorID uint32, fileName, s3Key, md5 string) error
	RecordDeleteFile(eventID uint32, name string) error
	DownloadFiles(w http.ResponseWriter, r *http.Request, eventID uint32, areEditedFiles bool, fileEventIDs []uint32, zpr vzipfiles.Zipper) error
	DeleteDataByEventFileID(eventFileID uint32) error
	DeleteDataByEditedFileID(editedFileID uint32) error
}
