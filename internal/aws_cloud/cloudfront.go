// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package awscloud

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

// given a distribution ID and paths, invalidate the cache
func InvalidateCache(cfg aws.Config, distributionID string, paths []string) (*cloudfront.CreateInvalidationOutput, error) {

	svc := cloudfront.NewFromConfig(cfg)

	// create the input
	input := &cloudfront.CreateInvalidationInput{
		DistributionId: &distributionID,
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String(time.Now().Format(time.RFC3339)),
			Paths: &types.Paths{
				Quantity: aws.Int32(int32(len(paths))),
				Items:    paths,
			},
		},
	}

	// create the invalidation
	return svc.CreateInvalidation(context.TODO(), input)
}

// get the dsitribtion by ID
func GetDistribution(cfg aws.Config, distributionID string) (*cloudfront.GetDistributionOutput, error) {

	svc := cloudfront.NewFromConfig(cfg)

	// create the input
	input := &cloudfront.GetDistributionInput{
		Id: &distributionID,
	}

	// get the distribution
	return svc.GetDistribution(context.TODO(), input)
}
