// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package awscloud

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type ConfigStruct struct {
	Service string
	ID      string
	Region  string
}

func getRegion(region *string) string {
	// user passed region if provided, which takes precedence since it is coming from the user
	if region != nil {
		return *region
	}

	// second check if region is set in environment variable
	regionEnv := os.Getenv("AWS_REGION")
	if regionEnv != "" {
		return regionEnv
	}

	// third check if region is set in environment variable
	regionEnv = os.Getenv("AWS_DEFAULT_REGION")

	if regionEnv != "" {
		return regionEnv
	}

	return "us-est-1" // Fallback to a default region if no region is provided
}

func GetConfig(region *string, cfg_files []string, crd_files []string) (aws.Config, error) {
	// Load the Shared AWS Configuration (~/.aws/config)
	// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/
	cfg, err := config.LoadDefaultConfig(context.TODO())
	// if region is not provided, try to get it from the environment variable or the default region or config region
	if region == nil || *region == "" {
		regionValue := getRegion(&cfg.Region)
		region = &regionValue
	}

	if len(cfg_files) > 0 {
		return config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFiles(cfg_files), config.WithRegion(*region))
	}

	if len(crd_files) > 0 {
		return config.LoadDefaultConfig(context.Background(), config.WithSharedCredentialsFiles(crd_files), config.WithRegion(*region))
	}

	cfg.Region = *region
	return cfg, err
}

// this assumes the parameter is in the format of service::id::region, while region is optional.
func ParseStr(param string) ConfigStruct {
	arr := strings.Split(param, "::")

	if len(arr) > 2 {
		return ConfigStruct{
			Service: arr[0],
			ID:      arr[1],
			Region:  arr[2],
		}
	}

	if len(arr) > 1 {
		return ConfigStruct{
			Service: arr[0],
			ID:      arr[1],
			Region:  "",
		}

	}

	if len(arr) > 0 {
		return ConfigStruct{
			Service: arr[0],
			ID:      "",
			Region:  "",
		}

	}

	return ConfigStruct{
		Service: "",
		ID:      "",
		Region:  "",
	}
}

// get caller identity.
func GetCallerIdentity(cfg aws.Config) (*sts.GetCallerIdentityOutput, error) {
	svc := sts.NewFromConfig(cfg)
	return svc.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})

}

func StringOrMap(s string) interface{} {

	isMap := (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) || (strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))

	if isMap {
		var m map[string]interface{}
		err := json.Unmarshal([]byte(s), &m)
		if err != nil {
			return nil
		}
		return m
	}

	return s

}
