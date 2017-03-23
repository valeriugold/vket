package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// const policy = `{ "expiration": "2017-12-01T12:00:00.000Z",
//   "conditions": [
//     {"bucket": "vket"},
//     ["starts-with", "$key", "user/vg/"],
//     ["starts-with", "$Content-Type", "multipart/"],
//     {"acl": "private"},
//     {"success_action_redirect": "http://localhost:9090/hello"},
//     {"x-amz-credential": "AKIAI3WUHSIINGF2M3RQ/20170311/us-east-1/s3/aws4_request"},
//     {"x-amz-algorithm": "AWS4-HMAC-SHA256"},
//     {"x-amz-date": "20170311T000000Z"}
//   ]
// }`

func getSignatureKey(secret string, dateStamp string, regionName string, serviceName string) []byte {
	dateKey := hmacSHA256([]byte(secret), []byte(dateStamp))
	dateRegionKey := hmacSHA256(dateKey, []byte(regionName))
	dateRegionServiceKey := hmacSHA256(dateRegionKey, []byte(serviceName))
	signingKey := hmacSHA256(dateRegionServiceKey, []byte("aws4_request"))
	return signingKey
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func getDateStampFromPolicy(policy []byte) (dateStamp string, err error) {
	dateStamp = ""
	st := struct {
		Conditions []interface{} `json:"conditions"`
	}{}
	if err = json.Unmarshal(policy, &st); err != nil {
		return
	}
	// fmt.Println("pol: ", policy)
	// fmt.Println("st: ", st)

	var dateTime string
	for _, i := range st.Conditions {
		// fmt.Println("i: ", i)
		if m, ok := i.(map[string]interface{}); ok {
			// fmt.Println("    type---ok, m=", m)
			if val, ok := m["x-amz-date"]; ok {
				if dateTime, ok = val.(string); ok {
					// fmt.Println("           all is ok, u=", dateTime)
					dateStamp = dateTime[0:8]
					// fmt.Printf("dateStamp=%s\n", dateStamp)
					return
				}
			}
		}
	}

	err = errors.New("could not find my conditions->x-amz-date")
	return
}

// GetSignedPolicy signes the policy with V4
func GetSignedPolicy(region string, awsSecretAccessKey string, policy []byte) (base64Policy, s3Signature string, err error) {
	// func GetSignedPolicy(region string, awsSecretAccessKey string, p io.Reader) (base64Policy, s3Signature string, err error) {
	service := "s3"
	base64Policy = ""
	s3Signature = ""

	// get "x-amz-date" from the policy, it is needed to sign the policy with signature V4
	dateStamp, err := getDateStampFromPolicy(policy)
	if err != nil {
		return
	}

	base64Policy = base64.StdEncoding.EncodeToString(policy)
	signingKey := getSignatureKey("AWS4"+awsSecretAccessKey, dateStamp, region, service)
	s3sigByte := hmacSHA256(signingKey, []byte(base64Policy))
	s3Signature = fmt.Sprintf("%x", s3sigByte)
	fmt.Printf("base64Policy : %s\n", base64Policy)
	fmt.Printf("S3 Signature: %s\n", s3Signature)
	return
}

// func main() {
// 	base64Policy, s3Signature, err := GetSignedPolicy("us-east-1", AWSSecretAccessKey, policy)
// 	if err != nil {
// 		fmt.Printf("err=%v", err)
// 		return
// 	}
// 	fmt.Printf("base64Policy : %s\n", base64Policy)
// 	fmt.Printf("S3 Signature: %x\n", s3Signature)
// }
//      <input type="hidden" name="AWSAccessKeyId" value="AKIAI3WUHSIINGF2M3RQ">

// http://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-UsingHTTPPOST.html
// http://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-HTTPPOSTForms.html
// https://play.golang.org/p/7PItbEVN2r

// {"expiration":"2017-03-19T02:39:41.617Z",
// "conditions":[{"acl":"private"},
//                {"bucket":"vket"},
//                {"Content-Type":"image/jpeg"},
//                {"success_action_status":"200"},
//                {"x-amz-algorithm":"AWS4-HMAC-SHA256"},
//                {"key":"85663d88-77dc-41b5-8446-3ec026471f42.jpg"},
//                {"x-amz-credential":"AKIAJB6BSMFWTAXC5M2Q/20170319/us-east-1/s3/aws4_request"},
//                {"x-amz-date":"20170319T023441Z"},
//                {"x-amz-meta-qqfilename":"0-weu-d3-f72f2f8f39efd5db4d0482818bbb5973.jpg"},
//                ["content-length-range","0","15000000"]]}
