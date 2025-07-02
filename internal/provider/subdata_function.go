// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	awscloud "terraform-provider-awsutils/internal/aws_cloud"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var (
	_ function.Function = &AwsVarFunction{}
)

func NewAwsVarFunction() function.Function {
	return &AwsVarFunction{}
}

type AwsVarFunction struct{}

func (r AwsVarFunction) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "sub_data"
}

func (f *AwsVarFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Parse an JSON string and replace ssm:: and secret:: references if found",
		Description: "Given an JSON string, will parse and return the values of ssm and secret references if found",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "json_file",
				Description: "A JSON file string to parse",
			},
			function.BoolParameter{
				Name:               "strict",
				Description:        "If true, will throw an error if a reference is not found",
				AllowNullValue:     true,
				AllowUnknownValues: true,
			},
		},
		Return: function.StringReturn{},
	}
}

// a function that iterates over each key and value recursively of map[string]interface{} and prints them.
//
//nolint:errcheck
func traverseMap(m map[string]interface{}, strict bool) (map[string]interface{}, error) {
	for k, v := range m {

		switch valType := v.(type) {
		case string:
			fmt.Println(k, v)

			valStr, ok := v.(string)
			if !ok {
				if strict {
					return nil, fmt.Errorf("value for key %s is not a string", k)
				}
			}
			if strings.Contains(valStr, "::") {
				configVal := awscloud.ParseStr(valStr)

				switch configVal.Service {
				case "ssm":
					param, err := awscloud.FetchSSMParameter(configVal.ID, configVal.Region)
					if err != nil || param == "" {
						if strict {
							return nil, fmt.Errorf("SSM parameter not found")
						}
					}

					if param != "" {
						m[k] = param
					}

				case "secret":
					secret, err := awscloud.FetchSecret(configVal.ID, configVal.Region)
					if err != nil || secret == "" {

						if strict {
							return nil, fmt.Errorf("secret not found")
						}
					}

					if secret != "" {
						m[k] = secret
					}

				default:
					fmt.Println(k, v)

				}
			}
		case map[string]interface{}:
			fmt.Println(k)
			vMap, ok := v.(map[string]interface{})

			if !ok {
				if strict {
					return nil, fmt.Errorf("value for key %s is not a map", k)
				}
			}
			//nolint:errcheck
			tMap, terr := traverseMap(vMap, strict)

			if terr != nil || tMap == nil {
				if strict {
					return nil, fmt.Errorf("error traversing map: %v", terr)
				}
			}

		default:
			fmt.Println(k, v, valType)
		}
	}
	return m, nil
}

func (f *AwsVarFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var json_file string
	var strict bool

	resp.Error = req.Arguments.Get(ctx, &json_file, &strict)

	// parse the JSON file into a map.
	var jsonMap map[string]interface{}

	err := json.Unmarshal([]byte(json_file), &jsonMap)

	if err != nil {
		resp.Error = function.NewArgumentFuncError(0, fmt.Sprintf("Error parsing JSON file: %q is not a valid JSON file", json_file))
		return
	}

	// traverse the map and replace ssm:: and secret:: references if found.
	//nolint:errcheck
	_, err = traverseMap(jsonMap, strict)

	if err != nil {
		resp.Error = function.NewArgumentFuncError(0, fmt.Sprintf("Error parsing JSON file: %q", json_file))
		if strict {
			return
		}
	}

	// convert the map back to a JSON string.
	var json_str []byte
	json_str, err = json.Marshal(jsonMap)

	if err != nil {
		resp.Error = function.NewArgumentFuncError(0, fmt.Sprintf("Error converting JSON map to string: %q", json_file))
		return
	}

	resp.Error = resp.Result.Set(ctx, string(json_str))
}
