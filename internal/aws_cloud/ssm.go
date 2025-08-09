// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package awscloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func getParameter(param string, region *string) (*ssm.GetParameterOutput, error) {
	cfg, err := GetConfig(region, nil, nil)
	if err != nil {
		return nil, err
	}

	svc := ssm.NewFromConfig(cfg)
	input := &ssm.GetParameterInput{
		Name:           aws.String(param),
		WithDecryption: aws.Bool(true),
	}

	return svc.GetParameter(context.TODO(), input)

}

// fetchSSMParameter fetches a parameter from AWS SSM Parameter Store by name.
func FetchSSMParameter(paramName string, region *string) (interface{}, error) {
	result, err := getParameter(paramName, region)

	if err != nil {
		return "", fmt.Errorf("unable to retrieve parameter, %v", err)
	}

	// if the parameter is a StringList, split it by comma and return it as a slice.
	if result.Parameter.Type == "StringList" {
		return strings.Split(aws.ToString(result.Parameter.Value), ","), nil
	}

	return StringOrMap(aws.ToString(result.Parameter.Value)), nil
}
