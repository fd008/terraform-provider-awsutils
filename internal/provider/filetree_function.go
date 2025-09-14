// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"terraform-provider-awsutils/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ function.Function = &FileTree{}
)

func FileTreeFunction() function.Function {
	return &FileTree{}
}

type FileTree struct{}

func (r FileTree) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "filetree"
}

func (f *FileTree) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Return the file tree given a root path upto the provided depth, which defaults to 1",
		Description: "Given a directory path, will return a list of (one level as default) files and folders in that directory. Returns an empty list if the directory does not exist or is not accessible.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "path",
				Description: "Root path where to start the tree",
			},
			function.Int64Parameter{
				Name:               "depth",
				Description:        "Depth of the tree, defaults to 1",
				AllowNullValue:     true,
				AllowUnknownValues: true,
			},
			function.SetParameter{
				Name:                "exclusion_list",
				Description:         "A list of file patterns to exclude from the result. Patterns are matched against file names.",
				AllowNullValue:      true,
				AllowUnknownValues:  true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of file patterns to exclude from the result. Patterns are matched against file names. For example, `*.txt` will exclude all text files.",
			},
		},
		Return: function.DynamicReturn{},
	}
}

func (f *FileTree) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var path string
	var depth int64
	var exclusionList []string

	resp.Error = req.Arguments.Get(ctx, &path, &depth, &exclusionList)

	if depth == 0 {
		depth = 1
	}

	fileTree, err := utils.FileTree(path, int(depth), exclusionList)

	if err != nil {
		resp.Error = function.NewArgumentFuncError(0, fmt.Sprintf("Error: %q", err.Error()))
		return
	}

	mapVal, mapErr := decode(ctx, fileTree)

	if mapErr != nil {
		resp.Error = function.NewArgumentFuncError(0, fmt.Sprintf("Error: %q", mapErr.Errors()))
		return
	}

	resp.Error = resp.Result.Set(ctx, types.DynamicValue(mapVal))
}
