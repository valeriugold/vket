package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// package vcloud

// func GetCloudFileInfo() (FileInfo, error) {
func GetCloudFileInfo() {
	sess := session.Must(session.NewSession())

	svc := s3.New(sess)
	// svc := s3.New(sess, &aws.Config{Region: aws.String("us-east-1")})

	params := &s3.ListObjectsInput{
		Bucket: aws.String("vket"), // Required
		// Delimiter:    aws.String("Delimiter"),
		// EncodingType: aws.String("EncodingType"),
		// Marker:       aws.String("Marker"),
		MaxKeys: aws.Int64(10),
		Prefix:  aws.String("user/vg/0-2-f627ab2d-62e0-4d18-9edf-4bf6420dbdc6--IMG_20160821_165110290.jpg"),
		// Prefix:  aws.String("user/vg/"),
		// RequestPayer: aws.String("RequestPayer"),
	}
	resp, err := svc.ListObjects(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}

func main() {
	GetCloudFileInfo()
}

// valmac:s3 valeriug$ ./vcloud
// {
//   Contents: [{
//       ETag: "\"4a895d9fe96fa4fc7f49458259eb0e88\"",
//       Key: "user/vg/0-2-f627ab2d-62e0-4d18-9edf-4bf6420dbdc6--IMG_20160821_165110290.jpg",
//       LastModified: 2017-03-27 01:15:09 +0000 UTC,
//       Owner: {
//         DisplayName: "valeriug",
//         ID: "4d6f081b2fbdc839ceca1b49347556e000233bee2c24b436e2568d9e5d40b120"
//       },
//       Size: 1134885,
//       StorageClass: "STANDARD"
//     }],
//   IsTruncated: false,
//   Marker: "",
//   MaxKeys: 10,
//   Name: "vket",
//   Prefix: "user/vg/0-2-f627ab2d-62e0-4d18-9edf-4bf6420dbdc6--IMG_20160821_165110290.jpg"
// }
