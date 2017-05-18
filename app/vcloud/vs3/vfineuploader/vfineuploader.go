package vfineuploader

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/valeriugold/vket/app/shared/vlog"
	"github.com/valeriugold/vket/app/vcloud"
	"github.com/valeriugold/vket/app/vmodel/vmodelcallbacks"
)

type uploader struct{}

func New() vcloud.Uploader {
	n := new(uploader)
	return n
}

func (u *uploader) UploadFileCallbackBefore(r *http.Request, vr vmodelcallbacks.VModelRecorder, vs vcloud.VCloud) (b []byte, err error) {
	if r.Method != "POST" {
		return b, errors.New("UploadFileCallbackBefore is not POST, but " + r.Method)
	}
	policyBuf := new(bytes.Buffer)
	_, err = policyBuf.ReadFrom(r.Body)
	if err != nil {
		return
	}
	base64Policy, s3Signature, err := vs.SignPolicy(policyBuf.Bytes())
	if err != nil {
		return
	}
	vlog.Trace.Printf("base64Policy=%s\n", base64Policy)
	vlog.Trace.Printf("s3Signature=%s\n", s3Signature)
	resp := struct {
		Policy    string `json:"policy"`
		Signature string `json:"signature"`
	}{Policy: base64Policy, Signature: s3Signature}
	b, err = json.Marshal(resp)
	if err != nil {
		return
	}
	return
}

func (u *uploader) UploadFileCallbackAfter(r *http.Request, vr vmodelcallbacks.VModelRecorder, vs vcloud.VCloud) (b []byte, err error) {
	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))
	// if r.Method != "POST" {
	// 	vlog.Warning.Printf("Method is not POST, but %s", r.Method)
	// 	// VG: show error page
	// 	vviews.Error(w, "Method is not POST, but "+r.Method)
	// 	return
	// }
	key := r.FormValue("key")
	// bucket := r.FormValue("bucket")
	name := r.FormValue("name")
	md5 := r.FormValue("etag")
	eventID := r.FormValue("eventID")
	editorID := r.FormValue("editorID")
	vlog.Trace.Printf("converting ev=%v", eventID)
	eid64, err := strconv.ParseUint(eventID, 10, 32)
	if err != nil {
		vlog.Warning.Printf("stringToUint32 s=%s, err=%v", eventID, err)
		return
	}
	eid := uint32(eid64)
	edID := uint32(0)
	if len(editorID) > 0 {
		vlog.Trace.Printf("converting editorID=%v", editorID)
		edID64, err64 := strconv.ParseUint(editorID, 10, 32)
		if err64 != nil {
			err = err64
			vlog.Warning.Printf("stringToUint32 s=%s, err=%v", editorID, err)
			return
		}
		edID = uint32(edID64)
	}
	// isBrowserPreviewCapable=true&key=user%2Fvg%2F4-4-3f525c61-ea46-4ae1-9ca1-fdf80cfa0839--IMG_20160821_165653814_HDR.jpg&
	// 	uuid=3f525c61-ea46-4ae1-9ca1-fdf80cfa0839&
	// 	name=IMG_20160821_165653814_HDR.jpg&
	// 	bucket=vket&
	// 	etag=%2201791e79c9ced416ad2b11c0979931ef%22

	vlog.Trace.Printf("calling vr.RecordUploadedFile k=%s, evid=%s eid=%d, md5=%s\n", key, eventID, eid, md5[1:33])
	if err = vr.RecordUploadedFile(eid, edID, name, key, md5[1:33]); err != nil {
		// vlog.Warning.Printf("err on SaveMultipart, err:%v", err)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vlog.Trace.Printf("success!\n")
	return
}
