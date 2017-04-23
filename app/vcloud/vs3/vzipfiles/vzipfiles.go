package vzipfiles

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
)

type S3NameAndObjectID struct {
	Name     string
	ObjectID string
}

type Zipper interface {
	Zip(w http.ResponseWriter, r *http.Request, nameZip string, files []S3NameAndObjectID) error
}

type zipMaker struct {
	aws_bucket *s3.Bucket
}

// Remove all other unrecognised characters apart from
var makeSafeFileName = regexp.MustCompile(`[#<>:"/\|?*\\]`)

// func InitAwsBucket(bucket string) {
func NewZipMaker(awsAccessKey, awsSecretAccessKey, awsRegion, bucket string) Zipper {
	expiration := time.Now().Add(time.Hour * 1)
	auth, err := aws.GetAuth(awsAccessKey, awsSecretAccessKey, "", expiration) //"" = token which isn't needed
	if err != nil {
		panic(err)
	}

	z := new(zipMaker)
	z.aws_bucket = s3.New(auth, aws.GetRegion(awsRegion)).Bucket(bucket)
	return z
}

// Zip creates a zip file named nameZip, containing files from files, using their .Name
func (z *zipMaker) Zip(w http.ResponseWriter, r *http.Request, nameZip string, files []S3NameAndObjectID) error {
	w.Header().Add("Content-Disposition", "attachment; filename=\""+nameZip+"\"")
	w.Header().Add("Content-Type", "application/zip")

	// Loop over files, add them to the
	zipWriter := zip.NewWriter(w)
	for _, file := range files {

		if file.ObjectID == "" {
			log.Printf("Missing path for file: %v", file)
			continue
		}

		// Build safe file file name
		safeFileName := makeSafeFileName.ReplaceAllString(file.Name, "")
		if safeFileName == "" { // Unlikely but just in case
			safeFileName = "file"
		}

		// Read file from S3, log any errors
		rdr, err := z.aws_bucket.GetReader(file.ObjectID)
		if err != nil {
			switch t := err.(type) {
			case *s3.Error:
				if t.StatusCode == 404 {
					log.Printf("File not found. %s", file.ObjectID)
				}
			default:
				log.Printf("Error downloading \"%s\" - %s", file.ObjectID, err.Error())
			}
			continue
		}

		// Build a good path for the file within the zip
		zipPath := safeFileName
		// zipPath := ""
		// // Prefix project Id and name, if any (remove if you don't need)
		// if file.ProjectId > 0 {
		// 	zipPath += strconv.FormatInt(file.ProjectId, 10) + "."
		// 	// Build Safe Project Name
		// 	file.ProjectName = makeSafeFileName.ReplaceAllString(file.ProjectName, "")
		// 	if file.ProjectName == "" { // Unlikely but just in case
		// 		file.ProjectName = "Project"
		// 	}
		// 	zipPath += file.ProjectName + "/"
		// }
		// // Prefix folder name, if any
		// if file.Folder != "" {
		// 	zipPath += file.Folder
		// 	if !strings.HasSuffix(zipPath, "/") {
		// 		zipPath += "/"
		// 	}
		// }
		// zipPath += safeFileName

		// We have to set a special flag so zip files recognize utf file names
		// See http://stackoverflow.com/questions/30026083/creating-a-zip-archive-with-unicode-filenames-using-gos-archive-zip
		h := &zip.FileHeader{
			Name:   zipPath,
			Method: zip.Deflate,
			Flags:  0x800,
		}

		// if file.Modified != "" {
		// 	h.SetModTime(file.ModifiedTime)
		// }

		f, _ := zipWriter.CreateHeader(h)

		io.Copy(f, rdr)
		rdr.Close()
	}

	zipWriter.Close()

	// log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
	return nil
}
