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

func (vs *vS3Account) GeneratePreviewCopy(key string) (newKey, newMd5 string, err error) {
	// VG: generate copy, next version add watermark/transcode
	newKey = key + ".preview"
	params := &s3.CopyObjectInput{
		Bucket:     aws.String(vs.aWSBucket),             // Required
		CopySource: aws.String(vs.aWSBucket + "/" + key), // Required
		Key:        aws.String(newKey),                   // Required
		// ACL:                            aws.String("ObjectCannedACL"),
		// CacheControl:                   aws.String("CacheControl"),
		// ContentDisposition:             aws.String("ContentDisposition"),
		// ContentEncoding:                aws.String("ContentEncoding"),
		// ContentLanguage:                aws.String("ContentLanguage"),
		// ContentType:                    aws.String("ContentType"),
		// CopySourceIfMatch:              aws.String("CopySourceIfMatch"),
		// CopySourceIfModifiedSince:      aws.Time(time.Now()),
		// CopySourceIfNoneMatch:          aws.String("CopySourceIfNoneMatch"),
		// CopySourceIfUnmodifiedSince:    aws.Time(time.Now()),
		// CopySourceSSECustomerAlgorithm: aws.String("CopySourceSSECustomerAlgorithm"),
		// CopySourceSSECustomerKey:       aws.String("CopySourceSSECustomerKey"),
		// CopySourceSSECustomerKeyMD5:    aws.String("CopySourceSSECustomerKeyMD5"),
		// Expires:                        aws.Time(time.Now()),
		// GrantFullControl:               aws.String("GrantFullControl"),
		// GrantRead:                      aws.String("GrantRead"),
		// GrantReadACP:                   aws.String("GrantReadACP"),
		// GrantWriteACP:                  aws.String("GrantWriteACP"),
		// Metadata: map[string]*string{
		// 	"Key": aws.String("MetadataValue"), // Required
		// 	// More values...
		// },
		// MetadataDirective:       aws.String("MetadataDirective"),
		// RequestPayer:            aws.String("RequestPayer"),
		// SSECustomerAlgorithm:    aws.String("SSECustomerAlgorithm"),
		// SSECustomerKey:          aws.String("SSECustomerKey"),
		// SSECustomerKeyMD5:       aws.String("SSECustomerKeyMD5"),
		// SSEKMSKeyId:             aws.String("SSEKMSKeyId"),
		// ServerSideEncryption:    aws.String("ServerSideEncryption"),
		// StorageClass:            aws.String("StorageClass"),
		// Tagging:                 aws.String("TaggingHeader"),
		// TaggingDirective:        aws.String("TaggingDirective"),
		// WebsiteRedirectLocation: aws.String("WebsiteRedirectLocation"),
	}
	out, err := vs.svc.CopyObject(params)
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		vlog.Error.Printf("Err CopyObject(%s): %v", key, err)
		return
	}
	newMd5 = *out.CopyObjectResult.ETag
	return
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
