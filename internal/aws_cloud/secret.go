// Copyright github.com/fd008 - All rights reserved
// SPDX-License-Identifier: MPL-2.0

package awscloud

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// fetchSecret fetches a secret from AWS Secrets Manager by ID.
func FetchSecret(secretID string, region string) (string, error) {
	// Create a Secrets Manager client
	cfg, err := getConfig(region, nil, nil)
	if err != nil {
		return "", fmt.Errorf("AWS Credentials Error, %v", err)
	}
	svc := secretsmanager.NewFromConfig(cfg)

	// Get the secret value
	result, err := svc.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	})
	if err != nil {
		return "", fmt.Errorf("unable to retrieve secret, %v", err)
	}

	return aws.ToString(result.SecretString), nil
}
