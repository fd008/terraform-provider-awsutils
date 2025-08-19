// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package awscloud

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

func BucketExists(bucketName string, region *string) bool {
	// Load the AWS configuration
	cfg, err := GetConfig(region, nil, nil)
	if err != nil {
		return false
	}
	// Create a new S3 client
	client := s3.NewFromConfig(cfg)

	_, err = client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		// var nf *types.NotFound // Specific error type for bucket not found
		if _, ok := err.(*types.NotFound); ok {
			fmt.Printf("Bucket '%s' does not exist.\n", bucketName)
			return false
		} else {
			// Handle other potential errors (e.g., permissions, network issues)
			fmt.Printf("Error checking bucket '%s': %v\n", bucketName, err)
			return false
		}
	}

	return true
}

func DeleteObjects(cfg aws.Config, bucket string, objects []string) error {
	client := s3.NewFromConfig(cfg)

	// Prepare the list of objects to delete
	var objectsToDelete []types.ObjectIdentifier
	for _, obj := range objects {
		objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
			Key: aws.String(obj),
		})
	}

	// Create the DeleteObjects input
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: objectsToDelete,
			Quiet:   aws.Bool(true), // Set to true to suppress the response for each deleted object
		},
	}

	// Call the DeleteObjects operation
	_, err := client.DeleteObjects(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to delete objects from bucket %s: %w", bucket, err)
	}

	return nil
}

// DeleteObjectsWithPrefix deletes all S3 objects within a specified bucket that share a common prefix.
func DeleteObjectsWithPrefix(cfg aws.Config, bucketName string, prefix string, region *string) error {

	if region != nil {
		cfg.Region = *region
	}

	s3Client := s3.NewFromConfig(cfg)

	// 1. List Objects with the Prefix
	var objectsToDelete []types.ObjectIdentifier
	var continuationToken *string

	log.Printf("Listing objects in bucket %q with prefix %q...", bucketName, prefix)
	for {
		listObjectsInput := &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucketName),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		}

		resp, err := s3Client.ListObjectsV2(context.TODO(), listObjectsInput)
		if err != nil {
			return fmt.Errorf("failed to list objects in bucket %q with prefix %q: %w", bucketName, prefix, err)
		}

		for _, obj := range resp.Contents {
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{Key: obj.Key})
		}

		if !aws.ToBool(resp.IsTruncated) {
			break // All objects have been listed
		}
		continuationToken = resp.NextContinuationToken
	}

	if len(objectsToDelete) == 0 {
		log.Printf("No objects found to delete in bucket %q with prefix %q.", bucketName, prefix)
		return nil
	}

	log.Printf("Found %d objects to delete in bucket %q with prefix %q.", len(objectsToDelete), bucketName, prefix)

	// 2. Delete Objects in Batches (up to 1000 per request)
	const maxObjectsPerDelete = 1000
	for i := 0; i < len(objectsToDelete); i += maxObjectsPerDelete {
		end := i + maxObjectsPerDelete
		if end > len(objectsToDelete) {
			end = len(objectsToDelete)
		}

		batch := objectsToDelete[i:end]

		deleteInput := &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: batch,
				Quiet:   aws.Bool(true), // Set to true to suppress detailed results of each object deletion
			},
		}

		deleteResp, err := s3Client.DeleteObjects(context.TODO(), deleteInput)
		if err != nil {
			return fmt.Errorf("failed to delete batch of objects from bucket %q with prefix %q: %w", bucketName, prefix, err)
		}

		if len(deleteResp.Errors) > 0 {
			for _, deleteErr := range deleteResp.Errors {
				log.Printf("Error deleting object %q (Code: %s, Message: %s)",
					aws.ToString(deleteErr.Key), aws.ToString(deleteErr.Code), aws.ToString(deleteErr.Message))
			}
			return fmt.Errorf("encountered errors during batch deletion from bucket %q with prefix %q", bucketName, prefix)
		}
		log.Printf("Deleted %d objects from bucket %q with prefix %q.", len(batch), bucketName, prefix)
	}

	log.Printf("Successfully deleted all objects in bucket %q with prefix %q.", bucketName, prefix)
	return nil
}
