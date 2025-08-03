// Copyright (c) HashiCorp, Inc.

package awscloud

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// GetS3BucketPolicy retrieves the bucket policy for a given S3 bucket name.
func GetS3BucketPolicy(bucketName string, region *string) (*string, error) {
	// Load the AWS configuration
	cfg, err := GetConfig(region, nil, nil)
	if err != nil {
		return nil, err
	}

	// Create a new S3 client
	client := s3.NewFromConfig(cfg)

	// Build the GetBucketPolicy input
	input := &s3.GetBucketPolicyInput{
		Bucket: &bucketName,
	}

	// Call the GetBucketPolicy operation
	resp, err := client.GetBucketPolicy(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 bucket policy for %s: %w", bucketName, err)
	}

	// Return the bucket policy document as a string
	return resp.Policy, nil
}
