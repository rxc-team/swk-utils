package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

var (
	publicPolicyTpl = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "",
				"Effect": "Allow",
				"Principal": "*",
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/%s/*"]
			}
		]
	}`
)

func (svc *Service) publicPolicy() string {
	return fmt.Sprintf(publicPolicyTpl, svc.BucketName, svc.PublicPath)
}

// CreateBucket create bucket with name, returns right away if exists
func createBucket(client *minio.Client, bucketName, region string) (found bool, err error) {
	found, err = client.BucketExists(context.Background(), bucketName)
	if err != nil {
		log.Infof("Checking exist bucket '%s': %v\nCreating Bucket now", bucketName, err)
	}
	if !found {
		// create the bucket if not found
		err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{
			Region:        region,
			ObjectLocking: false,
		})
		if err != nil {
			err = fmt.Errorf("Error creating bucket '%s': %v", bucketName, err)
		}
	}
	return
}
