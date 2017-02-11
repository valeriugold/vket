package vlocal

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type VFilesLocal struct {
	dir string
}

var config Configuration

type Configuration struct {
	DestDir string `json:"destDir"`
}

func InitConfiguration(c Configuration) *VFilesLocal {
	config = c
	if len(c.DestDir) < 2 {
		log.Fatalln("invalid DestDir=" + c.DestDir)
	}
	// check if the dir exists or needs to be created
	var vbox = VFilesLocal{c.DestDir}
	vbox.Init()
	return &vbox
}

func (x VFilesLocal) Init() {
	if err := os.MkdirAll(x.dir, 0777); err != nil {
		log.Fatalln("could not create dir " + x.dir + " err:" + err.Error())
	}
}

// implementing the vfiles.SaveLoader interface
// Save saves the file nameLocal to dir/nameBox
func (x VFilesLocal) Save(r io.Reader, nameHint string) (fileId string, size int64, savedMd5 string, err error) {
	// prepare hash
	hash := md5.New()
	f := io.TeeReader(r, hash)

	fstore := filepath.Join(x.dir, nameHint)
	out, err := os.Create(fstore)
	if err != nil {
		return nameHint, 0, "", err
	}
	size, err = io.Copy(out, f)
	out.Close()

	savedMd5 = fmt.Sprintf("%x", hash.Sum(nil))
	return nameHint, size, savedMd5, err
}

// Load copies the file dir/nameBox to nameLocal
func (x VFilesLocal) Load(w io.Writer, fileId string) error {
	fstore := filepath.Join(x.dir, fileId)
	r, err := os.Open(fstore)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	r.Close()
	return err
}

// Remove file from dir/
func (x VFilesLocal) Remove(fileId string) error {
	fstore := filepath.Join(x.dir, fileId)
	return os.Remove(fstore)
}

// DoesExist returns true if file exists in dir/
func (x VFilesLocal) DoesExist(fileId string) bool {
	fstore := filepath.Join(x.dir, fileId)
	if _, err := os.Stat(fstore); os.IsNotExist(err) {
		return false
	}
	return true
}

// // CopyFile copies a file from src to dst. If src and dst files exist, and are
// // the same, then return success. Otherise, attempt to create a hard link
// // between the two files. If that fail, copy the file contents from src to dst.
// func CopyFile(src, dst string) (err error) {
// 	sfi, err := os.Stat(src)
// 	if err != nil {
// 		// no file to copy
// 		return errors.New("there is no source " + src)
// 	}
// 	if !sfi.Mode().IsRegular() {
// 		// cannot copy non-regular files (e.g., directories,
// 		// symlinks, devices, etc.)
// 		return errors.New("CopyFile: non-regular source file " + sfi.Name() + " (" + sfi.Mode().String() + ")")
// 	}
// 	dfi, err := os.Stat(dst)
// 	if err != nil {
// 		if !os.IsNotExist(err) {
// 			return errors.New("destination " + dst + " returned error " + err.Error())
// 		}
// 	} else {
// 		if !(dfi.Mode().IsRegular()) {
// 			return fmt.Errorf("CopyFile: non-regular destination file " + dfi.Name() + " (" + dfi.Mode().String() + ")")
// 		}
// 		if os.SameFile(sfi, dfi) {
// 			return
// 		}
// 	}
// 	if err = os.Link(src, dst); err == nil {
// 		return
// 	}
// 	return copyFileContents(src, dst)
// }

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
// func copyFileContents(src, dst string) (err error) {
// 	in, err := os.Open(src)
// 	if err != nil {
// 		return
// 	}
// 	defer in.Close()
// 	out, err := os.Create(dst)
// 	if err != nil {
// 		return
// 	}
// 	defer func() {
// 		cerr := out.Close()
// 		if err == nil {
// 			err = cerr
// 		}
// 	}()
// 	if _, err = io.Copy(out, in); err != nil {
// 		return
// 	}
// 	err = out.Sync()
// 	return
// }
