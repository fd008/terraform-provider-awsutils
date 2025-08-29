// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"terraform-provider-awsutils/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ function.Function = &FileSet{}
)

func FileSetFunction() function.Function {
	return &FileSet{}
}

type FileSet struct{}

func (r FileSet) Metadata(_ context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "fileset"
}

func (f *FileSet) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Return a list of files/folders given a directory path or empty list if there are any errors",
		Description: "Given a directory path, will return a list of (one level only) files and folders in that directory. Returns an empty list if the directory does not exist or is not accessible.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "path",
				Description: "A Directory Path",
			},
			function.ListParameter{
				Name:                "exclusion_list",
				Description:         "A list of file patterns to exclude from the result. Patterns are matched against file names.",
				AllowNullValue:      true,
				AllowUnknownValues:  true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of file patterns to exclude from the result. Patterns are matched against file names. For example, `*.txt` will exclude all text files.",
			},
		},
		Return: function.ListReturn{
			ElementType: types.StringType,
		},
	}
}

// resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &path, &exclusionList))

func (f *FileSet) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var path string
	var exclusionList []string

	// Get arguments
	resp.Error = req.Arguments.Get(ctx, &path, &exclusionList)
	if resp.Error != nil {
		return
	}

	if len(exclusionList) == 0 {
		exclusionList = []string{}
	}

	// Call utility function
	fileList, err := utils.ConcurrentFileSet(path, exclusionList...)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &path, &exclusionList))
	}

	// Always return a list, even if empty
	resp.Error = resp.Result.Set(ctx, fileList)
}
