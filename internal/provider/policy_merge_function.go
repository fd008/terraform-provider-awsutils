// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var (
	_ function.Function = &MergePolicy{}
)

func MergePolicyFunction() function.Function {
	return &MergePolicy{}
}

type MergePolicy struct{}

// Policy represents the top-level IAM policy document.
type Policy struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

// Statement represents a single policy statement within the policy document.
type Statement struct {
	Effect    string                  `json:"Effect"`
	Principal Principal               `json:"Principal,omitempty"`
	Action    Action                  `json:"Action"`
	Resource  Resource                `json:"Resource"`
	Condition map[string]ConditionMap `json:"Condition,omitempty"`
}

// Principal can be a string or a map for different types (AWS, Service, etc.).
type Principal struct {
	AWS     interface{} `json:"AWS,omitempty"`     // Can be string or []string
	Service interface{} `json:"Service,omitempty"` // Can be string or []string
	// Add other principal types as needed, e.g., CanonicalUser, Federated, etc.
}

// Action can be a string or an array of strings.
type Action interface{}

// Resource can be a string or an array of strings.
type Resource interface{}

// ConditionMap represents the key-value pairs within a condition operator.
type ConditionMap map[string]interface{}

func MergePolicyStatements(existingPolicy string, newPolicy string) string {

	fmt.Println("Merging policies...")

	var policy Policy
	err := json.Unmarshal([]byte(existingPolicy), &policy)
	if err != nil {
		fmt.Printf("Error unmarshalling JSON: %v\n", err)
		return ""
	}

	var statements []Statement
	err = json.Unmarshal([]byte(newPolicy), &statements)
	if err != nil {
		fmt.Printf("Error unmarshalling new policy: %v\n", err)
		return ""
	}

	mergedPolicy := Policy{
		Version:   policy.Version,
		Statement: append(policy.Statement, statements...),
	}

	mergedPolicyJSON, err := json.MarshalIndent(mergedPolicy, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling merged policy: %v\n", err)
		return ""
	}

	fmt.Printf("Successfully merged policies:\n%s\n", string(mergedPolicyJSON))

	return string(mergedPolicyJSON)
}

func (r MergePolicy) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "merge_policy_statements"
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
				Name:        "statements",
				Description: "Statements in string format to merge with the existing policy",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *MergePolicy) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var existingPolicy string
	var statements string

	resp.Error = req.Arguments.Get(ctx, &existingPolicy, &statements)
	if resp.Error != nil {
		return
	}

	mergedPolicy := MergePolicyStatements(existingPolicy, statements)

	if mergedPolicy == "" {
		resp.Error = function.NewArgumentFuncError(0, "Failed to merge policies! Check existing_policy and statements arguments.")
		return
	}

	resp.Error = resp.Result.Set(ctx, mergedPolicy)

}
