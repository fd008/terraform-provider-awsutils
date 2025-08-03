// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ function.Function = &ShallowList{}
)

func ShallowListFunction() function.Function {
	return &ShallowList{}
}

type ShallowList struct{}

func (r ShallowList) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "shallow_list"
}

func (f *ShallowList) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Return a list of files/folders given a directory path or empty list if there are any errors",
		Description: "Given a directory path, will return a list of (one level only) files and folders in that directory. Returns an empty list if the directory does not exist or is not accessible.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "path",
				Description: "A Directory Path",
			},
		},
		Return: function.ListReturn{
			ElementType: types.StringType,
		},
	}
}

func ListDirectoryContents(path string) []string {
	entries, err := os.ReadDir(path) // Read directory entries
	if err != nil {
		return []string{} // Return an empty list if there's an error
	}

	var contents []string
	for _, entry := range entries {
		contents = append(contents, entry.Name())
	}
	return contents
}

func (f *ShallowList) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var path string

	resp.Error = req.Arguments.Get(ctx, &path)
	fileList := ListDirectoryContents(path)

	resp.Error = resp.Result.Set(ctx, fileList)
}
