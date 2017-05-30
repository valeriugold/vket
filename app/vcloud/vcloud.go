package vcloud

import (
	"net/http"

	"github.com/valeriugold/vket/app/vcloud/vs3/vzipfiles"
	"github.com/valeriugold/vket/app/vmodel/vmodelcallbacks"
)

// Uploader is implemented by third parties like fine-uploader to complement their upload to S3
type Uploader interface {
	UploadFileCallbackBefore(r *http.Request, vr vmodelcallbacks.VModelRecorder, vs VCloud) ([]byte, error)
	UploadFileCallbackAfter(r *http.Request, vr vmodelcallbacks.VModelRecorder, vs VCloud) ([]byte, error)
}

// type SaveLoader interface {
// 	Save(r io.Reader, nameHint string) (fileId string, size int64, savedMd5 string, err error)
// 	Load(w io.Writer, fileId string) error
// 	Remove(fileId string) error
// 	DoesExist(fileId string) bool
// }

type VCloud interface {
	SignPolicy(policy []byte) (base64Policy, s3Signature string, err error)
	GetZipper() vzipfiles.Zipper
	DeleteFile(key string) error
	GeneratePreviewCopy(key string) (newKey, newMd5 string, err error)
	// DeleteFile(key interface{}) error
	// // DeleteFileByEventIDName(eventID uint32, name string) error
	// // DeleteFileByEventFileID(eventFileID uint32) error
	// // DeleteFileByEventFile(ef vmodel.EventFile) error
	// GetDownloadLink(key interface{}) (string, error)
}
