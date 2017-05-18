package vs3

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/valeriugold/vket/app/shared/vlog"
	"github.com/valeriugold/vket/app/vcloud"
	"github.com/valeriugold/vket/app/vcloud/vs3/vsecret"
	"github.com/valeriugold/vket/app/vcloud/vs3/vzipfiles"
)

type vS3Account struct {
	aWSSecretAccessKey string
	aWSRegion          string
	aWSBucket          string
	zip                vzipfiles.Zipper
	svc                *s3.S3
	// sess               *session.Session
}

func New() vcloud.VCloud {
	vs := new(vS3Account)
	vs.aWSSecretAccessKey = vsecret.AWSSecretAccessKey
	vs.aWSRegion = vsecret.AWSRegion
	vs.aWSBucket = vsecret.AWSBucket
	// sess := session.New(&aws.Config{Region: aws.String(vsecret.AWSRegion)})
	// sess := session.NewSession(&aws.Config{Region: aws.String(vsecret.AWSRegion)})
	sess, err := session.NewSession(aws.NewConfig().WithRegion(vsecret.AWSRegion))
	if err != nil {
		vlog.Error.Printf("Err creating session: %v", err)
		os.Exit(1)
	}
	vs.svc = s3.New(sess)

	vs.zip = vzipfiles.NewZipMaker(vsecret.AWSAccessKey, vsecret.AWSSecretAccessKey, vsecret.AWSRegion, vsecret.AWSBucket)
	return vs
}

func (vs *vS3Account) SignPolicy(policy []byte) (base64Policy, s3Signature string, err error) {
	return SignPolicyV4(policy, vs.aWSSecretAccessKey, vs.aWSRegion)
}

func (vs *vS3Account) DeleteFile(key string) error {
	// delete file key
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(vs.aWSBucket), // Required
		Key:    aws.String(key),          // Required
	}
	_, err := vs.svc.DeleteObject(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		return err
	}
	return nil
}

func (vs *vS3Account) GetZipper() vzipfiles.Zipper {
	return vs.zip
}

// func (vs *vS3Account) Zip(w http.ResponseWriter, r *http.Request, nameZip string, files []vzipfiles.S3NameAndObjectID) error {
// 	return vs.zip.Zip(w, r, nameZip, files)
// }

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

// // Download data to user computer... TODO: use correct name and function
// func LoadData(eventID uint32, name string, w io.Writer) error {
// 	uf, err := vmodel.EventFileGetByEventIDName(eventID, name)
// 	if err != nil {
// 		return err
// 	}
// 	sf, err := vmodel.StoredFileGetByID(uf.StoredFileID)
// 	if err != nil {
// 		return err
// 	}
// 	if err = ???.Load(w, sf.Name); err != nil {
// 		return err
// 	}
// 	return nil
// }
