// Copyright (c) HashiCorp, Inc.

package awscloud

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

// GetKMSKeyPolicy retrieves the default key policy for a given KMS key ID.
func GetKMSKeyPolicy(kmsKeyID string, region *string) (*string, error) {
	cfg, err := GetConfig(region, nil, nil)
	if err != nil {
		return nil, err
	}

	// Create a new KMS client
	client := kms.NewFromConfig(cfg)

	// Build the GetKeyPolicy input
	input := &kms.GetKeyPolicyInput{
		KeyId:      &kmsKeyID,
		PolicyName: aws.String("default"), // The only valid policy name is "default"
	}

	// Call the GetKeyPolicy operation
	resp, err := client.GetKeyPolicy(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get KMS key policy for %s: %w", kmsKeyID, err)
	}

	// Return the key policy document as a string
	return resp.Policy, nil
}
