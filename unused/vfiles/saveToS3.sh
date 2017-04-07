#!/bin/bash

# pass: f@r@n1c!Un$3f
aws s3 ls --recursive --human-readable
aws s3 cp test.txt s3://valgbak/bak/poze/


# post
# http://docs.aws.amazon.com/AmazonS3/latest/dev/HTTPPOSTExamples.html

# multiple files upload
# https://fineuploader.com/
# https://github.com/blueimp/jQuery-File-Upload
