// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	awscloud "terraform-provider-awsutils/internal/aws_cloud"

	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
)

var (
	_ function.Function = &MergePolicy{}
)

func MergePolicyFunction() function.Function {
	return &MergePolicy{}
}

type MergePolicy struct{}

func (r MergePolicy) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "merge_policy"
}

func (f *MergePolicy) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Return a list of files/folders given a directory path or empty list if there are any errors",
		Description: "Given a directory path, will return a list of (one level only) files and folders in that directory. Returns an empty list if the directory does not exist or is not accessible.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "existing_policy",
				Description: "Existing policy in string format",
			},
			function.StringParameter{
				Name:        "new_policy",
				Description: "New policy in string format to merge with the existing policy",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *MergePolicy) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var existingPolicy string
	var newPolicy string

	resp.Error = req.Arguments.Get(ctx, &existingPolicy, &newPolicy)
	if resp.Error != nil {
		return
	}

	var policy1, policy2 awscloud.Policy

	err := json.Unmarshal([]byte(existingPolicy), &policy1)
	if err != nil {
		resp.Error = function.FuncErrorFromDiags(ctx, diag.Diagnostics{
			diag.NewErrorDiagnostic("Policy Unmarshal Error", fmt.Sprintf("Error unmarshalling policy 1: %s", err.Error())),
		})
		return
	}
	err = json.Unmarshal([]byte(newPolicy), &policy2)
	if err != nil {
		resp.Error = function.FuncErrorFromDiags(ctx, diag.Diagnostics{
			diag.NewErrorDiagnostic("Policy Unmarshal Error", fmt.Sprintf("Error unmarshalling policy 2: %s", err.Error())),
		})
		return
	}

	mergedPolicy := awscloud.MergePolicies(policy1, policy2)

	mergedJSON, err := json.Marshal(mergedPolicy)
	if err != nil {
		resp.Error = function.FuncErrorFromDiags(ctx, diag.Diagnostics{
			diag.NewErrorDiagnostic("Marshalling Error", fmt.Sprintf("Error marshalling merged policy: %s", err.Error())),
		})
		return
	}

	if string(mergedJSON) == "" {
		resp.Error = function.FuncErrorFromDiags(ctx, diag.Diagnostics{
			diag.NewErrorDiagnostic("Merge Policy Error", "Failed to merge policies! Check existing_policy and statements arguments."),
		})
		return
	}

	resp.Error = resp.Result.Set(ctx, string(mergedJSON))

}
